import { defineConfig } from "vitepress";

const enSidebar = [
  {
    text: "Guide",
    items: [
      { text: "Getting Started", link: "/guide/getting-started" },
      { text: "How to Read the Examples", link: "/guide/how-to-read" },
    ],
  },
  {
    text: "Fundamentals",
    items: [
      { text: "Overview", link: "/fundamentals/" },
      { text: "Runtime Evolution", link: "/fundamentals/runtime-evolution" },
      { text: "Go Runtime and Scheduler", link: "/fundamentals/go-runtime-scheduler" },
      { text: "Netpoller, Timers, and Syscalls", link: "/fundamentals/netpoller-timers-syscalls" },
      { text: "TCP, DNS, and Connection Lifecycles", link: "/fundamentals/tcp-dns-connection-lifecycles" },
      { text: "Channels, Select, and the Memory Model", link: "/fundamentals/channels-memory-model" },
      { text: "Channel Internals", link: "/fundamentals/channel-internals" },
      { text: "Mutex and Runtime Semaphore Internals", link: "/fundamentals/mutex-semaphore-internals" },
      { text: "Map Internals and Swiss Tables", link: "/fundamentals/map-internals" },
      { text: "Garbage Collector and Green Tea GC", link: "/fundamentals/garbage-collector" },
    ],
  },
  {
    text: "Internals",
    items: [
      { text: "Overview", link: "/internals/" },
      { text: "Compiler and Toolchain", link: "/internals/compiler-and-toolchain" },
      { text: "Allocator and Hybrid Write Barrier", link: "/internals/allocator-and-write-barrier" },
      { text: "Layout, Padding, and False Sharing", link: "/internals/layout-padding-false-sharing" },
      { text: "Generics and Interfaces", link: "/internals/generics-and-interfaces" },
      { text: "unsafe, cgo, and Pinner", link: "/internals/unsafe-cgo-pinner" },
      { text: "Kernel I/O Paths", link: "/internals/kernel-io-paths" },
      { text: "Modern Performance Tuning", link: "/internals/modern-performance-tuning" },
      { text: "Standard Library Anatomy", link: "/internals/stdlib-anatomy" },
    ],
  },
  {
    text: "Standard Library",
    items: [
      { text: "Overview", link: "/stdlib/" },
      { text: "context Package Internals", link: "/stdlib/context-internals" },
      { text: "time, Timers, and Tickers", link: "/stdlib/time-timers-tickers" },
      { text: "sync and atomic Primitives", link: "/stdlib/sync-and-atomic" },
      { text: "net, netip, and Dial Budgets", link: "/stdlib/net-and-netip" },
      { text: "Protocol Design with net.Conn and bufio", link: "/stdlib/protocol-design-net-conn-bufio" },
      { text: "crypto/tls in Production", link: "/stdlib/crypto-tls" },
      { text: "HTTP/2, ALPN, and Stream Multiplexing", link: "/stdlib/http2-alpn-stream-multiplexing" },
      { text: "net/http Server and Transport", link: "/stdlib/net-http-server-transport" },
      { text: "database/sql Pool Internals", link: "/stdlib/database-sql-pool" },
      { text: "Process Signals and Runtime Observability", link: "/stdlib/process-signals-and-observability" },
      { text: "os/exec and Subprocess Lifecycle", link: "/stdlib/os-exec-and-subprocesses" },
      { text: "io, bufio, and bytes", link: "/stdlib/io-bufio-bytes" },
      { text: "encoding/json in Production", link: "/stdlib/encoding-json" },
    ],
  },
  {
    text: "Playbooks",
    items: [
      { text: "Overview", link: "/playbooks/" },
      { text: "net/http Production Field Guide", link: "/playbooks/net-http-production-field-guide" },
      { text: "grpc-go Production Playbook", link: "/playbooks/grpc-go-production-playbook" },
      { text: "go-redis Production Playbook", link: "/playbooks/go-redis-production-playbook" },
      { text: "Kafka with IBM Sarama", link: "/playbooks/kafka-with-ibm-sarama" },
    ],
  },
  {
    text: "Patterns",
    items: [
      { text: "Overview", link: "/patterns/" },
      { text: "Worker Pool", link: "/patterns/worker-pool" },
      { text: "Pipeline", link: "/patterns/pipeline" },
      { text: "Fan-Out / Fan-In", link: "/patterns/fan-out-fan-in" },
      { text: "Context Cancellation", link: "/patterns/context-cancellation" },
      { text: "Channel of Channels", link: "/patterns/channel-of-channels" },
      { text: "Optional Values Across API Boundaries", link: "/patterns/optional-values-across-boundaries" },
      { text: "Graceful Shutdown", link: "/patterns/graceful-shutdown" },
      { text: "Or-Done, Tee, and Bridge", link: "/patterns/or-done-tee-bridge" },
    ],
  },
  {
    text: "Advanced",
    items: [
      { text: "Overview", link: "/advanced/" },
      { text: "Structured Concurrency", link: "/advanced/structured-concurrency" },
      { text: "Weighted Semaphore", link: "/advanced/weighted-semaphore" },
      { text: "Singleflight", link: "/advanced/singleflight" },
      { text: "Backpressure and Load Shedding", link: "/advanced/backpressure-load-shedding" },
      { text: "Actor Pattern", link: "/advanced/actor-pattern" },
      { text: "CSP Theory in Go", link: "/advanced/csp-theory" },
    ],
  },
  {
    text: "Production",
    items: [
      { text: "Overview", link: "/production/" },
      { text: "Large-Scale Go Systems", link: "/production/large-scale-go-systems" },
      { text: "Docker, containerd, and Kubernetes", link: "/production/docker-containerd-kubernetes" },
      { text: "Debugging Go Network Services", link: "/production/debugging-go-network-services" },
      { text: "Kubernetes Service Networking", link: "/production/kubernetes-service-networking" },
      { text: "Regulated Go Systems", link: "/production/regulated-systems" },
      { text: "Temporal and Durable Execution", link: "/production/temporal-durable-execution" },
      { text: "Open-Source Case Studies", link: "/production/open-source-case-studies" },
      { text: "Go Open-Source Histories", link: "/production/go-open-source-histories" },
    ],
  },
  {
    text: "Testing",
    items: [
      { text: "Overview", link: "/testing/" },
      { text: "AI-Assisted Go Safety", link: "/testing/ai-assisted-go-safety" },
      { text: "Race Detector", link: "/testing/race-detector" },
      { text: "Deterministic Tests with synctest", link: "/testing/synctest" },
      { text: "Integration Testing with Testcontainers", link: "/testing/integration-testcontainers" },
      { text: "Leak, Shutdown, and Timeout Testing", link: "/testing/leaks-and-shutdowns" },
      { text: "Tracing and Contention Observability", link: "/testing/tracing-and-profiling" },
    ],
  },
  {
    text: "Extras",
    items: [
      { text: "Overview", link: "/extras/" },
      { text: "Go Pitfalls Appendix", link: "/extras/go-pitfalls/" },
      { text: "Go CSP vs Rust Tokio", link: "/extras/go-csp-vs-rust-tokio" },
      { text: "Go vs Rust Decision Guide", link: "/extras/go-vs-rust-decision-guide" },
    ],
  },
] as const;

