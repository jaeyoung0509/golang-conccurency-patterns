---
title: grpc-go 실전 플레이북
description: grpc-go를 long-lived channel, deadline, keepalive discipline, streaming ownership 관점에서 운영하는 법을 설명합니다.
---

# grpc-go 실전 플레이북

`grpc-go`는 단순히 “바이너리 HTTP”가 아닙니다.

이 라이브러리는 transport, channel, RPC policy가 한꺼번에 들어 있는 스택이라서, connection과 deadline을 대충 다루면 금방 복잡해집니다.

## Mental model

첫 번째 규칙은 단순합니다.

- `ClientConn`은 long-lived channel이고,
- RPC는 그 채널 위에서 잠깐 흐르는 work입니다.

이 수명을 뒤집으면 비용도 커지고 고장도 잦아집니다.

## 안전한 기본 스케치

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

이 스케치는 다음 기본값을 강제합니다.

- deprecated `Dial`이 아니라 `grpc.NewClient`
- long-lived channel 하나
- per-RPC deadline
- 명시적인 transport credential
- bounded connect backoff

## 운영 규칙

### `grpc.NewClient`를 우선한다

공식 패키지 문서는 이제 `NewClient`를 중심으로 설명합니다. 이 함수는 즉시 I/O를 하지 않고 channel만 만듭니다. 대부분의 경우 이 동작이 맞습니다.

### `Dial`, 특히 `WithBlock`은 기본값으로 쓰지 않는다

`Dial`은 deprecated이고 `WithBlock`은 공식 문서에서도 권장하지 않습니다. dependency가 잠깐 불안정한 것을 서비스 전체 startup outage로 바꾸기 쉽기 때문입니다.

### authority 또는 policy boundary마다 channel 하나

destination과 policy set 단위로 `ClientConn`을 만들고, 그 위에 여러 RPC를 태우는 구조가 맞습니다. 요청마다 channel을 열면 안 됩니다.

### deadline은 startup이 아니라 RPC에 붙인다

channel이 이미 있어도 각 RPC는 deadline이 필요합니다. 그렇지 않으면 queueing, retry, network stall의 lifetime이 무제한이 됩니다.

### streaming은 ownership discipline이 필요하다

stream을 중간에 버릴 거면 context를 cancel해야 하고, 끝까지 읽을 거면 `Recv()`를 `io.EOF`까지 돌려야 합니다. stream을 반쯤 방치하는 것은 HTTP body leak과 비슷한 문제입니다.

### keepalive는 협의된 운영 정책이다

aggressive client keepalive는 harmless한 설정이 아닙니다. 공식 keepalive 가이드는 지원되지 않는 cadence가 서버에서 `GOAWAY` `too_many_pings`를 유발할 수 있다고 명시합니다.

## 실패 패턴

### RPC마다 channel 생성

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

이렇게 하면 connection setup, resolver work, transport warmup을 전부 hot path에 태우게 됩니다.

### RPC에 deadline이 없음

```go
resp, err := client.GetUser(context.Background(), req) // bad: lifetime 무제한
```

### startup 정책으로 `WithBlock`

```go
cc, err := grpc.Dial(target, grpc.WithTransportCredentials(creds), grpc.WithBlock()) // bad default
```

### 너무 공격적인 keepalive

서비스가 그 cadence를 명시적으로 지원하지 않으면, 오히려 reconnect churn을 만드는 설정이 됩니다.

### cancel 없이 stream 방치

```go
stream, _ := client.Subscribe(ctx, req)
return nil // bad: stream context가 transport resource를 계속 소유
```

## 주의해서 써야 할 것

- `WithBlock`
- 명시적 retry policy 없이 blanket default로 두는 `WaitForReady`
- 조용한 channel에 대한 공격적인 keepalive
- staging에서 이해하지 못한 채 켜는 service-config 기능
- 서로 다른 trust/credential boundary를 하나의 `ClientConn`에 섞는 것

## Observability와 테스트

- `deadline_exceeded`와 `unavailable` 비율을 분리해서 봅니다.
- interceptor나 stats handler로 method latency, size, status code를 수집합니다.
- 실제 네트워크 없이 RPC 레벨 테스트를 하려면 `bufconn`이나 별도 통합 환경을 씁니다.
- 장애 분석 때는 “gRPC가 느리다”가 아니라 resolver, connect, TLS, handler latency를 분리해서 봐야 합니다.

## 언제 잘 맞는가

typed internal API, streaming, deadline, cross-language protocol compatibility가 필요하면 `grpc-go`는 강한 기본값입니다.

## 언제 다른 도구를 봐야 하나

트래픽이 단순한 HTTP/JSON이고 복잡성 대부분이 브라우저/사람 대상 API에서 나온다면 raw HTTP가 더 운영하기 쉬울 수 있습니다. 반대로 fire-and-forget 이벤트 전달만 필요하다면 RPC 스택 자체가 과한 추상화일 수 있습니다.

## 공식 자료

- [`google.golang.org/grpc` 패키지 문서](https://pkg.go.dev/google.golang.org/grpc)
- [gRPC keepalive guide](https://grpc.io/docs/guides/keepalive/)
- [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)

## Practical takeaway

grpc-go는 “호출마다 connection”이 아니라 “long-lived channel 위에 짧은 RPC budget”이라고 보는 순간 훨씬 다루기 쉬워집니다.
