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

The mental model is:

- the bubble owns the goroutines it starts,
- the bubble owns its fake clock,
- time moves only when the bubble has reached a stable blocked point.

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

## Another useful example: `context.AfterFunc`

```go
func TestAfterFuncRunsAfterCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()

		called := false
		context.AfterFunc(ctx, func() {
			called = true
		})

		synctest.Wait()
		if called {
			t.Fatal("AfterFunc ran before cancellation")
		}

		cancel()
		synctest.Wait()
		if !called {
			t.Fatal("AfterFunc did not run after cancellation")
		}
	})
}
```

This is the kind of test that used to invite arbitrary sleeps and now does not need any.

## What "durably blocked" means

This definition matters.

According to the standard library docs, operations such as:

- blocking channel send/receive on bubble-created channels,
- `sync.WaitGroup.Wait` for bubble-associated wait groups,
- `sync.Cond.Wait`,
- `time.Sleep`

can durably block a goroutine inside the bubble.

By contrast, some things may block but are **not** durably blocked:

- `sync.Mutex` or `sync.RWMutex` lock acquisition,
- network I/O,
- system calls,
- events driven by goroutines outside the bubble.

That distinction is why `synctest` works best with self-contained tests and fake dependencies.

## What it is good at and where it is a poor fit

| Great fit | Poor fit |
| --- | --- |
| timeout and retry loops driven by `time` | real sockets and live network services |
| cancellation propagation | goroutines started outside the bubble |
| internal worker coordination | external processes and OS-driven events |
| `AfterFunc`, timers, tickers | tests that depend on `sync.Mutex` blocking semantics |

If you need protocol-level I/O but still want bubble control, prefer fake in-process transports such as `net.Pipe` over loopback sockets.

## Isolation rules that matter in practice

The bubble is stricter than it first appears:

- a channel, timer, or ticker created in the bubble is associated with that bubble,
- operating on those bubbled objects from outside the bubble panics,
- a `WaitGroup` becomes bubble-associated on its first `Add` or `Go`,
- package-global `WaitGroup` values are a technical limitation; prefer test-owned or heap-allocated wait groups instead.

This is one reason `synctest` pushes you toward cleaner ownership boundaries in tests.

## Deadlock behavior

Two rules are easy to miss:

1. if every goroutine is durably blocked and the next timer can fire, fake time advances,
2. if every goroutine is durably blocked and nothing can make progress, `synctest.Test` panics with a deadlock.

Also, fake time stops advancing once the root goroutine for the bubble exits. That means background work must still be owned and awaited properly.

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

Use it as a new layer in the stack, not as a universal concurrency test runner.

## Failure pattern

Wall-clock sleeps are still the most common reason concurrency tests stay flaky:

```go
go func() {
	time.Sleep(100 * time.Millisecond)
	done <- struct{}{}
}()

time.Sleep(10 * time.Millisecond) // bad: timing guess
```

`testing/synctest` exists so internal timers and scheduling behavior can be driven deterministically instead of by luck.

## Official reading

- [Testing Time (Go Blog)](https://go.dev/blog/testing-time)
- [testing/synctest package docs](https://pkg.go.dev/testing/synctest)

## Practical takeaway

If your concurrency tests still rely on arbitrary sleeps, `synctest` is one of the highest-leverage upgrades you can make.
