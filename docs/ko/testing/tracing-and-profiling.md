---
title: 트레이싱과 경합 관측
description: runtime trace, block profile, mutex profile, scheduler debug output으로 실제 동시성 동작을 관측하는 방법을 설명합니다.
---

# 트레이싱과 경합 관측

어떤 동시성 버그는 runtime behavior를 직접 보기 전에는 잘 드러나지 않습니다.

그 경우 unit test만으로는 부족합니다.

:::tip Quick takeaway
질문이 "runtime이 실제로 뭘 하고 있나"면 execution trace부터 시작하는 게 맞습니다. 질문이 "시간이 어디서 새나"라면 block, mutex, heap, CPU profile로 내려가면 됩니다.
:::

## 주요 도구

| 도구 | 잘 맞는 문제 | 대표 커맨드 |
| --- | --- | --- |
| `go test -trace` | goroutine state, scheduler activity, netpoll, syscall, GC | `go test -trace=trace.out ./pkg` |
| block profile | 어디서 goroutine이 block되는지 | `go test -blockprofile=block.out ./pkg` |
| mutex profile | lock contention hotspot | `go test -mutexprofile=mutex.out ./pkg` |
| heap / alloc profile | allocation pressure와 GC 연결 | `go test -memprofile=mem.out ./pkg` |
| `GODEBUG=schedtrace=...` | scheduler snapshot과 run queue pressure | `GODEBUG=schedtrace=1000,scheddetail=1 go test ./pkg` |

## 어떤 도구부터 볼까

| 증상 | 먼저 볼 것 | 이유 |
| --- | --- | --- |
| "요청이 멈추는데 어디가 문제인지 모르겠다" | trace | runnable, blocked, syscall, GC 흐름을 한 번에 봄 |
| "lock이 뜨거운 것 같다" | mutex profile | contended lock 지점을 바로 보여줌 |
| "worker가 영원히 기다리는 것 같다" | block profile | channel, select, cond 같은 wait 지점을 찍어줌 |
| "CPU는 괜찮은데 지연이 튄다" | trace + heap profile | GC나 queueing 문제일 수 있음 |
| "scheduler가 이상해 보인다" | `schedtrace` 또는 trace | run queue buildup과 P 사용 상태를 봄 |

## execution trace

`runtime/trace`와 `go test -trace=trace.out`은 다음을 보여줍니다.

- goroutine 생성과 block,
- syscall enter/exit,
- GC activity,
- processor activity,
- user region과 task.

즉, "runtime이 실제로 뭘 하고 있는가"를 보기에 가장 좋은 출발점입니다.

기본 workflow는 이렇습니다.

```bash
go test -trace=trace.out ./...
go tool trace trace.out
```

전체 모듈이 너무 크면 좁혀서 봅니다.

```bash
go test -trace=trace.out ./examples/workerpool -run TestGenerateQuotesCancelsSlowJobsAfterError
go tool trace trace.out
```

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

자주 쓰는 커맨드는 이렇습니다.

```bash
go test -blockprofile=block.out ./...
go tool pprof -http=:0 block.out

go test -mutexprofile=mutex.out ./...
go tool pprof -http=:0 mutex.out

go test -memprofile=mem.out ./...
go tool pprof -http=:0 mem.out
```

장기 실행 프로세스에서 직접 profile을 수집할 땐 runtime knob도 알아야 합니다.

```go
runtime.SetBlockProfileRate(1)
runtime.SetMutexProfileFraction(5)
```

다만 이 값들은 오버헤드가 있습니다. 평상시 상시 최고 강도로 두기보다, 짧은 진단 구간이나 디버그 모드에서 켜는 편이 낫습니다.

## scheduler debug output

`GODEBUG=schedtrace=1000,scheddetail=1`은 trace만큼 세밀하지는 않지만 빠른 점검에는 유용합니다.

- runnable goroutine buildup,
- idle / busy processor,
- spinning worker,
- 전체 scheduler pressure.

완전한 trace를 뜨기 전에 빠른 텍스트 스냅샷이 필요할 때 특히 쓸모가 있습니다.

## 실전 사용 순서

이 순서를 고정해두면 좋습니다.

1. 믿을 수 있는 가장 작은 재현 경로를 만듭니다.
2. 문제가 GC인지 queueing인지 syscall인지 아직 모르면 trace부터 뜹니다.
3. trace가 waiting/lock 쪽을 가리키면 block / mutex profile로 내려갑니다.
4. trace는 맞는데 읽기 어렵다면 task / region annotation을 추가하고 다시 봅니다.

## 실패 패턴

```go
// request latency metric만 보고 있고, trace도 block profile도 없는 상태
```

이 정도 관측으로는 scheduler state, lock contention, queue delay, GC interaction을 제대로 볼 수 없습니다. runtime 문제는 runtime-facing tool이 필요합니다.

또 다른 실패 패턴은 질문 없이 profile만 모으는 것입니다. lock 문제를 찾는지, queue delay를 찾는지, allocation churn을 찾는지 명확하지 않으면 결과가 그냥 노이즈가 됩니다.

## 공식 자료

- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)
- [net/http/pprof package docs](https://pkg.go.dev/net/http/pprof)

## Practical takeaway

문서가 깊어질수록 직감만으로는 부족합니다.

trace와 profile은 "어디선가 goroutine이 막힌 것 같아"를 "정확히 어디서 왜 막혔는지 안다"로 바꿔주는 도구입니다.