const koSidebar = [
  {
    text: "가이드",
    items: [
      { text: "시작하기", link: "/ko/guide/getting-started" },
      { text: "예제 읽는 법", link: "/ko/guide/how-to-read" },
    ],
  },
  {
    text: "기초 원리",
    items: [
      { text: "개요", link: "/ko/fundamentals/" },
      { text: "런타임 진화", link: "/ko/fundamentals/runtime-evolution" },
      { text: "Go 런타임과 스케줄러", link: "/ko/fundamentals/go-runtime-scheduler" },
      { text: "Netpoller, 타이머, 그리고 Syscall", link: "/ko/fundamentals/netpoller-timers-syscalls" },
      { text: "TCP, DNS, 그리고 Connection Lifecycle", link: "/ko/fundamentals/tcp-dns-connection-lifecycles" },
      { text: "Channels, Select, 그리고 Memory Model", link: "/ko/fundamentals/channels-memory-model" },
      { text: "채널 내부 동작", link: "/ko/fundamentals/channel-internals" },
      { text: "Mutex와 런타임 세마포어 내부", link: "/ko/fundamentals/mutex-semaphore-internals" },
      { text: "맵 내부 구조와 Swiss Tables", link: "/ko/fundamentals/map-internals" },
      { text: "가비지 컬렉터와 Green Tea GC", link: "/ko/fundamentals/garbage-collector" },
    ],
  },
  {
    text: "내부 구조",
    items: [
      { text: "개요", link: "/ko/internals/" },
      { text: "컴파일러와 툴체인", link: "/ko/internals/compiler-and-toolchain" },
      { text: "할당기와 하이브리드 write barrier", link: "/ko/internals/allocator-and-write-barrier" },
      { text: "레이아웃, 패딩, false sharing", link: "/ko/internals/layout-padding-false-sharing" },
      { text: "제네릭과 인터페이스", link: "/ko/internals/generics-and-interfaces" },
      { text: "unsafe, cgo, 그리고 Pinner", link: "/ko/internals/unsafe-cgo-pinner" },
      { text: "Kernel I/O Paths", link: "/ko/internals/kernel-io-paths" },
      { text: "현대 Go 성능 튜닝", link: "/ko/internals/modern-performance-tuning" },
      { text: "표준 라이브러리 해부", link: "/ko/internals/stdlib-anatomy" },
    ],
  },
  {
    text: "표준 라이브러리",
    items: [
      { text: "개요", link: "/ko/stdlib/" },
      { text: "context 패키지 내부", link: "/ko/stdlib/context-internals" },
      { text: "time, Timers, Tickers", link: "/ko/stdlib/time-timers-tickers" },
      { text: "sync와 atomic 프리미티브", link: "/ko/stdlib/sync-and-atomic" },
      { text: "net, netip, 그리고 dial budget", link: "/ko/stdlib/net-and-netip" },
      { text: "net.Conn과 bufio로 프로토콜 설계하기", link: "/ko/stdlib/protocol-design-net-conn-bufio" },
      { text: "프로덕션에서의 crypto/tls", link: "/ko/stdlib/crypto-tls" },
      { text: "HTTP/2, ALPN, 그리고 Stream Multiplexing", link: "/ko/stdlib/http2-alpn-stream-multiplexing" },
      { text: "net/http 서버와 Transport 내부", link: "/ko/stdlib/net-http-server-transport" },
      { text: "database/sql 풀 내부", link: "/ko/stdlib/database-sql-pool" },
      { text: "프로세스 신호와 런타임 관측", link: "/ko/stdlib/process-signals-and-observability" },
      { text: "os/exec와 subprocess lifecycle", link: "/ko/stdlib/os-exec-and-subprocesses" },
      { text: "io, bufio, bytes", link: "/ko/stdlib/io-bufio-bytes" },
      { text: "프로덕션에서의 encoding/json", link: "/ko/stdlib/encoding-json" },
    ],
  },
  {
    text: "실전 플레이북",
    items: [
      { text: "개요", link: "/ko/playbooks/" },
      { text: "net/http 실전 필드 가이드", link: "/ko/playbooks/net-http-production-field-guide" },
      { text: "grpc-go 실전 플레이북", link: "/ko/playbooks/grpc-go-production-playbook" },
      { text: "go-redis 실전 플레이북", link: "/ko/playbooks/go-redis-production-playbook" },
      { text: "Kafka with IBM Sarama", link: "/ko/playbooks/kafka-with-ibm-sarama" },
    ],
  },
  {
    text: "패턴",
    items: [
      { text: "개요", link: "/ko/patterns/" },
      { text: "워커 풀", link: "/ko/patterns/worker-pool" },
      { text: "파이프라인", link: "/ko/patterns/pipeline" },
      { text: "팬아웃 / 팬인", link: "/ko/patterns/fan-out-fan-in" },
      { text: "컨텍스트 취소", link: "/ko/patterns/context-cancellation" },
      { text: "Channel of Channels", link: "/ko/patterns/channel-of-channels" },
      { text: "API 경계에서의 Optional 값 패턴", link: "/ko/patterns/optional-values-across-boundaries" },
      { text: "Graceful Shutdown", link: "/ko/patterns/graceful-shutdown" },
      { text: "Or-Done, Tee, Bridge", link: "/ko/patterns/or-done-tee-bridge" },
    ],
  },
  {
    text: "고급 주제",
    items: [
      { text: "개요", link: "/ko/advanced/" },
      { text: "구조화된 동시성", link: "/ko/advanced/structured-concurrency" },
      { text: "가중 세마포어", link: "/ko/advanced/weighted-semaphore" },
      { text: "Singleflight", link: "/ko/advanced/singleflight" },
      { text: "역압력과 로드 셰딩", link: "/ko/advanced/backpressure-load-shedding" },
      { text: "액터 패턴", link: "/ko/advanced/actor-pattern" },
      { text: "Go에서의 CSP 이론", link: "/ko/advanced/csp-theory" },
    ],
  },
  {
    text: "프로덕션",
    items: [
      { text: "개요", link: "/ko/production/" },
      { text: "대규모 Go 시스템", link: "/ko/production/large-scale-go-systems" },
      { text: "Docker, containerd, 그리고 Kubernetes", link: "/ko/production/docker-containerd-kubernetes" },
      { text: "Go 네트워크 서비스 디버깅", link: "/ko/production/debugging-go-network-services" },
      { text: "Kubernetes 서비스 네트워킹", link: "/ko/production/kubernetes-service-networking" },
      { text: "규제 환경의 Go 시스템", link: "/ko/production/regulated-systems" },
      { text: "Temporal과 Durable Execution", link: "/ko/production/temporal-durable-execution" },
      { text: "오픈소스 사례", link: "/ko/production/open-source-case-studies" },
      { text: "Go 오픈소스 역사 읽기", link: "/ko/production/go-open-source-histories" },
    ],
  },
  {
    text: "테스트",
    items: [
      { text: "개요", link: "/ko/testing/" },
      { text: "AI 보조 Go 안전성", link: "/ko/testing/ai-assisted-go-safety" },
      { text: "Race Detector", link: "/ko/testing/race-detector" },
      { text: "synctest로 결정적 테스트", link: "/ko/testing/synctest" },
      { text: "Testcontainers로 통합 테스트하기", link: "/ko/testing/integration-testcontainers" },
      { text: "리크, 종료, 타임아웃 테스트", link: "/ko/testing/leaks-and-shutdowns" },
      { text: "트레이싱과 경합 관측", link: "/ko/testing/tracing-and-profiling" },
    ],
  },
  {
    text: "비교 / 확장",
    items: [
      { text: "개요", link: "/ko/extras/" },
      { text: "Go 함정 부록", link: "/ko/extras/go-pitfalls/" },
      { text: "Go CSP vs Rust Tokio", link: "/ko/extras/go-csp-vs-rust-tokio" },
      { text: "Go vs Rust 결정 가이드", link: "/ko/extras/go-vs-rust-decision-guide" },
    ],
  },
] as const;

