---
title: Kafka with IBM Sarama
description: Learn how to operate Kafka from Go with IBM Sarama across producer durability, consumer-group lifecycle, and offset discipline.
---

# Kafka with IBM Sarama

Kafka and Sarama belong in one topic.

The client API only makes sense if you respect Kafka's actual semantics around acks, rebalances, offsets, and broker-version compatibility.

## Mental model

There are three separate boundaries to keep straight:

- producer durability policy,
- consumer-group session lifecycle,
- offset ownership relative to your side effects.

Most Kafka incidents happen when a team collapses those three into “send messages” and “read messages.”

## Safe producer sketch

```go
config := sarama.NewConfig()
config.Version = sarama.V3_6_0_0
config.Producer.RequiredAcks = sarama.WaitForAll
config.Producer.Idempotent = true
config.Producer.Return.Successes = true
config.Producer.Return.Errors = true

producer, err := sarama.NewSyncProducer(brokers, config)
if err != nil {
	return err
}
defer producer.Close()
```

This sketch is not enough for exactly once by itself, but it is a sane durability baseline:

- explicit broker version,
- strongest normal ack mode,
- idempotent producer mode,
- the return channels enabled the way Sarama requires.

## Safe consumer-group sketch

```go
config := sarama.NewConfig()
config.Version = sarama.V3_6_0_0
config.Consumer.Return.Errors = true
config.Consumer.Offsets.Initial = sarama.OffsetNewest
config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{
	sarama.NewBalanceStrategyRange(),
}

group, err := sarama.NewConsumerGroup(brokers, "payments", config)
if err != nil {
	return err
}
defer group.Close()

for {
	if err := group.Consume(ctx, topics, handler); err != nil {
		return err
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
}
```

The outer `for` loop matters. Consumer-group sessions end and restart across rebalances. Treat `Consume` as a session boundary, not as a one-time setup call.

## Operating rules

### Set `Config.Version` deliberately

The official Sarama FAQ is explicit that some features change behavior based on `Config.Version`. If you leave cluster capability vague, the client can fall back to older behavior or reject newer features unexpectedly.

### A synchronous producer is not a magic durability switch

The official package docs say `SyncProducer` is generally less efficient than `AsyncProducer`, and even acknowledged messages can still be lost depending on `Producer.RequiredAcks`.

### Async producers require channel discipline

If you enable `Producer.Return.Successes`, you must read the `Successes()` channel or the producer can deadlock. You must also read `Errors()` or disable returned errors explicitly.

### Consumer groups are session-based, not endless loops

The `ConsumerGroup` docs describe `Consume` as a blocking session lifecycle. Rebalances, handler setup, cleanup, and claims are part of that lifecycle.

### Offsets follow side effects

If you mark offsets before your database write, RPC, or downstream publish is actually durable, you have acknowledged work the system has not really completed.

## Failure patterns

### No explicit Kafka version

```go
config := sarama.NewConfig() // bad: cluster capability left implicit
```

### Sync producer assumed to mean “safe enough”

```go
config.Producer.RequiredAcks = sarama.WaitForLocal // weaker than many teams realize
producer, _ := sarama.NewSyncProducer(brokers, config)
```

### Async producer without draining channels

```go
producer, _ := sarama.NewAsyncProducer(brokers, config)
producer.Input() <- msg // bad: no goroutine draining Successes/Errors
```

### Consumer group without rejoin loop

```go
_ = group.Consume(ctx, topics, handler) // bad: one session, no rebalance loop
```

### Marking offsets before the real side effect

```go
sess.MarkMessage(msg, "")
err := writeToDB(msg) // bad: offset moved first
```

## What to use carefully

- `SyncProducer` when throughput matters more than synchronous call style
- high retry counts without idempotent handlers
- auto-commit assumptions you have never tested under crash and rebalance
- large handler critical sections that extend rebalance or session timing

## Observability and testing

- Track consumer lag, rebalance count, rebalance duration, and handler error rates.
- Track producer error rate separately from application error rate.
- Test crash timing around offset commit and side effects with integration tests, not only mocks.
- Watch partition skew. Hot partitions can make consumer-group health look like handler slowness.

## When it is the right tool

Kafka with Sarama is a good fit when you need event durability, replay, partitioned throughput, and consumer-group coordination from Go.

## When it is not the right tool

If the workload is request-response RPC, ultra-low-latency cache access, or tiny in-process job dispatch, Kafka is often the wrong abstraction regardless of how good the client is.

## Official reading

- [Package docs for `github.com/IBM/sarama`](https://pkg.go.dev/github.com/IBM/sarama)
- [IBM Sarama FAQ](https://github.com/IBM/sarama/wiki/Frequently-Asked-Questions)

## Practical takeaway

Kafka correctness does not come from importing Sarama.

It comes from making version, acks, consumer-session lifecycle, and offset-after-side-effect policy explicit in your service.
