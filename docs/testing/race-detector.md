---
title: Race Detector
description: Use Go's race detector to catch unsynchronized shared-memory access and understand what it can and cannot prove.
---

# Race Detector

The race detector is the first tool you should run against nontrivial concurrency code.

## What it catches

It detects data races: concurrent accesses to the same memory location where at least one access is a write and there is no synchronization edge between them.

Typical examples:

- reading a map while another goroutine writes it,
- mutating shared state outside the channel protocol that is supposed to protect it,
- forgetting to lock around a cache update.

## Tiny race example

```go
func TestRacyMap(t *testing.T) {
    cache := map[string]int{}

    go func() { cache["a"] = 1 }()
    go func() { _ = cache["a"] }()
}
```

That code may appear to "work" in a casual run. Under `-race`, it is exactly the kind of bug you want flagged immediately.

## What it does not prove

A clean `-race` run does **not** prove:

- absence of deadlocks,
- absence of goroutine leaks,
- fairness,
- good shutdown behavior,
- good timeout behavior.

It also only catches executions that actually happen during the test run.

## How to use it

```bash
go test -race ./...
```

The detector is slower and more expensive than ordinary test runs. That is normal and worth the cost for concurrency-heavy code.

## The right workflow

Use the race detector to answer:

- did I accidentally share mutable memory,
- did I create a path that bypasses synchronization,
- did a refactor break a previously safe ownership boundary?

Do **not** use it as your only concurrency validation step.

## Failure pattern

```go
func TestSomething(t *testing.T) {
    go mutateSharedState()
    time.Sleep(10 * time.Millisecond)
}
```

Tests like this often pass just long enough to create false confidence. Run them with `-race`, then replace timing luck with real synchronization.

## Good design patterns under `-race`

- single owner goroutine per mutable state machine,
- `map + mutex` with short critical sections,
- explicit reply channels for one-shot responses,
- cancellation paths that cannot leave background goroutines mutating shared state after the test returns.

## Official reading

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)

## Practical takeaway

`-race` is the minimum bar, not the finish line.

Run it early, run it often, and combine it with deterministic schedule tests and shutdown tests.
