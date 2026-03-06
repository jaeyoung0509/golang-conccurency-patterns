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

:::tip Quick takeaway
If a component has `Start`, `Run`, `Submit`, or `Check`, it also has a lifetime contract. Good tests prove both the admission path and the exit path.
:::

## What to verify

For any nontrivial concurrent component, write tests that prove:

1. accepted work completes or is canceled by policy,
2. rejected work fails clearly once shutdown begins,
3. blocked goroutines have an escape path,
4. shutdown honors a hard deadline.

## A shutdown contract worth testing

Before writing tests, make the contract concrete. A good concurrent component usually answers all of these:

- what work is still accepted after shutdown begins,
- whether already accepted work drains, cancels, or is dropped,
- which error the caller sees after admission is closed,
- how long shutdown may wait,
- who owns background goroutines and when they are guaranteed to exit.

If the implementation cannot answer those questions, the tests will stay vague too.

## Example checklist

### Cancellation after first error

This is already demonstrated in `examples/workerpool/workerpool_test.go`.

The test does not only assert an error. It asserts that slow jobs observe cancellation and stop.

### Graceful drain on shutdown

`examples/gracefulshutdown/gracefulshutdown_test.go` verifies that admitted jobs finish before shutdown returns.

### Non-blocking broker replies after caller timeout

`examples/requestreply/requestreply_test.go` verifies that a timed-out caller does not wedge the broker on reply.

## Useful test shapes

### Admission closes once shutdown starts

```go
func TestSubmitRejectedAfterShutdown(t *testing.T) {
	processor := newProcessor(t)

	require.NoError(t, processor.Shutdown(context.Background()))

	err := processor.Submit(context.Background(), Event{OrderID: "ord-1"})
	if !errors.Is(err, ErrClosed) {
		t.Fatalf("Submit error = %v, want ErrClosed", err)
	}
}
```

This is the test that proves callers get a stable answer instead of "sometimes blocked, sometimes accepted."

### Deadline-bounded shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
defer cancel()

if err := processor.Shutdown(ctx); err == nil {
    t.Fatal("expected shutdown deadline error")
}
```

### Blocked sender must have an escape path

```go
func TestWorkerExitsOnCancel(t *testing.T) {
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		defer close(done)
		runWorker(ctx)
	}()

	cancel()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("worker did not exit after cancellation")
	}
}
```

The point is not the `100ms`. The point is the explicit completion signal. The timeout only prevents the test from hanging forever.

### Caller-abandoned reply path

```go
func TestReplyPathDoesNotLeakAfterCallerTimeout(t *testing.T) {
	release := make(chan struct{})
	broker := newBroker(func() Result {
		<-release
		return Result{}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	if err := broker.Check(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Check error = %v, want deadline exceeded", err)
	}

	close(release)
	if err := broker.Check(context.Background()); err != nil {
		t.Fatalf("follow-up request failed after timeout path: %v", err)
	}
}
```

This is the pattern behind `examples/requestreply/requestreply_test.go`: prove that one abandoned caller does not poison the whole broker.

### Leak detection by completion protocol

Instead of counting goroutines globally, prefer explicit completion signals:

- `WaitGroup`,
- closed `done` channel,
- `context` cancellation plus wait.

These are usually more robust than asserting raw goroutine counts.

Raw goroutine counts can still be useful as a coarse smoke test in CI, but they are a poor primary oracle because the runtime and unrelated tests create background activity of their own.

## Failure pattern

This kind of test is comforting and nearly useless:

```go
func TestWorkerStops(t *testing.T) {
	startWorker()
	time.Sleep(50 * time.Millisecond) // bad: sleep is not evidence of shutdown
}
```

A sleeping test can pass while goroutines remain parked on channels, timers, or locks. Good shutdown tests need an explicit completion signal or a bounded `Shutdown` contract they can assert on.

## Common mistakes

### Tests that only check the happy path

The happy path almost never finds lifecycle bugs.

### Using real sleeps as proof of termination

`time.Sleep(50 * time.Millisecond)` is not proof that a goroutine is gone.

### Forgetting the caller-abandoned path

Any background worker or broker that replies to a caller should be tested for the case where the caller times out or leaves first.

### Reusing `t.Context()` inside cleanup

`testing.T.Context()` is canceled just before cleanup runs. If teardown needs a live context, create a fresh bounded background context in the cleanup function instead of reusing `t.Context()`.

## A practical review checklist

When you review a lifecycle test, ask:

1. Does the test prove when admission closes?
2. Does it prove how accepted work ends?
3. Does every blocked goroutine have a cancellation or close path?
4. Does the test use explicit completion, not "sleep and hope"?
5. Does teardown finish within a bounded deadline?

## Practical takeaway

If a concurrent component has a start path, it also has an end path. Test both with equal seriousness.
