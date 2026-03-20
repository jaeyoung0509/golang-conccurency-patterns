---
layout: home

hero:
  name: Go Handbook
  text: 컴파일러 내부부터 프로덕션 시스템까지
  tagline: Go를 기초 원리, 시스템 내부 구조, 표준 라이브러리, 실전 패턴, 운영 가이드, 영문/국문 동시 문서로 깊게 학습합니다.
  actions:
    - theme: brand
      text: 기초 원리부터 시작
      link: /ko/fundamentals/
    - theme: alt
      text: 내부 구조 보기
      link: /ko/internals/
    - theme: alt
      text: 표준 라이브러리
      link: /ko/stdlib/
    - theme: alt
      text: 실전 플레이북
      link: /ko/playbooks/
    - theme: alt
      text: 패턴 둘러보기
      link: /ko/patterns/
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
  - title: 시스템 레벨 내부 구조 포함
    details: "Escape analysis, SSA, allocator, generics, interface layout, unsafe, cgo, PGO, sync.Pool 내부까지 추가로 다룹니다."
  - title: 표준 라이브러리까지 깊게
    details: "`context`, `net`, `crypto/tls`, `net/http`, `database/sql`, `os/exec`, `time`를 따로 빼서 실제 서비스 코드가 돌아가는 경계를 자세히 설명합니다."
  - title: 실전 라이브러리 플레이북
    details: "`net/http`, `grpc-go`, `go-redis`, Kafka with IBM Sarama를 safe default, observability, failure pattern 중심으로 운영 관점에서 설명합니다."
  - title: 실용 예제 중심
    details: "배송, 부정거래 분석, 재고, 중복 요청 억제처럼 실제 백엔드 문제에 가까운 예제를 사용합니다."
  - title: 테스트로 검증
    details: "`go test`로 순서 보장, 취소 전파, 동시성 제한, 실패 정책을 실제로 확인합니다."
  - title: 고급 운영 주제 포함
    details: "구조화된 동시성, 가중 세마포어, singleflight, 액터, 로드 셰딩까지 운영 관점의 주제를 다룹니다."
  - title: 프로덕션 운영 규칙 포함
    details: "admission control, queue budget, lifetime ownership, 대규모 시스템 동시성 tradeoff, 그리고 실제 오픈소스 역사까지 별도 섹션으로 다룹니다."
  - title: 테스트와 관측도 포함
    details: "Race detector, synctest, 실제 의존성 통합 테스트, leak test, trace, contention profile을 동시성 역량의 일부로 다룹니다."
  - title: 영어/한국어 지원
    details: "영문 `/`와 국문 `/ko/`가 같은 구조를 공유해서 팀 단위 학습에 맞습니다."
---

## 여기서 어떻게 시작하면 되나

<div class="lead-panel">
  <p>
    이 사이트는 Go 문법이나 goroutine/channel만 외우는 곳이 아니라,
    <strong>Go가 왜 이런 식으로 동작하는지, 런타임과 표준 라이브러리가 실제 시스템을 어떻게 만들게 하는지, 운영에서 어떻게 안전하게 쓰는지</strong>를 배우는 곳입니다.
  </p>
</div>

