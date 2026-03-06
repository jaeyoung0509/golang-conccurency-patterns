---
title: 프로덕션에서의 crypto/tls
description: tls.Config, handshake lifetime, verification hook, session behavior가 실제 Go 서비스에 어떤 영향을 주는지 설명합니다.
---

# 프로덕션에서의 crypto/tls

TLS는 단순히 “암호화 켜기”가 아닙니다.

Go에서 `crypto/tls`는 certificate policy, handshake lifetime, ALPN, resumption, hostname verification이 실행 가능한 설정이 되는 지점입니다.

## 왜 이 패키지가 중요한가

실제 TLS 장애는 cryptography 문제보다 configuration 문제인 경우가 많습니다.

- 잘못된 `ServerName`
- 위험한 `InsecureSkipVerify`
- 숨겨진 certificate selection 비용
- request/dial budget을 존중하지 않는 handshake
- config clone/resumption 사이의 예상 밖 동작

같은 것들입니다.

## Mental model

`crypto/tls`는 기존 `net.Conn` 위에 policy를 덧씌우는 패키지입니다.

즉 lifecycle은 보통:

1. TCP connection establish
2. TLS client/server state로 wrap
3. handshake 실행
4. negotiated state 확인
5. application data read/write

입니다.

중요한 의미는, TCP 성공이 TLS 성공을 뜻하지 않는다는 점입니다. handshake는 자체적인 실패면과 timeout 경계를 가집니다.

## 실전 코드 스케치

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

핵심은 handshake에 별도의 예산을 준다는 점입니다. 그래야 certificate validation과 ALPN이 background magic이 아니라 ordinary request lifetime의 일부가 됩니다.

## 단순화한 내부 코드 예시

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

정확한 원문은 아니지만 두 가지 중요한 사실을 잘 보여줍니다.

- policy는 `tls.Config`에 들어 있고,
- handshake lifetime은 독립적으로 의미 있는 취소/실패 경계입니다.

## `tls.Config`에서 특히 중요한 선택

### `ServerName`

클라이언트에서 `ServerName`은 ordinary verification이 켜져 있을 때 hostname verification을 결정합니다. IP로 연결하면서 named certificate를 기대하는데 `ServerName`을 안 넣으면, 그에 맞는 실패를 얻게 됩니다.

### `RootCAs`

`RootCAs`가 nil이면 host trust store를 씁니다. 대개 맞지만, internal PKI 환경에서는 항상 원하는 동작은 아닙니다.

### `NextProtos`

ALPN negotiation을 정의하는 필드입니다. HTTP/2나 다른 상위 프로토콜 선택이 중요하면 빠질 수 없습니다.

### `Certificates`, `GetCertificate`

서버에 여러 certificate가 있고 `Leaf`가 비어 있으면 handshake마다 숨은 selection cost가 커질 수 있습니다. 문서도 이 점을 명시합니다.

### `VerifyPeerCertificate` vs `VerifyConnection`

이 둘은 같지 않습니다.

- `VerifyPeerCertificate`는 normal verification 이후 실행되지만 resumed connection에서는 호출되지 않습니다.
- `VerifyConnection`은 resumption을 포함한 모든 connection에서 실행됩니다.

resumed session에도 반드시 적용되어야 하는 정책이 있다면 이 차이가 중요합니다.

### `InsecureSkipVerify`

ordinary chain/hostname verification을 끕니다. “개발 편의용 임시 설정”처럼 취급하면 안 됩니다. 테스트이거나, 정말로 이해한 custom verification을 붙이는 경우에만 써야 합니다.

## Handshake lifetime과 cancellation

Go 1.26 구현에서 가장 유용한 디테일 중 하나는, `HandshakeContext`가 입력 context가 handshake 완료 전에 취소되면 underlying connection을 닫는다는 점입니다.

즉:

- handshake timeout을 request/dial budget으로 둘 수 있고,
- stuck handshake를 cancellation로 멈출 수 있으며,
- handshake가 이미 실패했는데 뒤의 `Read`/`Write`가 성공하길 기대하면 안 됩니다.

## Session behavior와 resumption

TLS session resumption은 latency 최적화이지만 verification/policy 흐름도 바꿉니다.

이해할 만한 필드:

- `ClientSessionCache`
- `SessionTicketsDisabled`
- session ticket rotation behavior

성능에는 좋을 수 있지만, connection별 verification rule이 있다면 단순 implementation detail로 보면 안 됩니다.

## 런타임/소스 코드 워크

읽을 만한 진입점:

- [`crypto/tls/common.go` `Config`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/common.go#L566)
- [`crypto/tls/tls.go` `Client`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/tls.go#L59)
- [`crypto/tls/tls.go` `Server`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/tls.go#L46)
- [`crypto/tls/conn.go` `HandshakeContext`](https://github.com/golang/go/blob/go1.26.0/src/crypto/tls/conn.go#L1513)

특히 볼 점:

- `HandshakeContext`에는 이미 완료된 handshake를 위한 fast path가 있고,
- cancellation은 `context.AfterFunc`를 통해 underlying connection close로 이어지며,
- `Config`의 여러 필드가 성능과 trust policy를 동시에 바꿉니다.

## 실패 패턴

### `InsecureSkipVerify`를 영구 해결책처럼 사용

```go
cfg := &tls.Config{InsecureSkipVerify: true}
```

이건 명시적으로 대체 verification을 붙이지 않는 한 보안 구멍입니다.

### TCP dial 성공이 TLS 성공이라고 생각

certificate, ALPN, version, cipher, policy에서 handshake는 얼마든지 실패할 수 있습니다.

### `ServerName` 누락

특히 IP dial이나 custom dial path에서 hostname verification은 알아서 맞춰지지 않습니다.

### verification hook을 같은 걸로 취급

`VerifyPeerCertificate`와 `VerifyConnection`은 resumption 동작이 다릅니다.

### publish 후 config나 certificate mutate

활성 사용 중인 `Config`나 callback에서 반환한 certificate는 immutable하게 다뤄야 합니다.

## 프로덕션에서의 의미

- handshake에도 명시적인 context budget을 둡니다.
- TLS verification은 기본적으로 strict하게 유지합니다.
- 실제 트래픽에서 certificate selection과 resumption behavior를 점검합니다.
- transport issue를 볼 때 negotiated protocol과 peer state를 디버그 정보로 남깁니다.

## 공식 자료

- [`crypto/tls` 패키지 문서](https://pkg.go.dev/crypto/tls)
- [Go security and FIPS docs](https://go.dev/doc/security/fips140)

## Practical takeaway

Go에서 TLS는 대체로 configuration discipline 문제입니다. handshake lifetime과 verification policy를 명시하면 예측 가능하게 동작하고, 그렇지 않으면 그 나름대로 예측 가능하게 불리하게 동작합니다.
