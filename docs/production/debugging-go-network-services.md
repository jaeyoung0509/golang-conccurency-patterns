---
title: Debugging Go Network Services in Production
description: Learn how to decompose Go network latency into DNS, connect, TLS, request, and body phases and prove whether the problem is code, transport, or the network.
---

# Debugging Go Network Services in Production

When a Go network service is slow, “the network is bad” is usually not a useful diagnosis.

You need to prove where the time went.

## Mental model

A single outbound HTTP or RPC call often has at least these phases:

| Phase | Typical owner |
| --- | --- |
| DNS lookup | resolver / environment / service discovery |
| TCP connect | dialer, routing, remote accept path |
| TLS handshake | `crypto/tls`, certificate policy, ALPN |
| Request write | client, kernel buffers, peer read speed |
| First-byte wait | remote handler or upstream queue |
| Response body read | transport reuse discipline, peer throughput, caller ownership |

If you collapse all of that into one “request latency” number, debugging becomes guesswork.

## First debugging move: decompose the path

For HTTP clients, `httptrace` is the quickest way to stop guessing:

```go
trace := &httptrace.ClientTrace{
	DNSStart:           func(httptrace.DNSStartInfo) {},
	DNSDone:            func(httptrace.DNSDoneInfo) {},
	ConnectStart:       func(_, _ string) {},
	ConnectDone:        func(_, _ string, _ error) {},
	TLSHandshakeStart:  func() {},
	TLSHandshakeDone:   func(tls.ConnectionState, error) {},
	GotConn:            func(httptrace.GotConnInfo) {},
	GotFirstResponseByte: func() {},
}
```

The point is not to log every hook forever. The point is to prove which phase changed during an incident.

## A practical debugging workflow

1. Confirm whether the issue is inbound, outbound, or both.
2. Split latency into DNS, connect, TLS, handler, and body-read phases.
3. Check whether connections are being reused or churned.
4. Confirm whether time is spent in network wait, CPU, lock contention, or downstream queueing.
5. Check socket-level state such as `TIME_WAIT`, `ESTABLISHED`, or connection explosion.
6. Only then decide whether the problem is application code, transport policy, or the network.

## Tools and what they prove

### `httptrace`

Use it to prove:

- reuse vs new dial,
- DNS latency,
- connect latency,
- TLS handshake latency,
- first-byte delay.

### `runtime/trace`

Use it to prove:

- goroutines blocked on network wait vs locks,
- cancellation and wakeup timing,
- scheduler transitions around stalled requests.

### `pprof`

Use it to prove:

- CPU time is actually in handler logic or encoding,
- heap growth is due to body buffering or request fan-out,
- contention exists above the transport layer.

### `ss` and `lsof`

Use them to prove:

- socket counts,
- `TIME_WAIT` churn,
- per-process descriptor pressure,
- whether connections are piling up faster than reuse is happening.

### `tcpdump`

Use it when the application and socket metrics disagree, or when you need packet-level proof of retransmissions, resets, or missing responses.

### `GODEBUG`

Useful values in this track:

- `GODEBUG=netdns=go+2` for resolver choice and lookup behavior,
- `GODEBUG=http2debug=1` or `2` for HTTP/2 behavior,
- scheduler tracing when you suspect runtime starvation rather than transport delay.

## Failure patterns

### Treating `context deadline exceeded` as a root cause

That error only tells you a budget expired. It does not tell you which phase consumed the budget.

### Recreating clients and transports in hot paths

That makes connection churn look like network instability.

### Not draining and closing response bodies

That silently destroys reuse and makes later requests pay extra dial and handshake cost.

### Looking only at service-level latency

A flat “p95 = 800 ms” graph cannot tell you whether DNS, TLS, queueing, or body ownership is the problem.

### Ignoring socket state explosion

If `TIME_WAIT` or connection count explodes, the incident may be about transport lifecycle rather than handler CPU.

## How to prove which layer is guilty

### Signs the application layer is guilty

- high CPU in handlers or JSON encoding,
- blocked goroutines on mutexes or channels,
- stable network timings but slow first-byte latency,
- downstream fan-out or queue saturation.

### Signs transport policy is guilty

- new dial for almost every request,
- low reuse despite stable upstream,
- frequent body-close mistakes,
- HTTP/2 or gRPC channel misuse.

### Signs the network or environment is guilty

- DNS spikes before connect starts,
- retransmissions or resets on captures,
- connection attempts timing out before handler work even begins,
- cluster-level packet path or service discovery issues.

## Related docs in this handbook

- [net and netip](/stdlib/net-and-netip)
- [crypto/tls in Production](/stdlib/crypto-tls)
- [net/http Server and Transport Internals](/stdlib/net-http-server-transport)
- [HTTP/2, ALPN, and Stream Multiplexing](/stdlib/http2-alpn-stream-multiplexing)
- [grpc-go Production Playbook](/playbooks/grpc-go-production-playbook)
- [Tracing and Contention Observability](/testing/tracing-and-profiling)

## Practical takeaway

Production network debugging gets much easier once you stop asking “is the network slow?” and start asking “which phase of this connection lifecycle changed?”
