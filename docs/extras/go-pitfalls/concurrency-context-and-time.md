---
title: Concurrency, Context, and Time Pitfalls
description: Ten concurrency and lifecycle mistakes that create leaks, wedges, and noisy shutdowns.
---

# Concurrency, Context, and Time Pitfalls

## 31. Closing a channel from the receiver side

**Bad**

```go
func consume(work chan Job) {
	job := <-work
	handle(job)
	close(work)
}

func produce(work chan<- Job, job Job) {
	work <- job
}
```

Why it bites: the next sender panics. The receiver usually does not own the send lifecycle.

**Better**

```go
func produce(done <-chan struct{}, work chan<- Job, job Job) error {
	select {
	case work <- job:
		return nil
	case <-done:
		return context.Canceled
	}
}
```

Catch it early: tests, review.

Rule: the sending side should own closing the data channel.

## 32. Multiple senders racing to close the same channel

**Bad**

```go
func stop(done chan struct{}) {
	close(done)
}

go stop(done)
go stop(done)
```

Why it bites: only one close succeeds. The second one panics.

**Better**

```go
var once sync.Once

func stop(done chan struct{}) {
	once.Do(func() {
		close(done)
	})
}
```

Catch it early: tests, review.

Rule: if many goroutines may stop one signal channel, serialize the close.

## 33. Using `len(ch)` as a coordination signal

**Bad**

```go
if len(jobs) == 0 {
	return ErrIdle
}

job := <-jobs
handle(job)
```

Why it bites: `len(ch)` is only a snapshot. Another goroutine can change the channel before the receive.

**Better**

```go
select {
case job := <-jobs:
	handle(job)
default:
	return ErrIdle
}
```

Catch it early: tests, review.

Rule: coordinate with channel operations, not with `len`.

## 34. Goroutine leak from blocked reply send after caller timeout

**Bad**

```go
func request(ctx context.Context) error {
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

Why it bites: if the caller times out first, the goroutine can block forever on `reply <-`.

**Better**

```go
func request(ctx context.Context) error {
	reply := make(chan Result, 1)

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

Catch it early: leak tests, shutdown tests, review.

Rule: any reply path must still terminate when the caller leaves first.

## 35. Starting goroutines with no explicit owner or stop condition

**Bad**

```go
func StartMetricsLoop() {
	go func() {
		for range time.Tick(time.Second) {
			publishMetrics()
		}
	}()
}
```

Why it bites: this goroutine has no owner, no shutdown path, and no contract about when it ends.

**Better**

```go
func StartMetricsLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				publishMetrics()
			case <-ctx.Done():
				return
			}
		}
	}()
}
```

Catch it early: lifecycle tests, review.

Rule: every goroutine needs an owner, a stop condition, and a wait strategy.

## 36. Using `context.Background()` in request-scoped work

**Bad**

```go
func handleRequest(ctx context.Context, id string) error {
	go auditLog(context.Background(), id)
	return fetch(ctx, id)
}
```

Why it bites: the background work escapes the request lifetime. Timeouts and cancellations no longer control it.

**Better**

```go
func handleRequest(ctx context.Context, id string) error {
	go auditLog(ctx, id)
	return fetch(ctx, id)
}
```

Catch it early: tests, traces, review.

Rule: request-scoped work should inherit the request context unless escape is intentional.

## 37. Forgetting to call `cancel`

**Bad**

```go
func fetchAll(parent context.Context) error {
	ctx, _ := context.WithTimeout(parent, 2*time.Second)
	return callBackend(ctx)
}
```

Why it bites: the timer and child context resources live until the deadline expires even if the work finishes early.

**Better**

```go
func fetchAll(parent context.Context) error {
	ctx, cancel := context.WithTimeout(parent, 2*time.Second)
	defer cancel()
	return callBackend(ctx)
}
```

Catch it early: review, tests.

Rule: if you create a cancelable context, call its cancel function.

## 38. `time.After` in loops creating timer churn

**Bad**

```go
for {
	select {
	case msg := <-msgs:
		handle(msg)
	case <-time.After(time.Second):
		reportIdle()
	}
}
```

Why it bites: every loop iteration allocates a new timer. Under load, this creates avoidable churn.

**Better**

```go
timer := time.NewTimer(time.Second)
defer timer.Stop()

for {
	timer.Reset(time.Second)
	select {
	case msg := <-msgs:
		handle(msg)
	case <-timer.C:
		reportIdle()
	}
}
```

Catch it early: alloc profiles, review.

Rule: long-lived loops should usually reuse timers instead of recreating them.

## 39. Forgetting to `Stop` a `Ticker`

**Bad**

```go
func start() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			refreshCache()
		}
	}()
}
```

Why it bites: the ticker keeps firing until you stop it. If the loop exits or the owner dies, the ticker still needs cleanup.

**Better**

```go
func start(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				refreshCache()
			case <-ctx.Done():
				return
			}
		}
	}()
}
```

Catch it early: tests, review.

Rule: whoever creates a ticker should also define who stops it.

## 40. Calling `WaitGroup.Add` after goroutine start instead of before

**Bad**

```go
var wg sync.WaitGroup

go func() {
	defer wg.Done()
	work()
}()
wg.Add(1)
wg.Wait()
```

Why it bites: `Wait` can race with `Add`, and `Done` can happen before the counter was incremented.

**Better**

```go
var wg sync.WaitGroup

wg.Add(1)
go func() {
	defer wg.Done()
	work()
}()
wg.Wait()
```

Catch it early: tests, review.

Rule: increment the wait-group count before the goroutine becomes runnable.
