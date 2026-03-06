---
title: Tracing and Contention Observability
description: Use runtime traces, block profiles, mutex profiles, and scheduler output to inspect real concurrent behavior.
---

# Tracing and Contention Observability

Some concurrency bugs are visible only when you look at runtime behavior directly.

For those cases, unit tests are not enough by themselves.

:::tip Quick takeaway
If the question is "what is the runtime actually doing," start with an execution trace. If the question is "where is time being lost," move next to block, mutex, heap, or CPU profiles.
:::

## The main tools

| Tool | Best for | Typical command |
| --- | --- | --- |
| `go test -trace` | goroutine states, scheduler activity, netpoll, syscalls, GC | `go test -trace=trace.out ./pkg` |
| block profile | where goroutines wait on blocking operations | `go test -blockprofile=block.out ./pkg` |
| mutex profile | lock contention hotspots | `go test -mutexprofile=mutex.out ./pkg` |
| heap / alloc profile | allocation pressure that feeds GC and latency | `go test -memprofile=mem.out ./pkg` |
| `GODEBUG=schedtrace=...` | scheduler snapshots and run queue pressure | `GODEBUG=schedtrace=1000,scheddetail=1 go test ./pkg` |

## Which tool first?

| Symptom | Start with | Why |
| --- | --- | --- |
| "Requests stall, but I do not know where" | trace | shows runnable, blocked, syscall, GC, and wakeup flow together |
| "It feels lock-heavy" | mutex profile | shows where contended locks are held |
| "Workers are waiting forever" | block profile | points at channel, select, cond, and other blocking sites |
| "CPU is fine, but latency spikes during load" | trace plus heap profile | often GC or queueing, not raw CPU |
| "The scheduler looks unhealthy" | `schedtrace` or trace | shows run-queue growth and idle/busy processor state |

## Execution traces

The `runtime/trace` package and `go test -trace=trace.out` let you inspect:

- goroutine creation and blocking,
- syscall entry and exit,
- GC activity,
- processor activity,
- user regions and tasks.

This is the best first tool when the question is "what is the runtime actually doing?"

The basic workflow is:

```bash
go test -trace=trace.out ./...
go tool trace trace.out
```

When the full module is too noisy, narrow it:

```bash
go test -trace=trace.out ./examples/workerpool -run TestGenerateQuotesCancelsSlowJobsAfterError
go tool trace trace.out
```

## Tiny instrumentation sketch

```go
ctx, task := trace.NewTask(ctx, "rebuild-cache")
defer task.End()

trace.WithRegion(ctx, "load-metadata", func() {
    loadMetadata()
})
```

That kind of annotation makes traces dramatically easier to read once a workflow spans several goroutines or phases.

## Block and mutex profiles

If a service feels "concurrent but slow," the real issue may be:

- too much time waiting on channels,
- a hot mutex,
- shutdown blocking on a queue,
- backpressure in the wrong place.

Profiles make these visible.

Useful commands:

```bash
go test -blockprofile=block.out ./...
go tool pprof -http=:0 block.out

go test -mutexprofile=mutex.out ./...
go tool pprof -http=:0 mutex.out

go test -memprofile=mem.out ./...
go tool pprof -http=:0 mem.out
```

If you are collecting profiles in a long-running process instead of a test binary, the relevant runtime knobs are:

```go
runtime.SetBlockProfileRate(1)
runtime.SetMutexProfileFraction(5)
```

Use those deliberately. Full block profiling and aggressive mutex sampling have real overhead, so they are usually better as a targeted debug mode than a permanent production default.

## Scheduler debug output

`GODEBUG=schedtrace=1000,scheddetail=1` is crude compared with traces, but still useful for quick inspection:

- runnable goroutine buildup,
- idle vs busy processors,
- spinning workers,
- overall scheduler pressure.

It is especially useful when you want a fast text snapshot before taking a full trace, or when a workload is hard to reproduce under a test harness.

## How to use this in practice

Use a fixed escalation order:

1. Reproduce the problem with the narrowest test, benchmark, or workload you trust.
2. Take an execution trace first when you do not yet know whether the issue is GC, queueing, scheduler pressure, or syscalls.
3. Move to block or mutex profiles when the trace points toward waiting or contention.
4. Add task and region annotations if the trace is technically correct but hard to interpret.

## Failure pattern

```go
// Only looking at request latency metrics, no trace, no block profile.
```

That is not enough once the issue is scheduler state, lock contention, queue delay, or GC interaction. Runtime behavior needs runtime-facing tools.

Another common failure is collecting a profile without a question. If you do not know whether you are looking for lock contention, queue delay, or allocation churn, the output becomes noise.

## Official reading

- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)
- [net/http/pprof package docs](https://pkg.go.dev/net/http/pprof)

## Practical takeaway

The deeper your concurrency design becomes, the less you can rely on intuition alone.

Tracing and profiles are how you move from "I think a goroutine is stuck" to "I know exactly where and why it is stuck."