<div class="path-grid">
  <div class="path-card">
    <h3>1. 기초 원리</h3>
    <p>스케줄러, 메모리 모델, 채널 내부부터 이해해서 뒤의 패턴들이 왜 그렇게 생겼는지 연결합니다.</p>
    <p><a href="/ko/fundamentals/">기초 원리 보기</a></p>
  </div>
  <div class="path-card">
    <h3>2. 내부 구조</h3>
    <p>Escape analysis, SSA, allocator 설계, generics, interfaces, unsafe 경계, 현대 성능 도구를 같이 봅니다.</p>
    <p><a href="/ko/internals/">내부 구조 보기</a></p>
  </div>
  <div class="path-card">
    <h3>3. 표준 라이브러리</h3>
    <p>`context`, `net`, `crypto/tls`, `net/http`, `database/sql`, `os/exec`, `time`를 통해 런타임 보장이 실제 request lifetime, connection reuse, subprocess ownership, deadline behavior로 어떻게 드러나는지 봅니다.</p>
    <p><a href="/ko/stdlib/">표준 라이브러리 보기</a></p>
  </div>
  <div class="path-card">
    <h3>4. 실전 플레이북</h3>
    <p>`net/http`, `grpc-go`, `go-redis`, Kafka with IBM Sarama를 운영할 때 필요한 정책과 실패 패턴을 봅니다.</p>
    <p><a href="/ko/playbooks/">실전 플레이북 보기</a></p>
  </div>
  <div class="path-card">
    <h3>5. 실전 패턴</h3>
    <p>워커 풀, 파이프라인, 팬아웃/팬인, 컨텍스트 취소를 실제 workload 관점으로 읽습니다.</p>
    <p><a href="/ko/patterns/">패턴 보기</a></p>
  </div>
  <div class="path-card">
    <h3>6. 고급 주제</h3>
    <p>자원 예산, 수명 관리, 중복 억제, 소유권 모델, 과부하 제어까지 확장합니다.</p>
    <p><a href="/ko/advanced/">고급 주제 보기</a></p>
  </div>
  <div class="path-card">
    <h3>7. 테스트</h3>
    <p>cancellation, shutdown, race safety, timeout behavior를 lucky sleep 없이 검증하는 법을 익힙니다.</p>
    <p><a href="/ko/testing/">테스트 보기</a></p>
  </div>
  <div class="path-card">
    <h3>8. 프로덕션</h3>
    <p>queue budget, overload policy, goroutine ownership, Temporal류 durable execution, 그리고 왜 많은 인프라 시스템이 Go를 택했는지도 같이 봅니다.</p>
    <p><a href="/ko/production/">프로덕션 보기</a></p>
  </div>
  <div class="path-card">
    <h3>9. 비교 / 확장</h3>
    <p>Go의 CSP 계열 모델을 Rust Tokio와 비교하고, 언제 Go에 남고 언제 Rust로 옮겨야 하는지도 결정 가이드로 정리합니다.</p>
    <p><a href="/ko/extras/">비교 / 확장 보기</a></p>
  </div>
</div>

## 문제에 맞게 바로 들어가기

