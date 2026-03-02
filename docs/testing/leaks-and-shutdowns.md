---
title: Leak, Shutdown, and Timeout Testing
description: Test whether goroutines terminate, queues drain, and shutdown deadlines are enforced.
---

# Leak, Shutdown, and Timeout Testing

Many concurrency regressions are lifecycle bugs:

- a sender blocks forever,
- a worker keeps running after the parent returned,
- shutdown never completes,
- a timed-out caller leaves background work wedged.

These are not always race detector failures. You need explicit tests for them.

## What to verify

For any nontrivial concurrent component, write tests that prove:

1. accepted work completes or is canceled by policy,
2. rejected work fails clearly once shutdown begins,
3. blocked goroutines have an escape path,
4. shutdown honors a hard deadline.

## Example checklist

### Cancellation after first error

This is already demonstrated in `examples/workerpool/workerpool_test.go`.

The test does not only assert an error. It asserts that slow jobs observe cancellation and stop.

### Graceful drain on shutdown

`examples/gracefulshutdown/gracefulshutdown_test.go` verifies that admitted jobs finish before shutdown returns.

### Non-blocking broker replies after caller timeout

`examples/requestreply/requestreply_test.go` verifies that a timed-out caller does not wedge the broker on reply.

## Useful test shapes

### Deadline-bounded shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

if err := processor.Shutdown(ctx); err == nil {
    t.Fatal("expected shutdown deadline error")
}
```

### Leak detection by completion protocol

Instead of counting goroutines globally, prefer explicit completion signals:

- `WaitGroup`,
- closed `done` channel,
- `context` cancellation plus wait.

These are usually more robust than asserting raw goroutine counts.

## Common mistakes

### Tests that only check the happy path

The happy path almost never finds lifecycle bugs.

### Using real sleeps as proof of termination

`time.Sleep(50 * time.Millisecond)` is not proof that a goroutine is gone.

### Forgetting the caller-abandoned path

Any background worker or broker that replies to a caller should be tested for the case where the caller times out or leaves first.

## Practical takeaway

If a concurrent component has a start path, it also has an end path. Test both with equal seriousness.
