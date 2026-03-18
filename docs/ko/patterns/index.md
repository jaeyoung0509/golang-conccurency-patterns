---
title: 패턴 개요
description: 문제 형태, 실패 정책, 소유권 모델에 맞춰 어떤 Go 동시성 패턴부터 읽을지 고르는 페이지입니다.
---

# 패턴 개요

<div class="lead-panel">
  <p>
    실전 패턴 섹션은 추상 분류보다 <strong>문제의 형태</strong>를 기준으로 구성했습니다.
    병렬 수 제한이 필요한지, 단계 분리가 중요한지, 집계가 핵심인지, 요청 수명 주기가 중요한지부터 생각하면 됩니다.
  </p>
</div>

## 문제 형태로 패턴 고르기

| 문제가 이런 모습이라면 | 여기부터 읽기 |
| --- | --- |
| 독립 작업이 많고 동시 실행 수를 제한해야 함 | [워커 풀](/ko/patterns/worker-pool) |
| 데이터가 여러 단계를 순서 있게 통과함 | [파이프라인](/ko/patterns/pipeline) |
| 하나의 요청이 여러 백엔드 응답을 동시에 모아야 함 | [팬아웃 / 팬인](/ko/patterns/fan-out-fan-in) |
| 여러 goroutine이 하나의 요청 lifetime에 묶여 있음 | [컨텍스트 취소](/ko/patterns/context-cancellation) |
| 하나의 broker가 많은 caller를 받지만 각 caller가 자기 reply path를 가져야 함 | [Channel of Channels](/ko/patterns/channel-of-channels) |
| 응답 DTO와 PATCH 입력에서 absent, null, concrete value를 안전하게 구분해야 함 | [API 경계에서의 Optional 값 패턴](/ko/patterns/optional-values-across-boundaries) |
| 프로세스가 종료될 때 이미 받은 일을 함부로 버리면 안 됨 | [Graceful Shutdown](/ko/patterns/graceful-shutdown) |
| channel-heavy pipeline을 cancellation-safe하게 조합해야 함 | [Or-Done, Tee, Bridge](/ko/patterns/or-done-tee-bridge) |

## 이 예제들이 실용적인 이유

<div class="signal-strip">
  <div class="signal">
    <strong>실제 도메인</strong>
    <span>배송, 부정거래 점검, 재고, 요청 집계 같은 현실적인 예제를 씁니다.</span>
  </div>
  <div class="signal">
    <strong>테스트된 동작</strong>
    <span>제한, 순서 보장, 취소, 부분 실패 정책을 테스트로 검증합니다.</span>
  </div>
  <div class="signal">
    <strong>트레이드오프 설명</strong>
    <span>언제 맞는 패턴인지뿐 아니라 언제 틀린 추상화인지도 같이 설명합니다.</span>
  </div>
  <div class="signal">
    <strong>수명 관리 포함</strong>
    <span>shutdown, timeout, reply ownership, leak prevention을 패턴의 일부로 같이 설명합니다.</span>
  </div>
</div>

## 추천 읽는 순서

1. 패턴 문서를 읽습니다.
2. `examples/`의 해당 패키지를 엽니다.
3. 코드를 복사하기 전에 테스트를 먼저 읽습니다.
4. 자기 서비스에 필요한 실패 정책과 순서 보장 계약을 따로 결정합니다.
