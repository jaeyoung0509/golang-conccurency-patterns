---
title: net/http Production Field Guide
description: Learn how to run Go's HTTP client and server stack with deliberate timeout, pooling, and lifecycle policy.
---

# net/http Production Field Guide

This page is intentionally different from [net/http Server and Transport Internals](/stdlib/net-http-server-transport).

That page explains how the package works. This page explains how to operate it safely in production.

## Mental model

`net/http` gives you two long-lived owners:

- `http.Server` owns inbound connection lifetime,
- `http.Client` plus `Transport` owns outbound connection lifetime.

Most production mistakes happen when a team treats either one as a throwaway helper instead of a policy boundary.

## Safe default sketch

```go
transport := http.DefaultTransport.(*http.Transport).Clone()
transport.MaxIdleConns = 256
transport.MaxIdleConnsPerHost = 64
transport.MaxConnsPerHost = 128
transport.IdleConnTimeout = 90 * time.Second
transport.ResponseHeaderTimeout = 2 * time.Second
transport.ExpectContinueTimeout = 1 * time.Second

client := &http.Client{
	Transport: transport,
}

srv := &http.Server{
	Addr:              ":8080",
	Handler:           mux,
	ReadHeaderTimeout: 2 * time.Second,
	IdleTimeout:       60 * time.Second,
	BaseContext: func(net.Listener) context.Context {
		return appCtx
	},
}
```

This is not the only valid policy. It is a safe starting point because it makes ownership explicit:

- one long-lived transport,
- one long-lived client,
- explicit server-side read boundaries,
- request-scoped deadlines layered on top.

## Operating rules

### Reuse clients and transports

Constructing a new `Transport` per request destroys pooling and creates needless dial churn. Treat the transport as a singleton per outbound policy boundary.

### Use per-request context and transport timeouts together

`Client.Timeout` is a blunt whole-exchange cap. It is useful for simple cases, but in larger systems you usually want:

- request context deadlines for business budgets,
- transport timeouts for specific network edges,
- server-side read and idle timeouts for inbound defense.

### Response bodies are connection ownership

Closing the body is not just cleanup. It tells the transport whether the connection can be reused.

If you stop early on a response, decide consciously between:

- bounded drain plus reuse, or
- immediate close and loss of reuse.

Do not accidentally do neither.

### Streaming changes timeout strategy

Server `WriteTimeout` can be the wrong tool for long-lived streaming responses such as SSE or large downloads. Streaming endpoints usually need their own policy rather than generic “cap everything” defaults.

## Failure patterns

### New client in a hot path

```go
func fetch(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 2 * time.Second} // bad: new transport policy every call
	return client.Get(url)
}
```

### Hidden default client on a critical dependency

```go
resp, err := http.Get(url) // bad: no explicit transport budget, proxy policy, or reuse settings
```

### Forgetting body ownership

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
return decode(resp.Body) // bad: body never closed
```

### One giant timeout knob

```go
client := &http.Client{Timeout: 30 * time.Second}
```

This often hides whether you actually need a dial budget, a response-header budget, a handler deadline, or a shutdown contract.

## What to use carefully

- `http.DefaultClient` and `http.DefaultTransport` in library-like code paths
- `WriteTimeout` on streaming handlers
- blind `io.Copy(io.Discard, resp.Body)` on huge or attacker-controlled bodies
- `TimeoutHandler` as a substitute for real context discipline

## Observability and testing

- Use `httptest.Server` to test client timeout edges and body-close discipline.
- Use `Server.Shutdown` tests to verify in-flight request behavior during deploys.
- Use `httptrace` when you need proof of reuse versus redial behavior.
- Expose request duration, status code, inflight count, and outbound dependency latency separately. Mixing them hides transport churn.

## When it is the right tool

`net/http` is the right default for most Go services, control planes, APIs, webhooks, and internal service-to-service calls. The standard library gives you enough control for the vast majority of workloads if you treat its policy boundaries seriously.

## When it is not the right tool

If your problem is a highly specialized L7 proxy, custom wire protocol gateway, or an environment where every extra allocation on the hot path is existential, you may need a narrower or more specialized stack. Even then, you should first prove that the real problem is `net/http` itself and not your own timeout, reuse, or buffering discipline.

## Official reading

- [Package docs for `net/http`](https://pkg.go.dev/net/http)
- [net/http Server and Transport Internals](/stdlib/net-http-server-transport)

## Practical takeaway

Production `net/http` is mostly an ownership problem.

If you make client reuse, response bodies, and timeout edges explicit, the package is extremely reliable. If you leave those policies implicit, it becomes hard to reason about latency and resource usage.
