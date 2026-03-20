---
title: Fundamentals Overview
description: Start here if you want to understand why Go concurrency works before you focus on specific patterns.
---

# Fundamentals Overview

<div class="lead-panel">
  <p>
    Fundamentals is the part of the site that answers <strong>why these patterns work at all</strong>.
    If you skip this section, you can still copy code. You will have a harder time debugging it under load.
  </p>
</div>

## What this section covers

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/fundamentals/runtime-evolution">Runtime Evolution</a></h3>
    <p>See how async preemption, Swiss Table maps, and Green Tea GC changed the shape of modern Go concurrency.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/go-runtime-scheduler">Runtime and Scheduler</a></h3>
    <p>Understand G, M, P, run queues, netpoll, sysmon, stack growth, and why goroutines are cheap but not free.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/netpoller-timers-syscalls">Netpoller and Timers</a></h3>
    <p>Follow how network readiness, deadlines, and scheduler wakeups make direct-style I/O practical.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/tcp-dns-connection-lifecycles">TCP, DNS, and Connection Lifecycles</a></h3>
    <p>Connect listeners, DNS resolution, socket deadlines, keepalive, half-close, and request budgeting into one usable network mental model.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/channels-memory-model">Channels and Memory Model</a></h3>
    <p>See where synchronization guarantees come from and how channel communication creates correctness, not just coordination.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/channel-internals">Channel Internals</a></h3>
    <p>Follow `hchan`, `sudog`, direct handoff, queueing, `close`, and `select` so channel behavior stops feeling magical.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/mutex-semaphore-internals">Mutex and Runtime Semaphore</a></h3>
    <p>Understand fast paths, starvation mode, contention, and the wakeup machinery beneath `sync.Mutex`.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/map-internals">Map Internals</a></h3>
    <p>Update your mental model to Swiss Tables, extendible hashing, iteration complexity, and ownership implications.</p>
  </div>
  <div class="path-card">
    <h3><a href="/fundamentals/garbage-collector">Garbage Collector</a></h3>
    <p>Connect allocation rate, assist work, heap shape, and Green Tea GC to real service latency.</p>
  </div>
</div>

## Suggested reading order

1. Read [Runtime Evolution](/fundamentals/runtime-evolution) first to get the release-history context.
2. Continue with [Go Runtime and Scheduler](/fundamentals/go-runtime-scheduler).
3. Read [Netpoller, Timers, and Syscalls](/fundamentals/netpoller-timers-syscalls).
4. Continue with [TCP, DNS, and Connection Lifecycles in Go](/fundamentals/tcp-dns-connection-lifecycles) to connect the runtime model to actual socket ownership.
5. Move to [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model).
6. Go deeper with [Channel Internals](/fundamentals/channel-internals) and [Mutex and Runtime Semaphore Internals](/fundamentals/mutex-semaphore-internals).
7. Finish with [Map Internals and Swiss Tables](/fundamentals/map-internals) and [Garbage Collector and Green Tea GC](/fundamentals/garbage-collector).

## What you should be able to answer afterward

- Why does `GOMAXPROCS` affect CPU parallelism but not external-service safety?
- Why did Go 1.14 async preemption materially change fairness under CPU-heavy load?
- Why does the netpoller make goroutine-per-connection feasible?
- Why is DNS resolution part of connection lifetime instead of a separate afterthought?
- Why does `DialContext` stop mattering once the socket is already open?
- Why can channel send/receive establish visibility guarantees?
- Why is `select` helpful but not a correctness proof by itself?
- Why are modern Go maps faster but still unsafe for concurrent mutation?
- How do allocation rate and GC assist affect concurrency latency?
- Why can a mutex be the cleanest option in some Go codebases?
