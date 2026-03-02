---
title: 대규모 Go 시스템
description: latency, memory, admission control, lifecycle, observability 관점에서 대규모 Go 서비스가 알아야 할 동시성 규칙을 정리합니다.
---

# 대규모 Go 시스템

튜토리얼 코드와 프로덕션 코드 사이의 가장 큰 차이는 문법이 아니라 운영 압력입니다.

실제 부하를 받기 시작하면 핵심 질문은 한계, 소유권, 가시성, 실패 격리로 바뀝니다.

:::tip Quick takeaway
대규모 시스템에서 가장 비싼 동시성 버그는 보통 "goroutine 문법 실수"가 아닙니다. 무한 queue, 새는 lifetime, 빠진 deadline, hot shared state, 약한 observability가 더 위험합니다.
:::

## 규칙 1: 모든 goroutine에는 owner가 있어야 한다

다음 질문에 답할 수 있어야 합니다.

- 누가 이 goroutine을 시작했는가
- 무엇이 이 goroutine을 멈추게 하는가
- 누가 종료를 기다리는가
- parent request나 process가 먼저 끝나면 어떻게 되는가

이 질문에 답할 수 없으면 아직 production-grade concurrency design이 아닙니다.

## 규칙 2: cleanup보다 admission control이 먼저다

많은 시스템이 overload를 나중 cleanup으로 버티려 합니다.

그건 늦습니다.

더 좋은 구조는:

- queue를 제한하고,
- worker 수를 제한하고,
- 초기에 reject 또는 degrade하고,
- latency를 예측 가능하게 유지하는 것입니다.

```go
func Submit(ctx context.Context, job Job) error {
    select {
    case jobs <- job:
        return nil
    default:
        return ErrOverloaded
    }
}
```

이 스케치의 핵심은 overload policy가 인터페이스에 드러난다는 점입니다.

## 규칙 3: deadline은 전체 호출 트리로 흘러야 한다

경계에서 timeout을 걸었다고 끝이 아닙니다.

leaf goroutine이 파생 context를 받지 않으면, timeout은 control mechanism이 아니라 로그 이벤트에 불과합니다.

### 실패 패턴

```go
ctx, cancel := context.WithTimeout(parent, 100*time.Millisecond)
defer cancel()

go fetchFromBackend(context.Background(), id) // 잘못됨: request lifetime과 분리됨
```

이게 request timeout이 background leak로 바뀌는 전형적인 방식입니다.

## 규칙 4: queue는 단순 buffer가 아니라 latency object다

queue는 세 가지를 합니다.

- burst를 흡수하고,
- overload를 잠시 숨기고,
- 서비스 압력을 대기 시간으로 바꿉니다.

그래서 queue length, queue age, drop policy는 구현 디테일이 아니라 운영 신호입니다.

## 규칙 5: hot shared state는 ownership 결정을 요구한다

map, cache, state machine이 hot해지면 의도적으로 선택해야 합니다.

- `map + mutex`
- actor ownership
- sharding
- read-mostly atomic snapshot
- per-key worker / mailbox partitioning

이건 취향 문제가 아니라 tail latency와 failure behavior를 바꾸는 선택입니다.

## 규칙 6: cgo와 syscall은 scheduler 현실을 바꾼다

순수 Go 네트워크 서비스는 runtime netpoller의 이점을 강하게 받습니다.

하지만 cgo나 opaque blocking syscall이 많아지면:

- thread가 더 오래 block될 수 있고,
- `P` handoff가 더 중요해지며,
- pure Go 기반 직감이 덜 맞게 됩니다.

그래서 "그냥 I/O bound예요"는 충분한 진단이 아닙니다.

## 규칙 7: 메모리는 concurrency budget의 일부다

대규모 시스템에서는 heap shape 자체가 concurrency behavior를 바꿉니다.

allocation rate가 높으면:

- GC assist가 늘고,
- background marking pressure가 커지고,
- object churn에 대한 latency 민감도가 올라갑니다.

즉, 요청 하나가 많은 goroutine fan-out과 allocation을 만든다면 GC도 concurrency story의 일부입니다.

## 규칙 8: shutdown도 correctness의 일부다

deploy, rollout, node drain은 늘 일어납니다.

프로덕션 서비스는 shutdown contract가 있어야 합니다.

1. 새 작업 수락 중단
2. 이미 받은 작업은 정책대로 drain 또는 cancel
3. hard deadline 강제
4. shutdown이 깔끔히 끝나지 않을 때 metric / log 노출

## 규칙 9: trace와 profile은 선택이 아니다

시스템이 충분히 커지면 직감만으로는 부족합니다.

필요합니다.

- `go test -trace`와 `go tool trace`
- block / mutex profile
- heap / alloc profile
- queue metric
- timeout / cancellation metric
- 필요시 goroutine lifecycle counter

## 실전 체크리스트

- 비싼 외부 경계마다 bounded concurrency 존재
- queue capacity와 overload behavior가 명시적
- request-scoped context가 모든 하위 goroutine에 전달됨
- graceful shutdown에 실제 deadline 존재
- 정당화 없는 unbounded mailbox 없음
- hot shared map에 ownership model이 명시됨
- incident 전부터 runtime observability가 준비됨

## Practical takeaway

프로덕션 Go 동시성은 "더 많은 goroutine"의 문제가 아닙니다.

어디서 concurrency를 멈추고, 어디서 기다리게 하고, 어디서 fail fast할지를 정하는 문제입니다.
