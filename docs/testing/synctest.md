---
title: Deterministic Tests with synctest
description: Use testing/synctest to remove real-time sleeps and make timeout-heavy concurrent tests deterministic.
---

# Deterministic Tests with synctest

For years, one of the biggest pain points in Go concurrency testing was time.

Tests that relied on `time.Sleep` were often:

- slow,
- flaky,
- sensitive to machine load,
- unconvincing.

`testing/synctest` changes that.

:::tip Quick takeaway
`synctest` gives you a bubble-local concurrency world with a fake clock. Time advances when bubble goroutines are durably blocked, which makes many timeout and cancellation tests deterministic.
:::

## The core idea

Inside a `synctest.Test` bubble:

- goroutines started by the test belong to the bubble,
- the `time` package uses a fake clock,
- `synctest.Wait()` waits until all other goroutines are durably blocked.

This lets you test timeout logic without waiting in wall-clock time.

## A practical example

```go
func TestContextTimeout(t *testing.T) {
    synctest.Test(t, func(t *testing.T) {
        ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
        defer cancel()

        time.Sleep(5*time.Second - time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != nil {
            t.Fatalf("timeout fired too early: %v", err)
        }

        time.Sleep(time.Nanosecond)
        synctest.Wait()
        if err := ctx.Err(); err != context.DeadlineExceeded {
            t.Fatalf("timeout missing: %v", err)
        }
    })
}
```

That test is deterministic and immediate. No "sleep 200ms and hope."

## What "durably blocked" means

This definition matters.

According to the standard library docs, operations such as:

- blocking channel send/receive on bubble-created channels,
- `sync.WaitGroup.Wait` for bubble-associated wait groups,
- `sync.Cond.Wait`,
- `time.Sleep`

can durably block a goroutine inside the bubble.

By contrast, network I/O and external syscalls are not durably blocked in the same way. That is why `synctest` works best with self-contained tests and fake dependencies.

## Where it fits well

- timeout-heavy request code,
- cancellation propagation,
- background callback execution,
- internal worker coordination,
- retry loops driven by timers or deadlines.

## Where it does not replace other tools

It does not replace:

- `-race`,
- integration tests with real networking,
- trace and profile analysis,
- leak tests for code that interacts with external systems.

## Official reading

- [Testing Time (Go Blog)](https://go.dev/blog/testing-time)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)

## Practical takeaway

If your concurrency tests still rely on arbitrary sleeps, `synctest` is one of the highest-leverage upgrades you can make.