| 이런 게 궁금하면 | 여기부터 읽기 |
| --- | --- |
| 고루틴이 왜 싸고, 스케줄러가 실제로 뭘 하는지 | [Go 런타임과 스케줄러](/ko/fundamentals/go-runtime-scheduler) |
| 지역 값이 왜 여전히 힙으로 가는지 | [컴파일러와 툴체인](/ko/internals/compiler-and-toolchain) |
| 어떤 allocation 패턴이 왜 GC를 더 힘들게 하는지 | [할당기와 하이브리드 write barrier](/ko/internals/allocator-and-write-barrier) |
| request-scoped cancellation이 실제로 어떻게 전파되는지 | [context 패키지 내부](/ko/stdlib/context-internals) |
| channel이 shared state에 맞지 않을 때 무엇을 써야 하는지 | [sync와 atomic 프리미티브](/ko/stdlib/sync-and-atomic) |
| connection establish budget을 어떻게 잡고 `net.IP` footgun 없이 endpoint를 표현하는지 | [net과 netip](/ko/stdlib/net-and-netip) |
| TLS handshake, verification, ALPN이 request lifetime과 어떻게 연결되는지 | [프로덕션에서의 crypto/tls](/ko/stdlib/crypto-tls) |
| TCP, DNS, deadline, connection teardown이 어떻게 하나의 socket lifecycle을 이루는지 | [Go에서의 TCP, DNS, 그리고 Connection Lifecycle](/ko/fundamentals/tcp-dns-connection-lifecycles) |
| short-read/short-write 버그 없이 framed socket protocol을 어떻게 설계하는지 | [net.Conn과 bufio로 프로토콜 설계하기](/ko/stdlib/protocol-design-net-conn-bufio) |
| HTTP/2와 ALPN이 reuse를 socket에서 multiplexed stream으로 어떻게 바꾸는지 | [HTTP/2, ALPN, 그리고 Stream Multiplexing](/ko/stdlib/http2-alpn-stream-multiplexing) |
| Go HTTP 서버와 클라이언트 transport가 connection을 어떻게 소유하는지 | [net/http 서버와 Transport 내부](/ko/stdlib/net-http-server-transport) |
| `http.Client`와 `Transport`를 실제 timeout/reuse 정책으로 어떻게 운영하는지 | [net/http 실전 필드 가이드](/ko/playbooks/net-http-production-field-guide) |
| `Dial`과 `WithBlock` 함정 없이 gRPC channel을 어떻게 운영하는지 | [grpc-go 실전 플레이북](/ko/playbooks/grpc-go-production-playbook) |
| Redis client를 pool, protocol, timeout 정책으로 어떻게 다뤄야 하는지 | [go-redis 실전 플레이북](/ko/playbooks/go-redis-production-playbook) |
| Kafka semantics와 IBM Sarama 설정이 실제로 어떻게 연결되는지 | [Kafka with IBM Sarama](/ko/playbooks/kafka-with-ibm-sarama) |
| 왜 `sql.DB`는 connection이 아니라 pool인지 | [database/sql 풀 내부](/ko/stdlib/database-sql-pool) |
| 왜 `time.After`가 항상 좋은 loop primitive는 아닌지 | [time, Timers, Tickers](/ko/stdlib/time-timers-tickers) |
| subprocess cancellation, pipe, `WaitDelay`가 실제로 어떻게 동작하는지 | [os/exec와 subprocess lifecycle](/ko/stdlib/os-exec-and-subprocesses) |
| byte stream과 JSON을 hidden buffering mistake 없이 다루는 법 | [io, bufio, bytes](/ko/stdlib/io-bufio-bytes) |
| 프로세스 shutdown과 runtime observability를 Go 서비스에 어떻게 붙이는지 | [프로세스 신호와 런타임 관측](/ko/stdlib/process-signals-and-observability) |
| netpoll, epoll/kqueue, zero-copy path가 runtime boundary에서 어떻게 만나는지 | [Kernel I/O Paths: netpoll, epoll/kqueue, 그리고 Zero-Copy](/ko/internals/kernel-io-paths) |
| 채널이 왜 메모리 가시성 경계를 만드는지 | [Channels, Select, 그리고 Memory Model](/ko/fundamentals/channels-memory-model) |
| 많은 독립 작업의 병렬 수를 어떻게 제한하는지 | [워커 풀](/ko/patterns/worker-pool) |
| REST, gRPC, 메시지 경계에서 nullable output과 tri-state input을 어떻게 모델링해야 하는지 | [API 경계에서의 Optional 값 패턴](/ko/patterns/optional-values-across-boundaries) |
| 하나의 요청 안에서 여러 sibling task를 어떻게 관리하는지 | [구조화된 동시성](/ko/advanced/structured-concurrency) |
| 같은 cache miss 요청을 어떻게 하나로 합치는지 | [Singleflight](/ko/advanced/singleflight) |
| 과부하를 늦게 터뜨리지 않고 초기에 제어하는 법 | [역압력과 로드 셰딩](/ko/advanced/backpressure-load-shedding) |
| AI가 만든 Go 코드가 컴파일은 되는데 여전히 위험할 때 어떻게 조기 발견하는지 | [AI 보조 Go 안전성](/ko/testing/ai-assisted-go-safety) |
| timeout-heavy 코드를 실제 sleep 없이 어떻게 테스트하는지 | [synctest로 결정적 테스트](/ko/testing/synctest) |
| Postgres, Redis, container-backed 통합 테스트를 어떻게 깨끗하고 결정적으로 유지하는지 | [Testcontainers로 통합 테스트하기](/ko/testing/integration-testcontainers) |
| Docker, containerd, Kubernetes가 product UX, runtime lifecycle, control-plane ownership를 어떻게 나눠 갖는지 | [Docker, containerd, 그리고 Kubernetes](/ko/production/docker-containerd-kubernetes) |
| 네트워크 incident가 DNS, connect, TLS, transport reuse, network 중 어디서 발생했는지 어떻게 증명하는지 | [프로덕션에서 Go 네트워크 서비스 디버깅하기](/ko/production/debugging-go-network-services) |
| Service, proxy, readiness, draining이 Kubernetes 안에서 Go behavior를 어떻게 바꾸는지 | [Go 엔지니어를 위한 Kubernetes 서비스 네트워킹](/ko/production/kubernetes-service-networking) |
| Temporal 같은 durable workflow 엔진이 history, matching, worker를 가진 Go 시스템으로 어떻게 구성되는지 | [Temporal과 Durable Execution](/ko/production/temporal-durable-execution) |
| event sourcing, erasure, retention, audit 제약을 실제 Go 시스템에서 어떻게 함께 풀어야 하는지 | [규제 환경의 Go 시스템](/ko/production/regulated-systems) |
| 왜 중요한 인프라 오픈소스들이 하필 Go를 많이 택했는지 | [Go 오픈소스 역사 읽기](/ko/production/go-open-source-histories) |
| 토이 패턴을 넘어 대규모 Go 운영에서 무엇이 중요한지 | [대규모 Go 시스템](/ko/production/large-scale-go-systems) |
| 주요 Go 오픈소스가 queue, transport loop, stopper lifetime, pool을 어떻게 구현하는지 | [오픈소스 사례](/ko/production/open-source-case-studies) |
| `select`, typed nil, slice, context, stdlib contract 주변의 compile-clean footgun을 어떻게 피해야 하는지 | [Go 함정 부록](/ko/extras/go-pitfalls/) |
| Go와 Rust Tokio의 async runtime 모델이 어떻게 다른지 | [Go CSP vs Rust Tokio](/ko/extras/go-csp-vs-rust-tokio) |
| 언제 Go를 기본값으로 유지하고 언제 Rust로 좁게 옮겨야 하는지 | [Go vs Rust 결정 가이드](/ko/extras/go-vs-rust-decision-guide) |

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
    <span>`hchan`, `sudog`, run queue, starvation mode 같은 런타임 개념뿐 아니라 SSA, allocator tier, interface metadata까지 내려갑니다.</span>
  </div>
