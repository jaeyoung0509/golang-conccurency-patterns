---
title: 오픈소스 사례
description: 대표적인 Go 오픈소스가 실제로 concurrency를 어떻게 구조화하는지 소스 포인터와 함께 설명합니다.
---

# 오픈소스 사례

동시성 감각을 빠르게 끌어올리는 가장 좋은 방법 중 하나는 실제 production pressure를 버틴 Go 프로젝트를 읽는 것입니다.

이 문서는 프로젝트 전체를 요약하려는 것이 아니라, 특히 볼 가치가 있는 concurrency 구조를 짚습니다.

## Kubernetes client-go workqueue

소스:

- [kubernetes/client-go util/workqueue/queue.go](https://github.com/kubernetes/client-go/blob/master/util/workqueue/queue.go)

볼 포인트:

- `dirty`와 `processing` 집합
- `Get`, `Done`, `ShutDown`
- 중복 work를 dedupe하면서도 requeue semantics를 잃지 않는 방식

왜 중요한가:

이 queue 설계는 "같은 item이 처리 중일 때 다시 업데이트되면 어떻게 할 것인가"에 대한 production-grade 답입니다. 같은 item이 queue를 무한히 오염시키지 않으면서도 재처리 신호는 보존합니다.

단순화한 형태:

```go
if item already processing {
    mark dirty
    return
}
enqueue(item)
```

핵심 takeaway:

프로덕션 queue는 단순 slice + channel이 아닙니다. 재처리 semantics와 shutdown semantics를 함께 품습니다.

## etcd raft node loop

소스:

- [etcd-io/raft node.go](https://github.com/etcd-io/raft/blob/main/node.go)

볼 포인트:

- `Ready() <-chan Ready`
- `Advance()`
- `Tick()`
- `Step(ctx, msg)`

왜 중요한가:

Raft core와 storage / transport를 channel 기반 경계로 분리합니다. core는 "다음에 persist/send 해야 할 것"을 내보내고, 바깥 계층이 durability를 보장한 뒤 `Advance`로 진행시킵니다.

단순화한 형태:

```go
for {
    select {
    case rd := <-node.Ready():
        persist(rd)
        send(rd.Messages)
        node.Advance()
    }
}
```

핵심 takeaway:

communication boundary가 곧 correctness boundary가 되는 아주 좋은 예입니다.

## Prometheus scrape loop

소스:

- [prometheus/prometheus scrape/scrape.go](https://github.com/prometheus/prometheus/blob/main/scrape/scrape.go)

볼 포인트:

- `newScrapeLoop`
- `(*scrapeLoop).run`
- ticker-driven scheduling, cancellation, staleness handling이 한 lifecycle loop 안에 모여 있는 구조

왜 중요한가:

Prometheus는 단순히 "N초마다 goroutine 하나 돌린다"가 아닙니다. scrape loop가 시간, 에러 처리, append flow, shutdown semantics를 하나의 owner loop에서 책임집니다.

핵심 takeaway:

주기 작업은 helper 곳곳에 timer를 뿌리는 것보다, explicit owner loop 하나에 lifecycle을 모으는 편이 대규모 운영에서 훨씬 낫습니다.

## NATS server client read/write loops

소스:

- [nats-io/nats-server server/client.go](https://github.com/nats-io/nats-server/blob/main/server/client.go)

볼 포인트:

- `readLoop`
- `writeLoop`
- outbound flush를 위한 `sync.Cond` signaling

왜 중요한가:

NATS는 socket read와 write를 분리하고, busy waiting 대신 명시적 signaling을 사용합니다. 모든 것을 channel로만 밀어붙이지 않고, condition variable과 소유권이 더 잘 맞는 경우를 보여줍니다.

핵심 takeaway:

프로덕션 Go에서는 channel만이 정답이 아닙니다. `sync.Cond`와 명확한 owner loop가 더 적합한 경우가 분명히 있습니다.

## 실패 패턴

프로덕션 Go 코드를 잘못 읽으면 goroutine만 복사하고 그 주변 경계는 빠뜨리기 쉽습니다.

```go
for _, msg := range batch {
	go process(msg) // bad: queue도 limit도 shutdown contract도 없음
}
```

위에서 본 시스템들이 실제로 강한 이유는 queue, owner loop, signaling edge, lifecycle rule을 같이 갖고 있기 때문입니다. goroutine의 존재 자체보다 그 경계가 더 중요합니다.

## 어떻게 읽어야 하나

오픈소스 Go 코드를 읽을 때는 이렇게 물어보면 좋습니다.

1. lifetime owner는 어디인가
2. overload boundary는 어디인가
3. state owner는 누구인가
4. shutdown contract는 어디에 드러나는가
5. 어떤 metric / trace가 실패를 드러내는가

## Practical takeaway

오픈소스 Go 코드는 API보다 경계를 읽을 때 가장 많이 배울 수 있습니다.

좋은 프로젝트일수록 queue design, lifecycle loop, signaling edge, shutdown path 안에 concurrency policy가 잘 드러납니다.
