---
title: Interfaces and Types Pitfalls
description: Ten interface and type-system mistakes that compile cleanly and still surprise Go engineers.
---

# Interfaces and Types Pitfalls

## 11. Typed nil inside an interface not comparing equal to `nil`

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

## 12. Comparing interface values whose concrete values are not comparable

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

## 13. Pointer-vs-value receiver confusion in interface satisfaction

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

## 14. Nil receiver methods that still panic internally

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

## 15. Implicit interface satisfaction assumed without checking method set shape

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

## 16. Unchecked type assertions

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

## 17. Using `==` on wrapped errors instead of `errors.Is` / `errors.As`

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

## 18. Assuming `any` preserves intended domain shape

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

## 19. `json.Unmarshal` into `interface{}` producing `float64` for numbers

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

## 20. Interface fields hiding zero-value vs absent-state distinctions

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
