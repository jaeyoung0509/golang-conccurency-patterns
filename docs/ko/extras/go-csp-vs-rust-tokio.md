---
title: Go CSP vs Rust Tokio
description: Go의 goroutine/channel 모델과 Rust Tokio의 future/executor 모델을 비교합니다.
---

# Go CSP vs Rust Tokio

Go와 Rust는 둘 다 고성능 concurrent system을 만들 수 있지만, 거기에 도달하는 방식은 꽤 다릅니다.

이 문서는 누가 더 좋다를 말하려는 것이 아니라, trade를 명확히 하려는 것입니다.

:::tip Quick takeaway
Go는 goroutine, channel, `select`, netpoll integration, stackful execution처럼 더 많은 concurrency machinery를 언어와 런타임에 넣었습니다. Rust + Tokio는 future, executor, trait, async state machine, library-level composition 쪽에 더 많은 무게를 둡니다.
:::

## 가장 짧은 비교

| 질문 | Go | Rust + Tokio |
| --- | --- | --- |
| 동시성 단위 | goroutine | future / task |
| 실행 방식 | stackful goroutine | `async`가 만든 stackless state machine |
| 스케줄러 | runtime 내장 | Tokio executor |
| I/O 통합 | runtime netpoller | Tokio reactor / resource driver |
| 조합 방식 | channel과 `select`가 1급 | channel, `select!`, await 조합이 라이브러리 쪽 |
| cancellation | `context`, channel close | drop, abort, cancellation token, `select!` |

## Go에서 CSP가 어떻게 드러나나

Go는 CSP에 가까운 도구를 언어 표면에 직접 둡니다.

- `go`
- `chan T`
- `select`

그래서 process composition이 매우 자연스럽게 느껴집니다.

## Tokio는 같은 문제를 어떻게 푸나

Tokio는 Rust의 async runtime입니다. 중요한 mental model은:

- `async fn`은 future state machine으로 컴파일되고,
- future는 executor가 poll하기 전에는 아무 일도 하지 않으며,
- Tokio가 scheduler, timer driver, I/O driver를 제공해 task를 진행시킵니다.

그래서 Tokio의 동시성은 보통 이렇게 설명됩니다.

- task,
- `.await` yield point,
- executor,
- async I/O resource,
- `select!`,
- `mpsc`, `oneshot`, `watch` 같은 채널들.

## 가장 큰 구현 차이는 어디서 오나

Go goroutine은 stackful입니다.

그래서 ordinary blocking-looking code가 별도의 future state machine 사고 없이 suspend / resume될 수 있습니다.

Rust future는 stackless이고 cooperative polling 모델입니다.

이 차이 때문에 Rust는:

- ownership / lifetime 제약을 타입 시스템에 더 강하게 밀어넣고,
- 숨은 stackful coroutine 모델을 피하며,
- `.await`에서 suspension point를 더 명시적으로 드러냅니다.

대신 같은 workflow를 표현할 때 Tokio 쪽이 더 구조를 많이 요구하는 경우가 흔합니다.

## scheduling과 fairness

Go runtime은 asynchronous preemption과 netpoll wakeup을 scheduler에 직접 통합합니다.

Tokio task는 일반적으로 cooperative합니다. `.await` 같은 지점에서 yield하며 executor가 다시 poll해야 진행됩니다.

이 차이는 엔지니어링 스타일을 바꿉니다.

- Go는 direct-style code를 매우 싸게 느끼게 하고,
- Tokio는 suspension point를 더 명시적이고 분석 가능하게 만듭니다.

## channel과 request composition

Go는 channel 이야기가 언어 표면에 더 깊게 박혀 있습니다.

Rust + Tokio도 channel 기반 설계를 잘 지원하지만, channel은 library primitive입니다. 그래서 실무에서는:

- Go는 더 빨리 channel을 꺼내 들고,
- Tokio는 channel, mutex, atomics, typed ownership transfer를 더 많이 섞으며,
- request coordination도 `join!`, `select!`, task handle, `mpsc`, `oneshot` 중심으로 표현되는 경우가 많습니다.

## cancellation

이 차이는 cancellation에서도 뚜렷합니다.

Go에서는:

- `context.Context`가 request lifetime의 중심이고,
- channel close가 lifecycle signal이 되며,
- goroutine shutdown은 주로 application protocol 문제입니다.

Tokio에서는:

- future나 task handle의 drop이 중요하고,
- cancellation token이나 `select!` 패턴이 흔하며,
- structured composition이 더 library layer에서 드러납니다.

## Go가 더 편하게 느껴질 때

- direct-style network service code를 쓰고 싶을 때
- channels와 `select`를 기본 조합 도구로 쓰고 싶을 때
- poll-based futures mental model을 앞에 두고 싶지 않을 때

## Tokio가 더 편하게 느껴질 때

- ownership / lifetime 제약을 타입 시스템에 더 밀어 넣고 싶을 때
- async boundary를 더 명시적으로 보고 싶을 때
- Rust 생태계의 zero-cost abstraction 감각을 함께 가져가고 싶을 때

## 공식 자료

- [Rust Async Book](https://rust-lang.github.io/async-book/)
- [Tokio Runtime](https://docs.rs/tokio/latest/tokio/runtime/)
- [Tokio Tutorial](https://tokio.rs/tokio/tutorial)
- [Go에서의 CSP 이론](/ko/advanced/csp-theory)

## Practical takeaway

Go와 Tokio는 같은 모델의 두 스킨이 아닙니다.

둘 다 큰 concurrent system을 만들 수 있지만, 복잡성을 배치하는 층이 다릅니다.

- Go: 언어와 런타임에 더 많이
- Tokio: future, executor, library composition에 더 많이

이 차이를 이해하면 패턴 선택도 더 의도적으로 할 수 있고, Go가 실제로 무엇을 사다 주는지 더 명확하게 보입니다.
