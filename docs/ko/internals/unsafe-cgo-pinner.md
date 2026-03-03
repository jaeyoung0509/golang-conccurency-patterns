---
title: unsafe, cgo, 그리고 runtime.Pinner
description: raw pointer 규칙, cgo 경계, memory pinning semantics를 시스템 레벨에서 설명합니다.
---

# unsafe, cgo, 그리고 runtime.Pinner

여기는 Go가 기본적으로 보호해 주는 범위를 넘어서는 경계입니다.

보상은 분명합니다. zero-copy 트릭, C interop, 직접적인 메모리 뷰.

대가도 분명합니다. 코드가 C처럼 보여도 GC, 스케줄러, pointer 규칙은 그대로 살아 있습니다.

:::tip Quick takeaway
`unsafe.Pointer`는 여전히 Go pointer world 안에 있습니다. `uintptr`는 그냥 정수입니다. cgo crossing은 scheduler model을 바꿉니다. `runtime.Pinner`는 C가 Go pointer를 안전하게 유지해야 할 때 그 주소를 명시적으로 고정하기 위해 존재합니다.
:::

## `unsafe.Pointer`와 `uintptr`는 같지 않다

전문가 수준에서 꼭 기억해야 할 규칙은 이렇습니다.

- `unsafe.Pointer`는 pointer semantics 안에 있고,
- `uintptr`는 객체를 살려 두지 않으며 GC에 아무 신호도 주지 않습니다.

Go pointer를 정수로 바꿔 놓고 런타임이 그 주소 유효성을 알아서 지켜주길 기대하면 계약 밖입니다.

## 최근 helper API가 더 안전한 이유

현대 Go는 예전의 `reflect.SliceHeader` 트릭보다 덜 깨지기 쉬운 helper를 제공합니다.

- `unsafe.String`
- `unsafe.StringData`
- `unsafe.Slice`
- `unsafe.SliceData`
- `unsafe.Add`

여전히 unsafe입니다. 다만 header-casting보다 의도가 더 명확하고 실수 여지가 적습니다.

## 단순화한 pinning 예시

```go
buf := make([]byte, 4096)

var p runtime.Pinner
p.Pin(&buf[0])
defer p.Unpin()

// 여기서 buf 주소를 C에 넘기거나, 잠시 retained 하게 만들 수 있다.
runtime.KeepAlive(buf)
```

중요한 것은 문법이 아니라 계약입니다.

- `Unpin` 전까지 객체 주소가 고정되고,
- 그 객체 안에 C가 따라갈 Go pointer가 더 있으면 별도로 pin 해야 하며,
- `Pinner` 자체도 C가 쓰는 동안 살아 있어야 합니다.

## 왜 cgo는 concurrency 모델을 바꿔 놓는가

cgo 호출은 그냥 함수 호출이 아닙니다.

다음에 영향을 줍니다.

- scheduler accounting
- stack transition과 pointer check
- runtime이 blocking을 얼마나 투명하게 관측할 수 있는지
- Go trace만 보고 latency를 해석하기 쉬운지 여부

pure Go network I/O는 netpoll과 잘 맞지만, opaque C 호출은 그렇지 않습니다.

## 실전 cgo 규칙 하나

경계를 물방울처럼 자주 넘지 말고, 덩어리로 넘겨야 합니다.

ownership contract가 분명한 큰 호출 한 번이, 레코드마다 C 경계를 넘나드는 설계보다 batching과 observability 측면에서 훨씬 낫습니다.

## 실패 패턴

이건 대표적인 버그입니다.

```go
func firstByteAddr(buf []byte) uintptr {
	return uintptr(unsafe.Pointer(&buf[0]))
}
```

`uintptr`로 바꾸는 순간, 그 값은 더 이상 런타임 입장에서 Go pointer가 아닙니다. 그 정수가 liveness나 pinning을 보존한다고 생각하면 틀립니다.

## zero-copy를 안전하게 생각하는 방법

unsafe helper로 zero-copy 변환을 할 때는:

- 원본 backing storage를 살아 있게 두고,
- ownership을 명확히 하고,
- mutable/immutable view의 lifetime 가정을 섞지 말고,
- "copy가 없다"를 "lifetime 관리도 필요 없다"로 오해하지 말아야 합니다.

## `runtime.Pinner`가 필요한 정확한 상황

`runtime.Pinner`는 아주 좁은 문제를 위해 존재합니다.

- cgo call이 끝난 뒤에도 C가 Go pointer를 계속 들고 있어야 하거나,
- Go object 안에 든 Go pointer를 C가 따라가야 하는 경우.

이건 ownership discipline을 버리라는 뜻이 아닙니다. 오히려 더 명시적으로 하라는 뜻입니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [unsafe package docs](https://pkg.go.dev/unsafe)
- [cmd/cgo package docs](https://pkg.go.dev/cmd/cgo)
- [runtime/pinner.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/pinner.go)
- [runtime/mbarrier.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mbarrier.go)

## 실전 요약

`unsafe`를 제대로 쓰는 방법은 런타임이 사라졌다고 믿는 것이 아닙니다.

어떤 런타임 보장에 기대고 있고, 어떤 보장 밖으로 나갔는지를 정확히 아는 것이 핵심입니다.
