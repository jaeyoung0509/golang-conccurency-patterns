---
title: Internals Overview
description: Systems-level Go internals for compiler behavior, allocator design, type metadata, unsafe boundaries, and performance tooling.
---

# Internals Overview

<div class="lead-panel">
  <p>
    Internals is where the site stops asking only "how do I use concurrency?" and starts asking
    <strong>why the toolchain, runtime, allocator, and type system make certain designs fast, safe, or dangerous</strong>.
  </p>
</div>

## What this section covers

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/internals/compiler-and-toolchain">Compiler and Toolchain</a></h3>
    <p>Escape analysis, SSA, and Plan 9 assembly so you can connect source code to what the compiler actually emits.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/allocator-and-write-barrier">Allocator and Write Barrier</a></h3>
    <p>Follow `mcache`, `mcentral`, `mheap`, `mspan`, and the hybrid write barrier behind concurrent GC correctness.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/layout-padding-false-sharing">Layout, Padding, and False Sharing</a></h3>
    <p>See how alignment rules, struct layout, and cache lines affect contention even when your algorithm looks correct.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/generics-and-interfaces">Generics and Interfaces</a></h3>
    <p>Understand GC shape stenciling, dictionaries, `eface`/`iface`-style layouts, and dynamic dispatch tradeoffs.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/unsafe-cgo-pinner">unsafe, cgo, and Pinner</a></h3>
    <p>Learn the real boundary rules around raw pointers, scheduler handoff, pointer pinning, and zero-copy tricks.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/modern-performance-tuning">Modern Performance Tuning</a></h3>
    <p>Profile-guided optimization, execution tracing, flight recording, and zero-copy I/O in modern Go releases.</p>
  </div>
  <div class="path-card">
    <h3><a href="/internals/stdlib-anatomy">Standard Library Anatomy</a></h3>
    <p>Study why `sync.Pool` scales, why `reflect` costs what it costs, and when code generation is the better trade.</p>
  </div>
</div>

## Suggested reading order

1. Read [Compiler and Toolchain](/internals/compiler-and-toolchain) first.
2. Continue with [Allocator and Write Barrier](/internals/allocator-and-write-barrier).
3. Read [Layout, Padding, and False Sharing](/internals/layout-padding-false-sharing).
4. Move to [Generics and Interfaces](/internals/generics-and-interfaces).
5. Then read [unsafe, cgo, and Pinner](/internals/unsafe-cgo-pinner).
6. Finish with [Modern Performance Tuning](/internals/modern-performance-tuning) and [Standard Library Anatomy](/internals/stdlib-anatomy).

## What you should be able to answer afterward

- Why does a local variable sometimes still move to the heap?
- What does `GOSSAFUNC` show that `go test -bench` cannot?
- Why is the allocator hierarchy `mcache -> mcentral -> mheap` instead of one global lock?
- Why does the GC need a hybrid write barrier in a concurrent collector?
- Why can two correct atomic counters still fight each other on one cache line?
- Why are Go generics neither pure C++-style monomorphization nor Java-style erasure?
- Why is `uintptr` not a GC root?
- When does `io.Copy` reach a zero-copy fast path, and when does it silently fall back?
- Why does `sync.Pool` pad its per-P shards?
- When is reflection flexible enough, and when is it simply the wrong cost model?
