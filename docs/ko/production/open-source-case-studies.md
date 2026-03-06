---
title: 오픈소스 사례
description: 대표적인 Go 오픈소스가 실제로 concurrency를 어떻게 구조화하는지 소스 포인터와 함께 설명합니다.
---

# 오픈소스 사례

동시성 감각을 빠르게 끌어올리는 가장 좋은 방법 중 하나는 실제 production pressure를 버틴 Go 프로젝트를 읽는 것입니다.

이 문서는 프로젝트 소개 모음이 아니라, 실제로 훔쳐올 만한 concurrency boundary를 읽는 가이드입니다.

왜 이런 프로젝트들이 많이 Go로 왔는지 더 넓은 역사 맥락이 궁금하면 [Go 오픈소스 역사 읽기](/ko/production/go-open-source-histories)를 읽으면 됩니다.
현대 durable execution 시스템 하나를 깊게 보고 싶으면 [Temporal과 Durable Execution](/ko/production/temporal-durable-execution)을 같이 보면 됩니다.

## 빠른 지도

| 프로젝트 | 핵심 교훈 | 주 소스 |
| --- | --- | --- |
| Kubernetes `client-go` | 재처리 semantics와 drain-aware worker queue | [`util/workqueue/queue.go`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L167-L339) |
| etcd/raft | `Ready`/`Advance`를 correctness boundary로 쓰는 법 | [`node.go`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L297-L549) |
| Prometheus | periodic work를 owner loop 하나로 묶는 법 | [`scrape/scrape.go`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L822-L1308) |
| NATS Server | `sync.Cond`로 read/write socket ownership을 분리하는 법 | [`server/client.go`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L355-L355), [`writeLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1274-L1355), [`readLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1358-L1415) |
| gRPC-Go | single writer goroutine과 control buffer | [`internal/transport/controlbuf.go`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L307-L629) |
| CockroachDB | 서비스 전체의 task ownership과 quiesce contract | [`pkg/util/stop/stopper.go`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L152-L670) |
| go-redis | pool admission, wait budget, connection handoff | [`internal/pool/pool.go`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L306-L1078) |

## 어떻게 읽어야 하나

아래 코드를 읽기 전에 항상 먼저 물어야 할 질문은 이 다섯 가지입니다.

1. lifetime owner는 어디인가
2. overload boundary는 어디인가
3. state owner는 누구인가
4. shutdown contract는 어디에 드러나는가
5. 어떤 metric이나 trace가 이 설계의 이상 징후를 보여줄 수 있는가

API 이름보다 이 질문이 더 중요합니다.

## Kubernetes client-go workqueue

왜 읽을 가치가 있는가:

controller 코드는 흔히 “worker goroutine이 queue를 소비한다” 정도로 설명됩니다. 진짜 중요한 건 goroutine 수가 아니라, 처리 중 재변경 신호를 어떻게 잃지 않는가입니다.

주 소스:

- [`Typed` queue 상태](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L192-L203)
- [`Add`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L224-L249)
- [`Get`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L264-L283)
- [`Done`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L289-L302)
- [`ShutDown`, `ShutDownWithDrain`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L305-L340)

핵심 concurrency boundary:

queue는 queued item, `dirty`, `processing` 세 상태를 같이 소유하면서 재처리 semantics를 책임집니다.

단순화한 형태:

```go
func Add(item Key) {
	if shuttingDown {
		return
	}
	if dirty[item] {
		if !processing[item] {
			touch(item)
		}
		return
	}
	dirty[item] = true
	if processing[item] {
		return
	}
	queue.push(item)
	signalWorker()
}

func Done(item Key) {
	delete(processing, item)
	if dirty[item] {
		queue.push(item)
		signalWorker()
	}
}
```

왜 이게 중요한가:

worker가 아직 처리 중인 동안 같은 object가 다시 바뀌어도 그 신호가 사라지지 않습니다. `dirty`가 그것을 기억했다가 `Done` 뒤에 다시 큐잉합니다.

실패 패턴:

이걸 보고 “아무 큐나 두 개의 map만 붙이면 된다”고 읽으면 안 됩니다. 교훈은 더 좁습니다. in-flight 중복 업데이트가 가능한 시스템이라면 replay semantics를 queue가 명시적으로 가져야 한다는 점입니다.

핵심 takeaway:

이건 단순 channel이 아니라 상태 semantics를 가진 queue입니다.

## etcd/raft node loop

왜 읽을 가치가 있는가:

Go에서 communication boundary가 correctness boundary가 되는 가장 좋은 예 중 하나입니다.

주 소스:

- [`node` 메인 루프](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L297-L456)
- [`Tick`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L458-L466)
- [`Step`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L473-L545)
- [`Ready`, `Advance`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L547-L553)

