---
title: synctest로 결정적 테스트
description: testing/synctest를 사용해 실제 시간 sleep 없이 timeout-heavy concurrent 테스트를 결정적으로 만드는 방법을 설명합니다.
---

# synctest로 결정적 테스트

오랫동안 Go 동시성 테스트의 가장 큰 고통 중 하나는 시간 문제였습니다.

`time.Sleep`에 의존한 테스트는 대개:

- 느리고,
- flaky하며,
- 머신 부하에 민감하고,
- 설득력이 약했습니다.

`testing/synctest`는 이 부분을 크게 바꿉니다.

:::tip Quick takeaway
`synctest`는 bubble-local concurrency 세계와 fake clock을 제공합니다. bubble 안의 goroutine들이 durably blocked 되었을 때 시간이 전진하므로 timeout / cancellation 테스트를 결정적으로 만들 수 있습니다.
:::

## 핵심 아이디어

`synctest.Test` bubble 안에서는:

- 테스트가 만든 goroutine이 bubble에 속하고,
- `time` 패키지가 fake clock을 쓰며,
- `synctest.Wait()`가 다른 goroutine이 durably blocked 될 때까지 기다립니다.

즉, 벽시계 시간에 기대지 않고 timeout 로직을 검증할 수 있습니다.

## 실전 예시

```go
func TestContextTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
        defer cancel()

        time.Sleep(5*time.Second - time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != nil {
            t.Fatalf("timeout fired too early: %v", err)
        }

        time.Sleep(time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("timeout missing: %v", err)
        }
    })
}
```

이 테스트는 실제 5초를 기다리지 않습니다. 즉시, 그리고 결정적으로 실행됩니다.

## "durably blocked"가 왜 중요하나

표준 라이브러리 문서 기준으로:

- bubble 안에서 만든 channel에 대한 blocking send/receive,
- bubble에 연결된 `WaitGroup.Wait`,
- `sync.Cond.Wait`,
- `time.Sleep`

등은 durably blocked로 간주될 수 있습니다.

반면:

- network I/O,
- 외부 syscall,
- bubble 바깥 이벤트에 의존한 block

은 같은 방식으로 다뤄지지 않습니다.

그래서 `synctest`는 fake dependency와 self-contained 테스트에서 특히 강합니다.

## 어디에 잘 맞나

- timeout-heavy request code,
- cancellation propagation,
- background callback,
- internal worker coordination,
- timer / deadline 기반 retry loop.

## 무엇을 대체하지는 않나

다음을 대체하지 않습니다.

- `-race`
- real network integration test
- trace / profile 분석
- 외부 시스템과 얽힌 leak test

## 실패 패턴

wall-clock sleep에 기대는 순간 동시성 테스트는 다시 flaky해집니다.

```go
go func() {
	time.Sleep(100 * time.Millisecond)
	done <- struct{}{}
}()

time.Sleep(10 * time.Millisecond) // bad: 그냥 timing guess
```

`testing/synctest`는 이런 내부 timer / scheduling 동작을 운에 맡기지 않고 deterministic하게 다루기 위해 존재합니다.

## 공식 자료

- [Testing Time (Go Blog)](https://go.dev/blog/testing-time)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)

## Practical takeaway

동시성 테스트가 아직도 arbitrary sleep에 기대고 있다면, `synctest`는 가장 높은 레버리지의 개선 중 하나입니다.
