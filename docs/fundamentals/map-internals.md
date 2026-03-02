---
title: Map Internals and Swiss Tables
description: Understand Go's modern map structure, Swiss Table groups, extendible hashing, and why concurrent writes are still unsafe.
---

# Map Internals and Swiss Tables

Maps are one of the most important shared structures in real Go systems:

- caches,
- in-memory indexes,
- deduplication tables,
- actor-owned state,
- request-scoped accumulators.

If you want expert-level Go fundamentals, you need a current map model, not an outdated one.

:::tip Quick takeaway
In Go 1.24 and later, the built-in map implementation is based on Swiss Tables with extendible hashing. That improves locality and growth behavior, but it does **not** make maps safe for unsynchronized concurrent mutation.
:::

## The modern structure

The Go 1.26 runtime delegates most map work to `internal/runtime/maps`.

Important terms from the runtime:

| Term | Meaning |
| --- | --- |
| Slot | one key/value storage location |
| Group | 8 slots plus one control word |
| Control word | metadata bytes describing empty, deleted, or occupied slots |
| Table | one Swiss Table hash table |
| Directory | top-level structure selecting which table to use |
| H1 / H2 | upper and lower hash portions used for table selection and in-group matching |

## Why Swiss Tables help

The key idea is that lookup can reject many non-matching slots cheaply.

Instead of checking one slot at a time, a group's control word lets the runtime compare several candidate slots in parallel using the 7-bit `H2` fingerprint.

That improves:

- locality,
- branch behavior,
- the amount of useful work done per probe step.

## Simplified internal sketch

This is the mental model, not the exact runtime code:

```go
func lookup(m *Map, key K) (V, bool) {
	hash := hashKey(key, m.seed)
	table := selectTable(m.directory, hash)
	seq := probeSeq(h1(hash), table.groupMask)

	for {
		group := table.groups[seq.offset]
		matches := group.matchH2(h2(hash))

		for slot := range matches {
			if group.key(slot) == key {
				return group.value(slot), true
			}
		}

		if group.hasEmptySlot() {
			return zero, false
		}

		seq = seq.next()
	}
}
```

The real runtime handles deleted slots, indirect keys and values, iteration semantics, grow state, and hash seeding.

## Growth is no longer "one giant whole-map rehash"

The old mental model many engineers still carry is "Go maps grow by redistributing buckets."

Modern Go maps still need reorganization, but the structure now uses tables and a directory so growth can happen at a smaller unit than "rehash the entire world at once."

That matters because:

- grow latency is easier to control,
- the cost model is better for larger maps,
- the runtime can keep iteration semantics intact while tables are replaced.

## Iteration is the hardest part

The comments in `internal/runtime/maps/map.go` make this explicit: iteration is the gnarliest part of map correctness.

Why?

- entries must not be returned twice,
- updated entries should expose their latest value,
- deleted entries must disappear,
- growth can happen during iteration,
- order is intentionally unspecified.

This is one of the reasons unsynchronized concurrent mutation is so dangerous: there is a lot of internal state moving behind a tiny public API.

## Why concurrent writes are still unsafe

The map runtime keeps write-state bits and race-detection hooks because concurrent mutation can corrupt invariants or create undefined visibility.

The right mental model is:

- map operations are optimized for single-threaded ownership or externally synchronized access,
- the runtime may detect some bad concurrent access patterns,
- detection is not a substitute for correctness.

If multiple goroutines need to mutate a map, choose one of these on purpose:

- `map + sync.Mutex`,
- `map + sync.RWMutex`,
- single-owner goroutine with a channel protocol,
- sharded ownership.

## Practical consequences in production

### A faster map does not remove contention

If many goroutines hammer one hot map behind a lock, the bottleneck can still be lock contention rather than lookup cost.

### Actor-owned maps can be cleaner than shared maps

When updates are protocol-heavy, a single owner goroutine can be easier to reason about than a mutex around a complex state machine.

### Large maps still affect GC

Even with a better table design, a large pointer-rich map changes heap size, scan work, and cache behavior.

## Runtime source walk

Start here:

- [runtime/map.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/map.go)
- [internal/runtime/maps/map.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/map.go)
- [internal/runtime/maps/table.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/table.go)
- [internal/runtime/maps/group.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/group.go)

The package comments in `internal/runtime/maps/map.go` are especially worth reading because they document terminology and iteration semantics directly.

## Related official reading

- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)

## Practical takeaway

Modern Go maps are more sophisticated than the old bucket-only mental model suggests.

That sophistication improves performance, but it also reinforces the core concurrency lesson: state ownership still matters.

If your mental model for a hot in-memory map is "it is probably fine," you are operating below the level the runtime actually requires.
