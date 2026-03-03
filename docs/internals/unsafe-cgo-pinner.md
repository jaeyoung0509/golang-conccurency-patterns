---
title: unsafe, cgo, and runtime.Pinner
description: Raw-pointer rules, cgo boundaries, and pinning semantics for engineers working near the metal.
---

# unsafe, cgo, and runtime.Pinner

This is the boundary where Go stops protecting you by default.

The reward is power: zero-copy tricks, C interop, direct memory views.

The cost is that the garbage collector, scheduler, and pointer rules still exist even if your code looks like C.

:::tip Quick takeaway
`unsafe.Pointer` is still part of Go's pointer world. `uintptr` is just an integer. cgo crossings change scheduler behavior. `runtime.Pinner` exists so C can safely retain Go pointers when the referenced objects are explicitly pinned.
:::

## `unsafe.Pointer` and `uintptr` are not interchangeable

This is the expert rule to remember:

- `unsafe.Pointer` participates in pointer semantics,
- `uintptr` does not keep objects alive and does not tell the GC anything.

If you turn a Go pointer into an integer and expect the runtime to preserve object validity on that basis, you are outside the contract.

## Small but important safe helper APIs

Modern Go added helpers that are easier to reason about than older header-casting tricks:

- `unsafe.String`,
- `unsafe.StringData`,
- `unsafe.Slice`,
- `unsafe.SliceData`,
- `unsafe.Add`.

These are still unsafe. They are just more explicit and less fragile than hand-rolling `reflect.SliceHeader` manipulations.

## Simplified pinning sketch

```go
buf := make([]byte, 4096)

var p runtime.Pinner
p.Pin(&buf[0])
defer p.Unpin()

// Pass buf's address to C here, or let C retain it briefly.
runtime.KeepAlive(buf)
```

The important detail is not the syntax. It is the contract:

- the object is pinned until `Unpin`,
- if the object contains Go pointers that C will follow, those referents must be pinned separately,
- the `Pinner` itself must stay alive for the duration of C's use.

## Why cgo changes concurrency behavior

A cgo call is not just "another function call."

It affects:

- scheduler accounting,
- stack transitions and pointer checks,
- how transparently the runtime can observe blocking,
- how easy it is to reason about latency from Go-side traces alone.

Pure Go network I/O benefits from netpoll integration. Opaque C calls do not fit that model nearly as well.

## A practical cgo rule

Cross the boundary in chunks, not droplets.

One larger call with a clear ownership contract is usually easier to operate than a per-record or per-packet crossing that destroys batching and observability.

## Failure pattern

This is a classic bug:

```go
func firstByteAddr(buf []byte) uintptr {
	return uintptr(unsafe.Pointer(&buf[0]))
}
```

Once the pointer is reduced to `uintptr`, it is no longer a Go pointer in the eyes of the runtime. Treating that integer as if it preserved liveness or pinning is wrong.

## A safer zero-copy mental model

If you use unsafe helpers for zero-copy conversions:

- keep the original backing storage alive,
- keep ownership obvious,
- do not let mutable and immutable views outlive the same assumptions,
- do not mix "no-copy" with "no-lifetime-management."

## Where `runtime.Pinner` fits

`runtime.Pinner` exists for a narrow class of problems:

- C needs to retain a Go pointer after a cgo call returns,
- or C needs to follow Go pointers stored inside a Go object that was itself passed across the boundary.

It is not a general excuse to abandon ownership discipline.

## Source walk

Start here:

- [unsafe package docs](https://pkg.go.dev/unsafe)
- [cmd/cgo package docs](https://pkg.go.dev/cmd/cgo)
- [runtime/pinner.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/pinner.go)
- [runtime/mbarrier.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mbarrier.go)

## Practical takeaway

The right way to use `unsafe` is not to pretend the runtime disappeared.

The right way is to know exactly which runtime guarantees you are depending on, and which ones you just stepped outside.
