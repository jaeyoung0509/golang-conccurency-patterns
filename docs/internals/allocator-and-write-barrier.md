---
title: Allocator and Hybrid Write Barrier
description: Follow Go's allocator hierarchy and the hybrid write barrier that keeps concurrent GC correct.
---

# Allocator and Hybrid Write Barrier

You cannot reason well about production Go performance if "allocation" is still a single mental bucket.

Small-object allocation, span reuse, sweeping, large-object paths, and write barriers all interact with concurrency and GC latency.

:::tip Quick takeaway
Go's allocator is still inspired by TCMalloc, but the important production model is simpler: `mcache` is the fast per-P path, `mcentral` amortizes sharing by size class, `mheap` manages pages, and the hybrid write barrier keeps concurrent marking correct.
:::

## The allocator hierarchy

The runtime comment in `malloc.go` gives the most useful summary:

- `mcache`: per-P cache of spans with free slots,
- `mcentral`: shared spans for one size class,
- `mheap`: page-level heap manager,
- `mspan`: a run of pages serving objects of one class.

```mermaid
flowchart LR
    A["goroutine allocates"] --> B["P-local mcache"]
    B --> C["size-class mspan"]
    C --> D["shared mcentral"]
    D --> E["global mheap"]
    E --> F["OS pages"]
```

The design goal is obvious: keep the common path lock-free or close to it, and pay shared coordination only when local caches run dry.

## Small allocations and size classes

For small objects, the allocator rounds the request to a size class and tries to serve it from the current P's cache.

That makes tiny short-lived allocations surprisingly cheap in isolation.

It does **not** make a high-allocation-rate workload free once GC, pointer scanning, and cross-goroutine fan-out are involved.

## Simplified allocation sketch

```go
func alloc(size uintptr) unsafe.Pointer {
	if span := mcache.lookup(sizeClass(size)); span.hasFree() {
		return span.alloc()
	}
	if span := mcentral.refill(sizeClass(size)); span != nil {
		mcache.install(span)
		return span.alloc()
	}
	span := mheap.allocPages(...)
	mcentral.install(span)
	mcache.install(span)
	return span.alloc()
}
```

This is not runtime code, but it is the right mental model.

## Large objects are a different path

Large allocations bypass much of the small-object machinery and hit the page allocator more directly.

That means patterns like "each request allocates a fresh 64 KiB buffer in ten goroutines" are qualitatively different from churning many tiny structs.

## Why the write barrier belongs here

The allocator story and GC story are tied together.

During concurrent marking, the runtime must prevent mutators from hiding reachable objects from the collector while pointers are being updated.

That is what the write barrier is for.

## The hybrid write barrier

The runtime comment in `mbarrier.go` describes it directly:

```go
// conceptual, not the actual runtime function
func writePointer(slot *unsafe.Pointer, ptr unsafe.Pointer) {
	shade(*slot)
	if currentStackIsGrey() {
		shade(ptr)
	}
	*slot = ptr
}
```

The combination matters:

- the deletion side protects the old referent,
- the insertion side protects the new referent while a stack is still grey,
- once the stack is black, the insertion side is no longer needed for that stack.

## Why memory ordering matters

Go does not try to make the barrier conditional on the destination object's color in the fast path, because the memory-ordering cost would be too high.

This is one of those runtime choices that looks slightly wasteful in a tiny benchmark and obviously correct in a real concurrent collector.

## Practical consequence for application code

You do not write barriers manually in ordinary Go.

But you absolutely feel their consequences when you:

- copy pointer-rich values at high rates,
- mutate large shared heaps of graph-shaped data,
- use reflection or `unsafe` helpers that ultimately funnel through typed memory moves,
- create workloads where GC assist shows up in request latency.

## Failure pattern

This is the kind of code that looks merely "allocation heavy" and becomes a GC and allocator problem under load:

```go
for _, req := range batch {
	go func(req Request) {
		buf := make([]byte, 64<<10) // large allocation path
		_ = handle(req, buf)
	}(req)
}
```

The problem is not just bytes allocated. The problem is:

- large objects skip the cheapest path,
- fan-out multiplies the pressure,
- the collector and allocator become part of tail latency.

## Source walk

Start here:

- [runtime/malloc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/malloc.go)
- [runtime/mcache.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mcache.go)
- [runtime/mcentral.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mcentral.go)
- [runtime/mheap.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mheap.go)
- [runtime/mbarrier.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mbarrier.go)
- [Proposal: eliminate rescan with a hybrid write barrier](https://github.com/golang/proposal/blob/master/design/17503-eliminate-rescan.md)

## Practical takeaway

The cheapest Go allocation is not "free." It is "carefully engineered."

Once you understand the allocator hierarchy and the hybrid barrier, you stop treating heap churn as a vague smell and start treating it as a concrete runtime budget.
