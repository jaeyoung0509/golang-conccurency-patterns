---
title: Modern Performance Tuning
description: PGO, execution tracing, flight recording, and zero-copy paths in modern Go releases.
---

# Modern Performance Tuning

Modern Go performance work is less about folklore and more about using the toolchain the way it was built to be used.

That means profiles feeding the compiler, traces feeding runtime diagnosis, and understanding when the standard library can already reach kernel fast paths without bespoke code.

:::tip Quick takeaway
If you are still tuning hot paths with only microbenchmarks and intuition, you are under-using the modern Go toolchain. PGO, execution tracing, and zero-copy-aware standard-library paths are now first-class tools.
:::

## Profile-guided optimization

PGO lets the compiler use real profile data when making some optimization decisions.

The important outcome is not magic. It is better evidence for:

- inlining budgets,
- devirtualization opportunities,
- hot-path layout decisions.

## Practical PGO loop

```bash
go test -cpuprofile=cpu.pprof ./...
go build -pgo=cpu.pprof ./cmd/service
```

For main packages, modern `go build` also supports `-pgo=auto`, which looks for `default.pgo` in the main package directory.

## Execution tracing is for runtime truth, not just latency charts

`runtime/trace` captures:

- goroutine creation and blocking,
- scheduler transitions,
- syscall and netpoll activity,
- GC events,
- heap-goal changes,
- user tasks, regions, and logs.

That makes it the best tool when you need to answer "what was the runtime actually doing?"

## Tiny instrumentation sketch

```go
ctx, task := trace.NewTask(ctx, "replicate-batch")
defer task.End()

trace.WithRegion(ctx, "fetch-primary", func() {
	_ = fetchPrimary(ctx)
})

trace.Log(ctx, "tenant", tenantID)
```

Use user tasks and regions to put your application story on top of the runtime story.

## Trace v2 and flight recording

Recent Go releases use a v2 execution-trace wire format internally, and newer releases also added `trace.FlightRecorder` for a rolling in-memory window of recent runtime events.

That matters operationally:

- traces are more viable as ongoing diagnostics,
- you can snapshot recent runtime behavior after a spike instead of reproducing everything under a one-shot trace,
- you can connect runtime state to incidents with lower friction.

## Zero-copy I/O is sometimes already in the standard library

For the right file/socket combinations on supported platforms, standard-library paths can reach:

- `sendfile`,
- `copy_file_range`,
- `splice`.

In other words, sometimes `io.Copy` is already smarter than the hand-optimized code you were about to write.

## Simplified I/O sketch

```go
f, _ := os.Open("payload.bin")
defer f.Close()

_, _ = io.Copy(conn, f)
```

Whether this hits a zero-copy path depends on the concrete descriptors and platform support. The important part is that the fast path is library- and kernel-driven, not source-code-shaped by wishful thinking.

## Failure pattern

This is not a performance strategy:

```go
func build() {
	_ = exec.Command("go", "build").Run() // bad: no representative profile, no evidence
}
```

Likewise, assuming every `io.Copy` is zero-copy is just as wrong as assuming none of them are. Performance tuning without checking the actual path is cargo culting in both directions.

## Source walk

Start here:

- [Profile-guided optimization](https://go.dev/doc/pgo)
- [cmd/internal/pgo](https://github.com/golang/go/tree/go1.26.0/src/cmd/internal/pgo)
- [runtime/trace package docs](https://pkg.go.dev/runtime/trace)
- [internal/trace/tracev2/doc.go](https://github.com/golang/go/blob/go1.26.0/src/internal/trace/tracev2/doc.go)
- [Flight Recorder in Go 1.25](https://go.dev/blog/flight-recorder)
- [os/zero_copy_linux.go](https://github.com/golang/go/blob/go1.26.0/src/os/zero_copy_linux.go)
- [internal/poll/sendfile_unix.go](https://github.com/golang/go/blob/go1.26.0/src/internal/poll/sendfile_unix.go)
- [internal/poll/splice_linux.go](https://github.com/golang/go/blob/go1.26.0/src/internal/poll/splice_linux.go)

## Practical takeaway

The modern Go stack wants you to close the loop:

- collect real profiles,
- feed the compiler,
- capture traces around incidents,
- verify fast paths instead of assuming them.
