---
title: Layout, Padding, and False Sharing
description: Struct layout, alignment, and cache-line behavior for engineers who want to eliminate hidden contention.
---

# Layout, Padding, and False Sharing

Correct concurrent code can still perform badly because of memory layout.

Alignment rules, padding, and cache-line sharing can turn "independent" counters or shards into a coherence fight.

:::tip Quick takeaway
Go's alignment rules are simple enough to inspect with `unsafe`, but CPU cache behavior is where the real surprises live. A correct atomic update can still be slow if two hot fields bounce the same line between cores.
:::

## Start with what the compiler guarantees

Go lays out struct fields in declaration order, adding padding so each field satisfies its alignment.

You can inspect the result directly:

```go
type Bad struct {
	Flag  byte
	Count uint64
	Tag   byte
}

fmt.Println(unsafe.Sizeof(Bad{}))
fmt.Println(unsafe.Offsetof(Bad{}.Count))
fmt.Println(unsafe.Alignof(Bad{}.Count))
```

This is the first layer. It explains wasted bytes.

The second layer is cache lines, and that explains wasted throughput.

## Reordering fields can shrink objects

```go
type Better struct {
	Count uint64
	Flag  byte
	Tag   byte
}
```

Field order changes:

- total size,
- offsets,
- how many objects fit in cache,
- sometimes whether a hot field shares a line with another hot field.

Do not treat field order as cosmetic in performance-sensitive code.

## False sharing is a coherence problem

False sharing happens when two goroutines update different data that lives on the same cache line.

No race is required.

No lock bug is required.

The program can be perfectly correct and still stall on cache invalidation traffic.

## Simplified counter sketch

```go
type Counters struct {
	OK  atomic.Int64
	Err atomic.Int64
}
```

If two cores hammer those fields independently and they land on the same line, every write can force line ownership traffic even though the counters are logically unrelated.

## Padding is sometimes the right tool

```go
type PaddedCounter struct {
	Value atomic.Int64
	_     [56]byte // example padding for a 64-byte line minus the counter
}
```

This is intentionally blunt. You do it only when:

- the field is genuinely hot,
- the contention is real,
- profiling or benchmarking says the padding wins.

## A real standard-library clue: `sync.Pool`

`sync.Pool` uses padding in `poolLocal` precisely to avoid widespread false sharing across P-local shards.

That is a strong signal from the standard library: layout is not an academic detail.

## Failure pattern

This looks harmless and can still waste CPU:

```go
type Shard struct {
	Hits atomic.Int64
}

var shards [2]Shard

go func() {
	for {
		shards[0].Hits.Add(1)
	}
}()

go func() {
	for {
		shards[1].Hits.Add(1)
	}
}()
```

The fields are independent. The cache line may not be. If these land next to each other, throughput can collapse for reasons that do not show up in race detection.

## How to study layout without guessing

Use a combination of:

- `unsafe.Sizeof`,
- `unsafe.Alignof`,
- `unsafe.Offsetof`,
- focused benchmarks,
- CPU profiles or hardware-counter tooling when available.

For Go-only investigation, start with size/offset inspection and benchmark deltas before reaching for heroic low-level tools.

## Source walk

Start here:

- [unsafe package docs](https://pkg.go.dev/unsafe)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)
- [sync/pool.go](https://github.com/golang/go/blob/go1.26.0/src/sync/pool.go)

## Practical takeaway

When a concurrent design is logically clean but still burns CPU, do not stop at mutexes and goroutines.

Inspect the bytes.
