---
title: Go 함정 부록
description: 컴파일은 되지만 프로덕션에서 문제를 만드는 Go 함정 50가지를 정리한 부록입니다.
---

# Go 함정 부록

이 부록은 다음 같은 코드를 다룹니다.

- 컴파일은 되고
- 대충 보면 리뷰도 통과하고
- 운영에 들어가면 크게 문제를 일으키는 코드

목표는 문법 입문이 아닙니다. 숙련된 Go 엔지니어도 일정 압박 아래에서 자주 만드는 실수를 더 날카롭게 구분하는 데 있습니다.

## 무엇을 다루나

| 범주 | 초점 |
| --- | --- |
| [Control Flow and Evaluation](/ko/extras/go-pitfalls/control-flow-and-evaluation) | `select`, `break`, loop variable, `defer`, 평가 시점처럼 겉보기엔 단순하지만 자주 틀리는 지점 |
| [Interfaces and Types](/ko/extras/go-pitfalls/interfaces-and-types) | typed nil, type assertion, method set, wrapped error, `any`가 만드는 모호함 |
| [Collections and Memory](/ko/extras/go-pitfalls/collections-and-memory) | map/slice 함정, iteration 가정, backing array retention, 메모리 형태 관련 놀라움 |
| [Concurrency, Context, and Time](/ko/extras/go-pitfalls/concurrency-context-and-time) | channel ownership, goroutine lifetime, cancellation, timer, `WaitGroup` 순서 |
| [Stdlib and API Boundaries](/ko/extras/go-pitfalls/stdlib-and-api-boundaries) | `net/http`, `database/sql`, `io`, `json`, `time`, `os/exec`, `sync.Map` 경계에서 생기는 footgun |

## 각 항목을 읽는 법

모든 함정은 같은 형식을 따릅니다.

1. 나쁜 코드
2. 왜 깨지거나 오해를 부르는지
3. 더 나은 코드
4. 어떻게 빨리 잡는지
5. 한 줄 규칙

언어 튜토리얼처럼 읽기보다, 코드 리뷰 체크리스트처럼 읽는 편이 맞습니다.

## 더 깊게 보고 싶다면

이 부록은 의도적으로 짧고 날카롭게 썼습니다. 더 큰 런타임/운영 주제로 이어지는 항목은 이미 사이트의 다른 문서에서 깊게 다룹니다.

- [Race Detector](/ko/testing/race-detector)
- [synctest로 결정적 테스트](/ko/testing/synctest)
- [context 패키지 내부](/ko/stdlib/context-internals)
- [time, Timers, Tickers](/ko/stdlib/time-timers-tickers)
- [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport)
- [sync와 atomic 프리미티브](/ko/stdlib/sync-and-atomic)
- [Channels, Select, 그리고 Memory Model](/ko/fundamentals/channels-memory-model)

## 현대 Go 기준으로 한 가지 주의점

Go 함정을 다루는 오래된 글과 발표 중 일부는 이제 완전히 맞지는 않습니다.

대표적으로 loop variable capture는 Go 1.22에서 흔한 `for ... := range ...` 패턴이 개선됐습니다. 이 부록은 현대 Go 동작을 기준으로 쓰되, 오래된 코드나 predeclared loop variable에서 여전히 남는 함정은 따로 짚습니다.

## Practical takeaway

아주 아픈 Go 버그는 대개 희귀한 마법에서 나오지 않습니다.

대부분은 API 경계, ownership 경계, evaluation 경계에 대한 작은 오해에서 나옵니다. 이 부록은 바로 그 지점을 조여 주기 위해 존재합니다.
