---
title: Garbage Collector and Green Tea GC
description: Understand Go's concurrent GC, pacer, assist, scavenger, and the Green Tea changes enabled by default in Go 1.26.
---

# Garbage Collector and Green Tea GC

Concurrency quality in Go is tightly connected to GC behavior.

Why?

Because goroutines, channels, maps, timers, and request graphs all allocate memory, retain pointers, and create scan work.

:::tip Quick takeaway
Go's GC is not a stop-the-world toy collector. It is a concurrent mark-sweep collector with pacing, assists, barriers, scavenging, and now Green Tea locality improvements for small-object marking.
:::

## The main moving parts

For day-to-day engineering, you should know these components:

| Component | Job |
| --- | --- |
| Mark phase | find reachable heap objects |
| Sweep phase | reclaim unreachable spans |
| Write barrier | preserve tri-color correctness during concurrent marking |
| Pacer | decide how aggressively GC should run |
| Assist | make mutators contribute GC work under pressure |
| Scavenger | return physical memory to the OS when possible |

## Simplified internal sketch

This is the model to carry around:

```go
func gcCycle() {
	stopTheWorldBriefly()
	enableWriteBarrier()
	startConcurrentMark()

	for workRemaining() {
		runBackgroundMarkWorkers()
		chargeAssistWorkToAllocators()
	}

	stopTheWorldBriefly()
	disableWriteBarrier()
	startSweep()
}
```

The real runtime is significantly more complex, but this sketch is enough to explain why allocation rate, pointer density, and long-lived heaps change application behavior.

## Why assists matter to concurrency

Many teams learn about GC from pause times only. That is too shallow.

When allocation pressure rises, mutator assists make ordinary goroutines perform part of the marking work.

This means GC cost can surface as:

- lower throughput in request handlers,
- higher tail latency,
- scheduler pressure that looks like "the service is slower" even without giant pauses.

## What Green Tea GC changes

Green Tea focuses on scan locality, especially for small objects.

The core idea is to batch marking and scanning around spans so that the runtime:

- touches adjacent objects together,
- amortizes metadata access,
- improves cache behavior during marking.

The file `runtime/mgcmark_greenteagc.go` describes this directly: marks and scans are tracked separately so spans can accumulate objects and then be scanned more locality-efficiently.

## Why this matters in real services

The practical wins show up most where Go services spend a lot of time:

- allocating small objects,
- building request-scoped graphs,
- maintaining pointer-heavy heaps,
- serving lots of concurrent requests with moderate object churn.

This is exactly the profile of many API, queue, and stream-processing services.

## Observability you should actually use

### `gctrace`

Use:

```bash
GODEBUG=gctrace=1 go test ./...
```

That gives fast feedback on cycle timing, heap growth, and collector work.

### Heap and alloc profiles

Use `pprof` when you need to see allocation hot spots and long-lived retention.

### Execution traces

`runtime/trace` lets you correlate goroutine behavior, blocking, and GC activity in the same timeline.

## Runtime source walk

Start here:

- [runtime/mgc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgc.go)
- [runtime/mgcmark.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark.go)
- [runtime/mgcpacer.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcpacer.go)
- [runtime/mgcscavenge.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcscavenge.go)
- [runtime/mgcmark_greenteagc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark_greenteagc.go)

## Historical checkpoints

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Green Tea GC](https://go.dev/blog/greenteagc)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)

Go 1.25 introduced Green Tea as an experiment. Go 1.26 enabled it by default. That is a good example of the Go team shipping runtime innovation gradually and then promoting it once confidence is high.

## Production consequences

### High allocation rate is a concurrency concern

If every request fans out into many goroutines and allocates heavily, GC work becomes part of your concurrency budget.

### Pointer shape matters

Pointer-rich long-lived structures cost more to scan than flat numeric buffers.

### Memory limits matter

GC is easier to operate when the service has explicit memory expectations rather than accidental heap growth.

## Practical takeaway

Go's concurrency model is strong partly because the GC is engineered for always-on server workloads.

If you want to reason like a runtime-aware Go engineer, you should connect goroutine behavior, allocation rate, heap shape, and GC pacing as one system rather than separate topics.