핵심 concurrency boundary:

Raft core는 `Ready`를 통해 “지금 persist/send/apply 해야 하는 작업 묶음”을 내보내고, 바깥 계층은 실제로 그 경계를 만족한 뒤 `Advance`를 호출합니다.

단순화한 형태:

```go
for {
	select {
	case rd := <-node.Ready():
		persist(rd.HardState, rd.Entries, rd.Snapshot)
		send(rd.Messages)
		apply(rd.CommittedEntries)
		node.Advance()
	}
}
```

왜 이게 중요한가:

알고리즘 코어는 디스크 durability나 네트워크 전송을 직접 소유하지 않습니다. 대신 다음 진행에 필요한 correctness unit을 선언하고 바깥이 확인해주길 기다립니다.

실패 패턴:

이걸 그냥 “채널 기반 루프 하나” 정도로 읽으면 핵심을 놓칩니다. durability boundary를 만족하기 전에 `Advance`를 호출하면 설계 자체를 깨뜨립니다.

핵심 takeaway:

좋은 channel API는 데이터를 전달할 뿐 아니라, 상태 머신이 언제 앞으로 나가도 되는지도 정의합니다.

## Prometheus scrape loop

왜 읽을 가치가 있는가:

Prometheus는 periodic work를 timer 조각들로 흩뿌리지 않고, 시간과 종료를 하나의 owner loop에 모읍니다.

주 소스:

- [`scrapeLoop` 상태](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L822-L910)
- [`newScrapeLoop`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L1154-L1232)
- [`run`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L1234-L1308)

핵심 concurrency boundary:

하나의 loop가 startup offset, ticker cadence, cancellation, staleness marker, append flow를 모두 소유합니다.

단순화한 형태:

```go
func run() {
	waitInitialOffset()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-parentCtx.Done():
			return
		case <-loopCtx.Done():
			break
		default:
		}

		last = scrapeAndReport(last, nowRounded())

		select {
		case <-parentCtx.Done():
			return
		case <-loopCtx.Done():
			break
		case <-ticker.C:
		}
	}

	endOfRunStaleness(last)
}
```

왜 이게 중요한가:

주기 작업은 시간이 곧 state입니다. Prometheus는 그 시간을 owner loop 하나에 붙여서 shutdown과 append semantics를 같이 관리합니다. 소스에서 shutdown용 context와 append용 context를 분리하는 점도 눈여겨볼 만합니다.

실패 패턴:

helper 함수 곳곳에 `time.After`와 ticker를 흩뿌리는 식으로 periodic 시스템을 만들면 lifecycle 경계가 흐려집니다.

핵심 takeaway:

주기 작업이라면 time ownership을 하나의 loop에 모으는 편이 강합니다.

## NATS Server client read/write loops

왜 읽을 가치가 있는가:

Go에서 channels만이 정답이 아니라는 점을 아주 분명하게 보여줍니다.

주 소스:

- [`sync.Cond`가 붙은 outbound 상태](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L355-L355)
- [`writeLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1274-L1355)
- [`readLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1358-L1415)
- [`flushSignal`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1996-L2003)

핵심 concurrency boundary:

socket read와 write는 ownership이 다릅니다. write side는 buffered outbound state를 소유하고, 필요할 때 `sync.Cond`로 깨워집니다.

단순화한 형태:

```go
func writeLoop() {
	for {
		lock()
		for open && shouldSleep() {
			cond.Wait()
		}
		if closed {
			flushAndClose()
			unlock()
			return
		}
		flushOutbound()
		unlock()
	}
}

func flushSignal() {
	cond.Signal()
}
```

왜 이게 중요한가:

writer loop가 socket write path를 단독 소유하므로, 여러 goroutine이 직접 connection에 write하는 혼란을 피할 수 있습니다.

실패 패턴:

`go readLoop`, `go writeLoop` 모양만 복사하면 안 됩니다. 진짜 교훈은 `single-owner socket writing + explicit wake-up semantics`입니다.

핵심 takeaway:

hot socket path에서는 `sync.Cond`와 명확한 owner가 channel보다 더 정확한 모델일 수 있습니다.

## gRPC-Go transport control buffer

왜 읽을 가치가 있는가:

여러 RPC stream이 하나의 HTTP/2 connection을 공유할 때, 왜 single writer가 필요한지 아주 잘 보여줍니다.

주 소스:

- [`controlBuffer`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L307-L511)
- [`loopyWriter`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L513-L584)
- [`loopyWriter.run`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L586-L629)

핵심 concurrency boundary:

frame emission은 하나의 writer goroutine이 소유하고, 다른 goroutine은 control item과 data를 control buffer에 enqueue합니다.

단순화한 형태:

