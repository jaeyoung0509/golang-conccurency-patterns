---
title: net/http 실전 필드 가이드
description: Go HTTP client/server stack을 timeout, pooling, lifecycle 정책 관점에서 운영하는 법을 설명합니다.
---

# net/http 실전 필드 가이드

이 문서는 [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)와 의도가 다릅니다.

저 문서가 패키지 내부 동작을 설명한다면, 이 문서는 그 패키지에 실제 트래픽을 어떻게 안전하게 태울지 설명합니다.

## Mental model

`net/http`에는 두 개의 장기 소유자가 있습니다.

- `http.Server`: inbound connection lifetime 소유
- `http.Client` + `Transport`: outbound connection lifetime 소유

대부분의 프로덕션 문제는 둘 중 하나를 정책 경계가 아니라 일회성 helper처럼 다룰 때 시작됩니다.

## 안전한 기본 스케치

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

이 설정이 유일한 정답은 아니지만, 출발점으로는 안전합니다.

- long-lived transport 하나,
- long-lived client 하나,
- 명시적인 server-side read 경계,
- 그 위에 request-scoped deadline을 추가하는 구조

이기 때문입니다.

## 운영 규칙

### client와 transport를 재사용한다

요청마다 `Transport`를 새로 만들면 pooling을 스스로 깨는 셈입니다. transport는 outbound policy boundary 단위의 singleton으로 다루는 편이 맞습니다.

### request context와 transport timeout을 같이 쓴다

`Client.Timeout`은 전체 exchange를 한 번에 자르는 blunt tool입니다. 간단한 경우엔 유용하지만, 큰 시스템에서는 보통 다음을 같이 써야 합니다.

- business budget용 request context deadline
- 특정 네트워크 edge용 transport timeout
- inbound 방어용 server read/idle timeout

### response body는 connection ownership 문제다

body를 닫는 것은 단순 cleanup이 아닙니다. transport가 이 connection을 재사용해도 되는지 판단하는 신호입니다.

중간에 읽기를 포기할 때는 의식적으로 둘 중 하나를 택해야 합니다.

- bounded drain 후 재사용
- 즉시 close 후 재사용 포기

실수로 둘 다 안 하면 안 됩니다.

### streaming endpoint는 timeout 전략이 달라진다

SSE나 긴 다운로드 같은 streaming response에는 server `WriteTimeout`이 오히려 맞지 않을 수 있습니다. 이런 endpoint는 generic timeout 하나로 덮지 말고 별도 정책을 가져가는 편이 낫습니다.

## 실패 패턴

### hot path에서 client를 매번 생성

```go
func fetch(url string) (*http.Response, error) {
	client := &http.Client{Timeout: 2 * time.Second} // bad: 호출마다 새 transport 정책
	return client.Get(url)
}
```

### 중요한 dependency에서 숨은 default client 사용

```go
resp, err := http.Get(url) // bad: transport budget, proxy policy, reuse 설정이 보이지 않음
```

### body ownership 누락

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
return decode(resp.Body) // bad: body를 닫지 않음
```

### 모든 걸 큰 timeout 하나로 처리

```go
client := &http.Client{Timeout: 30 * time.Second}
```

이렇게 하면 dial budget이 필요한지, response-header budget이 필요한지, handler deadline이 필요한지, shutdown contract가 필요한지가 모두 흐려집니다.

## 주의해서 써야 할 것

- 라이브러리 성격의 코드에서 `http.DefaultClient`, `http.DefaultTransport`
- streaming handler에 대한 `WriteTimeout`
- huge body나 attacker-controlled body에 대한 무조건적인 `io.Copy(io.Discard, resp.Body)`
- 실제 context discipline 대신 `TimeoutHandler`로 덮는 방식

## Observability와 테스트

- `httptest.Server`로 client timeout edge와 body-close discipline을 검증합니다.
- `Server.Shutdown` 테스트로 deploy 중 in-flight request 동작을 검증합니다.
- reuse vs redial 증명이 필요하면 `httptrace`를 씁니다.
- request duration, status code, inflight count, outbound dependency latency를 분리해서 봅니다. 섞으면 transport churn이 가려집니다.

## 언제 잘 맞는가

대부분의 Go 서비스, control plane, API, webhook, 내부 서비스 간 호출은 `net/http`가 기본값으로 맞습니다. 표준 라이브러리는 policy boundary만 의식해서 쓰면 대부분의 workload를 충분히 커버합니다.

## 언제 다른 도구를 봐야 하나

매우 특수한 L7 proxy, custom wire protocol gateway, 혹은 hot path의 추가 allocation 하나도 치명적인 환경이라면 더 특화된 스택이 필요할 수 있습니다. 그래도 그 전에 진짜 병목이 `net/http` 자체인지, 아니면 timeout/reuse/buffering discipline이 없는 것인지부터 증명해야 합니다.

## 공식 자료

- [`net/http` 패키지 문서](https://pkg.go.dev/net/http)
- [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)

## Practical takeaway

프로덕션 `net/http`는 대부분 ownership 문제입니다.

client reuse, response body, timeout edge를 명시하면 이 패키지는 매우 믿을 만합니다. 그 정책을 암묵적으로 두면 latency와 resource usage를 설명하기 어려워집니다.
