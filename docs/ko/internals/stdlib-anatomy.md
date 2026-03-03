---
title: 표준 라이브러리 해부
description: sync.Pool 내부 구조, reflection 비용, 그리고 표준 라이브러리가 드러내는 런타임 tradeoff를 설명합니다.
---

# 표준 라이브러리 해부

표준 라이브러리는 Go 팀이 abstraction, cache locality, GC pressure, concurrency safety를 실제로 어떻게 균형 잡는지 보여주는 최고의 교재입니다.

그중 특히 좋은 사례가 `sync.Pool`과 `reflect`입니다.

:::tip Quick takeaway
`sync.Pool`은 global bag + lock이 아닙니다. per-P shard, stealing, victim cache 구조입니다. `reflect`가 느린 이유도 "마법이라서"가 아니라 type metadata, indirection, check, copy 비용을 실제로 수행하기 때문입니다.
:::

## `sync.Pool`: per-P가 먼저다

`sync.Pool`은 많은 concurrent caller가 잠깐 쓰는 temporary object의 allocation pressure를 줄이기 위한 구조입니다.

핵심 설계는 다음과 같습니다.

- P마다 local shard
- 가장 빠른 reuse를 위한 `private` slot
- 추가 값을 담는 `shared` chain
- cross-P stealing
- 한 GC cycle 동안 값을 살려 두는 victim cache

## 단순화한 pool 예시

```go
var bufPool = sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

func encode(v any) []byte {
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)

	_ = json.NewEncoder(buf).Encode(v)
	return append([]byte(nil), buf.Bytes()...)
}
```

중요한 점은 pooled object가 durable ownership이 아니라 temporary scratch라는 사실입니다.

## 왜 pool은 padding을 두는가

`sync.Pool`의 `poolLocal`에는 P-local shard 사이 false sharing을 피하기 위한 명시적 padding이 있습니다.

표준 라이브러리가 이런 선택을 했다는 것은, multicore cache behavior가 알고리즘보다 덜 중요한 문제가 아니라는 뜻입니다.

## Reflection 비용은 대부분 명시적 런타임 작업이다

`reflect.Value`는 다음을 들고 있습니다.

- type metadata
- pointer/data 표현
- addressable 여부, indirection, method value, read-only 상태를 나타내는 flag bit

따라서 reflective operation은 종종 다음을 수행합니다.

- kind check
- metadata traversal
- interface pack/unpack
- indirect load
- 안전한 materialization을 위한 추가 copy

즉, 느린 이유는 신비해서가 아니라 실제로 더 많은 런타임 작업을 하기 때문입니다.

## 단순화한 reflection 예시

```go
func nonZeroFields(v any) []string {
	rv := reflect.ValueOf(v)
	rt := rv.Type()
	out := make([]string, 0, rv.NumField())

	for i := 0; i < rv.NumField(); i++ {
		if !rv.Field(i).IsZero() {
			out = append(out, rt.Field(i).Name)
		}
	}
	return out
}
```

이건 control plane, admin tool, config, 저속 경로에서는 충분히 괜찮습니다.

하지만 per-record hot loop에서는 전혀 다른 cost model이 됩니다.

## 실전 대안: typed code나 code generation

reflection을 대체하는 가장 빠른 방법은 보통 교묘한 unsafe가 아닙니다.

대부분은 다음 중 하나입니다.

- concrete hand-written typed code
- compile-time shape만으로 충분한 generic code
- serializer, validator, mapper를 위한 generated code

핵심은 runtime inspection 비용을 compile-time structure로 옮기는 것입니다.

## 실패 패턴

다음 두 가지는 특히 자주 보입니다.

```go
var pool sync.Pool

func storeHuge(x *BigBuffer) {
	pool.Put(x) // bad: Pool을 permanent cache처럼 사용
}
```

```go
func hot(items []any) {
	for _, item := range items {
		_ = reflect.TypeOf(item).Kind() // bad: 작은 hot loop에 metadata path를 올림
	}
}
```

첫 번째는 lifetime semantics를 잘못 이해한 것이고,
두 번째는 flexibility를 zero-cost abstraction으로 착각한 것입니다.

## 소스 읽기 포인트

여기서 시작하면 됩니다.

- [sync/pool.go](https://github.com/golang/go/blob/go1.26.0/src/sync/pool.go)
- [sync/poolqueue.go](https://github.com/golang/go/blob/go1.26.0/src/sync/poolqueue.go)
- [reflect/value.go](https://github.com/golang/go/blob/go1.26.0/src/reflect/value.go)
- [internal/abi/type.go](https://github.com/golang/go/blob/go1.26.0/src/internal/abi/type.go)

## 실전 요약

표준 라이브러리를 자세히 읽으면 같은 패턴이 반복됩니다.

- common path는 local하게 두고,
- lifetime은 명확히 하고,
- flexibility가 정당화될 때만 metadata 비용을 지불합니다.

이건 라이브러리 구현 팁이 아니라 프로덕션 설계 원칙입니다.
