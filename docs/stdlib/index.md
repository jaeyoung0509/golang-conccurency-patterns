---
title: Standard Library Overview
description: Learn the standard library packages that shape real Go concurrency, lifetimes, I/O, and time.
---

# Standard Library Overview

Most production Go services spend more time in the standard library than in custom concurrency helpers.

This section focuses on the packages that quietly define how real systems behave under load:

- `context` for lifetime and cancellation,
- `net/http` for server and client I/O,
- `database/sql` for pooling and backpressure,
- `time` for deadlines, retries, and timer correctness.

## What this section covers

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/stdlib/context-internals">context</a></h3>
    <p>Follow cancellation trees, `cancelCtx`, `timerCtx`, `valueCtx`, `Cause`, and the mistakes that leak request work.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/net-http-server-transport">net/http</a></h3>
    <p>Understand the server connection model, request lifetimes, transport pooling, keep-alive reuse, and timeout boundaries.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/database-sql-pool">database/sql</a></h3>
    <p>See why `sql.DB` is a pool, how waiters and cleaners work, and where cancellation or pool sizing assumptions fail.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/time-timers-tickers">time</a></h3>
    <p>Build a correct model for monotonic time, timers, tickers, `Stop`/`Reset`, and retry loops that do not quietly rot.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/sync-and-atomic">sync and atomic</a></h3>
    <p>Choose between `RWMutex`, `WaitGroup`, `Once`, `sync.Map`, `sync.Cond`, and `atomic.Pointer` based on the invariant you need to protect.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/process-signals-and-observability">signals and observability</a></h3>
    <p>Connect OS signals, graceful process shutdown, runtime metrics, and pprof-based incident investigation.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/io-bufio-bytes">io, bufio, and bytes</a></h3>
    <p>Understand streaming interfaces, buffering, `io.Copy` fast paths, scanner limits, and byte-slice aliasing.</p>
  </div>
  <div class="path-card">
    <h3><a href="/stdlib/encoding-json">encoding/json</a></h3>
    <p>Handle strictness, stream decoding, number handling, and compatibility traps at JSON service boundaries.</p>
  </div>
</div>

## Suggested reading order

1. Start with [context](/stdlib/context-internals), because lifetime ownership affects every other package in this section.
2. Continue with [time, timers, and tickers](/stdlib/time-timers-tickers), because deadlines and retries depend on correct clock and timer usage.
3. Read [sync and atomic primitives](/stdlib/sync-and-atomic) before deciding whether channels, mutexes, maps, or atomic snapshots are the right fit for a given state boundary.
4. Read [net/http server and transport internals](/stdlib/net-http-server-transport), where context and timers meet real network I/O.
5. Continue with [database/sql pool internals](/stdlib/database-sql-pool), where cancellation, waiting, and resource limits become operating concerns.
6. Read [io, bufio, and bytes](/stdlib/io-bufio-bytes) and [encoding/json in production](/stdlib/encoding-json) together when you are working at streaming API or log boundaries.
7. Finish with [process signals and runtime observability](/stdlib/process-signals-and-observability) so package-level design connects to service-level operations.

## What you should be able to answer afterward

- Why does forgetting `cancel()` leak more than a timer?
- Why can `http.Client` reuse collapse if you mishandle response bodies?
- Why is `sql.DB` a long-lived shared handle instead of a per-request object?
- Why does `time.Time` carry both wall-clock and monotonic readings?
- When is a plain mutex cleaner than `sync.Map` or atomics?
- Why is `signal.NotifyContext` usually a better shutdown entry point than raw signal channels?
- Why can `io.Copy` outperform your hand-written loop?
- Why does `encoding/json` silently accept behaviors that a strict schema system would reject?
- Why did timer channel semantics change materially in Go 1.23?

## Practical takeaway

If fundamentals explain why Go concurrency is possible, this section explains how most production Go code actually expresses that concurrency day to day.
