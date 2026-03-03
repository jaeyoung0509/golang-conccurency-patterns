---
title: Generics and Interface Internals
description: How Go implements generics and interface values, and why that matters for dispatch and performance.
---

# Generics and Interface Internals

Go's type system now does much more than the pre-generics language did, but the runtime model is still intentionally pragmatic.

The key is to understand what Go did **not** choose:

- not pure C++-style full monomorphization for everything,
- not Java-style type erasure,
- not a magical "all interface calls are optimized away" model.

:::tip Quick takeaway
For generics, Go uses GC-shape-based code sharing plus dictionaries where type-specific operations are needed. For interfaces, the mental model is still two words: type metadata plus data, with an `itab` in the non-empty-interface case.
:::

## Generics: GC shape plus dictionaries

The most useful expert-level summary is:

- code may be shared across instantiations with the same relevant memory layout,
- dictionaries carry type-specific operations and metadata where needed,
- the result is a middle ground between code-size explosion and pure erasure.

This is why generic code can sometimes behave "more shared" than C++ templates and still preserve the operations Go needs at runtime.

## Simplified generic sketch

```go
func Clone[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	return out
}
```

Conceptually, the compiler may generate code around a GC shape for `T` and pass enough dictionary information to handle type-specific needs. The exact emitted form is compiler territory, but the right mental model is "shared where possible, specialized where necessary."

## Interfaces are still metadata plus data

The exact internal type names move between packages, but this conceptual model remains useful:

```go
type eface struct {
	typ  *abi.Type
	data unsafe.Pointer
}

type iface struct {
	tab  *itab
	data unsafe.Pointer
}
```

- empty interfaces carry direct type metadata,
- non-empty interfaces use an `itab` that connects a concrete type to an interface's method set,
- data may be stored directly or indirectly depending on the type.

That last point matters for copies, boxing cost, and escape behavior.

## Why dynamic dispatch sometimes stays expensive

Interface calls are indirect calls unless the compiler can prove more.

That proof may come from:

- static knowledge,
- devirtualization,
- inlining opportunities,
- sometimes profile-guided information in modern toolchains.

But if you write a tiny hot loop over a broad interface and expect it to optimize like a monomorphic concrete call, you are asking the compiler for evidence it may not have.

## Simplified hot-path sketch

```go
type Hasher interface {
	Hash([]byte) uint64
}

func SumHashes(h Hasher, xs [][]byte) uint64 {
	var total uint64
	for _, x := range xs {
		total += h.Hash(x)
	}
	return total
}
```

This is an elegant API shape. It is not automatically the cheapest possible dispatch shape.

## Why generics and interfaces are related but not the same

Generics let you keep type information available at compile time for each instantiation site.

Interfaces let you abstract over behavior at runtime.

In practice:

- generics often help when the algorithm is shape-stable but the concrete type varies,
- interfaces help when the implementation choice is a runtime policy decision,
- mixing them is powerful, but it does not remove the cost model of either.

## Failure pattern

This kind of code is often written as if interface dispatch were free:

```go
type Matcher interface {
	Match([]byte) bool
}

func Count(ms []Matcher, payload []byte) int {
	n := 0
	for _, m := range ms {
		if m.Match(payload) {
			n++
		}
	}
	return n
}
```

Sometimes that API is exactly right.

Sometimes it becomes the hottest indirect-call site in the process. The mistake is not using interfaces. The mistake is assuming abstraction level and dispatch cost are unrelated.

## Source walk

Start here:

- [Proposal: dictionaries for Go 1.18 generics](https://go.googlesource.com/proposal/+/master/design/generics-implementation-dictionaries-go1.18.md)
- [Proposal: stenciling for Go 1.18 generics](https://go.googlesource.com/proposal/+/master/design/generics-implementation-stenciling.md)
- [runtime/iface.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/iface.go)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)
- [reflect/value.go](https://github.com/golang/go/blob/go1.26.0/src/reflect/value.go)
- [cmd/compile/internal/devirtualize](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/devirtualize)

## Practical takeaway

Go's type machinery is intentionally not ideological.

It trades code size, runtime metadata, and optimizer opportunity against each other. If you understand that trade, you choose between generics, interfaces, and concrete code much more deliberately.
