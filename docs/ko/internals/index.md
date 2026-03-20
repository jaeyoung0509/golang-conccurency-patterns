---
title: 내부 구조 개요
description: 컴파일러, 할당기, 타입 메타데이터, unsafe 경계, 성능 도구까지 다루는 시스템 레벨 Go 내부 구조.
---

# 내부 구조 개요

<div class="lead-panel">
  <p>
    이 섹션은 "동시성을 어떻게 쓰는가"에서 한 단계 더 내려가
    <strong>컴파일러, 런타임, 할당기, 타입 시스템이 왜 특정 설계를 빠르게 혹은 위험하게 만드는가</strong>를 다룹니다.
  </p>
</div>

## 이 섹션이 다루는 것

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/ko/internals/compiler-and-toolchain">컴파일러와 툴체인</a></h3>
    <p>Escape analysis, SSA, Plan 9 assembly를 통해 소스 코드가 실제로 어떤 형태로 빌드되는지 봅니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/allocator-and-write-barrier">할당기와 하이브리드 write barrier</a></h3>
    <p>`mcache`, `mcentral`, `mheap`, `mspan`과 concurrent GC correctness를 지키는 barrier를 따라갑니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/layout-padding-false-sharing">레이아웃, 패딩, false sharing</a></h3>
    <p>정렬 규칙, struct layout, cache line이 알고리즘과는 별개로 성능을 어떻게 바꾸는지 봅니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/generics-and-interfaces">제네릭과 인터페이스</a></h3>
    <p>GC shape stenciling, dictionary, `eface`/`iface` 계열 레이아웃, dynamic dispatch tradeoff를 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/unsafe-cgo-pinner">unsafe, cgo, 그리고 Pinner</a></h3>
    <p>raw pointer 규칙, 스케줄러 handoff, memory pinning, zero-copy 기법의 실제 경계를 다룹니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/modern-performance-tuning">현대 Go 성능 튜닝</a></h3>
    <p>PGO, execution trace, flight recorder, zero-copy I/O를 최신 Go 기준으로 정리합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/kernel-io-paths">Kernel I/O Paths</a></h3>
    <p>`pollDesc`, readiness wait, epoll/kqueue wakeup, `sendfile`/`splice` fast path를 runtime boundary 관점에서 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/internals/stdlib-anatomy">표준 라이브러리 해부</a></h3>
    <p>`sync.Pool`이 왜 잘 확장되는지, `reflect`가 왜 느린지, code generation이 언제 더 나은지 봅니다.</p>
  </div>
</div>

## 추천 읽기 순서

1. [컴파일러와 툴체인](/ko/internals/compiler-and-toolchain)부터 읽습니다.
2. 이어서 [할당기와 하이브리드 write barrier](/ko/internals/allocator-and-write-barrier)를 봅니다.
3. [레이아웃, 패딩, false sharing](/ko/internals/layout-padding-false-sharing)으로 메모리 레벨 비용을 정리합니다.
4. [제네릭과 인터페이스](/ko/internals/generics-and-interfaces)로 타입 시스템 내부를 봅니다.
5. 그 다음 [unsafe, cgo, 그리고 Pinner](/ko/internals/unsafe-cgo-pinner)를 읽습니다.
6. 네트워킹과 fast-copy의 OS-facing view가 필요하면 [Kernel I/O Paths: netpoll, epoll/kqueue, 그리고 Zero-Copy](/ko/internals/kernel-io-paths)를 읽습니다.
7. 마지막으로 [현대 Go 성능 튜닝](/ko/internals/modern-performance-tuning)과 [표준 라이브러리 해부](/ko/internals/stdlib-anatomy)를 마무리합니다.

## 이 섹션을 읽고 답할 수 있어야 하는 질문

- 왜 지역 변수처럼 보여도 힙으로 가는 값이 있는가?
- `GOSSAFUNC`는 벤치마크만으로는 보이지 않는 무엇을 보여주는가?
- 왜 할당기 계층이 `mcache -> mcentral -> mheap`인가?
- 왜 concurrent GC는 hybrid write barrier를 필요로 하는가?
- race도 없고 lock bug도 없는데 왜 cache line 때문에 느려질 수 있는가?
- 왜 Go 제네릭은 순수 monomorphization도 아니고 type erasure도 아닌가?
- 왜 `uintptr`는 GC root가 아닌가?
- 왜 ready socket이 곧 빠른 application throughput을 의미하지 않는가?
- 언제 `io.Copy`가 zero-copy fast path를 타고, 언제 조용히 fallback 하는가?
- 왜 `sync.Pool`은 per-P shard에 padding을 넣는가?
- 언제 reflection은 충분히 유연하고, 언제 cost model 자체가 틀린 선택이 되는가?
