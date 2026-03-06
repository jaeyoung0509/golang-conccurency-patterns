---
title: Race Detector
description: Go race detector가 무엇을 잡고 무엇을 보장하지 않는지 설명합니다.
---

# Race Detector

nontrivial한 동시성 코드라면 가장 먼저 돌려야 하는 도구가 race detector입니다.

:::tip Quick takeaway
`go test -race`는 concurrent Go의 최소 기준이지 끝이 아닙니다. shared memory에 대한 빠진 동기화는 잘 잡지만, shutdown, leak, timeout correctness 자체를 증명해주지는 못합니다.
:::

## 머릿속 모델

race detector가 묻는 질문은 하나입니다.

> 같은 메모리를 두 goroutine이 동시에 만졌는데, 그 사이에 유효한 happens-before edge가 있었는가?

그래서 이 도구는 [Go Memory Model](https://go.dev/ref/mem)과 같이 봐야 합니다. channel ownership transfer, mutex, atomic, `WaitGroup` completion 같은 경계를 설계했다면, detector는 코드가 실제로 그 계약을 지키는지 확인하는 데 도움을 줍니다.

## 무엇을 잡나

같은 메모리 위치에 대한 concurrent access 중 하나 이상이 write이고, 그 사이에 동기화 경계가 없을 때 발생하는 data race를 잡습니다.

대표 예시는:

- 한 goroutine이 map을 읽고 다른 goroutine이 동시에 쓰는 경우,
- channel protocol이 보호해야 할 상태를 바깥에서 따로 읽거나 쓰는 경우,
- cache update에 lock을 빠뜨린 경우.

즉, "공유 mutable state를 실수로 열어둔 경우"를 잡는 데 특화된 도구이지, 모든 동시성 버그를 잡는 도구는 아닙니다.

## 작은 race 예시

```go
func TestRacyMap(t *testing.T) {
    cache := map[string]int{}

    go func() { cache["a"] = 1 }()
    go func() { _ = cache["a"] }()
}
```

이 코드는 casual run에서는 조용히 지나갈 수도 있습니다. `-race`를 돌리면 바로 잡아야 하는 버그라는 점이 드러납니다.

일부 map race는 `-race` 없이도 panic이 나지만, 대부분의 shared-state 버그는 그렇게 친절하지 않습니다. 그래서 detector가 필요합니다.

## `-race`가 깨끗해도 여전히 깨지는 예시

아래 코드는 race-free지만 여전히 잘못됐습니다.

```go
func fetchWithTimeout(ctx context.Context) error {
	reply := make(chan Result)

	go func() {
		reply <- slowQuery()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply:
		return nil
	}
}
```

caller가 먼저 timeout되면 goroutine은 `reply <- slowQuery()`에서 영원히 막힐 수 있습니다. shared-memory race는 없으니 race detector는 조용합니다. 그래서 leak / shutdown 테스트가 별도로 필요합니다.

## 무엇을 보장하지 않나

`-race`가 깨끗하다고 해서 다음이 보장되지는 않습니다.

- deadlock 부재,
- goroutine leak 부재,
- fairness,
- 좋은 shutdown behavior,
- 좋은 timeout behavior.

또한 실제로 실행된 경로만 관측할 수 있습니다.

## 사용법

기본은 전체 모듈입니다.

```bash
go test -race ./...
```

빠르게 돌릴 땐 이렇게 줄일 수 있습니다.

```bash
go test -race ./examples/workerpool -run TestGenerateQuotesCancelsSlowJobsAfterError
go test -race -count=20 -shuffle=on ./...
go test -race -cpu=1,4 ./...
```

자주 쓰는 형태는 다음과 같습니다.

| 커맨드 | 용도 |
| --- | --- |
| `go test -race ./...` | 전체 baseline |
| `go test -race ./pkg -run TestName` | 특정 테스트 빠른 반복 |
| `go test -race -count=20 -shuffle=on ./...` | 스케줄 민감한 경로 흔들기 |
| `go test -race -cpu=1,4 ./...` | 높은 병렬도에서만 깨지는 코드 찾기 |

느리고 무거운 건 정상입니다. 동시성 코드에선 그 비용을 감수할 가치가 충분합니다.

## race report는 어떻게 읽어야 하나

report에서 중요한 건 두 가지입니다.

1. 충돌한 access 두 개
2. 그 access를 만든 goroutine 생성 스택

보통 진짜 원인은 그 두 줄보다 한 단계 위에 있습니다.

- state owner가 우회됐거나
- mutex가 보호해야 할 map이 밖으로 새었거나
- background goroutine이 context lifetime 밖으로 나갔거나
- 리팩터링이 synchronization fast path를 깨뜨린 경우입니다.

## report를 본 뒤의 정석 순서

1. ownership을 잃은 mutable state가 무엇인지 찾습니다.
2. ownership, mutex, atomic, message passing 중 어떤 경계가 맞는지 결정합니다.
3. 임시 lock 추가보다 우회 경로 자체를 제거합니다.
4. 가장 좁은 `-race` 명령으로 다시 확인하고 마지막에 전체 스위트를 돌립니다.

## 좋은 사용 방식

이 도구는 이런 질문에 답합니다.

- mutable memory를 실수로 공유했는가
- synchronization boundary를 우회하는 경로가 생겼는가
- 리팩터링이 ownership contract를 깨뜨렸는가

하지만 이것만으로 validation을 끝내면 안 됩니다.

## 실패 패턴

```go
func TestSomething(t *testing.T) {
    go mutateSharedState()
    time.Sleep(10 * time.Millisecond)
}
```

이런 테스트는 한동안은 통과하면서 잘못된 자신감만 줍니다. `-race`로 돌리고, 그다음 timing luck을 실제 synchronization으로 바꿔야 합니다.

또 다른 실패 패턴은 "`-race`가 깨끗하니 프로토콜도 맞다"라고 생각하는 것입니다. 실제 버그가 goroutine leak, shutdown wedge, timeout 이후 reply block이라면 detector는 평생 초록일 수 있습니다.

## `-race` 친화적인 설계 패턴

- mutable state machine마다 single owner goroutine 두기
- 짧은 critical section을 가진 `map + mutex`
- one-shot response에는 explicit reply channel 쓰기
- 테스트를 sleep 기반이 아니라 실제 completion signal 기반으로 쓰기

## 공식 자료

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)

## Practical takeaway

`-race`는 최소 기준이지 종착점이 아닙니다.

초기에 자주 돌리고, deterministic test와 shutdown test와 함께 써야 합니다.
