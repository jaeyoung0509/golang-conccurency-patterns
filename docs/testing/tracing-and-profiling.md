---
title: Tracing and Contention Observability
description: Use runtime traces, block profiles, mutex profiles, and scheduler output to inspect real concurrent behavior.
---

# Tracing and Contention Observability

Some concurrency bugs are visible only when you look at runtime behavior directly.

For those cases, unit tests are not enough by themselves.

## The main tools

| Tool | Best for |
| --- | --- |
| `go test -trace` | goroutine states, scheduler activity, netpoll, syscalls, GC |
| block profile | where goroutines wait on blocking operations |
| mutex profile | lock contention hotspots |
| heap / alloc profile | allocation pressure that feeds GC and latency |
| `GODEBUG=schedtrace=...` | scheduler snapshots and run queue pressure |

## Execution traces

The `runtime/trace` package and `go test -trace=trace.out` let you inspect:

- goroutine creation and blocking,
- syscall entry and exit,
- GC activity,
- processor activity,
- user regions and tasks.

This is the best first tool when the question is "what is the runtime actually doing?"

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

## Scheduler debug output

`GODEBUG=schedtrace=1000,scheddetail=1` is crude compared with traces, but still useful for quick inspection:

- runnable goroutine buildup,
- idle vs busy processors,
- spinning workers,
- overall scheduler pressure.

## How to use this in practice

Start with:

```bash
go test -trace=trace.out ./...
go tool trace trace.out
```

Then escalate to block or mutex profiles when the problem smells like contention.

## Failure pattern

```go
// Only looking at request latency metrics, no trace, no block profile.
```

That is not enough once the issue is scheduler state, lock contention, queue delay, or GC interaction. Runtime behavior needs runtime-facing tools.

## Official reading

- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)

## Practical takeaway

The deeper your concurrency design becomes, the less you can rely on intuition alone.

Tracing and profiles are how you move from "I think a goroutine is stuck" to "I know exactly where and why it is stuck."
