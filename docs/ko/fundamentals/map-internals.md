---
title: 맵 내부 구조와 Swiss Tables
description: 현대 Go 맵의 구조, Swiss Table group, extendible hashing, 그리고 왜 concurrent write가 여전히 위험한지 설명합니다.
---

# 맵 내부 구조와 Swiss Tables

맵은 실제 Go 서비스에서 가장 자주 만나는 공유 구조 중 하나입니다.

- cache,
- in-memory index,
- dedup table,
- actor-owned state,
- request-scoped accumulator.

전문가 수준 fundamentals라면 최신 map mental model도 반드시 알아야 합니다.

:::tip Quick takeaway
Go 1.24 이후 built-in map은 Swiss Table과 extendible hashing 기반 구조를 사용합니다. locality와 growth behavior는 좋아졌지만, 동기화 없는 concurrent mutation이 안전해진 것은 아닙니다.
:::

## 현대 Go map의 구조

Go 1.26 런타임은 `internal/runtime/maps`에 대부분의 맵 로직을 둡니다.

중요 용어는 이렇습니다.

| 용어 | 의미 |
| --- | --- |
| Slot | key/value 한 쌍의 저장 위치 |
| Group | 8개 slot과 하나의 control word |
| Control word | empty/deleted/full 상태를 담는 메타데이터 |
| Table | 하나의 Swiss Table hash table |
| Directory | 어떤 table을 볼지 고르는 상위 구조 |
| H1 / H2 | hash의 상위/하위 비트 분할 |

## Swiss Table이 왜 빠른가

핵심은 group 단위 검사입니다.

전통적 방식처럼 slot을 하나씩 보는 대신, control word를 이용해 같은 group 안의 candidate를 훨씬 빠르게 걸러낼 수 있습니다.

그래서 좋아지는 것은:

- locality,
- 분기 예측,
- probe step당 실제 유효 작업량.

## 단순화한 내부 스케치

```go
func lookup(m *Map, key K) (V, bool) {
	hash := hashKey(key, m.seed)
	table := selectTable(m.directory, hash)
	seq := probeSeq(h1(hash), table.groupMask)

	for {
		group := table.groups[seq.offset]
		matches := group.matchH2(h2(hash))

		for slot := range matches {
			if group.key(slot) == key {
				return group.value(slot), true
			}
		}

		if group.hasEmptySlot() {
			return zero, false
		}

		seq = seq.next()
	}
}
```

실제 런타임은 tombstone, indirect key/elem, iteration semantics, grow state까지 처리하므로 더 복잡합니다.

## growth는 어떻게 달라졌나

예전 mental model은 흔히 "bucket을 재배치하며 grow한다"였습니다.

현대 Go map도 재배치는 필요하지만, 이제는 directory와 table 구조 덕분에 "맵 전체를 한 번에 크게 뒤집는 비용"을 더 세밀하게 나눌 수 있습니다.

이 점이 중요한 이유는:

- grow latency control이 더 좋아지고,
- 큰 맵에서 cost model이 더 세련되며,
- iteration semantics를 유지하기 쉬워지기 때문입니다.

## iteration이 제일 어렵다

`internal/runtime/maps/map.go` 주석도 분명하게 말합니다. iteration이 가장 복잡한 부분입니다.

이유는:

- 같은 entry를 두 번 반환하면 안 되고,
- 수정된 값은 최신 값이어야 하며,
- 삭제된 값은 보이면 안 되고,
- grow가 도중에 일어날 수 있으며,
- 순서는 원래부터 비결정적이기 때문입니다.

그래서 concurrent mutation이 더 위험합니다. 겉 API는 작아도 내부 상태는 꽤 복잡하게 움직입니다.

## 왜 concurrent write는 여전히 unsafe한가

런타임은 write-state bit와 race detector hook을 둡니다. 그만큼 동시 mutation이 내부 invariant를 깨뜨릴 가능성이 큽니다.

정확한 mental model은 이렇습니다.

- map은 single-owner 또는 외부 동기화를 전제로 최적화되어 있고,
- 런타임이 일부 나쁜 접근을 감지할 수는 있지만,
- 그것이 correctness를 보장하지는 않습니다.

여러 goroutine이 맵을 만져야 한다면 의도적으로 선택해야 합니다.

- `map + sync.Mutex`
- `map + sync.RWMutex`
- channel 기반 single-owner goroutine
- sharded ownership

## 실전 영향

### 더 빠른 map이 contention을 없애주지는 않는다

hot map 하나를 lock으로 감싸고 많은 goroutine이 때리면, 병목은 여전히 lock contention일 수 있습니다.

### actor-owned map이 더 깔끔할 때가 있다

업데이트 규칙이 복잡할수록, mutex보다 single-owner state machine이 읽기 쉬울 때가 많습니다.

### 큰 map은 GC에도 영향을 준다

좋아진 설계와 별개로, pointer-rich large map은 heap scan과 retention 비용을 바꿉니다.

## 실패 패턴

Swiss Table이 들어왔다고 concurrent mutation이 안전해진 것은 아닙니다.

```go
counts := map[string]int{}

go func() { counts["ok"]++ }()
go func() { counts["fail"]++ }()
```

이 코드는 여전히 race이며 concurrent map write panic으로 이어질 수 있습니다. 내부 구현이 빨라졌다는 사실은 ownership 규칙을 없애주지 않습니다.

## 런타임 소스 포인터

- [runtime/map.go](https://github.com/golang/go/blob/go1.26.0/src/runtime/map.go)
- [internal/runtime/maps/map.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/map.go)
- [internal/runtime/maps/table.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/table.go)
- [internal/runtime/maps/group.go](https://github.com/golang/go/blob/go1.26.0/src/internal/runtime/maps/group.go)

## 공식 자료

- [Faster Go maps with Swiss Tables](https://go.dev/blog/swisstable)
- [Go 1.24 Release Notes](https://go.dev/doc/go1.24)

## Practical takeaway

현대 Go map은 예전 bucket-only mental model보다 훨씬 정교합니다.

성능은 좋아졌지만, 동시에 ownership의 중요성도 더 분명하게 보여줍니다. hot map을 대할 때 "아마 괜찮겠지"는 런타임 수준 사고가 아닙니다.
