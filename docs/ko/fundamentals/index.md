---
title: 기초 원리 개요
description: 패턴을 외우기 전에 왜 Go 동시성이 이런 식으로 동작하는지 이해하고 싶다면 여기서 시작합니다.
---

# 기초 원리 개요

<div class="lead-panel">
  <p>
    기초 원리 섹션은 <strong>왜 이런 패턴이 가능한가</strong>를 설명하는 파트입니다.
    이 부분을 건너뛰어도 코드는 복사할 수 있지만, 운영 중 문제를 디버깅하기는 훨씬 어려워집니다.
  </p>
</div>

## 이 섹션이 다루는 것

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/ko/fundamentals/go-runtime-scheduler">런타임과 스케줄러</a></h3>
    <p>G, M, P, run queue, netpoll, sysmon, stack growth까지 이해해서 고루틴의 실제 비용 구조를 잡습니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/fundamentals/channels-memory-model">채널과 메모리 모델</a></h3>
    <p>동기화 보장이 어디서 생기는지 보고, 채널이 단순 coordination이 아니라 correctness 경계가 되는 이유를 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/fundamentals/channel-internals">채널 내부 동작</a></h3>
    <p>`hchan`, `sudog`, direct handoff, queueing, `close`, `select`가 실제로 어떻게 엮여 있는지 봅니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/fundamentals/mutex-semaphore-internals">Mutex와 런타임 세마포어</a></h3>
    <p>`sync.Mutex`의 fast path, starvation mode, contention, wakeup 메커니즘까지 이해합니다.</p>
  </div>
</div>

## 추천 읽기 순서

1. [Go 런타임과 스케줄러](/ko/fundamentals/go-runtime-scheduler)
2. [Channels, Select, 그리고 Memory Model](/ko/fundamentals/channels-memory-model)
3. [채널 내부 동작](/ko/fundamentals/channel-internals)
4. [Mutex와 런타임 세마포어 내부](/ko/fundamentals/mutex-semaphore-internals)

## 이 섹션을 읽고 답할 수 있어야 하는 질문

- 왜 `GOMAXPROCS`는 CPU 병렬성만 바꾸고 외부 서비스 안전성은 보장하지 않는가?
- 왜 channel send/receive가 메모리 가시성 경계를 만들 수 있는가?
- 왜 `select`는 유용하지만 그 자체로 correctness proof는 아닌가?
- 왜 어떤 코드에서는 mutex가 가장 Go다운 선택이 될 수 있는가?
