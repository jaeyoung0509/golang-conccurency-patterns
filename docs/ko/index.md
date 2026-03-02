---
layout: home

hero:
  name: Go Concurrency Patterns
  text: 실전 중심, 한영 동시 지원
  tagline: Go 동시성을 런타임 기초부터 고급 패턴까지, 실제 코드와 테스트, Mermaid 다이어그램으로 자세히 설명합니다.
  actions:
    - theme: brand
      text: 문서 시작하기
      link: /ko/guide/getting-started
    - theme: alt
      text: Read in English
      link: /
    - theme: alt
      text: GitHub 저장소
      link: https://github.com/jaeyoung0509/golang-conccurency-patterns

features:
  - title: 영어/한국어 지원
    details: "영문 루트와 `/ko/` 한글 문서를 같은 구조로 유지해서 팀 단위 학습에 맞춥니다."
  - title: 기초 원리 포함
    details: "GMP 스케줄러, 고루틴 비용 모델, channel, select, memory visibility 규칙까지 함께 설명합니다."
  - title: 실용 예제 중심
    details: "배송 견적, 결제 리스크 분석, 재고 조회, 대시보드 집계, 재고 액터처럼 실제 서비스에 가까운 예제를 사용합니다."
  - title: 테스트 포함
    details: "`go test`로 동시성 제한, 취소 전파, 데드라인 동작을 실제로 검증합니다."
  - title: Mermaid 다이어그램
    details: "채널 흐름과 취소 경계를 그림으로 먼저 이해한 뒤 코드를 읽을 수 있습니다."
  - title: 고급 주제 추가
    details: "액터 패턴과 CSP 이론을 단순 개념 소개가 아니라 Go 실전 트레이드오프와 연결해서 설명합니다."
  - title: 높은 가독성
    details: "긴 코드 나열보다 책임 경계, 실패 정책, 트레이드오프를 먼저 설명합니다."
---

## 이 사이트의 목적

많은 동시성 튜토리얼은 너무 작은 예제에서 멈춥니다. 이 저장소는 반대로 갑니다.

- 예제가 실제 백엔드 문제처럼 보이도록 구성했고,
- 테스트가 동시성 보장을 증명하도록 만들었고,
- 코드 조각만이 아니라 왜 이런 구조가 안전한지도 설명합니다.

<div class="custom-card-grid">
  <div class="custom-card">
    <h3>기초 원리</h3>
    <p>스케줄러, 채널, 메모리 모델을 먼저 이해해서 패턴을 왜 그렇게 짜는지까지 연결합니다.</p>
  </div>
  <div class="custom-card">
    <h3>워커 풀</h3>
    <p>병렬 수를 제한하면서도 입력 순서를 복원하고, 첫 실패 시 빠르게 중단합니다.</p>
  </div>
  <div class="custom-card">
    <h3>파이프라인</h3>
    <p>입력, 리스크 계산, 알림 생성처럼 서로 다른 단계를 명확하게 분리합니다.</p>
  </div>
  <div class="custom-card">
    <h3>팬아웃 / 팬인</h3>
    <p>여러 백엔드에 동시에 질의하고, 부분 실패를 버리지 않고 의미 있게 합칩니다.</p>
  </div>
  <div class="custom-card">
    <h3>컨텍스트 취소</h3>
    <p>요청 수명 주기에 묶인 여러 작업이 실패나 데드라인에 맞춰 즉시 멈추게 합니다.</p>
  </div>
  <div class="custom-card">
    <h3>고급 주제</h3>
    <p>구조화된 동시성, 가중 세마포어, singleflight, 액터 모델, 로드 셰딩까지 연결해서 고급 설계 감각을 잡게 합니다.</p>
  </div>
</div>

## 추천 읽기 순서

1. [시작하기](/ko/guide/getting-started)에서 저장소 구조와 실행 방법을 확인합니다.
2. [예제 읽는 법](/ko/guide/how-to-read)에서 동시성 코드를 보는 기준을 맞춥니다.
3. [기초 원리](/ko/fundamentals/go-runtime-scheduler)에서 런타임과 메모리 모델을 먼저 잡습니다.
4. 필요한 문제에 맞춰 실전 패턴 문서를 읽습니다.
5. 마지막에 [고급 주제](/ko/advanced/actor-pattern)에서 CSP와 액터 관점을 비교합니다.
