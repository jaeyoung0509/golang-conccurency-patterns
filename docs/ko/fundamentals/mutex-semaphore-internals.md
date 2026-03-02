---
title: Mutex와 런타임 세마포어 내부
description: sync.Mutex의 fast/slow path, starvation mode, 그리고 parking/wakeup에 쓰이는 런타임 세마포어를 설명합니다.
---

# Mutex와 런타임 세마포어 내부

Go는 채널로 유명하지만, 실제 시스템에서는 mutex도 매우 중요합니다.

Go 동시성을 깊게 이해하려면 다음 둘을 함께 이해해야 합니다.

- public API인 `sync.Mutex`
- contention 시 그 아래에서 동작하는 runtime semaphore 메커니즘

:::info 버전 기준
이 문서는 Go 1.26의 `internal/sync/mutex.go`, `runtime/sema.go` 기준으로 정리했습니다.
:::

## `sync.Mutex`는 표면은 작고 내부는 풍부하다

구현 수준에서 mutex는 대략 다음을 가집니다.

- `state` 비트 필드
- waiter를 park/wakeup하기 위한 semaphore word

중요한 state bit는 다음과 같습니다.

- `mutexLocked`
- `mutexWoken`
- `mutexStarving`
- waiter count 비트들

## Fast path: 낙관적 CAS

경합이 없는 경우 경로는 매우 단순합니다.

1. `0 -> mutexLocked` CAS 시도
2. 성공하면 즉시 반환

그래서 uncontended mutex는 매우 쌉니다.

## 단순화한 내부 스케치

```go
type Mutex struct {
	state int32
	sema  uint32
}

func Lock(m *Mutex) {
	if CAS(&m.state, 0, mutexLocked) {
		return
	}
	lockSlow(m)
}

func Unlock(m *Mutex) {
	if atomicAdd(&m.state, -mutexLocked) == 0 {
		return
	}
	unlockSlow(m)
}
```

핵심 mental model은 이것입니다. 아주 싼 fast path와, contention이 생겼을 때 훨씬 풍부한 slow path.

## Slow path: spin, queue, sleep

fast path가 실패하면 `lockSlow()`로 들어갑니다.

이 경로에서 런타임은 상황에 따라:

- 잠깐 spin하고,
- 자신이 woken waiter임을 표시하고,
- waiter count를 늘리고,
- runtime semaphore 위에서 sleep할 수 있습니다.

즉, 곧 풀릴 락이라면 비싼 park를 피하려 하고, 실제 contention이라면 결국 sleep으로 내려갑니다.

## Normal mode와 starvation mode

`internal/sync/mutex.go`에서 아주 중요한 점은 Go mutex가 두 가지 모드를 가진다는 사실입니다.

- normal mode
- starvation mode

### Normal mode

normal mode에서는:

- waiter가 큐에 들어가더라도,
- 깨어난 waiter가 자동으로 락을 소유하진 않고,
- 새로 도착한 goroutine이 먼저 가져갈 수 있습니다.

이 방식은 이미 CPU 위에 있는 goroutine이 빠르게 락을 잡을 수 있어서 throughput이 좋습니다.

### Starvation mode

waiter가 너무 오래 기다리면 mutex는 starvation mode로 들어갈 수 있습니다.

이 모드에서는:

- 맨 앞 waiter에게 ownership을 직접 넘기고,
- 새 goroutine은 끼어들지 못하며,
- throughput은 약간 떨어져도 tail latency 병목은 완화됩니다.

현재 소스는 starvation 판단 기준으로 대략 1ms를 사용합니다.

즉, Go는 완벽한 단순성보다 실용적 fairness를 선택합니다.

## Unlock도 단순히 0으로 만드는 게 아니다

unlock은 먼저 fast path로 locked bit를 내립니다.

waiter가 없으면 그걸로 끝납니다.

경합 중이라면 slow path에서:

- waiter 하나를 깨울지,
- starvation mode에서는 ownership을 직접 handoff할지

를 결정합니다.

이 handoff 정책 때문에 mutex 성능은 단순 spinlock보다 훨씬 미묘합니다.

## 런타임 세마포어는 일반-purpose semaphore API가 아니다

