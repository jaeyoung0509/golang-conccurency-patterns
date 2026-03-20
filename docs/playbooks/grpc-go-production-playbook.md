---
title: grpc-go Production Playbook
description: Learn how to operate grpc-go with long-lived channels, deadlines, keepalive discipline, and explicit streaming ownership.
---

# grpc-go Production Playbook

`grpc-go` is not “HTTP but binary.”

It is a transport, channel, and RPC policy stack that gets complicated quickly if you treat connections or deadlines casually.

## Mental model

The first rule is simple:

- a `ClientConn` is a long-lived channel,
- an RPC is short-lived work on top of that channel.

If you reverse those lifetimes, the library becomes expensive and unreliable.

## Safe default sketch

```go
cc, err := grpc.NewClient(
	"dns:///api.example.com:443",
	grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		MinVersion: tls.VersionTLS12,
	})),
	grpc.WithConnectParams(grpc.ConnectParams{
		MinConnectTimeout: 5 * time.Second,
		Backoff: backoff.Config{
			BaseDelay:  200 * time.Millisecond,
			Multiplier: 1.6,
			MaxDelay:   3 * time.Second,
		},
	}),
)
if err != nil {
	return err
}
defer cc.Close()

ctx, cancel := context.WithTimeout(parent, 2*time.Second)
defer cancel()

resp, err := pb.NewUsersClient(cc).GetUser(ctx, req)
```

This sketch bakes in the right defaults:

- `grpc.NewClient`, not deprecated `Dial`,
- one long-lived channel,
- per-RPC deadlines,
- explicit transport credentials,
- bounded connect backoff policy.

## Operating rules

### Prefer `grpc.NewClient`

The official package docs now center `NewClient`. It creates the channel without doing I/O immediately. RPCs will connect as needed. That is usually what you want.

### Avoid `Dial` and especially `WithBlock`

`Dial` is deprecated, and `WithBlock` is explicitly discouraged. Blocking process startup on connection establishment often turns transient dependency issues into global outages.

### One channel per authority or policy boundary

Build one `ClientConn` per destination and policy set, then reuse it for many RPCs. Do not open a new channel per request.

### Deadlines belong on RPCs, not only on startup

Even if the channel exists, each RPC still needs a deadline. Otherwise queueing, retries, or network stalls become unbounded.

### Streaming requires ownership discipline

If you abandon a stream early, cancel its context. If you intend to consume it, drive `Recv()` to `io.EOF`. Leaving streams half-owned is the streaming equivalent of leaking response bodies.

### Keepalive is a negotiated operating policy

Aggressive client keepalive is not harmless. The official keepalive guide warns that unsupported settings can make the server respond with `GOAWAY` carrying `too_many_pings`.

## Failure patterns

### One channel per RPC

```go
func call(ctx context.Context, req *pb.Request) error {
	cc, err := grpc.NewClient("dns:///api.example.com:443", grpc.WithTransportCredentials(creds))
	if err != nil {
		return err
	}
	defer cc.Close()
	_, err = pb.NewAPIClient(cc).Handle(ctx, req)
	return err
}
```

This burns connection setup, resolver work, and transport warmup on the hot path.

### No deadline on the RPC

```go
resp, err := client.GetUser(context.Background(), req) // bad: unbounded lifetime
```

### `WithBlock` as startup policy

```go
cc, err := grpc.Dial(target, grpc.WithTransportCredentials(creds), grpc.WithBlock()) // bad default
```

### Keepalive too aggressive

If the service does not explicitly support your keepalive cadence, you can create avoidable reconnect churn.

### Abandoning streams without canceling

```go
stream, _ := client.Subscribe(ctx, req)
return nil // bad: stream context still owns transport resources
```

## What to use carefully

- `WithBlock`
- `WaitForReady` as a blanket default instead of an explicit retry policy
- aggressive keepalive on quiet channels
- service-config features you do not actively understand in staging
- one `ClientConn` shared across unrelated trust or credential boundaries

## Observability and testing

- Track deadline-exceeded and unavailable rates separately.
- Use interceptors or stats handlers to record method latency, size, and status code.
- Use `bufconn` or dedicated integration environments for RPC-level tests without real networks.
- During incidents, distinguish resolver, connect, TLS, and RPC handler latency rather than blaming “gRPC” as one blob.

## When it is the right tool

`grpc-go` is a strong default when you need typed internal APIs, streaming, deadlines, and cross-language protocol compatibility.

## When it is not the right tool

If your traffic is simple HTTP/JSON and most complexity comes from human-facing APIs or browser compatibility, raw HTTP may be cheaper to operate. Likewise, if the system only needs fire-and-forget event transfer, an RPC stack may be the wrong abstraction.

## Official reading

- [Package docs for `google.golang.org/grpc`](https://pkg.go.dev/google.golang.org/grpc)
- [gRPC keepalive guide](https://grpc.io/docs/guides/keepalive/)
- [net/http Server and Transport Internals](/stdlib/net-http-server-transport)
- [HTTP/2, ALPN, and Stream Multiplexing](/stdlib/http2-alpn-stream-multiplexing)

## Practical takeaway

grpc-go gets much easier once you stop thinking “connection per call” and start thinking “long-lived channel with short-lived RPC budgets.”
