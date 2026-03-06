---
title: Kafka with IBM Sarama
description: IBM Sarama로 Kafka를 운영할 때 producer durability, consumer-group lifecycle, offset discipline을 어떻게 잡아야 하는지 설명합니다.
---

# Kafka with IBM Sarama

Kafka와 Sarama는 한 주제로 다루는 편이 맞습니다.

client API는 Kafka의 실제 semantics, 즉 acks, rebalance, offset, broker-version compatibility를 존중할 때만 의미가 있기 때문입니다.

## Mental model

아래 세 경계를 분리해서 봐야 합니다.

- producer durability policy
- consumer-group session lifecycle
- side effect와 offset ownership의 순서

많은 Kafka 장애는 이 셋을 그냥 “보내기”와 “읽기”로 뭉개는 순간 시작됩니다.

## 안전한 producer 스케치

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

이 설정만으로 exactly once가 되지는 않지만, durability 기준선으로는 괜찮습니다.

- 명시적인 broker version
- 가장 강한 일반 ack 모드
- idempotent producer 모드
- Sarama가 요구하는 return channel 설정

## 안전한 consumer-group 스케치

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

바깥 `for` 루프가 중요합니다. consumer-group session은 rebalance마다 끝나고 다시 시작됩니다. `Consume`를 한 번만 부르는 setup 함수처럼 보면 안 됩니다.

## 운영 규칙

### `Config.Version`을 명시한다

공식 FAQ는 일부 기능의 동작이 `Config.Version`에 달려 있다고 명시합니다. cluster capability를 흐리게 두면, 오래된 동작으로 fallback하거나 새 기능을 예상 밖에 거부할 수 있습니다.

### synchronous producer가 만능 durability 스위치는 아니다

공식 패키지 문서는 `SyncProducer`가 대체로 `AsyncProducer`보다 비효율적이고, acknowledged message도 `Producer.RequiredAcks`에 따라 여전히 유실될 수 있다고 설명합니다.

### async producer는 channel discipline이 필요하다

`Producer.Return.Successes`를 켰다면 `Successes()`를 반드시 읽어야 하고, 그렇지 않으면 producer가 deadlock될 수 있습니다. `Errors()`도 마찬가지입니다.

### consumer group은 endless loop가 아니라 session 단위다

`ConsumerGroup` 문서는 `Consume`를 blocking session lifecycle로 설명합니다. rebalance, handler setup/cleanup, claim 처리까지 전부 session의 일부입니다.

### offset은 side effect 뒤에 와야 한다

DB write나 downstream publish가 아직 durable하지 않은데 offset을 먼저 마킹하면, 시스템은 아직 끝나지 않은 일을 끝났다고 선언한 셈이 됩니다.

## 실패 패턴

### Kafka version을 명시하지 않음

```go
config := sarama.NewConfig() // bad: cluster capability가 암묵적
```

### sync producer를 “충분히 안전하다”로 오해

```go
config.Producer.RequiredAcks = sarama.WaitForLocal // 생각보다 약한 설정
producer, _ := sarama.NewSyncProducer(brokers, config)
```

### channel을 비우지 않는 async producer

```go
producer, _ := sarama.NewAsyncProducer(brokers, config)
producer.Input() <- msg // bad: Successes/Errors를 비우는 goroutine이 없음
```

### rejoin loop가 없는 consumer group

```go
_ = group.Consume(ctx, topics, handler) // bad: 한 세션만 돌고 끝
```

### side effect 전에 offset mark

```go
sess.MarkMessage(msg, "")
err := writeToDB(msg) // bad: offset이 먼저 움직임
```

## 주의해서 써야 할 것

- throughput이 중요한데도 무조건 `SyncProducer`
- idempotent handler 없이 높은 retry count
- crash/rebalance 테스트 없이 믿는 auto-commit 가정
- rebalance/session timing을 길게 잡아먹는 handler critical section

## Observability와 테스트

- consumer lag, rebalance count, rebalance duration, handler error rate를 봅니다.
- producer error rate는 application error rate와 분리해서 봅니다.
- offset commit과 side effect 사이 crash timing은 mock이 아니라 통합 테스트로 검증합니다.
- partition skew를 봐야 합니다. hot partition은 handler가 느린 것처럼 보이게 만듭니다.

## 언제 잘 맞는가

Go에서 event durability, replay, partitioned throughput, consumer-group coordination이 필요하면 Kafka + Sarama는 잘 맞습니다.

## 언제 다른 도구를 봐야 하나

request-response RPC, ultra-low-latency cache access, 작은 in-process job dispatch라면 Kafka는 클라이언트 품질과 별개로 맞지 않는 추상화일 가능성이 큽니다.

## 공식 자료

- [`github.com/IBM/sarama` 패키지 문서](https://pkg.go.dev/github.com/IBM/sarama)
- [IBM Sarama FAQ](https://github.com/IBM/sarama/wiki/Frequently-Asked-Questions)

## Practical takeaway

Kafka correctness는 Sarama를 import하는 것만으로 오지 않습니다.

version, acks, consumer-session lifecycle, offset-after-side-effect 정책을 서비스에 명시해야만 안전해집니다.
