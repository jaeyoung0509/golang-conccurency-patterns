---
title: 실전 플레이북
description: 실제 프로덕션 동작을 좌우하는 Go 라이브러리와 클라이언트 패키지를 운영 관점에서 설명합니다.
---

# 실전 플레이북

표준 라이브러리 섹션이 패키지가 어떻게 동작하는지 설명했다면,

이 섹션은 그 패키지와 라이브러리에 실제 트래픽을 어떻게 안전하게 태울지를 설명합니다.

이 섹션의 각 페이지는 공통적으로 다음을 다룹니다.

- config를 만지기 전에 필요한 mental model,
- 안전한 기본 설정 스케치,
- 팀을 실제로 아프게 하는 failure pattern,
- observability와 testing 포인트,
- 언제 잘 맞고 언제 아닌지.

## 무엇을 다루나

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/ko/playbooks/net-http-production-field-guide">net/http</a></h3>
    <p>transport reuse, timeout boundary, body ownership, streaming caveat를 운영 규칙으로 바꿉니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/playbooks/grpc-go-production-playbook">grpc-go</a></h3>
    <p>long-lived channel, RPC deadline, keepalive discipline, streaming ownership을 `Dial`/`WithBlock` 함정 없이 다룹니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/playbooks/go-redis-production-playbook">go-redis</a></h3>
    <p>pooling, timeout budget, RESP2/RESP3 선택, pipeline, instrumentation을 cache-heavy 서비스 관점에서 정리합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/playbooks/kafka-with-ibm-sarama">Kafka with IBM Sarama</a></h3>
    <p>consumer group, producer, acks, offset, rebalance loop를 Kafka의 실제 semantics에 맞게 운영하는 법을 설명합니다.</p>
  </div>
</div>

## 추천 읽기 순서

1. 먼저 [net/http](/ko/playbooks/net-http-production-field-guide)를 읽습니다. 거의 모든 Go 서비스가 이 lifetime/timeout 모델 위에 있기 때문입니다.
2. 다음으로 [grpc-go](/ko/playbooks/grpc-go-production-playbook)를 읽습니다. 여기서는 connection reuse, deadline, streaming, keepalive policy가 더 명시적으로 드러납니다.
3. 그 다음 [go-redis](/ko/playbooks/go-redis-production-playbook)를 읽습니다. cache-heavy request path나 background worker를 운영한다면 특히 중요합니다.
4. 마지막으로 [Kafka with IBM Sarama](/ko/playbooks/kafka-with-ibm-sarama)를 읽습니다. durable event flow, consumer-group ownership, offset discipline이 필요한 경우입니다.

## 읽고 나면 답할 수 있어야 하는 질문

- `http.Client.Timeout`은 언제 유용하고, 언제 너무 많은 경계를 한 번에 뭉개는가?
- 왜 gRPC `ClientConn`은 한 RPC보다 훨씬 오래 살아야 하는가?
- 왜 공격적인 gRPC keepalive 설정이 서버의 역반응을 부를 수 있는가?
- 왜 go-redis client는 command helper가 아니라 pooled resource owner로 봐야 하는가?
- RESP3가 있어도 왜 RESP2를 유지하는 편이 더 안전한 경우가 있는가?
- 왜 Sarama의 `Config.Version`은 장식이 아니라 운영 요구사항인가?
- 왜 synchronous Kafka producer도 end-to-end exactly once를 의미하지 않는가?
- 왜 offset mark는 side effect 뒤에 와야 하지 앞서면 안 되는가?

## Practical takeaway

이 플레이북은 언어/런타임 지식을 서비스 정책으로 바꾸는 구간입니다.

API 호출법보다, 트래픽, 재시도, 종료, 부분 실패 아래에서 시스템을 어떻게 정직하게 유지할지가 핵심입니다.
