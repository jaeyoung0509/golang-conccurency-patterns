---
title: net and netip
description: Learn how Dialer, Resolver, deadlines, listeners, and netip value types shape production Go networking.
---

# net and netip

The `net` package is where the Go runtime's blocking-I/O story becomes concrete.

It gives you direct-style networking APIs, but the real engineering work is in understanding where lifetime, timeout, address representation, and resolution behavior actually begin and end.

## Why these packages matter

Production network bugs are often not “TCP is hard” bugs.

They are:

- dialing without a real budget,
- assuming a connect timeout covers later reads and writes,
- using `net.IP` where immutable, comparable addresses would be cleaner,
- forgetting that DNS resolution is part of the request path.

## Example scenario

The `examples/dialbudget` package probes several replica endpoints using `net.Dialer` and `netip.AddrPort`, with one budget for connect and another deadline for the exchange after the socket is open.

```mermaid
flowchart LR
    A["overall request context"] --> B["per-attempt timeout"]
    B --> C["net.Dialer.DialContext"]
    C --> D["TCP connection established"]
    D --> E["SetDeadline for probe I/O"]
    E --> F["read ok / timeout / fallback"]
```

## Production sketch

```go
type Prober struct {
	Dialer            net.Dialer
	PerAttemptTimeout time.Duration
}

func (p Prober) Probe(ctx context.Context, endpoint netip.AddrPort) error {
	attemptCtx, cancel := context.WithTimeout(ctx, p.PerAttemptTimeout)
	defer cancel()

	conn, err := p.Dialer.DialContext(attemptCtx, "tcp", endpoint.String())
	if err != nil {
		return err
	}
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(p.PerAttemptTimeout))
	_, _ = fmt.Fprintf(conn, "health\n")
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		return err
	}
	if strings.TrimSpace(reply) != "ok" {
		return errors.New("unhealthy endpoint")
	}
	return nil
}
```

Two separate boundaries matter:

1. `DialContext` limits the connect path, including resolution and socket establishment.
2. `SetDeadline` limits later I/O on the already-open connection.

Many systems accidentally rely on only the first.

## Mental model

`net` and `netip` solve related but different problems:

| Package | Main job |
| --- | --- |
| `net` | sockets, listeners, resolution, deadlines, connection lifecycle |
| `netip` | compact immutable value types for IPs, address+port pairs, and prefixes |

Use `net` for behavior.

Use `netip` for address data you want to compare, store, or pass around without the footguns of `net.IP`.

## Simplified internal sketch

The key shape of `Dialer` is:

```go
type Dialer struct {
	Timeout       time.Duration
	Deadline      time.Time
	FallbackDelay time.Duration
	Resolver      *Resolver
}

func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	deadline := min(d.Timeout, d.Deadline, ctxDeadline(ctx))
	addrs := resolve(network, address)
	return racePrimaryAndFallbackConnections(deadline, addrs)
}
```

The real implementation is more careful, but the important ideas are visible:

- timeout and absolute deadline combine with the context,
- resolution is part of dialing,
- TCP may race primary and fallback address families.

`netip` is intentionally simpler:

```go
type Addr struct { /* small comparable value */ }
type AddrPort struct {
	Addr Addr
	Port uint16
}
```

That simplicity is the point.

## Dialing and deadlines

### `DialContext` is not a full request budget

Once a connection is established, later context expiration does not retroactively cancel ordinary `Read` or `Write` calls on that connection.

If you need I/O deadlines too, set them on the connection:

```go
_ = conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
```

### `Dialer.Timeout`, `Dialer.Deadline`, and context deadlines all interact

The dial path uses the earliest relevant deadline. That means layering many timeout mechanisms is possible, but also easy to make inconsistent.

### Fast fallback matters

For TCP, the dialer can race primary and fallback address families. The historical `DualStack` knob is deprecated; `FallbackDelay` is the relevant control.

## Resolution behavior matters more than many teams expect

The package docs are explicit:

- some systems prefer the pure Go resolver,
- others fall back to cgo/native resolution under specific conditions,
- blocked cgo lookups consume OS threads,
- `GODEBUG=netdns=...` can force or debug resolver choice.

This is not trivia. Under DNS stress, resolver behavior becomes latency behavior.

## Why `netip` is usually better than `net.IP` for new code

`netip.Addr` is:

- smaller,
- immutable,
- comparable,
- safe as a map key.

That makes it better for:

- endpoint tables,
- allowlists and blocklists,
- caches keyed by address,
- configuration snapshots.

You can still bridge to `net` with helpers like `TCPAddrFromAddrPort` and `UDPAddrFromAddrPort`.

## Runtime source walk

Useful entry points in the Go 1.26 tree:

- [`net/dial.go` `Dialer`](https://github.com/golang/go/blob/go1.26.0/src/net/dial.go#L126)
- [`net/dial.go` `DialContext`](https://github.com/golang/go/blob/go1.26.0/src/net/dial.go#L526)
- [`net/lookup.go` `Resolver`](https://github.com/golang/go/blob/go1.26.0/src/net/lookup.go#L134)
- [`net/netip/netip.go` `Addr`](https://github.com/golang/go/blob/go1.26.0/src/net/netip/netip.go#L37)
- [`net/netip/netip.go` `AddrPort`](https://github.com/golang/go/blob/go1.26.0/src/net/netip/netip.go#L1068)

Details worth noticing:

- the dialer derives one effective deadline from timeout, absolute deadline, and context,
- resolution is explicitly separated from connect tracing,
- `DialTCP`, `DialUDP`, and `DialIP` have `netip`-friendly overloads,
- `netip` exists specifically to avoid the allocation and comparability problems of `net.IP`.

## Failure patterns

### Assuming connect timeout also covers later socket I/O

```go
conn, _ := dialer.DialContext(ctx, "tcp", addr)
bufio.NewReader(conn).ReadString('\n') // bug: now unbounded unless a deadline is set
```

### Passing addresses around as raw strings forever

Stringly typed address handling makes validation and comparison harder than it needs to be.

### Using `net.IP` as a map key in new code

That often creates awkward normalization and equality behavior that `netip.Addr` avoids.

### Ignoring DNS behavior in latency debugging

Resolution can dominate tail latency even when the remote service is healthy.

### Forgetting that listener accept loops need shutdown ownership

The moment you create `go handle(conn)` inside an accept loop, you also need a clear listener shutdown contract.

## Production consequences

- Put a real dial budget on outbound connection establishment.
- Set explicit read/write deadlines or higher-level request timeouts after connect.
- Prefer `netip` types for configuration, identity, and lookup tables.
- Treat resolution behavior as part of the service path, not as an invisible prelude.

## Example and tests

- Example: `examples/dialbudget`
- The tests verify fallback from an unreachable endpoint to a healthy one and show that post-connect probe I/O still needs its own deadline.

## Official reading

- [Package docs for `net`](https://pkg.go.dev/net)
- [Package docs for `net/netip`](https://pkg.go.dev/net/netip)

## Practical takeaway

Good Go networking starts by separating address representation, connect lifetime, and post-connect I/O lifetime. `net` and `netip` make that separation explicit if you use them deliberately.
