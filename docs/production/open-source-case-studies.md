---
title: Open-Source Case Studies
description: Learn how major Go open-source projects structure concurrency in practice, with source pointers and design takeaways.
---

# Open-Source Case Studies

The fastest way to sharpen your concurrency taste is to read real Go systems that survived production pressure.

This page is not a project summary catalog. It is a source-reading guide focused on the concurrency boundaries that are actually worth stealing.

If you want the broader "why did so many of these projects end up in Go?" story, read [Go Open-Source Histories](/production/go-open-source-histories).
If you want one modern durable-execution system in detail, read [Temporal and Durable Execution](/production/temporal-durable-execution).

## Quick map

| Project | Main lesson | Primary source |
| --- | --- | --- |
| Kubernetes `client-go` | Requeue semantics and drain-aware worker queues | [`util/workqueue/queue.go`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L167-L339) |
| etcd/raft | `Ready`/`Advance` as a correctness boundary | [`node.go`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L297-L549) |
| Prometheus | One owner loop for periodic work and shutdown | [`scrape/scrape.go`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L822-L1308) |
| NATS Server | Split read/write socket ownership with `sync.Cond` signaling | [`server/client.go`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L355-L355), [`writeLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1274-L1355), [`readLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1358-L1415) |
| gRPC-Go | Single writer goroutine plus explicit control buffer | [`internal/transport/controlbuf.go`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L307-L629) |
| CockroachDB | Service-wide task ownership and quiesce semantics | [`pkg/util/stop/stopper.go`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L152-L670) |
| go-redis | Pool admission, wait budgets, and connection handoff | [`internal/pool/pool.go`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L306-L1078) |

## How to read these systems well

Before reading any source below, ask:

1. where is the lifetime owner,
2. where is the overload boundary,
3. where is the state owner,
4. where is the shutdown contract,
5. what metric or trace would prove this design is sick.

That lens is more useful than memorizing APIs.

## Kubernetes client-go workqueue

Why it is worth reading:

Controller code is usually described as “worker goroutines consuming a queue.” The interesting part is not the goroutines. It is how the queue preserves reprocessing signals while preventing duplicate flood.

Primary source:

