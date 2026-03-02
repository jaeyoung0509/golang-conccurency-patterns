---
title: Advanced Overview
description: Explore higher-leverage Go concurrency patterns for structured lifetimes, resource budgeting, duplicate suppression, ownership, and overload control.
---

# Advanced Overview

<div class="lead-panel">
  <p>
    Advanced topics are where Go concurrency stops being about “can I run this in parallel?”
    and starts being about <strong>lifetime control, resource budgeting, ownership, and overload survival</strong>.
  </p>
</div>

## Study map

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/advanced/structured-concurrency">Structured Concurrency</a></h3>
    <p>Use `errgroup` and context to keep goroutine lifetimes inside one parent task boundary.</p>
  </div>
  <div class="path-card">
    <h3><a href="/advanced/weighted-semaphore">Weighted Semaphore</a></h3>
    <p>Limit by memory or quota weight instead of naive goroutine count.</p>
  </div>
  <div class="path-card">
    <h3><a href="/advanced/singleflight">Singleflight</a></h3>
    <p>Collapse duplicate in-flight work to protect backends from cache-miss stampedes.</p>
  </div>
  <div class="path-card">
    <h3><a href="/advanced/actor-pattern">Actor Pattern</a></h3>
    <p>Serialize mutation around explicit ownership when a state machine deserves one clear owner.</p>
  </div>
  <div class="path-card">
    <h3><a href="/advanced/csp-theory">CSP Theory</a></h3>
    <p>Understand the communication-first ideas Go borrowed and where Go intentionally diverges.</p>
  </div>
  <div class="path-card">
    <h3><a href="/advanced/backpressure-load-shedding">Backpressure and Load Shedding</a></h3>
    <p>Keep systems stable by bounding queues, rejecting excess work, and degrading deliberately.</p>
  </div>
</div>

## When to come here

- You already understand goroutines, channels, and cancellation basics.
- You are designing production behavior under load, not just correctness in the happy path.
- You need to decide between ownership-first and communication-first designs.
