---
title: 런타임 진화
description: 선점, 스케줄러, 맵 구조, Green Tea GC를 중심으로 Go 런타임이 어떻게 진화했는지 따라갑니다.
---

# 런타임 진화

Go의 동시성 이야기는 처음부터 지금 모습이었던 것이 아닙니다.

초기부터 goroutine과 channel은 있었지만, 실제 운영 환경에서 그 약속을 지키기 위해 런타임은 계속 바뀌어 왔습니다.

:::tip Quick takeaway
핵심은 "예전 Go는 별로였고 지금은 좋다"가 아닙니다. 핵심은 런타임이 계속 latency cliff를 줄여 왔다는 점입니다. 선점, 맵 grow 비용, GC locality, 시간 기반 테스트 안정성이 그 예입니다.
:::

## 어떤 문제가 런타임을 바꿨나

| 문제 | 과거 한계 | 핵심 변화 |
| --- | --- | --- |
| CPU-bound goroutine이 fairness와 GC 진행을 늦춤 | 선점이 함수 호출 같은 safe point에 많이 의존 | Go 1.14의 asynchronous preemption |
| 큰 맵 grow 비용과 locality 한계 | 예전 bucket/overflow 중심 설계 | Go 1.24의 Swiss Table + extendible hashing |
| 작은 객체 marking의 cache locality 한계 | concurrent GC는 강했지만 locality 개선 여지 존재 | Go 1.25 Green Tea GC 실험, Go 1.26 기본 활성화 |
| 동시성 테스트가 sleep에 의존 | 시간 기반 테스트가 느리고 flaky | `testing/synctest` 정착 |

## 릴리스 기준 타임라인

| 릴리스 | 변화 | 의미 |
| --- | --- | --- |
| 2020년 2월 25일, Go 1.14 | asynchronous preemption | 긴 CPU loop가 프로세스를 독점하기 어려워짐 |
| 2025년 2월 11일, Go 1.24 | Swiss Table maps, `testing/synctest` 실험 | 맵 구조 개선과 동시성 테스트 도구 강화 |
| 2025년 8월 12일, Go 1.25 | Green Tea GC 실험 | 작은 객체 marking locality 개선 |
| 2026년 2월 10일, Go 1.26 | Green Tea GC 기본 활성화 | 개선된 marking 경로가 기본 런타임 동작이 됨 |

## 선점 이야기를 정확하게 표현하면

Go가 `선점형 -> 비선점형`으로 바뀐 것은 아닙니다.

더 정확한 설명은 이렇습니다.

- 초기 Go는 협조적이거나 safe-point 중심 선점 성격이 강했고,
- 긴 CPU loop가 fairness와 GC stop point를 늦출 수 있었으며,
- Go 1.14에서 asynchronous preemption이 들어오며 이 문제가 크게 줄었습니다.

즉, "언어가 바뀌었다"기보다는 "런타임이 더 공격적으로 안전한 선점을 할 수 있게 되었다"가 맞습니다.

## 단순화한 내부 스케치

```go
func schedule(p *P) {
	for {
		if g := runqget(p); g != nil {
			execute(g)
			continue
		}

		if g := globrunqget(); g != nil {
			execute(g)
			continue
		}

		if ready := netpoll(0); !ready.empty() {
			injectglist(ready)
			continue
		}

		if g := stealFromOtherP(p); g != nil {
			execute(g)
			continue
		}

		parkm()
	}
}
```

실제 런타임은 훨씬 복잡하지만, 큰 그림은 같습니다. 일을 계속 흘려보내되, thread management 자체가 병목이 되지 않게 하는 것입니다.

## 왜 map과 GC도 concurrency guide에 들어가야 하나

동시성 성능은 스케줄러만으로 결정되지 않습니다.

다음도 같이 봐야 합니다.

- 공유 구조 접근 비용,
- 메모리 locality,
- heap scan 비용,
- allocation-heavy workload에서의 GC 동작.

그래서 전문가 수준 문서라면:

- scheduler와 preemption,
- netpoller와 timers,
- maps와 memory layout,
- GC와 heap scanning

이 전부가 한 체계로 이어져야 합니다.

## 맵은 무엇이 바뀌었나

Go 1.24부터 built-in map은 Swiss Table과 extendible hashing 기반 구조를 사용합니다.

중요한 점은 "이제 concurrent write가 안전하다"가 아니라:

- lookup locality가 좋아졌고,
- growth cost model이 더 정교해졌으며,
- 런타임의 핵심 자료구조가 더 현대화되었다는 점입니다.

## GC는 무엇이 바뀌었나

Go의 GC는 Green Tea 이전에도 이미 concurrent tri-color mark-sweep collector였습니다.

Green Tea는 작은 객체 marking에서 span 단위 batching을 통해 locality를 더 끌어올리는 방향입니다.

즉:

- 차가운 metadata 접근을 줄이고,
- cache behavior를 개선하며,
- allocation-heavy 서버에서 marking 효율을 높입니다.

## 실패 패턴

오래된 스케줄러 folklore가 cargo-cult 형태로 남아 있는 경우가 많습니다.

```go
for {
	doCPUHeavyChunk()
	runtime.Gosched() // bad: 진짜 cancellation과 work budgeting을 대신할 수 없음
}
```

현대 Go에는 async preemption이 있지만, 그렇다고 끝없는 work loop가 좋은 설계가 되지는 않습니다. 런타임 발전을 배운다는 것은 예전 조언을 반복하는 게 아니라 mental model을 업데이트하는 일입니다.

## 런타임 소스 포인터

- [runtime/proc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/proc.go)
- [runtime/netpoll.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/netpoll.go)
- [runtime/map.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/map.go)
- [internal/runtime/maps/map.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/map.go)
- [runtime/mgc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgc.go)
- [runtime/mgcmark_greenteagc.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/mgcmark_greenteagc.go)

## 공식 자료

- [Go 1.14 Release Notes](https://go.dev/doc/go1.14)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)
- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable)
- [Go 1.25 Release Notes](https://go.dev/doc/go1.25)
- [Green Tea GC](https://go.dev/blog/greenteagc)
- [Go 1.26 Release Notes](https://go.dev/doc/go1.26)

## Practical takeaway

Go 동시성을 깊게 이해하려면 "goroutine이 싸다"에서 멈추면 안 됩니다.

런타임이 어떤 병목을 줄여 왔는지까지 같이 봐야, 현재 Go가 왜 이런 느낌으로 동작하는지 이해할 수 있습니다.
