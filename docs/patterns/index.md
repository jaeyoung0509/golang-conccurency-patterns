---
title: Patterns Overview
description: Choose the right Go concurrency pattern based on workload shape, failure policy, and ownership model.
---

# Patterns Overview

<div class="lead-panel">
  <p>
    The practical pattern section is organized by <strong>problem shape</strong>, not by abstract taxonomy.
    Start from what your system needs to control: parallelism, stages, aggregation, or request lifetime.
  </p>
</div>

## Choose a pattern by workload shape

| If your problem looks like this | Start here |
| --- | --- |
| Many independent tasks, but concurrency must be capped | [Worker Pool](/patterns/worker-pool) |
| Data moves through ordered stages | [Pipeline](/patterns/pipeline) |
| One request needs answers from many backends | [Fan-Out / Fan-In](/patterns/fan-out-fan-in) |
| Several goroutines belong to one request lifetime | [Context Cancellation](/patterns/context-cancellation) |

## What makes these examples useful

<div class="signal-strip">
  <div class="signal">
    <strong>Real domains</strong>
    <span>Shipping, fraud checks, inventory, and request aggregation instead of toy snippets.</span>
  </div>
  <div class="signal">
    <strong>Tested behavior</strong>
    <span>The tests verify limits, order guarantees, cancellation, and partial failure policy.</span>
  </div>
  <div class="signal">
    <strong>Clear tradeoffs</strong>
    <span>Each page explains when the pattern is the right fit and when it is the wrong abstraction.</span>
  </div>
</div>

## Recommended flow

1. Read the pattern page.
2. Open the matching package in `examples/`.
3. Read the tests before copying the implementation style.
4. Decide the failure policy and ordering contract you need in your own service.
