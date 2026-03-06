---
title: go-redis Production Playbook
description: Learn how to operate go-redis with explicit pooling, timeout, protocol, and instrumentation policy.
---

# go-redis Production Playbook

`go-redis` is easy to call and easy to misuse.

The client looks lightweight, but operationally it is a pool owner, a timeout boundary, and a protocol policy surface.

## Mental model

Treat a `go-redis` client as a long-lived resource boundary:

- it owns connection pooling,
- it owns dial, read, write, and pool wait budgets,
- it owns protocol mode and hooks,
- it should usually outlive individual requests.

## Safe default sketch

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

This sketch is safe because it makes the main policy levers explicit:

- connection reuse,
- short network budgets,
- bounded pool waiting,
- deliberate RESP version,
- instrumentation from day one.

## Operating rules

### Reuse the client

Creating a new client per request throws away pooling and turns Redis into a dial-heavy dependency. Keep one long-lived client per policy boundary.

### Pick RESP2 or RESP3 deliberately

The official repository supports both. That does not mean RESP3 is automatically the right default. If you depend on RediSearch or query responses whose RESP3 shapes are still unstable, RESP2 can be safer.

### Pipelines are batching, not transactions

Pipelines reduce round trips. They do not give transactional isolation. If you need server-side atomicity, use Redis transactions or Lua/Functions appropriately.

### Pool timeout matters as much as network timeout

A Redis dependency can fail because the server is slow, because the network is slow, or because your own callers are queued behind the pool. Those are different operational states.

### Tune buffers only when the workload proves it

The project defaults to 32KiB read/write buffers now. That is a good default. Increase them when large pipelines or high-throughput workloads justify it, not because bigger sounds faster.

## Failure patterns

### Client per request

```go
func get(ctx context.Context, key string) (string, error) {
	rdb := redis.NewClient(&redis.Options{Addr: "redis:6379"}) // bad
	return rdb.Get(ctx, key).Result()
}
```

### Using `context.Background()` in request paths

```go
val, err := rdb.Get(context.Background(), key).Result() // bad: no request budget
```

### Treating pipeline as atomic

```go
pipe := rdb.Pipeline()
pipe.Incr(ctx, "balance")
pipe.Set(ctx, "status", "ok", 0)
_, _ = pipe.Exec(ctx) // bad assumption: batching is not a transaction
```

### Blind RESP3 upgrade on search-heavy code

The official repo notes that some RediSearch/query response structures are still unstable under RESP3. Do not switch protocols without checking the command set you depend on.

### No pool visibility

If you never look at `PoolStats`, you can mistake self-inflicted queueing for remote Redis latency.

## What to use carefully

- very large pipelines without measuring memory and buffer impact
- RESP3 on command families with unstable structures
- high retry counts on non-idempotent workflows
- `context.Background()` and wide-open timeouts on user-facing paths

## Observability and testing

- Export `PoolStats()` regularly and watch hits, misses, timeouts, and total connections.
- Instrument tracing and metrics early with `redisotel`.
- Separate Redis command latency from pool-wait latency in dashboards.
- Test timeout and retry policy against a real Redis in integration tests before betting production load on it.

## When it is the right tool

go-redis is the default choice for most Go services that need Redis caching, lightweight state, rate limiting, queues, or ephemeral coordination.

## When it is not the right tool

If you need strict relational guarantees, large analytical scans, or multi-step business transactions with complex invariants, Redis plus go-redis is usually the wrong abstraction no matter how ergonomic the client feels.

## Official reading

- [go-redis repository](https://github.com/redis/go-redis)
- [Go-Redis is now an official Redis client](https://redis.io/blog/go-redis-official-redis-client/)
- [Package docs for `github.com/redis/go-redis/v9`](https://pkg.go.dev/github.com/redis/go-redis/v9)

## Practical takeaway

go-redis is safest when you treat it as a pooled network dependency with protocol and timeout policy, not as a tiny helper around Redis commands.
