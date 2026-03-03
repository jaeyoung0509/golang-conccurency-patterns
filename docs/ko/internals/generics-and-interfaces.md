---
title: 제네릭과 인터페이스 내부 구조
description: Go가 제네릭과 인터페이스 값을 어떻게 구현하는지, 그리고 그것이 dispatch와 성능에 왜 중요한지 설명합니다.
---

# 제네릭과 인터페이스 내부 구조

제네릭 이후 Go 타입 시스템은 훨씬 강해졌지만, 런타임 모델은 여전히 매우 실용주의적입니다.

핵심은 Go가 무엇을 선택하지 않았는지 이해하는 것입니다.

- 순수한 C++식 full monomorphization도 아니고,
- Java식 type erasure도 아니고,
- "인터페이스 호출은 알아서 다 최적화된다"는 모델도 아닙니다.

:::tip Quick takeaway
제네릭은 GC-shape 기반 code sharing과 dictionary를 함께 쓰고, 인터페이스는 여전히 type metadata + data라는 두 단어 모델이 핵심입니다. non-empty interface에서는 `itab`이 그 연결고리 역할을 합니다.
:::

## 제네릭: GC shape와 dictionary

실전에서 가장 유용한 요약은 이렇습니다.

- 같은 중요한 메모리 레이아웃을 공유하는 instantiation은 코드를 공유할 수 있고,
- type-specific operation이나 metadata가 필요할 때는 dictionary를 쓰며,
- 결과적으로 code size 폭증과 pure erasure 사이의 절충점을 택합니다.

그래서 Go 제네릭은 C++ template보다 더 공유적으로 보일 수 있고, 동시에 런타임이 필요한 연산은 그대로 지원할 수 있습니다.

## 단순화한 generic 예시

```go
func Clone[T any](in []T) []T {
	out := make([]T, len(in))
	copy(out, in)
	return out
}
```

개념적으로는 `T`의 GC shape를 기준으로 코드를 공유하고, 필요한 경우 dictionary 정보를 추가로 넘긴다고 이해하면 됩니다. 실제 emitted form은 컴파일러의 영역이지만, 머릿속 모델은 "가능한 한 공유하고, 필요할 때만 specialize"입니다.

## 인터페이스는 여전히 metadata + data다

정확한 내부 타입 이름은 패키지마다 조금씩 다르지만, 이 개념 모델은 계속 유효합니다.

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

- empty interface는 direct type metadata를 들고 있고,
- non-empty interface는 concrete type과 interface method set을 잇는 `itab`을 쓰고,
- 값은 type 특성에 따라 direct 또는 indirect로 저장됩니다.

이 마지막 점은 copy 비용, boxing, escape behavior와 직접 연결됩니다.

## 왜 dynamic dispatch 비용이 남는가

인터페이스 호출은 컴파일러가 더 많은 사실을 증명하지 못하면 간접 호출로 남습니다.

그 증거는 다음에서 올 수 있습니다.

- 정적인 타입 정보,
- devirtualization,
- inlining 기회,
- 최신 툴체인에서는 profile-guided 정보.

따라서 아주 작은 hot loop를 넓은 인터페이스 위에 얹고 concrete call처럼 최적화되길 기대하면, 컴파일러에게 없는 증거를 요구하는 셈이 됩니다.

## 단순화한 hot-path 예시

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

이건 우아한 API입니다. 하지만 자동으로 가장 싼 dispatch shape가 되지는 않습니다.

## 왜 제네릭과 인터페이스는 비슷하지만 다른가

제네릭은 type information을 compile-time에 더 오래 유지하게 해 줍니다.

인터페이스는 behavior abstraction을 runtime 정책으로 넘깁니다.

실전에서는:

- 알고리즘 구조는 같고 concrete type만 바뀌면 제네릭이 유리하고,
- 구현 선택 자체가 runtime policy면 인터페이스가 유리하고,
- 둘을 섞는 것은 강력하지만 각자의 cost model을 없애주지는 않습니다.

## 실패 패턴

다음 코드는 인터페이스 dispatch가 거의 공짜라고 생각할 때 자주 나옵니다.

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

이 API가 정확히 맞을 때도 있습니다.

하지만 때로는 프로세스 전체에서 가장 뜨거운 indirect call site가 됩니다. 인터페이스를 쓴 것이 문제라기보다, 추상화 수준과 dispatch 비용이 무관하다고 가정한 것이 문제입니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [Proposal: dictionaries for Go 1.18 generics](https://go.googlesource.com/proposal/+/master/design/generics-implementation-dictionaries-go1.18.md)
- [Proposal: stenciling for Go 1.18 generics](https://go.googlesource.com/proposal/+/master/design/generics-implementation-stenciling.md)
- [runtime/iface.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/iface.go)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)
- [reflect/value.go](https://github.com/golang/go/blob/go1.26.0/src/reflect/value.go)
- [cmd/compile/internal/devirtualize](https://github.com/golang/go/tree/go1.26.0/src/cmd/compile/internal/devirtualize)

## 실전 요약

Go의 타입 내부 구조는 이념적이지 않습니다.

code size, runtime metadata, optimizer opportunity 사이에서 현실적인 절충을 합니다. 이 절충을 이해하면 제네릭, 인터페이스, concrete code를 훨씬 더 의도적으로 고르게 됩니다.
