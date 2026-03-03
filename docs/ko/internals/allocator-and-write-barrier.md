---
title: 할당기와 하이브리드 write barrier
description: Go 할당기 계층과 concurrent GC correctness를 지키는 hybrid write barrier를 따라갑니다.
---

# 할당기와 하이브리드 write barrier

"allocation"을 아직도 하나의 뭉뚱그린 비용으로 생각하고 있다면, 프로덕션 Go 성능을 제대로 해석하기 어렵습니다.

작은 객체 할당, span 재사용, sweeping, 큰 객체 경로, write barrier는 concurrency와 GC latency 안에서 함께 움직입니다.

:::tip Quick takeaway
Go 할당기는 여전히 TCMalloc의 영향을 받았지만, 실전에서 필요한 모델은 더 단순합니다. `mcache`가 per-P fast path이고, `mcentral`이 size class별 공유 계층이며, `mheap`이 page를 관리하고, hybrid write barrier가 concurrent marking correctness를 지킵니다.
:::

## 할당기 계층

`malloc.go`의 런타임 주석이 가장 좋은 요약입니다.

- `mcache`: free slot이 남은 span을 가진 per-P cache
- `mcentral`: 한 size class를 위한 shared span 집합
- `mheap`: page 단위 heap manager
- `mspan`: 객체를 담는 page run

```mermaid
flowchart LR
    A["goroutine allocates"] --> B["P-local mcache"]
    B --> C["size-class mspan"]
    C --> D["shared mcentral"]
    D --> E["global mheap"]
    E --> F["OS pages"]
```

핵심 목표는 분명합니다. common path는 lock-free에 가깝게 두고, 공유 coordination 비용은 local cache가 비었을 때만 치르게 만드는 것입니다.

## 작은 할당과 size class

작은 객체는 요청 크기를 size class로 반올림한 뒤 현재 P의 cache에서 처리하려고 시도합니다.

이 덕분에 짧게 사는 작은 객체는 개별적으로 보면 꽤 싸게 보입니다.

하지만 GC, pointer scan, goroutine fan-out이 얽히면 "작다"는 사실이 비용을 없애주지는 않습니다.

## 단순화한 할당 스케치

```go
func alloc(size uintptr) unsafe.Pointer {
	if span := mcache.lookup(sizeClass(size)); span.hasFree() {
		return span.alloc()
	}
	if span := mcentral.refill(sizeClass(size)); span != nil {
		mcache.install(span)
		return span.alloc()
	}
	span := mheap.allocPages(...)
	mcentral.install(span)
	mcache.install(span)
	return span.alloc()
}
```

실제 runtime 코드는 아니지만, 머릿속 모델로는 이게 맞습니다.

## 큰 객체는 다른 경로를 탄다

큰 할당은 small-object machinery를 상당 부분 우회하고 page allocator에 더 직접적으로 닿습니다.

그래서 "요청마다 64 KiB buffer를 10개 goroutine에서 새로 만든다" 같은 패턴은 작은 struct churn과 질적으로 다른 문제입니다.

## 왜 write barrier를 같이 봐야 하나

할당기와 GC는 분리된 이야기가 아닙니다.

concurrent marking 중에는 mutator가 pointer를 바꾸는 동안 reachable object를 collector에게서 숨기지 못하게 해야 합니다.

그 역할이 write barrier입니다.

## 하이브리드 write barrier

`mbarrier.go`의 런타임 주석을 단순화하면 이렇습니다.

```go
// conceptual, not the actual runtime function
func writePointer(slot *unsafe.Pointer, ptr unsafe.Pointer) {
	shade(*slot)
	if currentStackIsGrey() {
		shade(ptr)
	}
	*slot = ptr
}
```

각 부분의 의미는 분명합니다.

- deletion 쪽은 old referent를 보호하고,
- insertion 쪽은 stack이 아직 grey일 때 new referent를 보호하고,
- stack이 black이 되면 그 stack에 대해서는 insertion 쪽이 더 이상 필요하지 않습니다.

## 왜 메모리 ordering이 중요한가

Go는 destination object 색을 fast path에서 매번 확인하는 조건부 barrier를 쓰지 않습니다. 그렇게 하면 memory-ordering 비용이 너무 커지기 때문입니다.

작은 벤치마크에서는 약간 낭비처럼 보여도, 실제 concurrent collector에서는 correctness와 전체 비용 구조가 더 중요합니다.

## 애플리케이션 코드에 주는 의미

보통의 Go 코드에서 barrier를 직접 쓰지는 않습니다.

하지만 다음 상황에서는 그 결과를 분명히 체감합니다.

- pointer-rich value를 고속으로 복사할 때
- 큰 shared graph를 계속 갱신할 때
- reflection이나 `unsafe` helper가 결국 typed memmove로 흘러갈 때
- GC assist가 request latency 안에 들어오는 workload를 만들 때

## 실패 패턴

다음 코드는 그저 "allocation-heavy"처럼 보이지만, 실제로는 GC와 allocator 문제를 함께 만듭니다.

```go
for _, req := range batch {
	go func(req Request) {
		buf := make([]byte, 64<<10) // large allocation path
		_ = handle(req, buf)
	}(req)
}
```

문제는 단순히 바이트 수가 아닙니다.

- 큰 객체는 가장 싼 경로를 타지 않고,
- fan-out이 pressure를 배수로 늘리고,
- collector와 allocator가 tail latency의 일부가 됩니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [runtime/malloc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/malloc.go)
- [runtime/mcache.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mcache.go)
- [runtime/mcentral.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mcentral.go)
- [runtime/mheap.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mheap.go)
- [runtime/mbarrier.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mbarrier.go)
- [Proposal: eliminate rescan with a hybrid write barrier](https://github.com/golang/proposal/blob/master/design/17503-eliminate-rescan.md)

## 실전 요약

Go에서 가장 싼 allocation은 "공짜"가 아니라 "정교하게 설계된 fast path"입니다.

할당기 계층과 hybrid barrier를 이해하면, heap churn을 막연한 냄새가 아니라 구체적인 런타임 예산으로 보게 됩니다.
