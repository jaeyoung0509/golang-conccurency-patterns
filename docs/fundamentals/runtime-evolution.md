---
title: Runtime Evolution
description: Follow how Go's concurrency runtime evolved through preemption, scheduler changes, map redesign, and Green Tea GC.
---

# Runtime Evolution

The Go concurrency story did not arrive fully formed in 2009.

The language promised cheap goroutines and channels early, but the runtime has kept changing to make those promises hold under real production workloads.

:::tip Quick takeaway
The important history is not "Go used to be bad and now it is good." The important history is that the runtime kept removing latency cliffs: scheduler monopolization, map grow costs, GC cache-miss costs, and flaky time-based testing.
:::

## The historical arc

The clearest way to understand Go's evolution is to track the problems the runtime had to solve:

| Problem | Earlier limitation | Key change |
| --- | --- | --- |
| CPU-bound goroutines could delay fairness and GC stop points | Preemption mostly happened at safe points such as function calls | Go 1.14 introduced asynchronous preemption |
| Large maps paid expensive whole-map grow costs | The old bucket/overflow design had different locality and grow tradeoffs | Go 1.24 moved maps to a Swiss Table design with extendible hashing |
| Marking small objects left CPU and cache locality on the table | Concurrent GC was already strong, but locality still mattered | Go 1.25 introduced experimental Green Tea GC, enabled by default in Go 1.26 |
| Concurrency tests often depended on sleeps and timing luck | Deterministic schedule control was weak | `testing/synctest` matured into a standard tool around Go 1.24/1.25 |

## A practical timeline

| Release | What changed | Why it mattered |
| --- | --- | --- |
| February 25, 2020, Go 1.14 | Asynchronous preemption | CPU-heavy loops became less able to monopolize the process |
| February 11, 2025, Go 1.24 | Swiss Table maps and `testing/synctest` experiments | Better map structure and stronger concurrency testing tools |
| August 12, 2025, Go 1.25 | Green Tea GC experiment | Better GC scan locality for small objects |
| February 10, 2026, Go 1.26 | Green Tea GC enabled by default | The improved marker path became the normal runtime behavior |

## The preemption story, correctly stated

Go did **not** move from "preemptive" to "non-preemptive."

The more accurate story is:

- early Go relied much more on cooperative or safe-point-driven scheduler behavior,
- long-running loops could delay scheduler fairness and GC progress,
- Go 1.14 added asynchronous preemption so the runtime could interrupt CPU-bound goroutines more reliably.

That change mattered because a production program does not control every call site. One tight loop without blocking communication should not be able to degrade the whole service.

## Simplified internal sketch

This is not runtime code, but it is close to the mental model you should carry:

```go
func schedule(p *P) {
	for {
		if g := runqget(p); g != nil {
			execute(g)
			continue
		}

		if g := globrunqget(); g != nil {
			execute(g)
			continue
		}

		if ready := netpoll(0); !ready.empty() {
			injectglist(ready)
			continue
		}

		if g := stealFromOtherP(p); g != nil {
			execute(g)
			continue
		}

		parkm()
	}
}
```

The runtime has far more detail than this, but the high-level job stays the same: keep work moving without turning thread management into a bottleneck.

## Why map and GC evolution belong in a concurrency guide

Concurrency performance is not only about scheduling.

It also depends on:

- how fast shared structures are accessed,
- how much memory traffic each operation generates,
- how expensive marking and scanning become when thousands of goroutines are active.

That is why a serious concurrency guide should include:

- scheduler and preemption,
- netpoller and timers,
- maps and their memory layout,
- GC and heap scanning behavior.

## What changed in maps

Go 1.24 moved built-in maps toward a design based on Swiss Tables plus extendible hashing.

The concurrency lesson is not "maps became safe for concurrent writes." They did not.

The lesson is:

- lookups became more locality-aware,
- growth became more incremental at the table level,
- the runtime improved the cost profile of a heavily used core data structure.

That matters for caches, routing tables, deduplication layers, and actor-owned state maps.

## What changed in GC

Go's GC was already concurrent, tri-color, and production-viable long before Green Tea.

Green Tea is about improving the marker's locality for small objects by batching scan work around spans more effectively.

That means:

- fewer cold metadata walks,
- better cache behavior during marking,
- more efficient throughput in allocation-heavy services.

## Runtime source walk

If you want to verify the story in code, start here:

- [runtime/proc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/proc.go)
- [runtime/netpoll.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/netpoll.go)
- [runtime/map.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/map.go)
- [internal/runtime/maps/map.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/map.go)
- [runtime/mgc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgc.go)
- [runtime/mgcmark_greenteagc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark_greenteagc.go)

## Release notes worth reading

- [Go 1.14 Release Notes](https://go.dev/doc/go1.14)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)
- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable)
- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Green Tea GC](https://go.dev/blog/greenteagc)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)

## Practical takeaway

If you want to master Go concurrency, do not freeze your mental model at "goroutines are cheap."

The real model is that Go keeps refining the runtime to preserve that convenience under modern workloads.

From here, continue with [Go Runtime and Scheduler](/fundamentals/go-runtime-scheduler), then [Netpoller, Timers, and Syscalls](/fundamentals/netpoller-timers-syscalls).
