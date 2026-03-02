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

## 무엇을 검증해야 하나

nontrivial concurrent component라면 최소한 다음을 테스트해야 합니다.

1. 수락한 작업이 정책대로 완료되거나 취소되는가
2. shutdown 시작 후 새 작업은 명확히 거부되는가
3. blocked goroutine에 escape path가 있는가
4. shutdown이 hard deadline을 존중하는가

## 이 저장소의 예시

### 첫 에러 이후 cancellation

`examples/workerpool/workerpool_test.go`는 단순히 에러를 확인하는 것이 아니라, 느린 작업이 cancellation을 관찰하는지까지 검증합니다.

### shutdown 시 drain

`examples/gracefulshutdown/gracefulshutdown_test.go`는 이미 받은 작업이 shutdown 전에 처리되는지 검증합니다.

### caller timeout 이후 broker non-blocking reply

`examples/requestreply/requestreply_test.go`는 caller가 떠난 뒤 broker가 reply send에서 멈추지 않는지 검증합니다.

## 유용한 테스트 형태

### deadline-bounded shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

if err := processor.Shutdown(ctx); err == nil {
    t.Fatal("expected shutdown deadline error")
}
```

### 명시적 completion protocol

goroutine 수를 전역적으로 세기보다:

- `WaitGroup`,
- closed `done` channel,
- `context` cancellation + wait

같은 completion protocol을 더 선호하는 편이 안정적입니다.

## 흔한 실수

### happy path만 테스트하는 것

lifecycle bug는 happy path에서는 잘 안 보입니다.

### `time.Sleep`를 종료 증거로 쓰는 것

`Sleep` 후 조용하다고 goroutine이 정말 끝났다는 뜻은 아닙니다.

### caller-abandoned path를 빼먹는 것

reply를 보내는 background worker / broker는 caller가 먼저 timeout되는 경우를 꼭 테스트해야 합니다.

## Practical takeaway

concurrent component에 start path가 있으면 end path도 있습니다. 둘 다 같은 무게로 테스트해야 합니다.
