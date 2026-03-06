---
title: net과 netip
description: Dialer, Resolver, deadline, listener, netip value type이 프로덕션 Go 네트워킹에 어떤 영향을 주는지 설명합니다.
---

# net과 netip

`net` 패키지는 Go 런타임의 blocking-I/O 모델이 실제 네트워크 코드로 드러나는 지점입니다.

direct-style API를 제공하지만, 실제 엔지니어링 포인트는 lifetime, timeout, address representation, resolution behavior가 어디서 시작되고 끝나는지 이해하는 데 있습니다.

## 왜 이 패키지들이 중요한가

프로덕션 네트워크 버그는 흔히 “TCP가 어렵다” 문제가 아닙니다.

- dial budget이 없음
- connect timeout이 이후 read/write까지 커버한다고 착각
- immutable하고 comparable한 주소 타입이 더 적합한데 계속 `net.IP` 사용
- DNS resolution도 요청 경로 일부라는 점을 무시

같은 문제입니다.

## 예제 시나리오

`examples/dialbudget` 패키지는 `net.Dialer`와 `netip.AddrPort`를 사용해 여러 replica endpoint를 probe합니다. connect path 예산과 소켓이 열린 뒤 probe exchange 예산을 분리해서 다룹니다.

```mermaid
flowchart LR
    A["overall request context"] --> B["per-attempt timeout"]
    B --> C["net.Dialer.DialContext"]
    C --> D["TCP connection established"]
    D --> E["SetDeadline for probe I/O"]
    E --> F["read ok / timeout / fallback"]
```

## 실전 코드 스케치

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

여기서 중요한 경계는 두 개입니다.

1. `DialContext`는 resolution과 socket establish를 포함한 connect path를 제한합니다.
2. `SetDeadline`은 이미 열린 connection의 이후 I/O를 제한합니다.

많은 시스템이 첫 번째만 걸고 두 번째를 잊습니다.

## Mental model

`net`과 `netip`는 관련 있지만 역할이 다릅니다.

| 패키지 | 주 역할 |
| --- | --- |
| `net` | socket, listener, resolution, deadline, connection lifecycle |
| `netip` | compact immutable IP/address+port/prefix value type |

동작은 `net`,

주소 데이터 표현은 `netip`가 맡는다고 생각하면 됩니다.

## 단순화한 내부 코드 예시

`Dialer`의 핵심 모양은 대략 이렇습니다.

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

실제 구현은 더 정교하지만 핵심 아이디어는 같습니다.

- timeout, absolute deadline, context deadline이 합쳐지고,
- resolution도 dialing의 일부이며,
- TCP에서는 primary/fallback address family를 race할 수 있습니다.

`netip`는 오히려 단순함이 핵심입니다.

```go
type Addr struct { /* small comparable value */ }
type AddrPort struct {
	Addr Addr
	Port uint16
}
```

## Dialing과 deadline

### `DialContext`는 full request budget이 아니다

connection이 성립된 뒤에는 ordinary `Read`, `Write`는 context 만료만으로 자동 취소되지 않습니다.

이후 I/O에도 예산이 필요하면 connection에 deadline을 직접 걸어야 합니다.

```go
_ = conn.SetDeadline(time.Now().Add(50 * time.Millisecond))
```

### `Dialer.Timeout`, `Dialer.Deadline`, context deadline은 함께 작동한다

dial path는 가장 이른 deadline을 씁니다. timeout 메커니즘을 여러 겹 두는 건 가능하지만, 서로 일관성이 없으면 오히려 헷갈리기 쉽습니다.

### Fast fallback도 중요하다

TCP에서는 primary/fallback address family를 race할 수 있습니다. 예전 `DualStack`은 deprecated이고, 지금은 `FallbackDelay`가 실제 제어 지점입니다.

## Resolution behavior는 생각보다 중요하다

패키지 문서는 명확합니다.

