---
title: net/http Server and Transport Internals
description: Learn how Go's standard HTTP server and client transport model map onto goroutines, deadlines, and connection reuse.
---

# net/http Server and Transport Internals

`net/http` is where many Go services spend most of their concurrency budget.

It works because the runtime makes blocking network I/O cheap enough to express in direct style, while the package itself manages connection ownership, keep-alive reuse, and request lifetimes.

## Why this package matters

If you operate a Go service, you almost certainly depend on:

- one goroutine tree per inbound request,
- one transport pool per outbound dependency,
- request-scoped context cancellation,
- connection reuse instead of one TCP dial per request.

Misunderstand any of those boundaries and the system becomes slower, leakier, or less predictable under load.

## End-to-end mental model

```mermaid
flowchart LR
    A["Listener accept loop"] --> B["Server connection goroutine"]
    B --> C["Request context"]
    C --> D["Handler tree"]
    D --> E["Outbound http.Client"]
    E --> F["Transport"]
    F --> G["Idle pool / persistConn"]
    G --> H["readLoop / writeLoop"]
```

On the server side, `net/http` turns sockets into request handlers.

On the client side, `Transport` turns requests into pooled, reusable connections.

Those are different halves of the same lifecycle problem.

## Production sketch

```go
srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 2 * time.Second,
	WriteTimeout:      10 * time.Second,
	IdleTimeout:       60 * time.Second,
	BaseContext: func(net.Listener) context.Context {
		return appCtx
	},
}

transport := &http.Transport{
	MaxIdleConns:        256,
	MaxIdleConnsPerHost: 64,
	MaxConnsPerHost:     128,
	IdleConnTimeout:     90 * time.Second,
	ResponseHeaderTimeout: 2 * time.Second,
}

client := &http.Client{
	Transport: transport,
	Timeout:   3 * time.Second,
}
```

This sketch is not interesting because it uses many fields. It is interesting because it makes lifetime and budget decisions explicit.

## Server mental model

The server side is not “one global handler goroutine.”

The package roughly does this:

1. accept a connection,
2. attach server and connection context,
3. read a request,
4. run the handler,
5. decide whether the connection stays alive,
6. repeat or close.

For HTTP/1.x, connection ownership is especially visible because one TCP connection is serially reused for multiple requests. For HTTP/2, multiplexing changes the shape, but the same lifetime and deadline questions still matter.

## Simplified internal sketch

```go
type Server struct {
	Handler           Handler
	ReadHeaderTimeout time.Duration
	WriteTimeout      time.Duration
	IdleTimeout       time.Duration
	BaseContext       func(net.Listener) context.Context
	ConnContext       func(context.Context, net.Conn) context.Context
}

type Transport struct {
	idleMu           sync.Mutex
	idleConn         map[connectMethodKey][]*persistConn
	reqCanceler      map[*Request]context.CancelCauseFunc
	connsPerHost     map[connectMethodKey]int
	connsPerHostWait map[connectMethodKey]wantConnQueue
}

type persistConn struct {
	conn net.Conn
	// one goroutine reads responses
	// one goroutine writes requests
}
```

The exact source has more bookkeeping, but the design pressure is visible already:

- server fields define timeout and ownership boundaries,
- transport fields define pooling and cancellation boundaries,
- `persistConn` is the reusable connection object that actual requests flow through.

## Source walk

### Server side

The `Server` type includes:

- `ReadHeaderTimeout`, `ReadTimeout`, `WriteTimeout`, and `IdleTimeout`,
- `BaseContext` and `ConnContext`,
- shutdown and active-connection bookkeeping.

These are not cosmetic settings. They are the core operating contract for inbound traffic.

Relevant source entry points:

- [`server.go` `Server`](https://github.com/golang/go/blob/go1.26.0/src/net/http/server.go#L2964)
- [`server.go` `Serve`](https://github.com/golang/go/blob/go1.26.0/src/net/http/server.go#L3444)

### Client side

`http.Client` is a thin wrapper. The real concurrency behavior mostly lives in `Transport`:

- idle connection maps,
- per-host connection limits,
- request cancellation bookkeeping,
- dialing and handshake timeouts,
- `persistConn` read and write loops.

Relevant source entry points:

- [`transport.go` `Transport`](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go#L97)
- [`transport.go` `roundTrip`](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go#L590)
- [`transport.go` `persistConn`](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go#L2115)
- [`transport.go` `readLoop`](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go#L2291)
- [`transport.go` `writeLoop`](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go#L2649)

## Response body ownership is a concurrency issue

This is one of the most common mistakes:

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
```

Closing the body is not just cleanup politeness.

It is how the transport learns whether the connection can be reused cleanly. Mishandle that and the pool churns connections instead of reusing them.

## Failure patterns

### Using `http.Get` in a hot production path

```go
resp, err := http.Get(url)
```

This hides transport configuration, timeout policy, per-host limits, proxy behavior, and connection budget from the reader. That is acceptable for a script, not for a service boundary.

### Forgetting to close the response body

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
return decode(resp.Body) // leak: body never closed
```

This breaks connection reuse and can exhaust file descriptors or sockets.

### Assuming one timeout knob is enough

`Client.Timeout`, transport header timeouts, request context deadlines, and server-side read/write timeouts cover different edges.

You usually need a deliberate combination, not one giant timeout applied everywhere.

### Treating `TimeoutHandler` as complete request control

`TimeoutHandler` is a blunt wrapper. It helps cap response time, but it does not replace handler-local context discipline, outbound deadlines, or shutdown design.

### Creating a new transport per request

That throws away pooling, creates dial churn, and usually makes latency and resource usage worse.

## Production consequences

- Keep one long-lived `Transport` per policy boundary, not per request.
- Reuse one `http.Client` per transport policy instead of constructing clients ad hoc in hot code.
- Make timeouts explicit at the server, request, and transport layers.
- Treat request context cancellation as a first-class signal inside handlers and outbound calls.
- Review body-closing discipline in code review with the same seriousness as mutex ownership.

## How to test and observe it

- Test client timeouts with `httptest.Server` and deliberate slow handlers.
- Test server shutdown with `Server.Shutdown` and in-flight requests.
- Use `httptrace` when you need to confirm reuse, dial timing, or connection setup behavior.
- Use runtime traces or socket-level metrics if you suspect connection churn rather than handler CPU.
- See `examples/httptransportlab` for focused tests around response-body drain and connection reuse.

## Networking track continuation

- For multiplexed reuse and ALPN negotiation, continue with [HTTP/2, ALPN, and Stream Multiplexing](/stdlib/http2-alpn-stream-multiplexing).
- For incident workflow, continue with [Debugging Go Network Services in Production](/production/debugging-go-network-services).

## Official reading

- [Package docs for `net/http`](https://pkg.go.dev/net/http)
- [Package docs for `net/http/httptrace`](https://pkg.go.dev/net/http/httptrace)
- [Go 1.26 `net/http` server source](https://github.com/golang/go/blob/go1.26.0/src/net/http/server.go)
- [Go 1.26 `net/http` transport source](https://github.com/golang/go/blob/go1.26.0/src/net/http/transport.go)

## Practical takeaway

`net/http` is not only a convenient API. It is the main place where Go's runtime, deadlines, and connection reuse policies become visible in production behavior.
