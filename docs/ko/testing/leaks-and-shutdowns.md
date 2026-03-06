---
title: 리크, 종료, 타임아웃 테스트
description: goroutine 종료, queue drain, shutdown deadline을 어떻게 검증할지 설명합니다.
---

# 리크, 종료, 타임아웃 테스트

많은 동시성 회귀는 lifecycle 버그입니다.

- sender가 영원히 block되고,
- worker가 parent가 끝난 뒤에도 계속 돌고,
- shutdown이 끝나지 않으며,
- timeout된 caller가 background work를 이상하게 남겨둡니다.

이건 race detector만으로 잘 안 잡힙니다. 명시적인 종료 테스트가 필요합니다.

:::tip Quick takeaway
component에 `Start`, `Run`, `Submit`, `Check`가 있다면 lifetime contract도 있습니다. 좋은 테스트는 admission path와 exit path를 둘 다 증명합니다.
:::

## 무엇을 검증해야 하나

nontrivial concurrent component라면 최소한 다음을 테스트해야 합니다.

1. 수락한 작업이 정책대로 완료되거나 취소되는가
2. shutdown 시작 후 새 작업은 명확히 거부되는가
3. blocked goroutine에 escape path가 있는가
4. shutdown이 hard deadline을 존중하는가

## 테스트 가능한 shutdown contract

테스트 전에 contract를 먼저 분명히 해야 합니다. 좋은 concurrent component라면 최소한 아래 질문에 답이 있어야 합니다.

- shutdown 시작 후 어떤 작업까지 받는가
- 이미 받은 작업은 drain되는가, cancel되는가, drop되는가
- admission이 닫힌 뒤 caller는 어떤 에러를 받는가
- shutdown은 얼마나 기다릴 수 있는가
- background goroutine의 owner는 누구이고, 언제 종료가 보장되는가

구현이 이 질문에 답하지 못하면 테스트도 흐려집니다.

## 이 저장소의 예시

### 첫 에러 이후 cancellation

`examples/workerpool/workerpool_test.go`는 단순히 에러를 확인하는 것이 아니라, 느린 작업이 cancellation을 관찰하는지까지 검증합니다.

### shutdown 시 drain

`examples/gracefulshutdown/gracefulshutdown_test.go`는 이미 받은 작업이 shutdown 전에 처리되는지 검증합니다.

### caller timeout 이후 broker non-blocking reply

`examples/requestreply/requestreply_test.go`는 caller가 떠난 뒤 broker가 reply send에서 멈추지 않는지 검증합니다.

## 유용한 테스트 형태

### shutdown 이후 admission 거부

```go
func TestSubmitRejectedAfterShutdown(t *testing.T) {
	processor := newProcessor(t)

	require.NoError(t, processor.Shutdown(context.Background()))

	err := processor.Submit(context.Background(), Event{OrderID: "ord-1"})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Submit error = %v, want ErrClosed", err)
	}
}
```

이 테스트는 "가끔 막히고 가끔 들어가는" 애매한 동작 대신, caller가 항상 같은 답을 받는지 증명합니다.

### deadline-bounded shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

if err := processor.Shutdown(ctx); err == nil {
    t.Fatal("expected shutdown deadline error")
}
```

### blocked sender의 escape path

```go
func TestWorkerExitsOnCancel(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(done)
		runWorker(ctx)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("worker did not exit after cancellation")
	}
}
```

핵심은 `100ms`가 아니라 explicit completion signal입니다. timeout은 테스트가 영원히 걸리지 않게 막는 장치일 뿐입니다.

### caller-abandoned reply path

```go
func TestReplyPathDoesNotLeakAfterCallerTimeout(t *testing.T) {
	release := make(chan struct{})
	broker := newBroker(func() Result {
		<-release
		return Result{}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := broker.Check(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Check error = %v, want deadline exceeded", err)
	}

	close(release)
	if err := broker.Check(context.Background()); err != nil {
		t.Fatalf("follow-up request failed after timeout path: %v", err)
	}
}
```

이게 `examples/requestreply/requestreply_test.go`가 증명하는 핵심입니다. caller 하나가 떠났다고 broker 전체가 오염되면 안 됩니다.

### 명시적 completion protocol

goroutine 수를 전역적으로 세기보다:

- `WaitGroup`,
- closed `done` channel,
- `context` cancellation + wait

같은 completion protocol을 더 선호하는 편이 안정적입니다.

`runtime.NumGoroutine` 같은 전역 카운트는 CI smoke test 정도로는 쓸 수 있지만, runtime background activity 때문에 1차 oracle로 쓰기엔 약합니다.

## 실패 패턴

이런 테스트는 마음만 편하게 해주고 실제 종료를 증명하지는 못합니다.

```go
func TestWorkerStops(t *testing.T) {
	startWorker()
	time.Sleep(50 * time.Millisecond) // bad: sleep은 shutdown 증거가 아님
}
```

잠깐 기다렸더니 조용하다는 사실은 goroutine이 채널, 타이머, lock에 여전히 parked 되어 있지 않다는 뜻이 아닙니다. 종료 테스트에는 explicit completion signal이나 bounded `Shutdown` contract가 필요합니다.

## 흔한 실수

### happy path만 테스트하는 것

lifecycle bug는 happy path에서는 잘 안 보입니다.

### `time.Sleep`를 종료 증거로 쓰는 것

`Sleep` 후 조용하다고 goroutine이 정말 끝났다는 뜻은 아닙니다.

### caller-abandoned path를 빼먹는 것

reply를 보내는 background worker / broker는 caller가 먼저 timeout되는 경우를 꼭 테스트해야 합니다.

### cleanup에서 `t.Context()`를 재사용하는 것

`testing.T.Context()`는 cleanup 직전에 cancel됩니다. teardown에 살아 있는 context가 필요하면 cleanup 안에서 별도 bounded background context를 새로 만들어야 합니다.

## 실전 리뷰 체크리스트

1. admission이 언제 닫히는지 증명하는가
2. 이미 받은 작업이 어떻게 끝나는지 증명하는가
3. 모든 blocked goroutine에 cancel/close path가 있는가
4. 테스트가 sleep이 아니라 explicit completion을 쓰는가
5. teardown이 bounded deadline 안에 끝나는가

## Practical takeaway

concurrent component에 start path가 있으면 end path도 있습니다. 둘 다 같은 무게로 테스트해야 합니다.
