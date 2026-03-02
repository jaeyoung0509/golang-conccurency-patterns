---
layout: home

hero:
  name: Go Concurrency Patterns
  text: 런타임 내부부터 실전 패턴까지
  tagline: Go 동시성을 기초 원리, 테스트된 예제, 패턴 선택 가이드, 영문/국문 동시 문서로 깊게 학습합니다.
  actions:
    - theme: brand
      text: 기초 원리부터 시작
      link: /ko/fundamentals/
    - theme: alt
      text: 패턴 둘러보기
      link: /ko/patterns/
    - theme: alt
      text: 프로덕션 가이드
      link: /ko/production/
    - theme: alt
      text: 테스트 플레이북
      link: /ko/testing/
    - theme: alt
      text: Read in English
      link: /

features:
  - title: 읽는 순서가 분명함
    details: "섹션 개요, 학습 경로, 패턴 선택 가이드를 추가해서 처음 들어와도 어디서 시작할지 바로 보이게 만들었습니다."
  - title: 런타임까지 깊게 설명
    details: "스케줄러, 메모리 모델, 채널 내부, mutex/runtime semaphore까지 Go 내부 원리를 설명합니다."
  - title: 실용 예제 중심
    details: "배송, 부정거래 분석, 재고, 중복 요청 억제처럼 실제 백엔드 문제에 가까운 예제를 사용합니다."
  - title: 테스트로 검증
    details: "`go test`로 순서 보장, 취소 전파, 동시성 제한, 실패 정책을 실제로 확인합니다."
  - title: 고급 운영 주제 포함
    details: "구조화된 동시성, 가중 세마포어, singleflight, 액터, 로드 셰딩까지 운영 관점의 주제를 다룹니다."
  - title: 프로덕션 운영 규칙 포함
    details: "admission control, queue budget, lifetime ownership, 대규모 시스템 동시성 tradeoff를 별도 섹션으로 다룹니다."
  - title: 테스트와 관측도 포함
    details: "Race detector, synctest, leak test, trace, contention profile을 동시성 역량의 일부로 다룹니다."
  - title: 영어/한국어 지원
    details: "영문 `/`와 국문 `/ko/`가 같은 구조를 공유해서 팀 단위 학습에 맞습니다."
---

## 여기서 어떻게 시작하면 되나

<div class="lead-panel">
  <p>
    이 사이트는 goroutine과 channel 문법을 외우는 곳이 아니라,
    <strong>왜 이런 패턴이 가능한지, 언제 어떤 패턴을 써야 하는지, 운영에서 어떻게 안전하게 유지할지</strong>를 배우는 곳입니다.
  </p>
</div>

<div class="path-grid">
  <div class="path-card">
    <h3>1. 기초 원리</h3>
    <p>스케줄러, 메모리 모델, 채널 내부부터 이해해서 뒤의 패턴들이 왜 그렇게 생겼는지 연결합니다.</p>
    <p><a href="/ko/fundamentals/">기초 원리 보기</a></p>
  </div>
  <div class="path-card">
    <h3>2. 실전 패턴</h3>
    <p>워커 풀, 파이프라인, 팬아웃/팬인, 컨텍스트 취소를 실제 workload 관점으로 읽습니다.</p>
    <p><a href="/ko/patterns/">패턴 보기</a></p>
  </div>
  <div class="path-card">
    <h3>3. 고급 주제</h3>
    <p>자원 예산, 수명 관리, 중복 억제, 소유권 모델, 과부하 제어까지 확장합니다.</p>
    <p><a href="/ko/advanced/">고급 주제 보기</a></p>
  </div>
  <div class="path-card">
    <h3>4. 테스트</h3>
    <p>cancellation, shutdown, race safety, timeout behavior를 lucky sleep 없이 검증하는 법을 익힙니다.</p>
    <p><a href="/ko/testing/">테스트 보기</a></p>
  </div>
  <div class="path-card">
    <h3>5. 프로덕션</h3>
    <p>queue budget, overload policy, goroutine ownership, 실제 오픈소스의 concurrency 구조를 같이 봅니다.</p>
    <p><a href="/ko/production/">프로덕션 보기</a></p>
  </div>
  <div class="path-card">
    <h3>6. 비교 / 확장</h3>
    <p>Go의 CSP 계열 모델을 Rust Tokio와 비교해 mental model을 더 넓힙니다.</p>
    <p><a href="/ko/extras/">비교 / 확장 보기</a></p>
  </div>
</div>

## 문제에 맞게 바로 들어가기

| 이런 게 궁금하면 | 여기부터 읽기 |
| --- | --- |
| 고루틴이 왜 싸고, 스케줄러가 실제로 뭘 하는지 | [Go 런타임과 스케줄러](/ko/fundamentals/go-runtime-scheduler) |
| 채널이 왜 메모리 가시성 경계를 만드는지 | [Channels, Select, 그리고 Memory Model](/ko/fundamentals/channels-memory-model) |
| 많은 독립 작업의 병렬 수를 어떻게 제한하는지 | [워커 풀](/ko/patterns/worker-pool) |
| 하나의 요청 안에서 여러 sibling task를 어떻게 관리하는지 | [구조화된 동시성](/ko/advanced/structured-concurrency) |
| 같은 cache miss 요청을 어떻게 하나로 합치는지 | [Singleflight](/ko/advanced/singleflight) |
| 과부하를 늦게 터뜨리지 않고 초기에 제어하는 법 | [역압력과 로드 셰딩](/ko/advanced/backpressure-load-shedding) |
| timeout-heavy 코드를 실제 sleep 없이 어떻게 테스트하는지 | [synctest로 결정적 테스트](/ko/testing/synctest) |
| 토이 패턴을 넘어 대규모 Go 운영에서 무엇이 중요한지 | [대규모 Go 시스템](/ko/production/large-scale-go-systems) |
| Go와 Rust Tokio의 async runtime 모델이 어떻게 다른지 | [Go CSP vs Rust Tokio](/ko/extras/go-csp-vs-rust-tokio) |

## 이 사이트가 다른 이유

<div class="signal-strip">
  <div class="signal">
    <strong>토이 예제가 아님</strong>
    <span>실제 서비스에서 볼 법한 백엔드 문제 형태로 예제를 구성했습니다.</span>
  </div>
  <div class="signal">
    <strong>코드만 던지지 않음</strong>
    <span>소유권, 순서, 실패 정책, 테스트가 무엇을 보장하는지까지 설명합니다.</span>
  </div>
  <div class="signal">
    <strong>이론도 표면적이지 않음</strong>
    <span>`hchan`, `sudog`, run queue, starvation mode 같은 런타임 개념까지 내려갑니다.</span>
  </div>
</div>

## 추천 읽기 흐름

1. [시작하기](/ko/guide/getting-started)에서 저장소 구조와 검증 명령을 확인합니다.
2. [예제 읽는 법](/ko/guide/how-to-read)으로 읽는 기준을 맞춥니다.
3. [기초 원리 개요](/ko/fundamentals/)부터 읽습니다.
4. 자기 workload에 맞는 패턴을 [패턴 개요](/ko/patterns/)에서 고릅니다.
5. concurrent component를 production-ready로 보기 전에 [테스트 개요](/ko/testing/)를 읽습니다.
6. 운영 규칙과 오픈소스 사례를 보려면 [프로덕션 개요](/ko/production/)를 읽습니다.
7. 운영 설계와 비교 관점까지 확장할 때 [고급 주제 개요](/ko/advanced/), [비교 / 확장 개요](/ko/extras/)로 넘어갑니다.
