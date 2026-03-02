---
title: 가비지 컬렉터와 Green Tea GC
description: Go의 concurrent GC, pacer, assist, scavenger, 그리고 Go 1.26에서 기본 활성화된 Green Tea GC를 설명합니다.
---

# 가비지 컬렉터와 Green Tea GC

Go 동시성 품질은 GC와 아주 강하게 연결되어 있습니다.

왜냐하면 goroutine, channel, map, timer, request graph는 모두 메모리를 잡고, pointer를 유지하고, scan work를 만들어내기 때문입니다.

:::tip Quick takeaway
Go의 GC는 단순 stop-the-world 수집기가 아닙니다. concurrent mark-sweep collector이고, pacing, assist, write barrier, scavenging, 그리고 이제 Green Tea locality 개선까지 포함한 시스템입니다.
:::

## 핵심 구성 요소

| 구성 요소 | 역할 |
| --- | --- |
| Mark phase | 도달 가능한 heap object 찾기 |
| Sweep phase | 도달 불가능한 span 회수 |
| Write barrier | concurrent marking 중 tri-color invariant 유지 |
| Pacer | GC 강도를 조절 |
| Assist | mutator가 marking work를 일부 부담 |
| Scavenger | 가능할 때 OS에 메모리 반환 |

## 단순화한 내부 스케치

```go
func gcCycle() {
	stopTheWorldBriefly()
	enableWriteBarrier()
	startConcurrentMark()

	for workRemaining() {
		runBackgroundMarkWorkers()
		chargeAssistWorkToAllocators()
	}

	stopTheWorldBriefly()
	disableWriteBarrier()
	startSweep()
}
```

실제 런타임은 훨씬 복잡하지만, 이 정도만 알아도 allocation rate, pointer density, long-lived heap이 왜 서비스 latency에 영향을 주는지 설명할 수 있습니다.

## 왜 assist가 concurrency와 연결되나

많은 팀은 GC를 pause time만으로 이해합니다. 그건 얕습니다.

allocation pressure가 커지면 mutator assist 때문에 일반 request goroutine이 marking work를 일부 떠안습니다.

그래서 GC 비용은 다음처럼 체감될 수 있습니다.

- request handler throughput 저하,
- tail latency 증가,
- giant pause 없이도 "서비스가 전체적으로 느려짐"으로 관측.

## Green Tea GC는 무엇을 바꾸나

Green Tea의 핵심은 locality입니다. 특히 small object marking에서 span 주변 객체를 묶어서 더 locality-friendly하게 scan하려는 방향입니다.

즉:

- 인접 객체를 같이 만지고,
- metadata 접근 비용을 amortize하고,
- marking 시 cache behavior를 개선합니다.

`runtime/mgcmark_greenteagc.go`의 주석도 이를 직접 설명합니다.

## 왜 서버에서 중요하나

이득이 잘 드러나는 곳은 보통 이런 서비스입니다.

- 작은 객체를 많이 할당하고,
- request-scoped object graph가 많으며,
- pointer-rich heap을 유지하고,
- 많은 동시 요청을 처리하는 API / queue / stream processor.

즉, Go가 많이 쓰이는 서버 영역과 정확히 겹칩니다.

## 실제 관측 도구

### `gctrace`

```bash
GODEBUG=gctrace=1 go test ./...
```

이것만으로도 cycle timing, heap growth, collector work를 빠르게 볼 수 있습니다.

### heap / alloc profile

allocation hotspot과 long-lived retention을 찾을 때는 `pprof`가 필요합니다.

### execution trace

`runtime/trace`는 goroutine behavior, blocking, GC activity를 같은 timeline에서 같이 보게 해줍니다.

## 런타임 소스 포인터

- [runtime/mgc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgc.go)
- [runtime/mgcmark.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark.go)
- [runtime/mgcpacer.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcpacer.go)
- [runtime/mgcscavenge.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcscavenge.go)
- [runtime/mgcmark_greenteagc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark_greenteagc.go)

## 역사적 체크포인트

- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Green Tea GC](https://go.dev/blog/greenteagc)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)

Go 1.25에서는 Green Tea가 실험 기능이었고, Go 1.26에서 기본 활성화됐습니다. 이 흐름은 Go 팀이 런타임 혁신을 어떻게 점진적으로 올리는지 잘 보여줍니다.

## 운영 설계에 주는 영향

### allocation rate는 concurrency 문제다

요청 하나가 많은 goroutine fan-out과 allocation을 일으키면, GC work도 concurrency budget의 일부가 됩니다.

### pointer 구조가 중요하다

숫자 위주의 flat buffer보다 pointer-rich long-lived 구조가 scan 비용이 큽니다.

### memory limit도 중요하다

accidental heap growth보다 명시적 메모리 예산이 있는 서비스가 운영하기 쉽습니다.

## Practical takeaway

Go의 동시성 모델이 실용적인 이유 중 하나는 GC가 늘 켜져 있는 서버 workload를 염두에 두고 설계되었기 때문입니다.

goroutine behavior, allocation rate, heap shape, GC pacing을 한 시스템으로 묶어 생각해야 런타임 감각이 생깁니다.
