---
title: 트레이싱과 경합 관측
description: runtime trace, block profile, mutex profile, scheduler debug output으로 실제 동시성 동작을 관측하는 방법을 설명합니다.
---

# 트레이싱과 경합 관측

어떤 동시성 버그는 runtime behavior를 직접 보기 전에는 잘 드러나지 않습니다.

그 경우 unit test만으로는 부족합니다.

## 주요 도구

| 도구 | 잘 맞는 문제 |
| --- | --- |
| `go test -trace` | goroutine state, scheduler activity, netpoll, syscall, GC |
| block profile | 어디서 goroutine이 block되는지 |
| mutex profile | lock contention hotspot |
| heap / alloc profile | allocation pressure와 GC 연결 |
| `GODEBUG=schedtrace=...` | scheduler snapshot과 run queue pressure |

## execution trace

`runtime/trace`와 `go test -trace=trace.out`은 다음을 보여줍니다.

- goroutine 생성과 block,
- syscall enter/exit,
- GC activity,
- processor activity,
- user region과 task.

즉, "runtime이 실제로 뭘 하고 있는가"를 보기에 가장 좋은 출발점입니다.

## 작은 instrumentation 스케치

```go
ctx, task := trace.NewTask(ctx, "rebuild-cache")
defer task.End()

trace.WithRegion(ctx, "load-metadata", func() {
    loadMetadata()
})
```

이 정도 annotation만 있어도 여러 goroutine과 단계가 섞인 trace를 읽기가 훨씬 쉬워집니다.

## block / mutex profile

서비스가 "동시성은 많은데 느리다"면 실제 문제는 다음일 수 있습니다.

- channel wait,
- hot mutex,
- shutdown queue 대기,
- 잘못된 backpressure 위치.

profile은 이런 걸 눈에 보이게 만듭니다.

## scheduler debug output

`GODEBUG=schedtrace=1000,scheddetail=1`은 trace만큼 세밀하지는 않지만 빠른 점검에는 유용합니다.

- runnable goroutine buildup,
- idle / busy processor,
- spinning worker,
- 전체 scheduler pressure.

## 실전 사용 순서

먼저:

```bash
go test -trace=trace.out ./...
go tool trace trace.out
```

그다음 contention 냄새가 나면 block / mutex profile로 넘어가면 됩니다.

## 실패 패턴

```go
// request latency metric만 보고 있고, trace도 block profile도 없는 상태
```

이 정도 관측으로는 scheduler state, lock contention, queue delay, GC interaction을 제대로 볼 수 없습니다. runtime 문제는 runtime-facing tool이 필요합니다.

## 공식 자료

- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)

## Practical takeaway

문서가 깊어질수록 직감만으로는 부족합니다.

trace와 profile은 "어디선가 goroutine이 막힌 것 같아"를 "정확히 어디서 왜 막혔는지 안다"로 바꿔주는 도구입니다.
