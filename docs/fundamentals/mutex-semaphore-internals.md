---
title: Mutex and Runtime Semaphore Internals
description: Understand sync.Mutex fast and slow paths, starvation mode, and the runtime semaphore system used for parking and wakeups.
---

# Mutex and Runtime Semaphore Internals

Go is famous for channels, but serious Go systems also depend heavily on mutexes.

To understand Go concurrency deeply, you need to understand both:

- the public `sync.Mutex` API,
- the runtime semaphore mechanism used underneath contended synchronization.

:::info Version note
This page is aligned with Go 1.26, especially `internal/sync/mutex.go` and `runtime/sema.go`.
:::

## `sync.Mutex` has a tiny surface and a rich implementation

At the implementation level, a mutex stores:

- `state` bits,
- a semaphore word used by the runtime for parking and waking waiters.

Key state bits include:

- `mutexLocked`,
- `mutexWoken`,
- `mutexStarving`,
- waiter count bits.

## Fast path: optimistic CAS

The uncontended path is simple:

1. compare-and-swap the state from `0` to `mutexLocked`,
2. return immediately if it succeeds.

This is why mutexes are extremely cheap when uncontended.

## Slow path: spin, queue, or sleep

If the fast path fails, the mutex enters `lockSlow()`.

There the runtime may:

- spin briefly if spinning is likely to help,
- mark itself woken,
- increment waiter count,
- sleep on the runtime semaphore.

The point is to avoid expensive parking when the lock will probably become available very soon, while still falling back to sleeping when contention is real.

## Normal mode vs starvation mode

One of the most important details in `internal/sync/mutex.go` is that Go mutexes have two regimes:

- normal mode,
- starvation mode.

### Normal mode

In normal mode:

- waiters are queued,
- a woken waiter does not automatically own the lock,
- newly arriving goroutines may win the race.

This gives high throughput because the goroutine already running on CPU often acquires the lock quickly.

### Starvation mode

If a waiter has been delayed for long enough, the mutex can switch into starvation mode.

In starvation mode:

- ownership is handed directly to the front waiter,
- new arrivals stop trying to barge ahead,
- throughput drops somewhat, but tail-latency pathologies improve.

The source currently uses a starvation threshold of about 1 millisecond.

That is a concrete example of Go choosing practical fairness rather than perfect simplicity.

## Unlock is not just "set to zero"

Unlock starts with a fast path that clears the locked bit.

If there are no waiters, that is enough.

Under contention, the slow path decides whether to:

- wake one waiter,
- or directly hand off ownership in starvation mode.

This handoff behavior is one reason mutex performance under contention is much more nuanced than a naive spinlock.

## The runtime semaphore is not a general-purpose semaphore API

The runtime comments in `runtime/sema.go` say something important: these semaphores are really a sleep-and-wakeup primitive used to build higher-level synchronization.

The goal is similar to a futex:

- every sleep is paired with one wakeup,
- wakeup should not be lost even if races occur around the sleep,
- blocking primitives like mutexes and wait groups can build on top of it.

## How waiters are tracked

The runtime hashes wait addresses into a semaphore table.

Each root contains:

- a lock,
- a waiter count,
- a tree of waiting `sudog` records keyed by address.

That structure matters because runtime primitives need to scale across many independently contended addresses without a single giant global waiter list.

## Why channels and mutexes both matter

Channels and mutexes solve different problems well:

| Tool | Best at |
| --- | --- |
| Channel | communication protocol, work distribution, ownership transfer |
| Mutex | protecting local shared state with minimal ceremony |

Go's concurrency story is strongest when you choose the primitive that matches the shape of the problem, not when you force everything through one ideology.

## Common misunderstandings

### "Channels are always more Go-like than mutexes"

No.

If you just need to protect a small in-memory map or short critical section, a mutex is usually simpler, faster, and easier to maintain than inventing a channel protocol.

### "Mutex contention only affects throughput"

Also no.

Because mutexes deliberately balance throughput and fairness, contention affects:

- tail latency,
- goroutine parking behavior,
- scheduler interaction,
- sometimes entire-service responsiveness.

## Practical implications

Understanding these internals should change how you design:

- keep critical sections short,
- avoid holding locks across slow I/O,
- do not assume fairness is automatic,
- choose actors or channels when explicit ownership is clearer than shared-memory locking,
- choose mutexes when the state is local and the protocol is trivial.

## Practical takeaway

Go's public synchronization primitives look small because the runtime absorbs a lot of complexity on your behalf.

That complexity is exactly what lets mutexes be fast in the common case and survivable under contention.

Next, move to [Structured Concurrency](/advanced/structured-concurrency) or [Weighted Semaphore](/advanced/weighted-semaphore) to see how higher-level patterns build on these foundations.
