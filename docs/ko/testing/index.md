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
| 동기화 없는 shared memory 접근 | [Race Detector](/ko/testing/race-detector) |
| timeout/timer 로직을 실제 sleep 없이 검증 | [synctest로 결정적 테스트](/ko/testing/synctest) |
| goroutine leak와 shutdown behavior | [리크, 종료, 타임아웃 테스트](/ko/testing/leaks-and-shutdowns) |
| scheduler, blocking, GC, contention behavior | [트레이싱과 경합 관측](/ko/testing/tracing-and-profiling) |

## 테스트 철학

좋은 동시성 테스트는 보통 이런 걸 증명해야 합니다.

1. 프로토콜이 안전한가
2. 실패 경로가 끝나는가
3. timeout 경로가 결정적인가
4. 구현이 goroutine이나 작업을 leak하지 않는가

## 추천 읽기 순서

1. [Race Detector](/ko/testing/race-detector)
2. [synctest로 결정적 테스트](/ko/testing/synctest)
3. [리크, 종료, 타임아웃 테스트](/ko/testing/leaks-and-shutdowns)
4. [트레이싱과 경합 관측](/ko/testing/tracing-and-profiling)

## Practical takeaway

동시성 테스트는 한 가지 도구로 끝나지 않습니다.

- `-race`
- `synctest`
- leak / shutdown 테스트
- trace / profile

이 스택을 같이 써야 실제 품질이 올라갑니다.
