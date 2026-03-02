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
    <h3><a href="/fundamentals/go-runtime-scheduler">Runtime and Scheduler</a></h3>
    <p>Understand G, M, P, run queues, netpoll, sysmon, stack growth, and why goroutines are cheap but not free.</p>
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
</div>

## Suggested reading order

1. Read [Go Runtime and Scheduler](/fundamentals/go-runtime-scheduler) first.
2. Move to [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model).
3. Go deeper with [Channel Internals](/fundamentals/channel-internals).
4. Finish with [Mutex and Runtime Semaphore Internals](/fundamentals/mutex-semaphore-internals).

## What you should be able to answer afterward

- Why does `GOMAXPROCS` affect CPU parallelism but not external-service safety?
- Why can channel send/receive establish visibility guarantees?
- Why is `select` helpful but not a correctness proof by itself?
- Why can a mutex be the cleanest option in some Go codebases?
