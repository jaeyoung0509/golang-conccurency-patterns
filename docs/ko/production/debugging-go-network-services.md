---
title: 프로덕션에서 Go 네트워크 서비스 디버깅하기
description: DNS, connect, TLS, request, body 단계로 Go 네트워크 지연을 분해하고 문제가 코드인지 transport policy인지 네트워크인지 증명하는 방법을 설명합니다.
---

# 프로덕션에서 Go 네트워크 서비스 디버깅하기

Go 네트워크 서비스가 느릴 때, “네트워크가 느리다”는 진단은 거의 도움이 되지 않습니다.

시간이 정확히 어디서 쓰였는지를 증명해야 합니다.

## Mental model

하나의 outbound HTTP 혹은 RPC 호출도 보통 최소 다음 단계를 가집니다.

| 단계 | 주 owner |
| --- | --- |
| DNS lookup | resolver / 환경 / service discovery |
| TCP connect | dialer, routing, remote accept path |
| TLS handshake | `crypto/tls`, certificate policy, ALPN |
| request write | client, kernel buffer, peer read speed |
| first-byte wait | remote handler 혹은 upstream queue |
| response body read | transport reuse discipline, peer throughput, caller ownership |

이걸 전부 하나의 “request latency”로 뭉개면 디버깅은 추측이 됩니다.

## 첫 번째 디버깅 동작: 경로를 분해하라

HTTP client라면 `httptrace`가 가장 빠른 출발점입니다.

```go
trace := &httptrace.ClientTrace{
	DNSStart:             func(httptrace.DNSStartInfo) {},
	DNSDone:              func(httptrace.DNSDoneInfo) {},
	ConnectStart:         func(_, _ string) {},
	ConnectDone:          func(_, _ string, _ error) {},
	TLSHandshakeStart:    func() {},
	TLSHandshakeDone:     func(tls.ConnectionState, error) {},
	GotConn:              func(httptrace.GotConnInfo) {},
	GotFirstResponseByte: func() {},
}
```

목적은 모든 hook을 영원히 로깅하는 게 아니라, incident 동안 어느 단계가 변했는지를 증명하는 것입니다.

## 실전 디버깅 워크플로

1. 문제 범위가 inbound인지 outbound인지, 혹은 둘 다인지 확인한다
2. latency를 DNS, connect, TLS, handler, body-read 단계로 분해한다
3. connection reuse가 일어나는지, churn이 나는지 확인한다
4. 시간의 대부분이 network wait인지, CPU인지, lock contention인지, downstream queueing인지 확인한다
5. `TIME_WAIT`, `ESTABLISHED`, connection explosion 같은 socket state를 본다
6. 그다음에야 application code, transport policy, network 중 어디가 원인인지 판단한다

## 도구별로 무엇을 증명하는가

### `httptrace`

다음을 증명할 때 씁니다.

- reuse vs new dial
- DNS latency
- connect latency
- TLS handshake latency
- first-byte delay

### `runtime/trace`

다음을 증명할 때 씁니다.

- goroutine이 network wait에 막혔는지 lock에 막혔는지
- cancellation과 wakeup timing
- stalled request 주변의 scheduler transition

### `pprof`

다음을 증명할 때 씁니다.

- 시간이 실제로 handler logic이나 encoding에 쓰였는지
- heap growth가 body buffering이나 request fan-out 때문인지
- transport 바깥의 contention이 있는지

### `ss`와 `lsof`

다음을 증명할 때 씁니다.

- socket 개수
- `TIME_WAIT` churn
- process별 descriptor pressure
- reuse보다 connection 생성이 빠른지 여부

### `tcpdump`

application metric과 socket metric이 엇갈리거나 retransmission, reset, missing response를 packet 레벨로 확인해야 할 때 씁니다.

### `GODEBUG`

이 트랙에서 유용한 값:

- `GODEBUG=netdns=go+2` for resolver choice and lookup behavior
- `GODEBUG=http2debug=1` 또는 `2` for HTTP/2 behavior
- runtime starvation이 의심될 때 scheduler trace

## 실패 패턴

### `context deadline exceeded`를 root cause로 착각

그 에러는 budget이 만료됐다는 뜻일 뿐, 어떤 단계가 budget을 먹었는지는 알려주지 않습니다.

### hot path에서 client나 transport를 계속 새로 만듦

그러면 network instability처럼 보이는 것이 사실은 connection churn일 수 있습니다.

### response body를 drain/close하지 않음

reuse가 무너지고 다음 요청이 dial/handshake cost를 다시 내게 됩니다.

### 서비스 레벨 latency만 봄

“p95 = 800ms” 그래프 하나로는 DNS, TLS, queueing, body ownership 문제를 구분할 수 없습니다.

### socket state 폭증을 무시

`TIME_WAIT`나 connection count가 폭증하면 문제는 handler CPU가 아니라 transport lifecycle일 수 있습니다.

## 어느 레이어가 원인인지 증명하는 법

### application layer가 원인일 때의 징후

- handler나 JSON encoding에서 CPU가 높음
- goroutine이 mutex나 channel에 막힘
- network timing은 안정적인데 first-byte latency가 큼
- downstream fan-out이나 queue saturation이 보임

### transport policy가 원인일 때의 징후

- 거의 모든 요청이 new dial
- upstream이 안정적인데 reuse가 낮음
- body close discipline이 엉망
- HTTP/2 혹은 gRPC channel misuse

### network/environment가 원인일 때의 징후

- connect 전에 DNS spike가 먼저 보임
- packet capture에 retransmission이나 reset이 보임
- handler 실행 전 connect 단계에서 timeout
- cluster-level packet path나 service discovery 문제가 드러남

## 이 핸드북에서 같이 읽을 문서

- [net과 netip](/ko/stdlib/net-and-netip)
- [프로덕션에서의 crypto/tls](/ko/stdlib/crypto-tls)
- [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)
- [HTTP/2, ALPN, 그리고 Stream Multiplexing](/ko/stdlib/http2-alpn-stream-multiplexing)
- [grpc-go 실전 플레이북](/ko/playbooks/grpc-go-production-playbook)
- [트레이싱과 경합 관측](/ko/testing/tracing-and-profiling)

## Practical takeaway

프로덕션 네트워크 디버깅은 “네트워크가 느린가?”를 묻는 순간부터 어려워집니다. 대신 “이 connection lifecycle의 어느 단계가 변했는가?”를 물어야 합니다.
