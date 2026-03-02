---
title: 고급 주제 개요
description: 구조화된 수명 관리, 자원 예산, 중복 억제, 소유권 모델, 과부하 제어 같은 고급 Go 동시성 주제를 안내합니다.
---

# 고급 주제 개요

<div class="lead-panel">
  <p>
    고급 주제는 “이걸 병렬로 돌릴 수 있나?”를 넘어서
    <strong>수명 관리, 자원 예산, 상태 소유권, 과부하에서의 생존</strong>을 다루는 섹션입니다.
  </p>
</div>

## 학습 지도

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/ko/advanced/structured-concurrency">구조화된 동시성</a></h3>
    <p>`errgroup`과 context로 goroutine lifetime을 부모 작업 경계 안에 묶습니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/advanced/weighted-semaphore">가중 세마포어</a></h3>
    <p>단순한 goroutine 수 대신 메모리나 quota weight 기준으로 admission을 제어합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/advanced/singleflight">Singleflight</a></h3>
    <p>같은 키에 대한 중복 in-flight 작업을 합쳐 cache miss herd를 줄입니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/advanced/actor-pattern">액터 패턴</a></h3>
    <p>상태 기계를 하나의 명확한 owner 주위로 직렬화합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/advanced/csp-theory">CSP 이론</a></h3>
    <p>Go가 빌려온 communication-first 사고방식과, Go가 의도적으로 다르게 만든 부분을 이해합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/advanced/backpressure-load-shedding">역압력과 로드 셰딩</a></h3>
    <p>bounded queue, 거절 정책, degrade 전략으로 시스템을 무너지지 않게 유지합니다.</p>
  </div>
</div>

## 언제 이 섹션으로 오면 좋은가

- goroutine, channel, cancellation 기본은 이미 이해한 상태
- happy path correctness보다 운영 중 behavior를 설계해야 하는 상태
- ownership-first 설계와 communication-first 설계 중 무엇이 맞는지 판단해야 하는 상태
