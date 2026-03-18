---
layout: home

hero:
  name: Go Concurrency Patterns
  text: From compiler internals to production-safe concurrency
  tagline: Learn Go through deep fundamentals, systems internals, tested examples, decision guides, and bilingual English/Korean docs.
  actions:
    - theme: brand
      text: Start With Fundamentals
      link: /fundamentals/
    - theme: alt
      text: Open Internals
      link: /internals/
    - theme: alt
      text: Standard Library
      link: /stdlib/
    - theme: alt
      text: Playbooks
      link: /playbooks/
    - theme: alt
      text: Browse Patterns
      link: /patterns/
    - theme: alt
      text: Testing Playbook
      link: /testing/
    - theme: alt
      text: 한국어 보기
      link: /ko/

features:
  - title: Clear learning path
    details: "The site now has section overviews, reading tracks, and pattern selection guides instead of throwing readers into dense pages."
  - title: Runtime-first depth
    details: "The fundamentals section explains the scheduler, memory model, channel internals, and mutex/runtime semaphore behavior."
  - title: Systems-level internals
    details: "The site now covers escape analysis, SSA, allocator internals, generics, interface layout, unsafe, cgo, PGO, and sync.Pool internals."
  - title: Standard library deep dives
    details: "The site now treats `context`, `net`, `crypto/tls`, `net/http`, `database/sql`, `os/exec`, and `time` as first-class learning tracks instead of assuming they are already understood."
  - title: Production library playbooks
    details: "The site now documents how to operate `net/http`, `grpc-go`, `go-redis`, and Kafka with IBM Sarama using safe defaults, observability hooks, and failure patterns."
  - title: Practical examples
    details: "Examples use realistic backend domains such as shipping, fraud analysis, inventory coordination, and cache-miss suppression."
  - title: Tested behavior
    details: "Every example is backed by `go test` so ordering, cancellation, limits, and failure policy are verified."
  - title: Advanced production topics
    details: "Structured concurrency, weighted semaphores, singleflight, actors, and load shedding are documented alongside the basics."
  - title: Production operating rules
    details: "Admission control, queue budgets, lifecycle ownership, large-system concurrency tradeoffs, and real open-source history are documented as first-class topics."
  - title: Testing and observability
    details: "Race detection, synctest, real dependency integration tests, leak testing, traces, and contention profiles are treated as first-class concurrency skills."
  - title: English and Korean
    details: "The site is mirrored across `/` and `/ko/` so mixed-language teams can study the same structure."
---

## Start Here

<div class="lead-panel">
  <p>
    This site is built for engineers who want to do more than memorize goroutines and channels.
    The goal is to understand <strong>why Go concurrency works, when each pattern is the right fit, and how to keep it safe in production</strong>.
  </p>
</div>