- 어떤 시스템은 pure Go resolver를 선호하고,
- 어떤 조건에서는 cgo/native resolver로 바뀌고,
- 막힌 cgo lookup은 OS thread를 소비하고,
- `GODEBUG=netdns=...`로 resolver 선택을 강제하거나 디버깅할 수 있습니다.

이건 trivia가 아닙니다. DNS 압박이 생기면 바로 latency 문제로 이어집니다.

## 왜 새 코드에서는 `netip`가 `net.IP`보다 더 나은가

`netip.Addr`는:

- 더 작고,
- immutable이며,
- comparable하고,
- map key로 안전합니다.

즉:

- endpoint table,
- allowlist/blocklist,
- address-keyed cache,
- config snapshot

에 더 잘 맞습니다.

`TCPAddrFromAddrPort`, `UDPAddrFromAddrPort` 같은 helper로 `net` 쪽과 연결할 수 있습니다.

## 런타임/소스 코드 워크

Go 1.26에서 읽을 만한 진입점:

- [`net/dial.go` `Dialer`](https://github.com/golang/go/blob/go1.26.0/src/net/dial.go#L126)
- [`net/dial.go` `DialContext`](https://github.com/golang/go/blob/go1.26.0/src/net/dial.go#L526)
- [`net/lookup.go` `Resolver`](https://github.com/golang/go/blob/go1.26.0/src/net/lookup.go#L134)
- [`net/netip/netip.go` `Addr`](https://github.com/golang/go/blob/go1.26.0/src/net/netip/netip.go#L37)
- [`net/netip/netip.go` `AddrPort`](https://github.com/golang/go/blob/go1.26.0/src/net/netip/netip.go#L1068)

특히 볼 만한 점:

- dialer는 timeout, absolute deadline, context에서 하나의 effective deadline을 만듭니다.
- resolution은 connect trace와 분리되어 다뤄집니다.
- `DialTCP`, `DialUDP`, `DialIP`에는 `netip` 친화적인 overload가 있습니다.
- `netip`는 `net.IP`의 allocation/comparability 문제를 피하려고 만들어졌습니다.

## 실패 패턴

### connect timeout이 이후 socket I/O도 커버한다고 생각

```go
conn, _ := dialer.DialContext(ctx, "tcp", addr)
bufio.NewReader(conn).ReadString('\n') // bug: deadline 없으면 여기선 무제한
```

### 주소를 끝까지 raw string으로만 다룸

validation과 comparison이 필요 이상으로 어렵고 약해집니다.

### 새 코드에서 `net.IP`를 map key처럼 사용

정규화와 equality 처리가 `netip.Addr`보다 훨씬 까다롭습니다.

### latency 디버깅에서 DNS를 무시

remote service가 멀쩡해도 resolution이 tail latency를 지배할 수 있습니다.

### listener accept loop의 shutdown ownership이 없음

`go handle(conn)`를 쓴 순간부터 listener shutdown contract도 같이 설계해야 합니다.

## 프로덕션에서의 의미

- outbound connection establish에는 실제 dial budget을 둡니다.
- connect 뒤에도 read/write deadline 또는 상위 request timeout을 명시합니다.
- configuration, identity, lookup table에는 `netip`를 우선합니다.
- resolution behavior도 서비스 경로 일부로 봐야 합니다.

## 예제와 테스트

- 예제: `examples/dialbudget`
- 테스트는 unreachable endpoint에서 healthy endpoint로 fallback하는 경로와, connect 후 probe I/O에 별도 deadline이 필요하다는 점을 검증합니다.

## 공식 자료

- [`net` 패키지 문서](https://pkg.go.dev/net)
- [`net/netip` 패키지 문서](https://pkg.go.dev/net/netip)

## Practical takeaway

좋은 Go 네트워킹 코드는 address representation, connect lifetime, post-connect I/O lifetime을 분리해서 다루는 데서 시작합니다. `net`과 `netip`는 그 분리를 명시적으로 드러내는 도구입니다.
