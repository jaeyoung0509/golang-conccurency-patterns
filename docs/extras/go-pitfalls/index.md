---
title: Go Pitfalls Appendix
description: A single-page bilingual appendix of 50 Go pitfalls that compile cleanly and still create bugs.
---

# Go Pitfalls Appendix

This appendix is about code that:

- compiles,
- often passes casual review,
- and still bites hard in production.

The goal is not to teach syntax from scratch. The goal is to sharpen judgment around the kinds of mistakes experienced Go engineers still make under deadline pressure.

## What this appendix covers

| Category | What it focuses on |
| --- | --- |
| [Control Flow and Evaluation](#control-flow-and-evaluation) | `select`, `break`, loop variables, `defer`, and evaluation rules that look obvious until they are not |
| [Interfaces and Types](#interfaces-and-types) | typed nil, assertions, method sets, wrapped errors, and `any`-shaped ambiguity |
| [Collections and Memory](#collections-and-memory) | map/slice pitfalls, iteration assumptions, backing-array retention, and memory-shape surprises |
| [Concurrency, Context, and Time](#concurrency-context-and-time) | channel ownership, goroutine lifetime, cancellation, timers, and `WaitGroup` sequencing |
| [Stdlib and API Boundaries](#stdlib-and-api-boundaries) | `net/http`, `database/sql`, `io`, `json`, `time`, `os/exec`, and `sync.Map` footguns |

## How to read this page

Every pitfall follows the same shape:

1. a bad snippet,
2. why it breaks or misleads,
3. a better snippet,
4. how to catch it early,
5. one rule of thumb.

Read it like a code-review checklist, not like a language tutorial.

## Where to go deeper

This appendix is intentionally sharp and practical. When an item touches a bigger runtime or operational topic, the deeper explanation already exists elsewhere in the site:

- [Race Detector](/testing/race-detector)
- [Deterministic Tests with synctest](/testing/synctest)
- [context Package Internals](/stdlib/context-internals)
- [time, Timers, and Tickers](/stdlib/time-timers-tickers)
- [net/http Server and Transport](/stdlib/net-http-server-transport)
- [sync and atomic Primitives](/stdlib/sync-and-atomic)
- [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model)

## One important modern-Go note

Some classic blog posts and talks about Go pitfalls are now partially outdated.

In particular, loop-variable capture changed in Go 1.22 for the common `for ... := range ...` case. This appendix uses modern semantics and calls out where old advice still matters.

## Jump by category

- [Control Flow and Evaluation](#control-flow-and-evaluation)
- [Interfaces and Types](#interfaces-and-types)
- [Collections and Memory](#collections-and-memory)
- [Concurrency, Context, and Time](#concurrency-context-and-time)
- [Stdlib and API Boundaries](#stdlib-and-api-boundaries)

<a id="control-flow-and-evaluation"></a>

## Control Flow and Evaluation

#### 1. `select` with `default` causing accidental busy-spin

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

#### 2. Nil channel branches silently disabling `select` cases

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

#### 3. Receiving from a closed channel and misreading zero values as real data

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

#### 4. Sending on a closed channel

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

#### 5. Assuming `select` is fair

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

#### 6. `break` inside `select` or `switch` not breaking the outer loop

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

#### 7. Loop variable capture in goroutines

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

#### 8. Loop variable capture in `defer`

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

#### 9. `:=` shadowing `err`

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

#### 10. `defer` inside loops holding resources too long

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

<a id="interfaces-and-types"></a>

## Interfaces and Types

#### 11. Typed nil inside an interface not comparing equal to `nil`

**Bad**

```go
func maybeReader() io.Reader {
	var f *os.File
	return f
}

r := maybeReader()
if r == nil {
	return
}
```

Why it bites: the interface value has a concrete type `*os.File`, so the interface itself is not nil even though the concrete pointer is nil.

**Better**

```go
func maybeReader() io.Reader {
	var f *os.File
	if f == nil {
		return nil
	}
	return f
}
```

Catch it early: tests, review.

Rule: an interface is nil only when both its type and value are nil.

#### 12. Comparing interface values whose concrete values are not comparable

**Bad**

```go
var left any = []int{1, 2}
var right any = []int{1, 2}

if left == right {
	log.Println("equal")
}
```

Why it bites: the comparison panics because slices are not comparable, and interfaces use the concrete type's comparability rules.

**Better**

```go
left := []int{1, 2}
right := []int{1, 2}

if slices.Equal(left, right) {
	log.Println("equal")
}
```

Catch it early: tests, review.

Rule: compare the concrete type with the right helper, not the boxed interface.

#### 13. Pointer-vs-value receiver confusion in interface satisfaction

**Bad**

```go
type Writer struct{}

func (*Writer) Write(p []byte) (int, error) {
	return len(p), nil
}

var _ io.Writer = Writer{}
```

Why it bites: `Writer{}` does not implement `io.Writer`; only `*Writer` does. This fails at compile time.

**Better**

```go
type Writer struct{}

func (*Writer) Write(p []byte) (int, error) {
	return len(p), nil
}

var _ io.Writer = (*Writer)(nil)
```

Catch it early: compiler, compile-time interface assertions.

Rule: if the method has a pointer receiver, the interface is satisfied by the pointer type.

#### 14. Nil receiver methods that still panic internally

**Bad**

```go
type Config struct {
	Timeout time.Duration
}

func (c *Config) EffectiveTimeout() time.Duration {
	return c.Timeout
}
```

Why it bites: method dispatch on a nil receiver is legal, but dereferencing fields inside the method still panics.

**Better**

```go
func (c *Config) EffectiveTimeout() time.Duration {
	if c == nil {
		return 5 * time.Second
	}
	return c.Timeout
}
```

Catch it early: tests, review.

Rule: if nil is an allowed receiver state, handle it intentionally inside the method.

#### 15. Implicit interface satisfaction assumed without checking method set shape

**Bad**

```go
type Saver interface {
	Save(context.Context, string) error
}

type Store struct{}

func (Store) Save(context.Context, string) error { return nil }
```

Why it bites: this looks fine today, but a refactor can quietly drift and the break only appears when some distant wiring site tries to assign `Store` to `Saver`.

**Better**

```go
type Saver interface {
	Save(context.Context, string) error
}

type Store struct{}

func (Store) Save(context.Context, string) error { return nil }

var _ Saver = Store{}
```

Catch it early: compiler, compile-time interface assertions.

Rule: when interface conformance is part of the design, assert it near the type.

#### 16. Unchecked type assertions

**Bad**

```go
func handle(msg any) {
	order := msg.(OrderPlaced)
	process(order)
}
```

Why it bites: one unexpected input shape turns this into a panic instead of a controlled error.

**Better**

```go
func handle(msg any) error {
	order, ok := msg.(OrderPlaced)
	if !ok {
		return fmt.Errorf("unexpected message type %T", msg)
	}
	process(order)
	return nil
}
```

Catch it early: tests, review.

Rule: use the one-value assertion only when panic is truly the right policy.

#### 17. Using `==` on wrapped errors instead of `errors.Is` / `errors.As`

**Bad**

```go
if err := load(); err != nil {
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}
```

Why it bites: wrapped errors stop direct equality from working the way you expect.

**Better**

```go
if err := load(); err != nil {
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	return err
}
```

Catch it early: tests, review.

Rule: if an error might be wrapped, compare semantically, not by direct equality.

#### 18. Assuming `any` preserves intended domain shape

**Bad**

```go
func setLimit(v any) int {
	limit := v.(int)
	if limit < 0 {
		panic("negative")
	}
	return limit
}
```

Why it bites: `any` does not preserve domain intent. `int64(10)` or `json.Number("10")` now becomes a runtime problem.

**Better**

```go
func setLimit(limit int) (int, error) {
	if limit < 0 {
		return 0, errors.New("negative limit")
	}
	return limit, nil
}
```

Catch it early: compiler, tests, review.

Rule: prefer typed inputs over `any` when the domain is already known.

#### 19. `json.Unmarshal` into `interface{}` producing `float64` for numbers

**Bad**

```go
var payload map[string]any
_ = json.Unmarshal([]byte(`{"retries":3}`), &payload)

retries := payload["retries"].(int)
fmt.Println(retries)
```

Why it bites: generic JSON decoding uses `float64` for numbers by default, so the assertion to `int` panics.

**Better**

```go
type Payload struct {
	Retries int `json:"retries"`
}

var payload Payload
_ = json.Unmarshal([]byte(`{"retries":3}`), &payload)
fmt.Println(payload.Retries)
```

Catch it early: tests, review.

Rule: decode into typed structs whenever the shape is known.

#### 20. Interface fields hiding zero-value vs absent-state distinctions

**Bad**

```go
type Query struct {
	Limit any `json:"limit"`
}

var q Query
_ = json.Unmarshal(body, &q)
```

Why it bites: `nil`, `0`, `0.0`, and wrong concrete types all get folded into one vague `any` field. Validation becomes late and messy.

**Better**

```go
type Query struct {
	Limit *int `json:"limit"`
}

var q Query
_ = json.Unmarshal(body, &q)
```

Catch it early: tests, review.

Rule: use typed fields, and use pointers when "absent" is semantically different from zero.

<a id="collections-and-memory"></a>

## Collections and Memory

#### 21. Writing to a nil map

**Bad**

```go
func record(hit map[string]int, key string) {
	hit[key]++
}

func main() {
	var hit map[string]int
	record(hit, "cache")
}
```

Why it bites: reading from a nil map is fine, writing to one panics.

**Better**

```go
func record(hit map[string]int, key string) {
	if hit == nil {
		panic("record called with nil map")
	}
	hit[key]++
}
```

Catch it early: tests, review.

Rule: if a map must be writable, initialize it before you hand it around.

#### 22. Assuming nil slice and empty slice are interchangeable in APIs or JSON

**Bad**

```go
type Response struct {
	Items []string `json:"items"`
}

func emptyResponse() Response {
	return Response{}
}
```

Why it bites: nil and empty slices often behave the same in Go code, but they encode differently in JSON: `null` versus `[]`.

**Better**

```go
type Response struct {
	Items []string `json:"items"`
}

func emptyResponse() Response {
	return Response{Items: []string{}}
}
```

Catch it early: JSON tests, API contract review.

Rule: if clients expect an empty list, return an empty slice, not nil.

#### 23. Assuming nil map and empty map are interchangeable in APIs or JSON

**Bad**

```go
type Response struct {
	Labels map[string]string `json:"labels"`
}

func emptyResponse() Response {
	return Response{}
}
```

Why it bites: a nil map marshals as `null`, while an empty map marshals as `{}`.

**Better**

```go
type Response struct {
	Labels map[string]string `json:"labels"`
}

func emptyResponse() Response {
	return Response{Labels: map[string]string{}}
}
```

Catch it early: JSON tests, API contract review.

Rule: if callers expect an object, initialize an object.

#### 24. `append` aliasing through shared backing arrays

**Bad**

```go
base := make([]int, 0, 4)

left := append(base, 1, 2)
right := append(base, 9)

fmt.Println(left, right)
```

Why it bites: `left` and `right` can share the same backing array. One append path can overwrite another.

**Better**

```go
base := []int{1, 2}

left := append([]int(nil), base...)
right := append([]int(nil), base...)

left = append(left, 3)
right = append(right, 9)
```

Catch it early: tests, review.

Rule: if two slices must evolve independently, copy before branching.

#### 25. Subslices retaining huge backing arrays

**Bad**

```go
func readHeader() []byte {
	buf := make([]byte, 10<<20)
	fill(buf)
	return buf[:32]
}
```

Why it bites: the returned 32-byte slice keeps the entire 10 MB backing array alive.

**Better**

```go
func readHeader() []byte {
	buf := make([]byte, 10<<20)
	fill(buf)

	header := make([]byte, 32)
	copy(header, buf[:32])
	return header
}
```

Catch it early: heap profile, review.

Rule: copy out the small piece if the large backing array should die.

#### 26. Taking the address of a `range` variable

**Bad**

```go
var user User
var ptrs []*User

for _, user = range users {
	ptrs = append(ptrs, &user)
}
```

Why it bites: all pointers refer to the same loop variable. In modern Go the common `:= range` case is fixed, but predeclared variables and older code still have this trap.

**Better**

```go
var ptrs []*User

for i := range users {
	ptrs = append(ptrs, &users[i])
}
```

Catch it early: tests, review.

Rule: if you need element addresses, take the address of the slice element, not the loop variable.

#### 27. Mutating a slice while ranging over it

**Bad**

```go
for i, id := range ids {
	if shouldDrop(id) {
		ids = append(ids[:i], ids[i+1:]...)
	}
}
```

Why it bites: `range` uses the original slice header. Removing elements while iterating can skip items or read stale positions.

**Better**

```go
filtered := ids[:0]

for _, id := range ids {
	if !shouldDrop(id) {
		filtered = append(filtered, id)
	}
}
ids = filtered
```

Catch it early: tests, review.

Rule: build the next slice explicitly instead of mutating the one you are ranging over.

#### 28. Assuming map iteration order is stable

**Bad**

```go
func pickOne(m map[string]Route) Route {
	for _, route := range m {
		return route
	}
	panic("empty")
}
```

Why it bites: map iteration order is deliberately not stable. "First item" is not a contract.

**Better**

```go
func pickOne(keys []string, m map[string]Route) Route {
	sort.Strings(keys)
	return m[keys[0]]
}
```

Catch it early: tests that run multiple times, review.

Rule: if order matters, create order explicitly.

#### 29. Concurrent map access without synchronization

**Bad**

```go
cache := map[string]int{}

go func() { cache["a"] = 1 }()
go func() { _ = cache["a"] }()

time.Sleep(10 * time.Millisecond)
```

Why it bites: ordinary Go maps are not safe for unsynchronized concurrent reads and writes.

**Better**

```go
var (
	mu    sync.Mutex
	cache = map[string]int{}
)

mu.Lock()
cache["a"] = 1
mu.Unlock()
```

Catch it early: `-race`, tests, review.

Rule: plain maps need ownership or synchronization.

#### 30. Assuming `string` and `[]byte` conversion is free or shares storage safely

**Bad**

```go
s := "prod"
b := []byte(s)

b[0] = 'P'
fmt.Println(s)
```

Why it bites: converting a string to `[]byte` copies the data. Mutating `b` does not mutate `s`, and repeated conversions can allocate in hot paths.

**Better**

```go
b := []byte("prod")
b[0] = 'P'

s := string(b)
fmt.Println(s)
```

Catch it early: benchmarks, alloc profiles, review.

Rule: convert when you need a different representation, not because you assume the storage is shared.

<a id="concurrency-context-and-time"></a>

## Concurrency, Context, and Time

#### 31. Closing a channel from the receiver side

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

#### 32. Multiple senders racing to close the same channel

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

#### 33. Using `len(ch)` as a coordination signal

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

#### 34. Goroutine leak from blocked reply send after caller timeout

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

#### 35. Starting goroutines with no explicit owner or stop condition

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

#### 36. Using `context.Background()` in request-scoped work

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

#### 37. Forgetting to call `cancel`

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

#### 38. `time.After` in loops creating timer churn

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

#### 39. Forgetting to `Stop` a `Ticker`

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

#### 40. Calling `WaitGroup.Add` after goroutine start instead of before

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

<a id="stdlib-and-api-boundaries"></a>

## Stdlib and API Boundaries

#### 41. Creating a new `http.Client` or `Transport` per request

**Bad**

```go
func fetch(ctx context.Context, url string) error {
	client := &http.Client{
		Transport: &http.Transport{},
	}
	_, err := client.Get(url)
	return err
}
```

Why it bites: connection pooling and transport reuse disappear. You pay for handshakes and sockets over and over again.

**Better**

```go
var client = &http.Client{
	Timeout: 5 * time.Second,
	Transport: &http.Transport{
		MaxIdleConnsPerHost: 16,
	},
}
```

Catch it early: load tests, metrics, review.

Rule: `http.Client` and especially `Transport` are long-lived policy objects, not throwaway helpers.

#### 42. Not draining and closing `http.Response.Body`

**Bad**

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
if resp.StatusCode != http.StatusOK {
	return fmt.Errorf("bad status: %s", resp.Status)
}
```

Why it bites: the body stays open, and the transport may not be able to reuse the connection.

**Better**

```go
resp, err := client.Do(req)
if err != nil {
	return err
}
defer resp.Body.Close()
_, _ = io.Copy(io.Discard, resp.Body)
```

Catch it early: tests, transport metrics, review.

Rule: if you got a response body, close it, and drain it when connection reuse matters.

#### 43. Assuming `io.Reader` fills the whole buffer in one read

**Bad**

```go
buf := make([]byte, 32)
n, err := r.Read(buf)
if err != nil {
	return err
}
return decode(buf[:n])
```

Why it bites: `Read` can return fewer bytes than you asked for and still be perfectly correct.

**Better**

```go
buf := make([]byte, 32)
if _, err := io.ReadFull(r, buf); err != nil {
	return err
}
return decode(buf)
```

Catch it early: tests with short readers, review.

Rule: `io.Reader` promises a stream, not full-buffer delivery.

#### 44. Forgetting `Rows.Close` and `Rows.Err` in `database/sql`

**Bad**

```go
rows, err := db.QueryContext(ctx, q)
if err != nil {
	return err
}
for rows.Next() {
	if err := rows.Scan(&item.ID); err != nil {
		return err
	}
}
```

Why it bites: you leak resources if you never close rows, and you can miss iteration errors if you never check `Rows.Err`.

**Better**

```go
rows, err := db.QueryContext(ctx, q)
if err != nil {
	return err
}
defer rows.Close()

for rows.Next() {
	if err := rows.Scan(&item.ID); err != nil {
		return err
	}
}
return rows.Err()
```

Catch it early: integration tests, review.

Rule: with `Rows`, always think in the trio `Query`, `Close`, and `Err`.

#### 45. Using `Query` when `QueryRow` or `Exec` matches the contract better

**Bad**

```go
rows, err := db.QueryContext(ctx, `
	INSERT INTO audit_log(actor, action) VALUES (?, ?)
`, actor, action)
if err != nil {
	return err
}
defer rows.Close()
```

Why it bites: `Query` implies result-set handling. For statements that do not return rows, it makes the contract noisier and easier to misuse.

**Better**

```go
_, err := db.ExecContext(ctx, `
	INSERT INTO audit_log(actor, action) VALUES (?, ?)
`, actor, action)
return err
```

Catch it early: review.

Rule: choose the narrowest `database/sql` API that matches the result shape.

#### 46. `json.Decoder` silently accepting unknown fields when strictness was intended

**Bad**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
if err := dec.Decode(&req); err != nil {
	return err
}
```

Why it bites: typoed or unexpected fields are ignored by default. Clients think they sent data the server never used.

**Better**

```go
var req CreateUserRequest

dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
if err := dec.Decode(&req); err != nil {
	return err
}
```

Catch it early: request-validation tests, review.

Rule: if input contracts are strict, make the decoder strict too.

#### 47. `time.Parse` layout mistakes from not using the reference time correctly

**Bad**

```go
ts, err := time.Parse("YYYY-MM-DD", rawDate)
if err != nil {
	return err
}
use(ts)
```

Why it bites: Go layouts are not symbolic format strings. They must be written using the reference time.

**Better**

```go
ts, err := time.Parse("2006-01-02", rawDate)
if err != nil {
	return err
}
use(ts)
```

Catch it early: tests, review.

Rule: in Go time layouts, the magic date is the format language.

#### 48. Assuming `os/exec.CommandContext` replaces explicit `Wait`

**Bad**

```go
cmd := exec.CommandContext(ctx, "worker", "--once")
if err := cmd.Start(); err != nil {
	return err
}

<-ctx.Done()
return ctx.Err()
```

Why it bites: `CommandContext` ties cancellation to the process, but if you call `Start`, you still need `Wait` to release resources and wait for I/O copying to finish.

**Better**

```go
cmd := exec.CommandContext(ctx, "worker", "--once")
if err := cmd.Start(); err != nil {
	return err
}

if err := cmd.Wait(); err != nil {
	return err
}
return nil
```

Catch it early: tests, review.

Rule: `CommandContext` handles cancellation policy, not `Wait` ownership.

#### 49. Using `context.Value` for optional parameters or config

**Bad**

```go
func ListUsers(ctx context.Context) error {
	limit, _ := ctx.Value("limit").(int)
	return query(limit)
}
```

Why it bites: the API contract becomes hidden, weakly typed, and hard to grep or validate.

**Better**

```go
type ListUsersOptions struct {
	Limit int
}

func ListUsers(ctx context.Context, opt ListUsersOptions) error {
	return query(opt.Limit)
}
```

Catch it early: review.

Rule: use `context.Value` for request-scoped metadata, not ordinary function parameters.

#### 50. Using `sync.Map` as the default map choice instead of a specialized tool

**Bad**

```go
var users sync.Map

func SetUser(id string, u User) {
	users.Store(id, u)
}
```

Why it bites: `sync.Map` is specialized. Most code loses type safety and invariant clarity when it reaches for it by default.

**Better**

```go
type UserStore struct {
	mu    sync.Mutex
	users map[string]User
}

func (s *UserStore) SetUser(id string, u User) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.users[id] = u
}
```

Catch it early: review, benchmarks when contention is the reason.

Rule: start with `map + mutex`; reach for `sync.Map` only when its specialization actually matches the workload.

## Practical takeaway

Most painful Go bugs are not exotic.

They come from small misunderstandings at API boundaries, ownership boundaries, and evaluation boundaries. That is exactly what this appendix is meant to tighten.