```go
func transportWriter() {
	for {
		item := controlBuffer.get(block=true)
		handleControl(item)
		processOneActiveStream()

		for controlBuffer.hasMore() || activeStreamsNotEmpty() {
			handleMoreItems()
			processOneActiveStream()
		}

		flushWire()
	}
}
```

왜 이게 중요한가:

control buffer가 transport mutation을 serialize하고, loopy writer는 flow-control quota 아래에서 active stream을 round-robin으로 처리합니다. 즉 하나의 connection 위에 여러 논리 stream을 안전하게 multiplex합니다.

실패 패턴:

stream goroutine이 shared `net.Conn`에 직접 HTTP/2 frame을 쓰게 두면 안 됩니다. single-writer boundary는 구현 디테일이 아니라 설계 핵심입니다.

핵심 takeaway:

multiplexed transport는 대개 command queue + writer owner 구조가 맞습니다.

## CockroachDB stopper

왜 읽을 가치가 있는가:

CockroachDB는 goroutine lifetime을 서비스 차원의 경계로 다룹니다. background task가 “언젠가 끝나겠지”에 맡겨져 있지 않습니다.

주 소스:

- [`Stopper` 상태](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L152-L214)
- [`WithCancelOnQuiesce`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L272-L289)
- [`RunAsyncTask`, `RunAsyncTaskEx`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L358-L446)
- [`Stop`, `Quiesce`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L562-L655)

핵심 concurrency boundary:

stopper가 task admission, task counting, quiesce signal, shutdown wait를 중앙에서 소유합니다.

단순화한 형태:

```go
func RunAsyncTask(ctx, opts, f) error {
	handle, err := GetHandle(ctx, opts)
	if err != nil {
		return err
	}
	go func() {
		defer handle.Release()
		f(handleCtx)
	}()
	return nil
}

func Stop(ctx) {
	markStopping()
	close(quiescer)
	cancelQuiesceContexts()
	waitUntilNumTasksIsZero()
	runClosersInReverseOrder()
}
```

왜 이게 중요한가:

background work에 admission control, observability, coordinated shutdown path를 붙이면 서비스 전체의 lifetime이 예측 가능해집니다.

실패 패턴:

이걸 “조금 더 좋은 wait group”으로 읽으면 안 됩니다. 핵심은 quiescing 상태에서 새 작업을 거부하고, 기존 task에 shutdown-tied cancellation을 준다는 점입니다.

핵심 takeaway:

큰 서비스는 goroutine lifetime을 중앙에서 소유하는 장치가 필요합니다.

## go-redis connection pool

왜 읽을 가치가 있는가:

이건 서버가 아니라 라이브러리지만, admission control과 handoff를 아주 깔끔하게 보여주는 예입니다.

주 소스:

- [`ConnPool` 상태](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L306-L339)
- [`Get`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L701-L819)
- [`queuedNewConn`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L821-L897)
- [`waitTurn`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L954-L978)
- [`Put`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L1049-L1078)

핵심 concurrency boundary:

caller는 무제한으로 dial하지 않습니다. 먼저 pool에서 permission을 얻고, 그 다음 idle connection을 재사용하거나 new dial 경로로 들어갑니다.

단순화한 형태:

```go
func Get(ctx) (*Conn, error) {
	if err := waitTurn(ctx); err != nil {
		return nil, err
	}

	if cn := popHealthyIdle(); cn != nil {
		return cn, nil
	}

	return queuedNewConn(ctx)
}

func waitTurn(ctx) error {
	if semaphore.TryAcquire() {
		return nil
	}
	return semaphore.Acquire(ctx, poolTimeout)
}
```

왜 이게 중요한가:

pool은 명시적인 admission boundary를 가집니다. saturation 시 caller는 기다리거나 fail하지, 무한 dial을 새로 만들지 않습니다.

실패 패턴:

pool을 두면서도 timeout 이후 caller가 각자 직접 connection을 열 수 있게 두면 실제로는 아무 것도 bounded하지 않은 것입니다.

핵심 takeaway:

진짜 pool은 idle connection cache가 아니라 admission controller입니다.

## 공통적인 오해

프로덕션 Go 코드를 잘못 읽으면 goroutine만 복사하고 그 경계는 빠뜨리기 쉽습니다.

```go
for _, msg := range batch {
	go process(msg) // bad: queue도 limit도 shutdown contract도 없음
}
```

위 시스템들이 강한 이유는 goroutine 수가 아니라, queue state, ownership loop, admission control, flow control, shutdown policy가 같이 있기 때문입니다.

## Practical takeaway

오픈소스 Go 코드는 API보다 경계를 읽을 때 가장 많이 배울 수 있습니다.

좋은 프로젝트일수록 queue design, lifecycle loop, signaling edge, backpressure, shutdown path 안에 concurrency policy가 드러납니다.
