---
title: Control Flow and Evaluation Pitfalls
description: Ten Go control-flow and evaluation mistakes that still trap experienced engineers.
---

# Control Flow and Evaluation Pitfalls

## 1. `select` with `default` causing accidental busy-spin

**Bad**

```go
for {
	select {
	case job := <-jobs:
		handle(job)
	default:
		continue
	}
}
```

Why it bites: this loop never blocks. When `jobs` is empty, it burns CPU instead of waiting for work.

**Better**

```go
for {
	select {
	case job, ok := <-jobs:
		if !ok {
			return
		}
		handle(job)
	case <-ctx.Done():
		return
	}
}
```

Catch it early: CPU profile, execution trace, review.

Rule: a long-lived `select` should block unless polling is truly intentional.

## 2. Nil channel branches silently disabling `select` cases

**Bad**

```go
var updates <-chan Event

if cfg.LiveReload {
	updates = watcher.Events()
}

select {
case ev := <-updates:
	apply(ev)
case <-ctx.Done():
	return ctx.Err()
}
```

Why it bites: a nil channel case never fires. That can be useful on purpose, but it is easy to forget that one branch is now permanently disabled.

**Better**

```go
if !cfg.LiveReload {
	<-ctx.Done()
	return ctx.Err()
}

updates := watcher.Events()
select {
case ev := <-updates:
	apply(ev)
case <-ctx.Done():
	return ctx.Err()
}
```

Catch it early: focused tests, review.

Rule: use nil-channel disabling only when the code makes that choice explicit.

## 3. Receiving from a closed channel and misreading zero values as real data

**Bad**

```go
for {
	id := <-ids
	if id == "" {
		return nil
	}
	process(id)
}
```

Why it bites: a receive on a closed channel returns the zero value immediately. Here `""` can mean either "channel closed" or "real empty value."

**Better**

```go
for {
	id, ok := <-ids
	if !ok {
		return nil
	}
	process(id)
}
```

Catch it early: tests, review.

Rule: if channel closure matters, always receive the `ok` flag.

## 4. Sending on a closed channel

**Bad**

```go
func consume(work <-chan Job) {
	for job := range work {
		handle(job)
	}
}

func stop(work chan Job) {
	close(work)
}
```

Why it bites: whoever still sends to `work` after `stop` runs will panic.

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

## 5. Assuming `select` is fair

**Bad**

```go
for {
	select {
	case job := <-hotQueue:
		handle(job)
	case <-ticker.C:
		flushMetrics()
	}
}
```

Why it bites: `select` does not guarantee the kind of fairness people often imagine. A hot case can dominate and delay less-busy cases.

**Better**

```go
go func() {
	for range ticker.C {
		flushMetrics()
	}
}()

for job := range hotQueue {
	handle(job)
}
```

Catch it early: load tests, trace, review.

Rule: if one branch must make progress independently, give it an explicit owner.

## 6. `break` inside `select` or `switch` not breaking the outer loop

**Bad**

```go
for {
	select {
	case <-ctx.Done():
		break
	case job := <-jobs:
		handle(job)
	}
}
cleanup()
```

Why it bites: `break` exits the `select`, not the surrounding `for`. The loop keeps running.

**Better**

```go
loop:
for {
	select {
	case <-ctx.Done():
		break loop
	case job := <-jobs:
		handle(job)
	}
}
cleanup()
```

Catch it early: tests, review.

Rule: use `return` or a labeled `break` when you mean to leave the outer loop.

## 7. Loop variable capture in goroutines

**Bad**

```go
var id int

for _, id = range ids {
	go func() {
		log.Println(id)
	}()
}
```

Why it bites: the goroutines capture the same `id` variable. In modern Go, the common `for _, id := range ids` case is fixed, but predeclared loop variables and older code still have this trap.

**Better**

```go
var id int

for _, id = range ids {
	id := id
	go func() {
		log.Println(id)
	}()
}
```

Catch it early: tests, review.

Rule: if a goroutine closes over a loop variable, make the per-iteration binding explicit.

## 8. Loop variable capture in `defer`

**Bad**

```go
var name string

for _, name = range files {
	defer func() {
		log.Println("closing", name)
	}()
}
```

Why it bites: all deferred closures observe the same variable, so they print the last value.

**Better**

```go
var name string

for _, name = range files {
	name := name
	defer func() {
		log.Println("closing", name)
	}()
}
```

Catch it early: tests, review.

Rule: `defer` captures variables, not snapshots, unless you create one.

## 9. `:=` shadowing `err`

**Bad**

```go
func run() error {
	err := start()
	if err != nil {
		return err
	}

	if data, err := fetch(); err != nil {
		log.Println("fetch failed:", err)
	} else {
		use(data)
	}

	return err
}
```

Why it bites: the `err` inside the `if` is a different variable. The function returns the outer `err`, which is still `nil`.

**Better**

```go
func run() error {
	if err := start(); err != nil {
		return err
	}

	data, err := fetch()
	if err != nil {
		return err
	}
	use(data)
	return nil
}
```

Catch it early: tests, review.

Rule: if the error matters after the block, do not hide it inside a short declaration.

## 10. `defer` inside loops holding resources too long

**Bad**

```go
for _, name := range names {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	scan(f)
}
```

Why it bites: every file stays open until the whole function returns. A large loop can exhaust file descriptors.

**Better**

```go
for _, name := range names {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	scan(f)
	if err := f.Close(); err != nil {
		return err
	}
}
```

Catch it early: tests under load, review.

Rule: `defer` is function-scoped, not loop-iteration-scoped.
