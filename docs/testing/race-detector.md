---
title: Race Detector
description: Use Go's race detector to catch unsynchronized shared-memory access and understand what it can and cannot prove.
---

# Race Detector

The race detector is the first tool you should run against nontrivial concurrency code.

:::tip Quick takeaway
`go test -race` is the minimum bar for concurrent Go, not the finish line. It is excellent at catching missing synchronization around shared memory, and useless for proving shutdown, leak, or timeout correctness by itself.
:::

## Mental model

The detector is asking one narrow question:

> Did two goroutines touch the same memory concurrently without a valid happens-before edge between them?

That is why it pairs naturally with the [Go Memory Model](https://go.dev/ref/mem). If your design relies on ownership transfer through channels, mutexes, atomics, or `WaitGroup` completion, the detector helps confirm that the code actually follows that contract.

## What it catches

It detects data races: concurrent accesses to the same memory location where at least one access is a write and there is no synchronization edge between them.

Typical examples:

- reading a map while another goroutine writes it,
- mutating shared state outside the channel protocol that is supposed to protect it,
- forgetting to lock around a cache update.

The practical reading is simple: it catches unsynchronized shared mutable memory, not general concurrency bugs.

## Tiny race example

```go
func TestRacyMap(t *testing.T) {
    cache := map[string]int{}

    go func() { cache["a"] = 1 }()
    go func() { _ = cache["a"] }()
}
```

That code may appear to "work" in a casual run. Under `-race`, it is exactly the kind of bug you want flagged immediately.

Some map races panic even without `-race`, but many other races do not. The detector matters because most unsafe shared-state bugs are not that obvious.

## A design bug that still passes `-race`

This code is race-free and still broken:

```go
func fetchWithTimeout(ctx context.Context) error {
	reply := make(chan Result)

	go func() {
		reply <- slowQuery()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-reply:
		return nil
	}
}
```

If the caller times out first, the goroutine can block forever on `reply <- slowQuery()`. The race detector stays clean because there is no shared-memory race. This is why leak and shutdown tests still matter.

## What it does not prove

A clean `-race` run does **not** prove:

- absence of deadlocks,
- absence of goroutine leaks,
- fairness,
- good shutdown behavior,
- good timeout behavior.

It also only catches executions that actually happen during the test run.

## How to use it

Start with the full suite:

```bash
go test -race ./...
```

Then narrow it when you need faster iteration:

```bash
go test -race ./examples/workerpool -run TestGenerateQuotesCancelsSlowJobsAfterError
go test -race -count=20 -shuffle=on ./...
go test -race -cpu=1,4 ./...
```

Useful patterns:

| Command | Use it for |
| --- | --- |
| `go test -race ./...` | module-wide baseline |
| `go test -race ./pkg -run TestName` | quick fix/verify loop on one test |
| `go test -race -count=20 -shuffle=on ./...` | shaking out schedule-sensitive paths |
| `go test -race -cpu=1,4 ./...` | catching code that only breaks at higher parallelism |

The detector is slower and more expensive than ordinary test runs. That is normal and worth the cost for concurrency-heavy code.

## How to read a race report

A report gives you two things that matter:

1. the conflicting accesses,
2. the goroutine creation stacks that made those accesses possible.

Do not stop at the two lines that raced. The real bug is usually one layer higher:

- a state owner got bypassed,
- a map escaped the mutex that was meant to guard it,
- a background goroutine outlived the context that defined its lifetime,
- a refactor added a fast path that skipped synchronization.

## What to do after a report

Use a fixed workflow:

1. Identify the piece of mutable state that has lost a clear owner.
2. Decide whether it should be guarded by ownership, a mutex, atomics, or message passing.
3. Remove the accidental side path instead of stacking random locks on top.
4. Re-run the smallest relevant `-race` command first, then the full suite.

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

Another common failure pattern is assuming a clean race run means the protocol is correct. If the bug is "goroutine never exits" or "reply path wedges after timeout," `-race` may stay green forever.

## Good design patterns under `-race`

- single owner goroutine per mutable state machine,
- `map + mutex` with short critical sections,
- explicit reply channels for one-shot responses,
- cancellation paths that cannot leave background goroutines mutating shared state after the test returns,
- tests that drive real protocol completion instead of sleeping and hoping.

## Official reading

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)

## Practical takeaway

`-race` is the minimum bar, not the finish line.

Run it early, run it often, and combine it with deterministic schedule tests and shutdown tests.
