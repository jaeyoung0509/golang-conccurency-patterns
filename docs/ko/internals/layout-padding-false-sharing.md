---
title: 레이아웃, 패딩, false sharing
description: Struct layout, alignment, cache line이 숨어 있는 contention을 어떻게 만드는지 설명합니다.
---

# 레이아웃, 패딩, false sharing

동시성 코드가 논리적으로 올바르다고 해서 빠르다는 뜻은 아닙니다.

alignment 규칙, padding, cache line sharing은 "독립적인" 카운터나 shard조차 coherence 비용으로 묶어버릴 수 있습니다.

:::tip Quick takeaway
Go의 alignment 규칙은 `unsafe`로 직접 확인할 수 있을 정도로 단순하지만, 진짜 놀라운 비용은 CPU cache behavior에서 나옵니다. race가 없는 atomic update도 같은 cache line을 튕기면 느려질 수 있습니다.
:::

## 먼저 컴파일러가 보장하는 것부터 본다

Go는 struct field를 선언 순서대로 배치하되, 각 field의 alignment를 만족시키기 위해 padding을 삽입합니다.

직접 확인할 수 있습니다.

```go
type Bad struct {
	Flag  byte
	Count uint64
	Tag   byte
}

fmt.Println(unsafe.Sizeof(Bad{}))
fmt.Println(unsafe.Offsetof(Bad{}.Count))
fmt.Println(unsafe.Alignof(Bad{}.Count))
```

이 단계는 "낭비된 바이트"를 설명합니다.

그 다음 단계인 cache line은 "낭비된 throughput"을 설명합니다.

## field 순서를 바꾸면 객체 크기가 줄 수 있다

```go
type Better struct {
	Count uint64
	Flag  byte
	Tag   byte
}
```

field 순서는 다음을 바꿉니다.

- 전체 크기,
- offset,
- cache에 들어가는 객체 수,
- 경우에 따라 hot field가 다른 hot field와 같은 line을 쓰는지 여부.

성능 민감한 코드에서 field 순서를 cosmetic으로 보면 안 됩니다.

## False sharing은 coherence 문제다

false sharing은 두 goroutine이 다른 데이터를 갱신하는데도 같은 cache line을 두고 싸우는 현상입니다.

race가 없어도 되고,
lock bug가 없어도 되고,
프로그램은 완전히 올바를 수 있습니다.

그래도 cache invalidation traffic 때문에 느려질 수 있습니다.

## 단순화한 counter 예시

```go
type Counters struct {
	OK  atomic.Int64
	Err atomic.Int64
}
```

두 코어가 이 field를 독립적으로 계속 두드리고, 둘이 같은 line에 놓이면 write마다 line ownership traffic이 생길 수 있습니다.

## padding이 맞는 도구일 때가 있다

```go
type PaddedCounter struct {
	Value atomic.Int64
	_     [56]byte // 예시: 64-byte line 기준 counter 나머지 공간
}
```

이런 패딩은 일부러 둔합니다. 다음 조건에서만 씁니다.

- field가 정말 hot하고,
- contention이 실제로 있고,
- 프로파일과 벤치마크가 padding 이득을 보여줄 때.

## 표준 라이브러리의 단서: `sync.Pool`

`sync.Pool`은 `poolLocal`에 명시적 padding을 둡니다. per-P shard 사이 false sharing을 피하기 위해서입니다.

표준 라이브러리가 이렇게까지 한다는 사실은, layout이 결코 사소한 구현 디테일이 아니라는 강한 신호입니다.

## 실패 패턴

다음 코드는 멀쩡해 보여도 CPU를 낭비할 수 있습니다.

```go
type Shard struct {
	Hits atomic.Int64
}

var shards [2]Shard

go func() {
	for {
		shards[0].Hits.Add(1)
	}
}()

go func() {
	for {
		shards[1].Hits.Add(1)
	}
}()
```

필드는 논리적으로 독립적이지만, cache line은 아닐 수 있습니다. 둘이 붙어 배치되면 race detector가 알려주지 않는 이유로 throughput이 무너질 수 있습니다.

## 추측하지 않고 layout을 보는 방법

다음을 조합해서 보면 됩니다.

- `unsafe.Sizeof`
- `unsafe.Alignof`
- `unsafe.Offsetof`
- focused benchmark
- 가능하다면 CPU profile이나 hardware counter 도구

Go 내부만으로 조사할 때는 size/offset 확인과 benchmark 변화부터 보는 것이 가장 현실적입니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [unsafe package docs](https://pkg.go.dev/unsafe)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)
- [sync/pool.go](https://github.com/golang/go/blob/go1.26.0/src/sync/pool.go)

## 실전 요약

동시성 설계가 논리적으로 깔끔한데도 CPU를 많이 쓴다면, mutex와 goroutine에서 멈추지 말고 바이트 배치를 봐야 합니다.
