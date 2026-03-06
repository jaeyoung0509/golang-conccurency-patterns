---
title: 테스트 개요
description: race, leak, time, cancellation, scheduler behavior까지 Go 동시성 테스트를 어떻게 구성할지 설명합니다.
---

# 테스트 개요

동시성 코드에 테스트 전략이 없으면, 대부분은 그냥 운에 기대는 코드입니다.

어려운 버그는 대개 문법 에러가 아니라:

- 빠진 happens-before edge,
- shutdown leak,
- 잘못된 시간 가정,
- cancellation race,
- 부하에서만 드러나는 contention

같은 것들입니다.

## 이 섹션이 다루는 것

| 이런 걸 검증하고 싶다면 | 여기부터 |
| --- | --- |
| AI가 만든 Go 코드가 컴파일은 되는데 안전한지 검증 | [AI 보조 Go 안전성](/ko/testing/ai-assisted-go-safety) |
| 동기화 없는 shared memory 접근 | [Race Detector](/ko/testing/race-detector) |
| timeout/timer 로직을 실제 sleep 없이 검증 | [synctest로 결정적 테스트](/ko/testing/synctest) |
| Postgres, Redis 같은 실제 의존성과 container-backed 테스트를 teardown까지 깔끔하게 검증 | [Testcontainers로 통합 테스트하기](/ko/testing/integration-testcontainers) |
| goroutine leak와 shutdown behavior | [리크, 종료, 타임아웃 테스트](/ko/testing/leaks-and-shutdowns) |
| scheduler, blocking, GC, contention behavior | [트레이싱과 경합 관측](/ko/testing/tracing-and-profiling) |

## 테스트 철학

좋은 동시성 테스트는 보통 이런 걸 증명해야 합니다.

1. 프로토콜이 안전한가
2. 실패 경로가 끝나는가
3. timeout 경로가 결정적인가
4. 구현이 goroutine이나 작업을 leak하지 않는가
5. AI/LLM이 만든 회귀를 조기에 잡는 검증 체인이 있는가

## 추천 읽기 순서

1. [Race Detector](/ko/testing/race-detector)
2. [AI 보조 Go 안전성](/ko/testing/ai-assisted-go-safety)
3. [synctest로 결정적 테스트](/ko/testing/synctest)
4. [Testcontainers로 통합 테스트하기](/ko/testing/integration-testcontainers)
5. [리크, 종료, 타임아웃 테스트](/ko/testing/leaks-and-shutdowns)
6. [트레이싱과 경합 관측](/ko/testing/tracing-and-profiling)

## Practical takeaway

동시성 테스트는 한 가지 도구로 끝나지 않습니다.

- compile/unit test와 정적 검사를 먼저 두고
- `-race`
- `synctest`
- `testcontainers-go`와 typed fixture로 실제 의존성 계약을 검증하고
- leak / shutdown 테스트
- trace / profile

이 스택을 같이 써야 실제 품질이 올라갑니다.
