---
title: Production Playbooks
description: Operator-oriented guides for the Go libraries and client packages that define real production behavior.
---

# Production Playbooks

The standard-library section explains how the packages work.

This section explains how to bet production traffic on them without guessing.

Every page in this section is organized around:

- the mental model you need before touching config,
- a safe default sketch,
- the failure patterns that actually hurt teams,
- observability and testing hooks,
- when the tool is a good fit and when it is not.

## What this section covers

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/playbooks/net-http-production-field-guide">net/http</a></h3>
    <p>Turn transport reuse, timeout boundaries, body ownership, and streaming caveats into explicit operating rules.</p>
  </div>
  <div class="path-card">
    <h3><a href="/playbooks/grpc-go-production-playbook">grpc-go</a></h3>
    <p>Use long-lived channels, RPC deadlines, keepalive discipline, and streaming ownership without falling into `Dial` and `WithBlock` traps.</p>
  </div>
  <div class="path-card">
    <h3><a href="/playbooks/go-redis-production-playbook">go-redis</a></h3>
    <p>Understand pooling, timeout budgets, RESP2/RESP3 choices, pipelines, and instrumentation for cache-heavy services.</p>
  </div>
  <div class="path-card">
    <h3><a href="/playbooks/kafka-with-ibm-sarama">Kafka with IBM Sarama</a></h3>
    <p>Operate consumer groups, producers, acks, offsets, and rebalance loops with a client model that matches Kafka's actual semantics.</p>
  </div>
</div>

## Suggested reading order

1. Start with [net/http](/playbooks/net-http-production-field-guide), because almost every Go service depends on its lifetime and timeout model.
2. Continue with [grpc-go](/playbooks/grpc-go-production-playbook), where connection reuse, deadlines, streaming, and keepalive policy become more explicit.
3. Read [go-redis](/playbooks/go-redis-production-playbook) next if you operate cache-heavy request paths or background workers.
4. Finish with [Kafka with IBM Sarama](/playbooks/kafka-with-ibm-sarama) when you need durable event flow, consumer-group ownership, and offset discipline.

## What you should be able to answer afterward

- When is `http.Client.Timeout` useful, and when does it blur too many edges together?
- Why should a gRPC `ClientConn` live much longer than one RPC?
- Why can aggressive gRPC keepalive settings make servers push back?
- Why is a go-redis client primarily a pooled resource owner, not just a command helper?
- When should you keep RESP2 even if RESP3 exists?
- Why is `Config.Version` in Sarama an operating requirement, not optional decoration?
- Why does a synchronous Kafka producer still not imply end-to-end exactly once?
- Why does offset marking have to follow side effects, not lead them?

## Practical takeaway

These playbooks are where language knowledge becomes service policy.

They are less about calling the API correctly and more about keeping the system honest under traffic, retries, shutdown, and partial failure.
