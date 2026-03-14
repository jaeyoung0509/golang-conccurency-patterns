---
title: Collections and Memory Pitfalls
description: Ten collection and memory-shape mistakes that create subtle bugs and unnecessary allocations.
---

# Collections and Memory Pitfalls

## 21. Writing to a nil map

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

## 22. Assuming nil slice and empty slice are interchangeable in APIs or JSON

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

## 23. Assuming nil map and empty map are interchangeable in APIs or JSON

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

## 24. `append` aliasing through shared backing arrays

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

## 25. Subslices retaining huge backing arrays

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

## 26. Taking the address of a `range` variable

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

## 27. Mutating a slice while ranging over it

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

## 28. Assuming map iteration order is stable

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

## 29. Concurrent map access without synchronization

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

## 30. Assuming `string` and `[]byte` conversion is free or shares storage safely

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