- [`Typed` queue state](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L192-L203)
- [`Add`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L224-L249)
- [`Get`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L264-L283)
- [`Done`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L289-L302)
- [`ShutDown` and `ShutDownWithDrain`](https://github.com/kubernetes/client-go/blob/c826020ed98be668f4b5a8f140e7a4b8621e2a76/util/workqueue/queue.go#L305-L340)

Main concurrency boundary:

The queue owns requeue semantics with three pieces of state: queued items, `dirty`, and `processing`.

Simplified shape:

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

Why this works:

If an object changes again while a worker is still processing it, the second change is not lost. It is remembered in `dirty` and replayed after `Done`.

Failure pattern to avoid:

Do not cargo-cult this into “I should always have a queue plus two maps.” The lesson is narrower: if updates can arrive while work is in flight, your queue must encode replay semantics explicitly.

Takeaway:

This is a queue with state semantics, not a channel with extra bookkeeping.

## etcd/raft node loop

Why it is worth reading:

This is one of the clearest examples in Go of communication boundaries doubling as correctness boundaries.

Primary source:

- [`node` loop and select`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L297-L456)
- [`Tick`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L458-L466)
- [`Step`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L473-L545)
- [`Ready` and `Advance`](https://github.com/etcd-io/raft/blob/f39329d54771fb68f39e26c440c8cbd02215f6c5/node.go#L547-L553)

Main concurrency boundary:

The Raft core emits `Ready` values that the outer system must persist and send before calling `Advance`.

Simplified shape:

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

Why this works:

The algorithm core does not own disk durability or transport delivery. It declares the next required unit of work and waits for the embedding system to confirm progress.

Failure pattern to avoid:

Never read this as “just another goroutine with channels.” Calling `Advance` before the durability boundary is actually satisfied breaks the design.

Takeaway:

The best channel APIs do not just move data. They codify when a state machine is allowed to move forward.

## Prometheus scrape loop

Why it is worth reading:

Prometheus shows how periodic work becomes robust when one loop owns time, cancellation, staleness markers, and append flow together.

Primary source:

- [`scrapeLoop` state](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L822-L910)
- [`newScrapeLoop`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L1154-L1232)
- [`run`](https://github.com/prometheus/prometheus/blob/3b44b8b463be9291a5605b935a5fbeda88db1366/scrape/scrape.go#L1234-L1308)

Main concurrency boundary:

One loop owns the target lifecycle. It offsets startup, manages ticker cadence, handles shutdown, and only then appends results.

Simplified shape:

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

Why this works:

The loop makes “periodic work” a lifecycle-owned component, not a loose collection of timers. The source even separates shutdown-sensitive context from append context because those lifetimes are not identical.

Failure pattern to avoid:

Do not scatter `time.After` or ad hoc tickers across helper functions and call it a scheduler. Periodic systems need one owner loop with one shutdown contract.

Takeaway:

If work is periodic, give time ownership to a single loop.

## NATS Server client read/write loops

Why it is worth reading:

NATS is an excellent reminder that channels are not the only clean concurrency tool in Go.

Primary source:

- [`sync.Cond` on the outbound state](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L355-L355)
- [`writeLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1274-L1355)
- [`readLoop`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1358-L1415)
- [`flushSignal`](https://github.com/nats-io/nats-server/blob/bae1d6c752e0550901d0a79a76fb2fabe79c1eff/server/client.go#L1996-L2003)

Main concurrency boundary:

Socket reads and writes have different ownership. The write side waits on a condition variable and flushes buffered outbound state when signaled.

Simplified shape:

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

Why this works:

The writer loop owns the socket write path. Other goroutines mutate shared outbound state and signal the writer. That avoids multiple goroutines racing to write directly to the connection.

Failure pattern to avoid:

Do not copy only the `go readLoop` / `go writeLoop` shape. The real lesson is single-owner socket writing plus explicit wake-up semantics.

Takeaway:

A `sync.Cond` plus a clear owner can be more exact than channel-heavy designs for hot socket paths.

## gRPC-Go transport control buffer

Why it is worth reading:

gRPC-Go has to multiplex many RPC streams over one HTTP/2 connection without letting every stream goroutine write frames directly to the socket.

Primary source:

- [`controlBuffer`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L307-L511)
- [`loopyWriter`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L513-L584)
- [`loopyWriter.run`](https://github.com/grpc/grpc-go/blob/b9f7967353473f3b585092ac5d961de3878e1f12/internal/transport/controlbuf.go#L586-L629)

Main concurrency boundary:

One writer goroutine owns frame emission. Other goroutines enqueue control items and data into the control buffer.

Simplified shape:

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

Why this works:

The control buffer serializes transport mutations, and the loopy writer round-robins active streams under flow-control limits. That is how many logical streams share one connection without frame corruption or fairness collapse.

Failure pattern to avoid:

Do not let per-stream goroutines write HTTP/2 frames directly to a shared `net.Conn`. The single-writer boundary is not an implementation detail. It is the design.

Takeaway:

Multiplexed transports usually want a command queue plus one writer owner.

## CockroachDB stopper

Why it is worth reading:

CockroachDB makes goroutine lifetime a first-class service boundary instead of hoping background tasks stop eventually.

Primary source:

- [`Stopper` state](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L152-L214)
- [`WithCancelOnQuiesce`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L272-L289)
- [`RunAsyncTask` and `RunAsyncTaskEx`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L358-L446)
- [`Stop` and `Quiesce`](https://github.com/cockroachdb/cockroach/blob/41c5df3d4a4cd26e8852677aebb513184e78ccd9/pkg/util/stop/stopper.go#L562-L655)

Main concurrency boundary:

The stopper owns task admission, task counting, quiesce signaling, and shutdown waiting.

Simplified shape:

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

Why this works:

Background work is not free-floating. It has admission control, observability, and a coordinated shutdown path.

Failure pattern to avoid:

Do not reduce this pattern to “a fancy wait group.” The important design move is refusing new work while quiescing and giving tasks cancellation that is tied to system shutdown.

Takeaway:

Large services need a first-class owner for goroutine lifetime, not just `context.Background()` and hope.

## go-redis connection pool

Why it is worth reading:

Even though this is a library, it is a clean study in admission control and handoff under load.

Primary source:

- [`ConnPool` state](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L306-L339)
- [`Get`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L701-L819)
- [`queuedNewConn`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L821-L897)
- [`waitTurn`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L954-L978)
- [`Put`](https://github.com/redis/go-redis/blob/6193e530d612d362359a6709b46dc2c159e1455d/internal/pool/pool.go#L1049-L1078)

Main concurrency boundary:

Callers do not dial unboundedly. They first acquire permission from the pool, then either reuse an idle connection or join the path that dials new ones.

Simplified shape:

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

Why this works:

The pool has an explicit admission boundary. Once saturated, callers wait or fail according to budget instead of spawning unlimited dials.

Failure pattern to avoid:

Do not copy a pool and leave dialing outside the same admission boundary. If every caller can still open its own connection on timeout, you have not actually bounded anything.

Takeaway:

Real pools are admission controllers, not just idle-connection caches.

## Common misread

A common misread of production Go code is to copy the goroutines and forget the boundary around them:

```go
for _, msg := range batch {
	go process(msg) // bad: no queue, no limit, no shutdown contract
}
```

The systems above work because they wrap concurrency in queue state, ownership loops, admission control, flow control, and shutdown policy.

## Practical takeaway

Reading open-source Go code is most useful when you read it for boundaries, not just APIs.

The strongest projects make concurrency policy visible in queue design, lifecycle loops, signaling edges, backpressure, and shutdown paths.
