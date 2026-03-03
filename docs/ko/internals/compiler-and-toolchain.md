---
title: 컴파일러와 툴체인 내부 구조
description: Escape analysis, SSA, Plan 9 assembly를 통해 Go 소스가 머신 코드로 내려가는 과정을 이해합니다.
---

# 컴파일러와 툴체인 내부 구조

런타임 내부 구조가 goroutine과 channel이 왜 동작하는지 설명한다면, 컴파일러 내부 구조는 왜 사소해 보이는 소스 변경이 heap churn, inlining 실패, dispatch 비용으로 이어지는지 설명합니다.

:::tip Quick takeaway
컴파일러는 단순한 코드 출력기가 아닙니다. stack vs heap 배치, 고수준 문법 분해, SSA lowering, assembler에 넘길 semi-abstract instruction stream까지 모두 결정합니다.
:::

## 가장 짧은 기본 모델

현대 Go 컴파일러는 대략 이런 단계로 갑니다.

1. parsing
2. type checking
3. compiler IR 구성
4. inlining, devirtualization, escape analysis
5. walk/lowering
6. SSA 생성과 최적화
7. machine-specific lowering과 object code 생성

이 순서는 중요합니다. 예를 들어 escape analysis는 SSA 이전에 일어나므로, assembly를 보기 시작할 때쯤이면 stack vs heap 결정은 이미 상당 부분 끝난 상태입니다.

## Escape analysis는 lifetime 문제다

가장 흔한 오해는 "지역 변수면 stack"이라는 생각입니다.

그렇지 않습니다.

핵심은 그 값이 현재 frame 밖으로 나갈 수 있는지, 혹은 컴파일러가 그렇지 않음을 증명할 수 있는지입니다.

```go
type User struct {
	ID   int
	Name string
}

func NewUser() *User {
	u := User{ID: 7, Name: "ops"}
	return &u // escapes: frame 밖으로 반환됨
}
```

문법상 지역 변수이지만, pointee는 frame-local 하지 않습니다. 따라서 힙으로 이동해야 합니다.

## 실전에서 가장 유용한 명령 조합

직관을 빠르게 키우는 방법은 컴파일러 출력을 직접 보는 것입니다.

```bash
go build -gcflags='-m=2' ./...
GOSSAFUNC=HandleBatch go build ./...
go build -gcflags='-S' ./...
```

- `-m=2`는 escape/inlining 결정을 보여주고,
- `GOSSAFUNC`는 함수 하나의 `ssa.html`을 만들고,
- `-S`는 lowering 이후의 assembly 비슷한 출력을 보여줍니다.

이 셋은 따로 노는 요령이 아니라 하나의 분석 루프입니다.

## 단순화한 escape 예시

이 패턴은 backing array를 frame 밖으로 밀어냅니다.

```go
func HeaderBytes() []byte {
	buf := [64]byte{}
	return buf[:] // 반환된 slice는 frame 이후에도 살아 있어야 함
}
```

slice header 자체는 stack에 남을 수 있어도, 실제 데이터는 반환 이후에도 도달 가능해야 하므로 heap으로 갑니다.

## SSA는 최적화가 보이는 곳이다

SSA, 즉 Static Single Assignment form은 각 논리적 값이 한 번씩만 할당되는 표현입니다.

이 표현이 중요한 이유는 다음 최적화가 쉬워지기 때문입니다.

- dead code elimination
- bounds check elimination
- nil check elimination
- constant propagation
- devirtualization
- machine-specific rewrite

## 단순화한 SSA 관점 예시

```go
func Sum(xs []int) int {
	total := 0
	for _, x := range xs {
		total += x
	}
	return total
}
```

소스 수준에서는 작아 보이지만, SSA에서는 explicit block, loop-carried phi value, bounds reasoning, architecture-specific op로 바뀝니다. 그래서 `GOSSAFUNC=Sum`이 강력합니다. 컴파일러가 실제로 reasoning 하는 control-flow graph를 보여주기 때문입니다.

## Plan 9 assembly는 raw ISA dump가 아니다

Go assembly 문법은 Plan 9 계열이지만, 핵심은 문법보다 도구체인 인터페이스라는 점입니다.

Go assembler는 툴체인에 맞는 semi-abstract instruction set 위에서 동작합니다. 최종 링크 후 `objdump`가 보여주는 실제 기계어와 1:1 대응이라고 생각하면 안 됩니다.

```asm
TEXT ·add64(SB), NOSPLIT, $0-24
	MOVQ a+0(FP), AX
	ADDQ b+8(FP), AX
	MOVQ AX, ret+16(FP)
	RET
```

여기서 중요한 포인트는:

- `SB`는 package-level symbol,
- `FP`는 virtual frame에서 argument/result를 가리키고,
- `NOSPLIT`은 stack growth behavior를 바꾸며 매우 조심해서 써야 하고,
- 주변 instruction은 툴체인이 추가로 재작성하거나 annotate 할 수 있다는 점입니다.

## 왜 런타임 중심 코드에서 이게 중요한가

컴파일러 동작은 위로 새어 나옵니다.

- escape가 많아지면 heap traffic이 늘고,
- inlining이 막히면 interface/closure 비용이 남고,
- devirtualization을 놓치면 hot loop에 간접 호출이 남고,
- assembly 경계는 일부 최적화를 막기도 합니다.

그래서 assembly를 직접 쓰지 않더라도, 프로덕션 Go 엔지니어는 컴파일러 출력을 읽을 줄 알아야 합니다.

## 실패 패턴

다음 코드는 "stack 같아 보인다"는 직관을 쉽게 깨뜨립니다.

```go
func ParseHeader() []byte {
	header := [32]byte{}
	return header[:]
}
```

겉으로는 작고 지역적이지만, 반환되는 slice 때문에 backing storage는 heap으로 갑니다. escape report를 그냥 컴파일러 잡음으로 취급하면, 나중에 GC 비용으로 돌아올 힙 pressure를 놓치게 됩니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [cmd/compile/README.md](https://github.com/golang/go/blob/go1.26.0/src/cmd/compile/README.md)
- [cmd/compile/internal/escape](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/escape)
- [cmd/compile/internal/ssa](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/ssa)
- [cmd/compile/internal/ssagen](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/ssagen)
- [A Quick Guide to Go's Assembler](https://go.dev/doc/asm)

## 실전 요약

Go 코드가 왜 할당하고, 왜 inline되고, 왜 특정 방식으로 dispatch되는지 이해하고 싶다면 컴파일러 내부 구조는 선택 사항이 아닙니다.

이 레이어가 source-level intent를 runtime-level cost로 바꿉니다.