export default defineConfig({
  title: "Go Handbook",
  description: "Deep, practical Go handbook covering internals, standard library, patterns, testing, and production systems.",
  base: "/golang-handbook/",
  cleanUrls: true,
  lastUpdated: true,
  head: [
    ["meta", { name: "theme-color", content: "#115e59" }],
  ],
  markdown: {
    config(md) {
      const fence = md.renderer.rules.fence;

      md.renderer.rules.fence = (tokens, idx, options, env, self) => {
        const token = tokens[idx];

        if (token.info.trim() === "mermaid") {
          return `<MermaidDiagram :chart='${JSON.stringify(token.content)}' />`;
        }

        return fence ? fence(tokens, idx, options, env, self) : self.renderToken(tokens, idx, options);
      };
    },
  },
  themeConfig: {
    search: {
      provider: "local",
    },
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/jaeyoung0509/golang-handbook",
      },
    ],
    footer: {
      message: "Built with VitePress and real Go examples.",
      copyright: "MIT",
    },
  },
  locales: {
    root: {
      label: "English",
      lang: "en-US",
      title: "Go Handbook",
      description: "Deep, practical Go handbook covering internals, standard library, patterns, testing, and production systems.",
      themeConfig: {
        nav: [
          { text: "Start", link: "/guide/getting-started" },
          {
            text: "Core",
            items: [
              { text: "Fundamentals", link: "/fundamentals/" },
              { text: "Internals", link: "/internals/" },
              { text: "Standard Library", link: "/stdlib/" },
            ],
          },
          {
            text: "Practice",
            items: [
              { text: "Playbooks", link: "/playbooks/" },
              { text: "Patterns", link: "/patterns/" },
              { text: "Advanced", link: "/advanced/" },
            ],
          },
          {
            text: "Operate",
            items: [
              { text: "Production", link: "/production/" },
              { text: "Testing", link: "/testing/" },
            ],
          },
          { text: "Extras", link: "/extras/" },
        ],
        sidebar: enSidebar,
        outline: {
          level: [2, 3],
        },
        outlineTitle: "On this page",
        docFooter: {
          prev: "Previous page",
          next: "Next page",
        },
        darkModeSwitchLabel: "Appearance",
        sidebarMenuLabel: "Menu",
        returnToTopLabel: "Back to top",
        langMenuLabel: "Languages",
      },
    },
    ko: {
      label: "한국어",
      lang: "ko-KR",
      link: "/ko/",
      title: "Go Handbook",
      description: "기초 원리, 내부 구조, 표준 라이브러리, 패턴, 테스트, 프로덕션 시스템까지 다루는 실전 Go 핸드북.",
      themeConfig: {
        nav: [
          { text: "시작", link: "/ko/guide/getting-started" },
          {
            text: "핵심",
            items: [
              { text: "기초 원리", link: "/ko/fundamentals/" },
              { text: "내부 구조", link: "/ko/internals/" },
              { text: "표준 라이브러리", link: "/ko/stdlib/" },
            ],
          },
          {
            text: "실전",
            items: [
              { text: "실전 플레이북", link: "/ko/playbooks/" },
              { text: "패턴", link: "/ko/patterns/" },
              { text: "고급 주제", link: "/ko/advanced/" },
            ],
          },
          {
            text: "운영",
            items: [
              { text: "프로덕션", link: "/ko/production/" },
              { text: "테스트", link: "/ko/testing/" },
            ],
          },
          { text: "비교 / 확장", link: "/ko/extras/" },
        ],
        sidebar: koSidebar,
        outline: {
          level: [2, 3],
        },
        outlineTitle: "이 페이지에서",
        docFooter: {
          prev: "이전 페이지",
          next: "다음 페이지",
        },
        darkModeSwitchLabel: "테마",
        sidebarMenuLabel: "메뉴",
        returnToTopLabel: "맨 위로",
        langMenuLabel: "언어",
      },
    },
  },
});