`runtime/sema.go`의 주석은 중요한 사실을 말합니다. 이 세마포어는 일반적인 counting semaphore라기보다, 상위 동기화 primitive를 구현하기 위한 sleep-and-wakeup 메커니즘입니다.

의도는 futex와 비슷합니다.

- 각 sleep은 정확한 wakeup과 짝이 맞아야 하고,
- race가 있어도 wakeup이 유실되면 안 되며,
- mutex, wait group 같은 상위 primitive가 그 위에 구축될 수 있어야 합니다.

## 런타임 소스 포인터

- [internal/sync/mutex.go](https://github.com/golang/go/blob/go1.26.0/src/internal/sync/mutex.go)
- [runtime/sema.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/sema.go)

mutex 소스는 throughput vs fairness trade를 이해하기 가장 좋고, `runtime/sema.go`는 sleeping / waking의 기반을 보여줍니다.

## contention은 어떻게 관측하나

- mutex profile
- block profile
- `go test -trace=trace.out ./...`

코드 리뷰상 멀쩡해 보여도 tail latency가 높다면, contention이 빠진 차원인 경우가 많습니다.

## Waiter는 어떻게 추적되는가

런타임은 대기 주소를 semaphore table로 해시합니다.

각 root는 대략 다음을 가집니다.

- lock
- waiter count
- 주소 기준으로 정리된 waiter `sudog` 트리

즉, 모든 경합 주소를 하나의 전역 리스트로 관리하는 대신, 여러 독립 경합 지점을 확장 가능하게 처리하려는 설계입니다.

## 채널과 mutex가 둘 다 중요한 이유

채널과 mutex는 잘 맞는 문제가 다릅니다.

| 도구 | 잘하는 문제 |
| --- | --- |
| Channel | 통신 프로토콜, 작업 분배, ownership transfer |
| Mutex | 작은 shared state 보호, 짧은 critical section |

Go의 강점은 둘 중 하나를 종교처럼 고집하는 데 있지 않고, 문제 형태에 맞는 primitive를 고를 수 있다는 데 있습니다.

## 흔한 오해

### "채널이 항상 mutex보다 더 Go답다"

아닙니다.

작은 in-memory map 보호나 짧은 critical section이라면 mutex가 더 단순하고 빠르며 유지보수도 쉽습니다.

### "mutex contention은 throughput만 낮춘다"

그것도 아닙니다.

mutex는 throughput과 fairness를 동시에 조절하므로 contention은 다음에도 영향을 줍니다.

- tail latency
- goroutine parking behavior
- scheduler interaction
- 전체 서비스 responsiveness

## 실전 설계에 주는 의미

이 내부 구조를 이해하면 설계가 달라집니다.

- critical section을 짧게 유지하고,
- lock을 잡은 채 느린 I/O를 하지 말고,
- fairness를 자동 보장으로 착각하지 말고,
- 공유 메모리보다 explicit ownership이 낫다면 actor나 channel을 고려하고,
- 상태가 로컬하고 프로토콜이 단순하면 mutex를 쓰는 편이 낫습니다.

## 실패 패턴

이 코드는 컴파일은 잘 되지만 latency를 망가뜨리기 딱 좋습니다.

```go
mu.Lock()
defer mu.Unlock()

resp, err := http.Get(url) // bad: lock을 잡은 채 느린 I/O 수행
if err != nil {
	return err
}
defer resp.Body.Close()
```

이런 코드 뒤에 waiter가 쌓이기 시작하면, 문제는 단순한 코드 냄새가 아니라 scheduler와 tail latency 전체에 영향을 주는 contention 이슈가 됩니다.

## 실전 요약

Go의 public synchronization primitive가 간단해 보이는 이유는, 런타임이 상당한 복잡도를 대신 흡수하고 있기 때문입니다.

그 복잡도가 있기에 mutex는 common case에서 빠르고, contention 상황에서도 완전히 무너지지 않습니다.

다음으로 [구조화된 동시성](/ko/advanced/structured-concurrency)이나 [가중 세마포어](/ko/advanced/weighted-semaphore)를 보면 이런 기반 위에 어떤 고급 패턴을 쌓을 수 있는지 이어서 볼 수 있습니다.
