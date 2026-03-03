---
title: Compiler and Toolchain Internals
description: Escape analysis, SSA, and Plan 9 assembly for engineers who want to see how Go source becomes machine code.
---

# Compiler and Toolchain Internals

If runtime internals explain why goroutines and channels work, compiler internals explain why a seemingly harmless source change can add heap churn, block inlining, or change dispatch behavior.

:::tip Quick takeaway
The compiler is not just a code printer. It decides stack vs heap placement, simplifies high-level constructs, lowers them into SSA, and hands a semi-abstract instruction stream to the assembler.
:::

## The shortest mental model

The modern Go compiler roughly moves through these phases:

1. parsing,
2. type checking,
3. compiler IR construction,
4. inlining, devirtualization, and escape analysis,
5. walk/lowering of higher-level constructs,
6. SSA construction and optimization,
7. machine-specific lowering and final object code emission.

That ordering matters. For example, escape analysis runs before SSA code generation, so by the time you look at assembly, many stack-vs-heap decisions are already made.

## Escape analysis is about lifetime and reachability

The most important misconception is that "local variable" means "stack variable."

Not true.

What matters is whether the value can outlive the current frame or whether the compiler cannot prove that it does not.

```go
type User struct {
	ID   int
	Name string
}

func NewUser() *User {
	u := User{ID: 7, Name: "ops"}
	return &u // escapes: returned beyond the frame
}
```

The variable is syntactically local. The pointee is not frame-local from the compiler's point of view, so it must move to the heap.

## A practical command loop

The fastest way to build intuition is to inspect the compiler directly:

```bash
go build -gcflags='-m=2' ./...
GOSSAFUNC=HandleBatch go build ./...
go build -gcflags='-S' ./...
```

- `-m=2` shows escape and inlining decisions.
- `GOSSAFUNC` writes `ssa.html` for one function.
- `-S` shows assembly-like output after lowering.

Treat these as one workflow, not three unrelated tricks.

## Simplified escape sketch

This pattern forces the backing array to survive beyond the frame:

```go
func HeaderBytes() []byte {
	buf := [64]byte{}
	return buf[:] // the returned slice needs storage that outlives the frame
}
```

The compiler may keep the slice header on the stack, but the array storage must be reachable after return, so the data moves to the heap.

## SSA is where optimizations become visible

SSA, or Static Single Assignment form, gives the compiler a representation where each logical value is assigned once and transformed through explicit dataflow.

That makes many optimizations easier:

- dead code elimination,
- bounds check elimination,
- nil check elimination,
- constant propagation,
- devirtualization,
- machine-specific rewrites.

## Simplified SSA-oriented sketch

```go
func Sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}
```

At the source level this is tiny. In SSA it becomes explicit blocks, phi values for loop-carried state, bounds reasoning, and eventually architecture-specific operations. That is why `GOSSAFUNC=Sum` is so useful: it lets you see the control-flow graph the optimizer actually reasons about.

## Plan 9 assembly is a toolchain interface, not a raw ISA dump

Go's assembly syntax is based on the Plan 9 family, but the crucial point is deeper than syntax.

The assembler works on a semi-abstract instruction set that fits the Go toolchain. It is not a 1:1 mirror of whatever `objdump` prints after linking.

```asm
TEXT ·add64(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), AX
	ADDQ b+8(FP), AX
	MOVQ AX, ret+16(FP)
	RET
```

What matters here:

- `SB` names package-level symbols,
- `FP` addresses arguments/results in the virtual frame,
- `NOSPLIT` changes stack-growth behavior and must be used with real care,
- the toolchain may still rewrite or annotate surrounding instructions.

## Why this matters for runtime-heavy code

Compiler behavior leaks upward into performance work:

- more escaping means more heap traffic,
- less inlining can mean more interface or closure overhead,
- missed devirtualization keeps indirect calls in hot loops,
- assembly boundaries can block some optimizations entirely.

That is why a production Go engineer should care about compiler output even if they never write assembly by hand.

## Failure pattern

A common mistake is assuming the source shape alone determines stack allocation:

```go
func ParseHeader() []byte {
	header := [32]byte{}
	return header[:]
}
```

This "looks stacky" but is not. The returned slice forces the backing storage to outlive the frame. If you treat escape reports as compiler noise, you will miss heap pressure that shows up later as GC cost.

## Source walk

Start here:

- [cmd/compile/README.md](https://github.com/golang/go/blob/go1.26.0/src/cmd/compile/README.md)
- [cmd/compile/internal/escape](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/escape)
- [cmd/compile/internal/ssa](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/ssa)
- [cmd/compile/internal/ssagen](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/ssagen)
- [A Quick Guide to Go's Assembler](https://go.dev/doc/asm)

## Practical takeaway

If you want to understand why Go code allocates, inlines, or dispatches the way it does, compiler internals are not optional trivia.

They are the layer that turns source-level intent into runtime-level cost.
