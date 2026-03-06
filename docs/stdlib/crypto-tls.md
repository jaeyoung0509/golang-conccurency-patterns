---
title: crypto/tls in Production
description: Learn how tls.Config, handshake lifetimes, verification hooks, and session behavior affect real Go services.
---

# crypto/tls in Production

TLS is not just “turn encryption on.”

In Go, `crypto/tls` is where certificate policy, handshake lifetime, ALPN, resumption, and hostname verification become executable configuration.

## Why this package matters

Many real outages around TLS are configuration problems, not cryptography problems:

- wrong `ServerName`,
- unsafe `InsecureSkipVerify`,
- certificate selection cost,
- handshakes that do not respect request or dial budgets,
- surprising resumption behavior across config clones.

## Mental model

`crypto/tls` layers policy on top of an existing `net.Conn`.

That means the lifecycle often looks like this:

1. establish a TCP connection,
2. wrap it with TLS client or server state,
3. run the handshake,
4. inspect negotiated state,
5. read and write application data.

The important consequence is that TCP success is not TLS success. The handshake has its own failure and timeout surface.

## Production sketch

```go
rawConn, err := dialer.DialContext(ctx, "tcp", "api.example.com:443")
if err != nil {
	return err
}

tlsConn := tls.Client(rawConn, &tls.Config{
	ServerName: "api.example.com",
	NextProtos: []string{"h2", "http/1.1"},
	RootCAs:    roots,
})

if err := tlsConn.HandshakeContext(ctx); err != nil {
	return err
}

state := tlsConn.ConnectionState()
_ = state.NegotiatedProtocol
```

The key is that the handshake is explicitly budgeted. That is what turns certificate validation and ALPN from background magic into an ordinary part of request lifetime.

## Simplified internal sketch

```go
type Config struct {
	Certificates       []Certificate
	GetCertificate     func(*ClientHelloInfo) (*Certificate, error)
	RootCAs            *x509.CertPool
	ServerName         string
	NextProtos         []string
	InsecureSkipVerify bool
	VerifyConnection   func(ConnectionState) error
}

func (c *Conn) HandshakeContext(ctx context.Context) error {
	if ctx is canceled before handshake completes {
		closeUnderlyingConn()
	}
	return runHandshakeMachine(ctx)
}
```

This is not the literal source, but it captures the two most important facts:

- policy lives in `tls.Config`,
- handshake lifetime is independently meaningful and can be canceled.

## `tls.Config` choices that matter a lot

### `ServerName`

For clients, `ServerName` drives hostname verification unless you disable verification. If you connect by IP to a named certificate without setting `ServerName`, you often get exactly the breakage you deserve.

### `RootCAs`

If `RootCAs` is nil, the host trust store is used. That is often correct, but not always what internal PKI environments need.

### `NextProtos`

This is where ALPN negotiation is configured. If you care about HTTP/2 or other protocol selection, this field is not optional.

### `Certificates` and `GetCertificate`

Server-side certificate selection can become a hidden handshake cost if multiple certificates are present and `Leaf` fields are missing. The docs call this out directly.

### `VerifyPeerCertificate` vs `VerifyConnection`

These hooks are not interchangeable:

- `VerifyPeerCertificate` runs after normal verification, but is skipped on resumed connections.
- `VerifyConnection` runs for all connections, including resumptions.

That distinction matters if your security policy must apply to resumed sessions too.

### `InsecureSkipVerify`

It disables ordinary chain and hostname verification. It is not a harmless way to “make development easier.” Use it only for testing or in combination with explicit custom verification that you actually understand.

## Handshake lifetime and cancellation

One of the most useful implementation details in Go 1.26 is that `HandshakeContext` uses cancellation to close the underlying connection if the input context is canceled before the handshake completes.

That means:

- handshake timeouts can be first-class request or dial budgets,
- cancellation can stop a stuck handshake,
- a later successful `Read` or `Write` should not be assumed if the handshake path already failed.

## Session behavior and resumption

TLS session resumption is a latency feature, but it also changes verification and policy flow.

Fields worth understanding:

- `ClientSessionCache`
- `SessionTicketsDisabled`
- session ticket rotation behavior

Resumption can be good for performance and subtle for policy. Do not treat it as just an implementation detail if you have per-connection verification rules.

## Runtime source walk

Useful entry points:

- [`crypto/tls/common.go` `Config`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/common.go#L566)
- [`crypto/tls/tls.go` `Client`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/tls.go#L59)
- [`crypto/tls/tls.go` `Server`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/tls.go#L46)
- [`crypto/tls/conn.go` `HandshakeContext`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/conn.go#L1513)

Details worth remembering:

- `HandshakeContext` has a fast path when the handshake is already complete,
- cancellation uses `context.AfterFunc` to close the underlying connection when needed,
- `Config` is full of hooks and fields that influence both performance and trust policy.

## Failure patterns

### Using `InsecureSkipVerify` as a permanent fix

```go
cfg := &tls.Config{InsecureSkipVerify: true}
```

This is a security hole unless you replace the skipped verification with something explicit and correct.

### Assuming TCP dial success means TLS success

The handshake can still fail on certificate, ALPN, version, cipher, or policy issues.

### Forgetting `ServerName`

Especially when dialing by IP or via a custom dial path, hostname verification will not guess the right policy for you.

### Treating verification hooks as equivalent

`VerifyPeerCertificate` and `VerifyConnection` have materially different resumption behavior.

### Mutating configs or certificates after publication

Once a `Config` or returned certificate is in active use, treat it as immutable.

## Production consequences

- Budget the handshake explicitly with context.
- Keep TLS verification strict by default.
- Review certificate selection and resumption behavior under real traffic.
- Surface negotiated protocol and peer state in debugging output when transport issues matter.

## Official reading

- [Package docs for `crypto/tls`](https://pkg.go.dev/crypto/tls)
- [Go security and FIPS docs](https://go.dev/doc/security/fips140)

## Practical takeaway

TLS in Go is mostly a configuration discipline problem. If you make handshake lifetime and verification policy explicit, the package behaves predictably. If you hand-wave those fields, it still behaves predictably, just not in your favor.