<div class="path-grid">
  <div class="path-card">
    <h3>1. Fundamentals</h3>
    <p>Start with the scheduler, memory model, and channel internals so the rest of the site has a solid mental foundation.</p>
    <p><a href="/fundamentals/">Open fundamentals</a></p>
  </div>
  <div class="path-card">
    <h3>2. Internals</h3>
    <p>Study escape analysis, SSA, allocator design, generics, interfaces, unsafe boundaries, and modern performance tooling.</p>
    <p><a href="/internals/">Open internals</a></p>
  </div>
  <div class="path-card">
    <h3>3. Standard Library</h3>
    <p>Learn how `context`, `net`, `crypto/tls`, `net/http`, `database/sql`, `os/exec`, and `time` turn runtime guarantees into request lifetimes, connection reuse, subprocess ownership, and deadline behavior.</p>
    <p><a href="/stdlib/">Open standard library</a></p>
  </div>
  <div class="path-card">
    <h3>4. Playbooks</h3>
    <p>Move from package internals to operator-facing guidance for `net/http`, `grpc-go`, `go-redis`, and Kafka with IBM Sarama.</p>
    <p><a href="/playbooks/">Open playbooks</a></p>
  </div>
  <div class="path-card">
    <h3>5. Practical Patterns</h3>
    <p>Move into worker pools, pipelines, fan-out/fan-in, and context cancellation when you are mapping code to real workloads.</p>
    <p><a href="/patterns/">Browse patterns</a></p>
  </div>
  <div class="path-card">
    <h3>6. Advanced Topics</h3>
    <p>Study resource budgeting, structured lifetimes, duplicate suppression, ownership models, and overload behavior.</p>
    <p><a href="/advanced/">Go deeper</a></p>
  </div>
  <div class="path-card">
    <h3>7. Testing</h3>
    <p>Learn how to prove cancellation, shutdown, race safety, and timeout behavior instead of relying on lucky sleeps.</p>
    <p><a href="/testing/">Open testing</a></p>
  </div>
  <div class="path-card">
    <h3>8. Production</h3>
    <p>Study queue budgets, overload policy, goroutine ownership, Temporal-style durable execution, and why so many real infrastructure systems ended up in Go.</p>
    <p><a href="/production/">Open production</a></p>
  </div>
  <div class="path-card">
    <h3>9. Extras</h3>
    <p>Compare Go's CSP-flavored model with Rust Tokio and use the Go-vs-Rust decision guide when language choice becomes an engineering question.</p>
    <p><a href="/extras/">Open extras</a></p>
  </div>
</div>

## Choose The Right Starting Point

| If you need to understand... | Start with |
| --- | --- |
| Why goroutines are cheap and how the scheduler actually runs them | [Go Runtime and Scheduler](/fundamentals/go-runtime-scheduler) |
| Why a local value still ends up on the heap | [Compiler and Toolchain](/internals/compiler-and-toolchain) |
| Why one allocation pattern hurts GC more than another | [Allocator and Hybrid Write Barrier](/internals/allocator-and-write-barrier) |
| How request-scoped cancellation actually propagates | [context Package Internals](/stdlib/context-internals) |
| When channels are the wrong tool for shared state | [sync and atomic Primitives](/stdlib/sync-and-atomic) |
| How to budget connection establishment and represent endpoints without `net.IP` footguns | [net and netip](/stdlib/net-and-netip) |
| How TLS handshake, verification, and ALPN fit into request lifetime | [crypto/tls in Production](/stdlib/crypto-tls) |
| How Go's HTTP server and client transport really own connections | [net/http Server and Transport](/stdlib/net-http-server-transport) |
| How to operate `http.Client` and `Transport` with real timeout and reuse policy | [net/http Production Field Guide](/playbooks/net-http-production-field-guide) |
| How to use gRPC channels without `Dial` and `WithBlock` footguns | [grpc-go Production Playbook](/playbooks/grpc-go-production-playbook) |
| How to run Redis clients with explicit pool, protocol, and timeout policy | [go-redis Production Playbook](/playbooks/go-redis-production-playbook) |
| How Kafka semantics and IBM Sarama config actually fit together | [Kafka with IBM Sarama](/playbooks/kafka-with-ibm-sarama) |
| Why `sql.DB` is a pool instead of a connection | [database/sql Pool Internals](/stdlib/database-sql-pool) |
| Why `time.After` is not always the right loop primitive | [time, Timers, and Tickers](/stdlib/time-timers-tickers) |
| How subprocess cancellation, pipes, and `WaitDelay` actually behave | [os/exec and Subprocess Lifecycle](/stdlib/os-exec-and-subprocesses) |
| How to stream bytes and JSON without hidden buffering mistakes | [io, bufio, and bytes](/stdlib/io-bufio-bytes) |
| How process shutdown and runtime observability fit into Go services | [Process Signals and Runtime Observability](/stdlib/process-signals-and-observability) |
| Why channels synchronize memory visibility | [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model) |
| How to cap parallelism across many independent tasks | [Worker Pool](/patterns/worker-pool) |
| How to model nullable outputs and tri-state inputs across REST, gRPC, and messages | [Optional Values Across API Boundaries](/patterns/optional-values-across-boundaries) |
| How to structure one request with several sibling tasks | [Structured Concurrency](/advanced/structured-concurrency) |
| How to stop duplicate cache-miss fetches | [Singleflight](/advanced/singleflight) |
| How to survive overload instead of just failing later | [Backpressure and Load Shedding](/advanced/backpressure-load-shedding) |
| How to keep AI-generated Go code from compiling successfully and still shipping bugs | [AI-Assisted Go Safety](/testing/ai-assisted-go-safety) |
| How to test timeout-heavy code without real sleeps | [Deterministic Tests with synctest](/testing/synctest) |
| How to keep Postgres, Redis, and container-backed integration tests clean, isolated, and deterministic | [Integration Testing with Testcontainers](/testing/integration-testcontainers) |
| How Docker, containerd, and Kubernetes divide product UX, runtime lifecycle, and control-plane ownership | [Docker, containerd, and Kubernetes](/production/docker-containerd-kubernetes) |
| How a durable workflow engine like Temporal becomes a Go system of history, matching, and workers | [Temporal and Durable Execution](/production/temporal-durable-execution) |
| How to combine event sourcing, erasure, retention, and audit constraints in a real Go system | [Regulated Go Systems](/production/regulated-systems) |
| Why so many influential infrastructure projects ended up in Go in the first place | [Go Open-Source Histories](/production/go-open-source-histories) |
| What large Go production systems care about beyond toy patterns | [Large-Scale Go Systems](/production/large-scale-go-systems) |
| How major Go projects actually build queues, transport loops, stopper lifetimes, and pools | [Open-Source Case Studies](/production/open-source-case-studies) |
| How to avoid compile-clean Go footguns around `select`, typed nil, slices, contexts, and stdlib contracts | [Go Pitfalls Appendix](/extras/go-pitfalls/) |
| How Go differs from Rust Tokio's async runtime model | [Go CSP vs Rust Tokio](/extras/go-csp-vs-rust-tokio) |
| When Go should stay the default and when a subsystem should move to Rust | [Go vs Rust Decision Guide](/extras/go-vs-rust-decision-guide) |

