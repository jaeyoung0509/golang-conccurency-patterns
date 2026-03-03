---
title: Standard Library Anatomy
description: sync.Pool internals, reflection cost, and the runtime tradeoffs hiding inside familiar packages.
---

# Standard Library Anatomy

The standard library is one of the best places to study how the Go team itself balances abstraction, cache locality, GC pressure, and concurrency safety.

Two especially useful case studies are `sync.Pool` and `reflect`.

:::tip Quick takeaway
`sync.Pool` is not a global bag with a lock. It is a per-P design with stealing and victim caches. `reflect` is not "slow because magic." It is slower because it carries runtime type metadata, indirection, checks, and often extra copying.
:::

## `sync.Pool`: per-P first, shared second

`sync.Pool` is designed to reduce allocation pressure for temporary objects shared across many concurrent callers.

Its critical design choices are:

- one local shard per P,
- a `private` slot for the fastest reuse,
- a `shared` chain for additional values,
- cross-P stealing,
- a victim cache that lets values survive one GC cycle before aging out.

## Simplified pool sketch

```go
var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func encode(v any) []byte {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	_ = json.NewEncoder(buf).Encode(v)
	return append([]byte(nil), buf.Bytes()...)
}
```

The important nuance is that pooled objects are temporary scratch, not durable ownership.

## Why the pool is padded

`sync.Pool`'s `poolLocal` includes explicit padding to avoid false sharing between P-local shards.

That is a concrete example of the standard library optimizing for multicore cache behavior, not just algorithmic big-O.

## Reflection cost is mostly explicit runtime work

`reflect.Value` carries:

- type metadata,
- pointer/data representation,
- flag bits describing addressability, indirection, method values, and read-only state.

Every reflective operation may involve:

- kind checks,
- metadata traversal,
- interface packing/unpacking,
- indirect loads,
- extra copying when values must be materialized safely.

None of that is mysterious. It is simply not free.

## Simplified reflection sketch

```go
func nonZeroFields(v any) []string {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	out := make([]string, 0, rv.NumField())

	for i := 0; i < rv.NumField(); i++ {
		if !rv.Field(i).IsZero() {
			out = append(out, rt.Field(i).Name)
		}
	}
	return out
}
```

This is fine for control planes, admin tools, configuration, or low-rate paths.

It is a very different proposition in a tight per-record hot loop.

## A practical alternative: typed code or generation

Often the fastest replacement for reflection is not clever unsafe code.

It is one of:

- concrete hand-written typed code,
- generic code where compile-time shape is enough,
- generated code for serializers, validators, or mappers.

The point is to shift work from runtime inspection to compile-time structure.

## Failure pattern

These are both common mistakes:

```go
var pool sync.Pool

func storeHuge(x *BigBuffer) {
	pool.Put(x) // bad: using Pool like a permanent cache
}
```

```go
func hot(items []any) {
	for _, item := range items {
		_ = reflect.TypeOf(item).Kind() // bad: metadata path in a tiny hot loop
	}
}
```

The first mistakes lifetime and cache semantics.

The second mistakes flexibility for zero-cost abstraction.

## Source walk

Start here:

- [sync/pool.go](https://github.com/golang/go/blob/go1.26.0/src/sync/pool.go)
- [sync/poolqueue.go](https://github.com/golang/go/blob/go1.26.0/src/sync/poolqueue.go)
- [reflect/value.go](https://github.com/golang/go/blob/go1.26.0/src/reflect/value.go)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)

## Practical takeaway

When you read the standard library closely, you see the same pattern over and over:

- keep the common path local,
- make lifetime explicit,
- pay metadata cost only when flexibility justifies it.

That is as much a production design rule as it is a library implementation detail.
