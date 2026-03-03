---
title: 현대 Go 성능 튜닝
description: PGO, execution trace, flight recorder, zero-copy 경로를 최신 Go 기준으로 설명합니다.
---

# 현대 Go 성능 튜닝

현대 Go 성능 튜닝은 예전처럼 folklore나 감에 기대는 일이 아닙니다.

프로파일을 컴파일러에 먹이고, trace로 런타임 진실을 보고, 표준 라이브러리가 이미 커널 fast path를 타는지 확인하는 일이 더 중요합니다.

:::tip Quick takeaway
아직도 microbenchmark와 감만으로 hot path를 튜닝하고 있다면, 현대 Go 툴체인을 절반밖에 안 쓰고 있는 것입니다. PGO, execution tracing, zero-copy-aware standard-library path는 이제 1급 도구입니다.
:::

## Profile-guided optimization

PGO는 실제 프로파일 데이터를 바탕으로 컴파일러가 일부 최적화 결정을 더 잘 하게 만듭니다.

핵심은 마법이 아니라 더 좋은 증거입니다.

- inlining budget
- devirtualization 기회
- hot-path layout

## 실전 PGO 루프

```bash
go test -cpuprofile=cpu.pprof ./...
go build -pgo=cpu.pprof ./cmd/service
```

현대 `go build`는 `-pgo=auto`도 지원합니다. main package 디렉터리에 `default.pgo`가 있으면 이를 자동으로 적용합니다.

## Execution trace는 runtime truth를 보는 도구다

`runtime/trace`는 다음을 잡습니다.

- goroutine 생성/블로킹
- scheduler 전환
- syscall과 netpoll 활동
- GC 이벤트
- heap goal 변화
- user task, region, log

그래서 "런타임이 실제로 무엇을 하고 있었는가?"를 묻는 상황에서는 가장 강력한 도구입니다.

## 작은 instrumentation 예시

```go
ctx, task := trace.NewTask(ctx, "replicate-batch")
defer task.End()

trace.WithRegion(ctx, "fetch-primary", func() {
	_ = fetchPrimary(ctx)
})

trace.Log(ctx, "tenant", tenantID)
```

user task와 region을 쓰면 애플리케이션 관점의 이야기 위에 런타임 관점의 이야기를 겹쳐서 볼 수 있습니다.

## Trace v2와 flight recording

최근 Go는 내부적으로 v2 execution-trace wire format을 쓰고, 더 최근 릴리스에서는 `trace.FlightRecorder`로 최근 런타임 이벤트의 rolling window를 메모리에 유지할 수 있게 됐습니다.

운영에서 중요한 이유는 분명합니다.

- trace가 지속 진단 도구에 더 가까워지고,
- 스파이크 직후 최근 몇 초를 snapshot 할 수 있고,
- one-shot trace를 미리 켜 두지 않아도 incident와 runtime 상태를 연결하기 쉬워집니다.

## Zero-copy I/O는 이미 표준 라이브러리에 있을 수 있다

올바른 file/socket 조합과 지원 플랫폼에서는 표준 라이브러리 경로가 다음에 도달할 수 있습니다.

- `sendfile`
- `copy_file_range`
- `splice`

즉, 때로는 `io.Copy`가 당신이 직접 쓰려던 hand-optimized code보다 이미 더 똑똑합니다.

## 단순화한 I/O 예시

```go
f, _ := os.Open("payload.bin")
defer f.Close()

_, _ = io.Copy(conn, f)
```

이게 실제로 zero-copy fast path를 타는지는 concrete descriptor 조합과 플랫폼에 달려 있습니다. 중요한 건, fast path는 소스 코드 모양이 아니라 라이브러리와 커널 조건에 의해 결정된다는 점입니다.

## 실패 패턴

다음은 성능 전략이 아닙니다.

```go
func build() {
	_ = exec.Command("go", "build").Run() // bad: 대표성 있는 profile도 없고 근거도 없음
}
```

마찬가지로 모든 `io.Copy`가 zero-copy일 거라고 믿는 것도, 아무 것도 zero-copy가 아닐 거라고 믿는 것도 둘 다 cargo cult입니다. 실제 경로를 확인하지 않는 튜닝은 튜닝이 아닙니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [Profile-guided optimization](https://go.dev/doc/pgo)
- [cmd/internal/pgo](https://github.com/golang/go/tree/go1.26.0/src/cmd/internal/pgo)
- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)
- [internal/trace/tracev2/doc.go](https://github.com/golang/go/blob/go1.26.0/src/internal/trace/tracev2/doc.go)
- [Flight Recorder in Go 1.25](https://go.dev/blog/flight-recorder)
- [os/zero_copy_linux.go](https://github.com/golang/go/blob/go1.26.0/src/os/zero_copy_linux.go)
- [internal/poll/sendfile_unix.go](https://github.com/golang/go/blob/go1.26.0/src/internal/poll/sendfile_unix.go)
- [internal/poll/splice_linux.go](https://github.com/golang/go/blob/go1.26.0/src/internal/poll/splice_linux.go)

## 실전 요약

현대 Go 스택은 분명한 루프를 요구합니다.

- 실제 프로파일을 수집하고,
- 그것을 컴파일러에 넣고,
- incident 주변 trace를 잡고,
- fast path를 추측하지 말고 확인해야 합니다.
