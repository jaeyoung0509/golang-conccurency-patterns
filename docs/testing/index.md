---
title: Testing Overview
description: Learn how to test Go concurrency for races, leaks, time, cancellation, and scheduler behavior.
---

# Testing Overview

Concurrency code without a testing strategy is mostly wishful thinking.

The hard bugs are rarely syntax errors. They are:

- missing happens-before edges,
- shutdown leaks,
- timing assumptions,
- cancellation races,
- contention that only appears under load.

## What this section covers

| If you need to test... | Start here |
| --- | --- |
| AI-generated Go code that compiles but may still be unsafe | [AI-Assisted Go Safety](/testing/ai-assisted-go-safety) |
| unsynchronized shared memory access | [Race Detector](/testing/race-detector) |
| timeout and timer logic without sleeping in real time | [Deterministic Tests with synctest](/testing/synctest) |
| real Postgres, Redis, and container-backed dependency behavior with clean teardown | [Integration Testing with Testcontainers](/testing/integration-testcontainers) |
| goroutine leaks and shutdown behavior | [Leak, Shutdown, and Timeout Testing](/testing/leaks-and-shutdowns) |
| scheduler, blocking, GC, and contention behavior | [Tracing and Contention Observability](/testing/tracing-and-profiling) |

## Testing philosophy

Good concurrency tests try to prove:

1. the protocol is safe,
2. the failure path terminates,
3. the timeout path is deterministic,
4. the implementation does not leak goroutines or work.
5. the verification stack would catch likely regressions early, even for AI-generated changes.

That is a higher bar than "it passed once on my laptop."

## Recommended reading order

1. [Race Detector](/testing/race-detector)
2. [AI-Assisted Go Safety](/testing/ai-assisted-go-safety)
3. [Deterministic Tests with synctest](/testing/synctest)
4. [Integration Testing with Testcontainers](/testing/integration-testcontainers)
5. [Leak, Shutdown, and Timeout Testing](/testing/leaks-and-shutdowns)
6. [Tracing and Contention Observability](/testing/tracing-and-profiling)

## Practical takeaway

Testing concurrency is not one tool. It is a stack:

- compile and unit tests as the first gate,
- `go vet` and static analysis for suspicious constructs,
- `-race` for memory safety,
- `synctest` for deterministic time and bubble-local goroutines,
- `testcontainers-go` and typed fixture setup for real dependency contracts,
- targeted leak and shutdown tests for lifecycle,
- traces and profiles for runtime behavior.