## What Makes This Site Different

<div class="signal-strip">
  <div class="signal">
    <strong>Not toy examples</strong>
    <span>The examples are shaped like backend systems you would actually maintain.</span>
  </div>
  <div class="signal">
    <strong>Not just code dumps</strong>
    <span>The pages explain ownership, ordering, failure policy, and what the tests are proving.</span>
  </div>
  <div class="signal">
    <strong>Not surface-level theory</strong>
    <span>The site now goes from runtime source concepts such as `hchan`, `sudog`, run queues, and starvation mode down to compiler SSA, allocator tiers, and interface metadata.</span>
  </div>
</div>

## Recommended Reading Flow

1. Read [Getting Started](/guide/getting-started) to understand the repo layout and validation commands.
2. Read [How to Read the Examples](/guide/how-to-read) to set the review lens.
3. Work through [Fundamentals Overview](/fundamentals/) before jumping into implementation patterns.
4. Read [Internals Overview](/internals/) when you want compiler, allocator, and type-system cost models rather than only runtime APIs.
5. Read [Standard Library Overview](/stdlib/) when you want to understand how real Go services express lifetime, I/O, SQL, and time semantics.
6. Read [Playbooks Overview](/playbooks/) when you want safe defaults and operator-facing rules for libraries you actually deploy.
7. Pick the practical pattern that matches your workload in [Patterns Overview](/patterns/).
8. Read [Testing Overview](/testing/) before treating any concurrent component as production-ready.
9. Read [Production Overview](/production/) for operating rules and open-source case studies.
10. Finish with [Advanced Overview](/advanced/) and [Extras Overview](/extras/) for deeper design and ecosystem comparison.
