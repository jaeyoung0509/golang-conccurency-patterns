---
title: go-redis 실전 플레이북
description: go-redis를 pooling, timeout, protocol, instrumentation 정책 관점에서 운영하는 법을 설명합니다.
---

# go-redis 실전 플레이북

`go-redis`는 호출은 쉽지만, 잘못 쓰기도 쉽습니다.

겉으로는 가벼운 클라이언트처럼 보이지만 운영 관점에서는 pool owner이자 timeout boundary이고 protocol policy surface입니다.

## Mental model

`go-redis` client는 long-lived resource boundary로 다루는 편이 맞습니다.

- connection pooling 소유
- dial/read/write/pool wait budget 소유
- protocol mode와 hook 소유
- 개별 request보다 오래 살아야 하는 객체

## 안전한 기본 스케치

```go
rdb := redis.NewClient(&redis.Options{
	Addr:            "redis:6379",
	Protocol:        2,
	DialTimeout:     1 * time.Second,
	ReadTimeout:     500 * time.Millisecond,
	WriteTimeout:    500 * time.Millisecond,
	PoolTimeout:     1 * time.Second,
	PoolSize:        64,
	MinIdleConns:    8,
	ReadBufferSize:  32 * 1024,
	WriteBufferSize: 32 * 1024,
	MaxRetries:      2,
})

if err := errors.Join(
	redisotel.InstrumentTracing(rdb),
	redisotel.InstrumentMetrics(rdb),
); err != nil {
	return err
}
```

이 스케치가 안전한 이유는 주요 정책 레버를 전부 드러내기 때문입니다.

- connection reuse
- 짧은 network budget
- bounded pool wait
- 의도적인 RESP version
- 처음부터 instrumentation

## 운영 규칙

### client를 재사용한다

request마다 client를 새로 만들면 pooling을 버리고 Redis를 dial-heavy dependency로 바꿔버립니다. policy boundary마다 long-lived client 하나가 맞습니다.

### RESP2/RESP3를 의식적으로 고른다

공식 저장소는 둘 다 지원하지만, 그게 RESP3가 무조건 기본값이라는 뜻은 아닙니다. RediSearch나 query response처럼 RESP3 shape가 아직 불안정한 경로를 쓴다면 RESP2가 더 안전할 수 있습니다.

### pipeline은 batching이지 transaction이 아니다

pipeline은 round trip을 줄여줍니다. transactional isolation을 주지는 않습니다. 서버 측 atomicity가 필요하면 Redis transaction이나 Lua/Functions를 별도로 써야 합니다.

### pool timeout은 network timeout만큼 중요하다

Redis dependency는 서버가 느려서도 실패하지만, 네트워크가 느려서도 실패하고, 내 caller가 pool 뒤에서 오래 대기해서도 실패합니다. 이 세 상태는 운영 의미가 다릅니다.

### buffer tuning은 workload가 증명할 때만 한다

프로젝트는 현재 32KiB read/write buffer를 기본값으로 둡니다. 이건 좋은 시작점입니다. 더 큰 buffer는 large pipeline이나 high-throughput workload가 실제로 요구할 때만 올리는 편이 낫습니다.

## 실패 패턴

### request마다 client 생성

```go
func get(ctx context.Context, key string) (string, error) {
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"}) // bad
	return rdb.Get(ctx, key).Result()
}
```

### request path에서 `context.Background()` 사용

```go
val, err := rdb.Get(context.Background(), key).Result() // bad: request budget이 없음
```

### pipeline을 atomic하게 오해

```go
pipe := rdb.Pipeline()
pipe.Incr(ctx, "balance")
pipe.Set(ctx, "status", "ok", 0)
_, _ = pipe.Exec(ctx) // bad assumption: batching은 transaction이 아님
```

### search-heavy 코드에서 blind RESP3 전환

공식 저장소는 일부 RediSearch/query 응답 구조가 RESP3에서 아직 unstable하다고 명시합니다. command set 확인 없이 protocol을 바꾸면 안 됩니다.

### pool 가시성이 없음

`PoolStats`를 한 번도 안 보면 self-inflicted queueing을 remote Redis latency로 착각하게 됩니다.

## 주의해서 써야 할 것

- memory와 buffer impact를 측정하지 않은 큰 pipeline
- unstable structure가 있는 command family에서의 RESP3
- non-idempotent workflow에서의 높은 retry count
- user-facing path에서의 `context.Background()`와 wide-open timeout

## Observability와 테스트

- `PoolStats()`를 주기적으로 수집하고 hits, misses, timeouts, total connections를 봅니다.
- `redisotel`로 tracing과 metrics를 초기에 붙입니다.
- Redis command latency와 pool-wait latency를 대시보드에서 분리합니다.
- 실제 Redis 통합 테스트로 timeout/retry 정책을 확인한 뒤 운영에 태웁니다.

## 언제 잘 맞는가

go-redis는 caching, lightweight state, rate limiting, queue, ephemeral coordination 같은 대부분의 Go+Redis workload에서 기본값으로 맞습니다.

## 언제 다른 도구를 봐야 하나

strict relational guarantee, 대형 analytical scan, 복잡한 multi-step business transaction이 필요하다면 Redis + go-redis는 아무리 사용성이 좋아도 맞는 추상화가 아닐 가능성이 큽니다.

## 공식 자료

- [go-redis 저장소](https://github.com/redis/go-redis)
- [Go-Redis is now an official Redis client](https://redis.io/blog/go-redis-official-redis-client/)
- [`github.com/redis/go-redis/v9` 패키지 문서](https://pkg.go.dev/github.com/redis/go-redis/v9)

## Practical takeaway

go-redis는 Redis command helper로 보면 위험하고, pooled network dependency이자 protocol/timeout policy surface로 보면 훨씬 안전해집니다.