</div>

## 추천 읽기 흐름

1. [시작하기](/ko/guide/getting-started)에서 저장소 구조와 검증 명령을 확인합니다.
2. [예제 읽는 법](/ko/guide/how-to-read)으로 읽는 기준을 맞춥니다.
3. [기초 원리 개요](/ko/fundamentals/)부터 읽습니다.
4. runtime API 너머의 cost model이 궁금하면 [내부 구조 개요](/ko/internals/)를 읽습니다.
5. 실제 Go 서비스가 lifetime, I/O, SQL, time을 어떻게 표현하는지 보려면 [표준 라이브러리 개요](/ko/stdlib/)를 읽습니다.
6. 실제로 배포하는 라이브러리의 안전한 기본값이 궁금하면 [실전 플레이북 개요](/ko/playbooks/)를 읽습니다.
7. 자기 workload에 맞는 패턴을 [패턴 개요](/ko/patterns/)에서 고릅니다.
8. concurrent component를 production-ready로 보기 전에 [테스트 개요](/ko/testing/)를 읽습니다.
9. 운영 규칙과 오픈소스 사례를 보려면 [프로덕션 개요](/ko/production/)를 읽습니다.
10. 운영 설계와 비교 관점까지 확장할 때 [고급 주제 개요](/ko/advanced/), [비교 / 확장 개요](/ko/extras/)로 넘어갑니다.
