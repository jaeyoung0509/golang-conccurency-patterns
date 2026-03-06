---
title: 표준 라이브러리 개요
description: 실전 Go 서비스의 동시성, 수명 관리, I/O, 시간을 결정하는 표준 라이브러리 패키지를 설명합니다.
---

# 표준 라이브러리 개요

프로덕션 Go 서비스는 커스텀 동시성 헬퍼보다 표준 라이브러리 안에서 더 많은 시간을 보냅니다.

이 섹션은 실제 시스템의 동작을 조용히 결정하는 패키지에 집중합니다.

- `context`: 수명 관리와 취소 전파
- `net/http`: 서버와 클라이언트 I/O
- `database/sql`: 풀링과 역압력
- `time`: deadline, retry, timer correctness

## 이 섹션이 다루는 것

<div class="path-grid">
  <div class="path-card">
    <h3><a href="/ko/stdlib/context-internals">context</a></h3>
    <p>cancellation tree, `cancelCtx`, `timerCtx`, `valueCtx`, `Cause`, 그리고 request 작업을 leak시키는 실수를 봅니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/net-http-server-transport">net/http</a></h3>
    <p>서버 connection 모델, request lifetime, transport pool, keep-alive reuse, timeout boundary를 이해합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/database-sql-pool">database/sql</a></h3>
    <p>`sql.DB`가 왜 pool인지, waiter와 cleaner가 어떻게 동작하는지, cancellation과 pool sizing이 어디서 깨지는지 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/time-timers-tickers">time</a></h3>
    <p>monotonic time, timers, tickers, `Stop`/`Reset`, retry loop를 안전하게 다루는 mental model을 잡습니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/sync-and-atomic">sync와 atomic</a></h3>
    <p>`RWMutex`, `WaitGroup`, `Once`, `sync.Map`, `sync.Cond`, `atomic.Pointer` 중 무엇이 어떤 invariant에 맞는지 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/process-signals-and-observability">신호와 관측</a></h3>
    <p>OS signal, graceful process shutdown, runtime metrics, pprof 기반 incident investigation을 연결합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/io-bufio-bytes">io, bufio, bytes</a></h3>
    <p>streaming interface, buffering, `io.Copy` fast path, scanner limit, byte-slice aliasing을 설명합니다.</p>
  </div>
  <div class="path-card">
    <h3><a href="/ko/stdlib/encoding-json">encoding/json</a></h3>
    <p>JSON 서비스 경계에서 strictness, stream decoding, number handling, compatibility trap을 정리합니다.</p>
  </div>
</div>

## 추천 읽기 순서

1. 먼저 [context](/ko/stdlib/context-internals)를 읽습니다. lifetime ownership이 이 섹션의 나머지 주제 전부에 영향을 주기 때문입니다.
2. 다음으로 [time, timers, tickers](/ko/stdlib/time-timers-tickers)를 읽습니다. deadline과 retry는 시간 모델이 정확해야 안전합니다.
3. 그 다음 [sync와 atomic 프리미티브](/ko/stdlib/sync-and-atomic)를 읽고, 특정 상태 경계에 channels, mutex, map, atomic snapshot 중 무엇이 맞는지 판단합니다.
4. 그 다음 [net/http 서버와 transport 내부](/ko/stdlib/net-http-server-transport)를 읽습니다. 여기서 context와 timer가 실제 네트워크 I/O와 만납니다.
5. 이어서 [database/sql pool 내부](/ko/stdlib/database-sql-pool)를 읽습니다. 여기서는 cancellation, waiting, resource limit이 운영 문제로 드러납니다.
6. streaming API나 로그 경계를 다룰 때는 [io, bufio, bytes](/ko/stdlib/io-bufio-bytes)와 [프로덕션에서의 encoding/json](/ko/stdlib/encoding-json)을 같이 읽습니다.
7. 마지막으로 [프로세스 신호와 런타임 관측](/ko/stdlib/process-signals-and-observability)을 읽어 서비스 운영 관점으로 연결합니다.

## 읽고 나면 답할 수 있어야 하는 질문

- 왜 `cancel()`을 빠뜨리면 timer만이 아니라 더 많은 것이 leak되는가?
- 왜 `http.Client`는 response body 처리를 잘못하면 reuse가 무너지는가?
- 왜 `sql.DB`는 per-request 객체가 아니라 long-lived shared handle인가?
- 왜 `time.Time`은 wall-clock과 monotonic reading을 같이 들고 있는가?
- 언제 plain mutex가 `sync.Map`나 atomics보다 더 깔끔한가?
- 왜 `signal.NotifyContext`가 raw signal channel보다 더 좋은 shutdown entry point가 되는가?
- 왜 `io.Copy`가 hand-written loop보다 빠를 수 있는가?
- 왜 `encoding/json`은 strict schema 시스템이라면 거부할 동작을 조용히 허용하는가?
- 왜 Go 1.23의 timer channel semantics 변화가 중요한가?

## Practical takeaway

기초 원리가 Go 동시성이 왜 가능한지 설명한다면, 이 섹션은 실제 프로덕션 Go 코드가 그 동시성을 어떻게 매일 표현하는지 설명합니다.
