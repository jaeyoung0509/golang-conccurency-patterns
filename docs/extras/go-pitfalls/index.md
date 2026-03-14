---
title: Go Pitfalls Appendix
description: A bilingual appendix of 50 Go pitfalls that compile cleanly and still create bugs.
---

# Go Pitfalls Appendix

This appendix is about code that:

- compiles,
- often passes casual review,
- and still bites hard in production.

The goal is not to teach syntax from scratch. The goal is to sharpen judgment around the kinds of mistakes experienced Go engineers still make under deadline pressure.

## What this appendix covers

| Category | What it focuses on |
| --- | --- |
| [Control Flow and Evaluation](/extras/go-pitfalls/control-flow-and-evaluation) | `select`, `break`, loop variables, `defer`, and evaluation rules that look obvious until they are not |
| [Interfaces and Types](/extras/go-pitfalls/interfaces-and-types) | typed nil, assertions, method sets, wrapped errors, and `any`-shaped ambiguity |
| [Collections and Memory](/extras/go-pitfalls/collections-and-memory) | map/slice pitfalls, iteration assumptions, backing-array retention, and memory-shape surprises |
| [Concurrency, Context, and Time](/extras/go-pitfalls/concurrency-context-and-time) | channel ownership, goroutine lifetime, cancellation, timers, and `WaitGroup` sequencing |
| [Stdlib and API Boundaries](/extras/go-pitfalls/stdlib-and-api-boundaries) | `net/http`, `database/sql`, `io`, `json`, `time`, `os/exec`, and `sync.Map` footguns |

## How to read each entry

Every pitfall follows the same shape:

1. a bad snippet,
2. why it breaks or misleads,
3. a better snippet,
4. how to catch it early,
5. one rule of thumb.

Read them like a code-review checklist, not like a language tutorial.

## Where to go deeper

This appendix is intentionally sharp and practical. When an item touches a bigger runtime or operational topic, the deeper explanation already exists elsewhere in the site:

- [Race Detector](/testing/race-detector)
- [Deterministic Tests with synctest](/testing/synctest)
- [context Package Internals](/stdlib/context-internals)
- [time, Timers, and Tickers](/stdlib/time-timers-tickers)
- [net/http Server and Transport](/stdlib/net-http-server-transport)
- [sync and atomic Primitives](/stdlib/sync-and-atomic)
- [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model)

## One important modern-Go note

Some classic blog posts and talks about Go pitfalls are now partially outdated.

In particular, loop-variable capture changed in Go 1.22 for the common `for ... := range ...` case. This appendix uses modern semantics and calls out where old advice still matters.

## Practical takeaway

Most painful Go bugs are not exotic.

They come from small misunderstandings at API boundaries, ownership boundaries, and evaluation boundaries. That is exactly what this appendix is meant to tighten.
