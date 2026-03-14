---
title: 비교 / 확장 개요
description: Go 바깥의 런타임과 비교해 mental model을 넓히는 보조 섹션입니다.
---

# 비교 / 확장 개요

이 사이트의 중심은 Go입니다.

하지만 extra 섹션은 두 가지 이유로 필요합니다.

- Go의 선택을 다른 생태계와 비교하기 위해
- 숙련된 엔지니어가 다른 런타임 mental model을 Go와 연결하기 위해

## 현재 주제

| 주제 | 왜 중요한가 |
| --- | --- |
| [Go 함정 부록](/ko/extras/go-pitfalls/) | 컴파일은 되지만 운영에서 문제를 만드는 Go 실수 50가지를 bad/better 코드와 함께 체크리스트로 정리합니다 |
| [Go CSP vs Rust Tokio](/ko/extras/go-csp-vs-rust-tokio) | 비슷한 목표를 가진 두 시스템이 왜 전혀 다른 설계를 택했는지 보여줍니다 |
| [Go vs Rust 결정 가이드](/ko/extras/go-vs-rust-decision-guide) | 서비스를 Go에 두고 특정 서브시스템만 Rust로 옮겨야 하는지 판단 기준을 제공합니다 |

## 추천 읽기 순서

1. 먼저 [Go 함정 부록](/ko/extras/go-pitfalls/)으로 컴파일은 되지만 위험한 실수들을 가로질러 훑습니다.
2. 그다음 [Go CSP vs Rust Tokio](/ko/extras/go-csp-vs-rust-tokio)로 동시성 모델 차이를 비교합니다.
3. 마지막으로 [Go vs Rust 결정 가이드](/ko/extras/go-vs-rust-decision-guide)에서 언어 선택을 실제 설계 질문으로 가져갑니다.

## Practical takeaway

다른 생태계를 모르는 척한다고 Go를 더 잘 이해하게 되지는 않습니다.

오히려 Go가 무엇을 언어/런타임에 넣었고, Rust + Tokio가 무엇을 library/runtime 조합으로 풀었는지 비교할 때 Go의 선택이 더 선명해집니다.
