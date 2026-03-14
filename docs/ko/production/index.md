---
title: 프로덕션 개요
description: Go 코드가 실제 대규모 트래픽을 받을 때 중요해지는 동시성 관점을 다룹니다.
---

# 프로덕션 개요

앞선 섹션은 Go 동시성이 왜 동작하는지, 각 패턴을 언제 써야 하는지를 설명합니다.

이 섹션은 코드가 예제를 넘어 실제 트래픽을 받기 시작했을 때 무엇이 달라지는지를 다룹니다.

## 무엇을 다루나

| 주제 | 왜 중요한가 |
| --- | --- |
| [대규모 Go 시스템](/ko/production/large-scale-go-systems) | 런타임 지식을 실제 latency, memory, overload, shutdown 규칙으로 바꿉니다 |
| [Docker, containerd, 그리고 Kubernetes](/ko/production/docker-containerd-kubernetes) | Go daemon, runtime core, controller loop가 현대 컨테이너 플랫폼을 어떻게 이루는지 설명합니다 |
| [규제 환경의 Go 시스템](/ko/production/regulated-systems) | event sourcing, cryptographic erase, retention, audit 제약이 Go 시스템 설계를 어떻게 바꾸는지 설명합니다 |
| [Temporal과 Durable Execution](/ko/production/temporal-durable-execution) | 현대 workflow 엔진이 history shard, task queue, worker polling을 가진 Go 시스템으로 어떻게 구성되는지 보여줍니다 |
| [오픈소스 사례](/ko/production/open-source-case-studies) | Kubernetes, etcd, Prometheus, NATS, gRPC-Go, CockroachDB, go-redis가 concurrency policy를 어떻게 코드에 드러내는지 봅니다 |
| [Go 오픈소스 역사 읽기](/ko/production/go-open-source-histories) | 왜 Go가 인프라 소프트웨어에 강했는지, 주요 프로젝트의 공개 시점과 형태를 통해 읽습니다 |

## 프로덕션에서 질문이 달라진다

예제에서는 보통 "이 패턴이 맞는가?"가 핵심입니다.

프로덕션에서는 질문이 이렇게 바뀝니다.

- 이 goroutine의 lifetime owner는 누구인가
- 어디서 load를 reject할 것인가
- deploy와 shutdown 때 어떤 일이 일어나는가
- heap pressure 아래에서 runtime이 어떻게 반응하는가
- tail latency를 만드는 hot lock / hot queue는 무엇인가
- 이걸 추측이 아니라 관측으로 어떻게 볼 것인가

## Practical takeaway

fundamentals가 Go 동시성이 왜 가능한지 설명한다면, production 섹션은 그 장점을 운영에서 잃지 않는 법을 설명합니다.
