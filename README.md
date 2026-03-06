# Go Concurrency Patterns

Expert-level Go concurrency fundamentals, standard library deep dives, practical patterns, testing guidance, and bilingual English/Korean documentation built with VitePress.

## Where to view it

- GitHub repository: `https://github.com/jaeyoung0509/golang-handbook`
- Expected GitHub Pages site: `https://jaeyoung0509.github.io/golang-handbook/`
- English docs source: `docs/`
- Korean docs source: `docs/ko/`
- Runnable examples and tests: `examples/`

If the GitHub Pages URL still returns `404`, go to:

1. GitHub repository `Settings`
2. `Pages`
3. `Build and deployment`
4. Set the source to `GitHub Actions`

After that, pushes to `develop` will deploy the site automatically through the `Deploy docs` workflow.

## Run locally

```bash
npm install
npm run docs:dev
go test ./...
```

VitePress will print a local preview URL in the terminal, usually `http://localhost:5173/`.

## Deployment

- Deployment branch flow: push to `develop`
- Deployment workflow: `.github/workflows/deploy-docs.yml`
- GitHub Actions page: `https://github.com/jaeyoung0509/golang-handbook/actions`

## Included patterns

- Worker pool
- Pipeline
- Fan-out / fan-in
- Context cancellation
- Channel of channels
- Graceful shutdown
- Or-done / tee / bridge
- Structured concurrency (`errgroup`)
- Weighted semaphore
- Singleflight request coalescing
- Actor pattern

## Included foundations

- Go runtime evolution and release-history checkpoints
- Go runtime and scheduler
- Netpoller, timers, and syscalls
- Channels, `select`, and the memory model
- Channel internals
- Mutex and runtime semaphore internals
- Map internals and Swiss Tables
- Garbage collector and Green Tea GC

## Included internals

- Compiler and toolchain internals: escape analysis, SSA, and Plan 9 assembly
- Allocator hierarchy and hybrid write barrier
- Memory layout, padding, and false sharing
- Generics implementation and interface runtime layout
- unsafe, cgo, and `runtime.Pinner`
- Modern performance tuning with PGO, execution tracing, and zero-copy I/O
- Standard library anatomy for `sync.Pool` and `reflect`

## Included standard library topics

- `context`: cancellation trees, `cancelCtx`, `timerCtx`, `Cause`, `AfterFunc`, and `WithoutCancel`
- `net` + `netip`: dial budgets, deadlines, resolver behavior, `netip.Addr` identity semantics, and listener ownership
- `crypto/tls`: handshake lifetime, `tls.Config`, ALPN, verification hooks, resumption, and hostname policy
- `net/http`: server timeout boundaries, request lifetimes, transport pooling, `persistConn`, and response-body ownership
- `database/sql`: pool internals, waiters, cleaner lifecycle, `DB.Stats`, and cancellation boundaries
- `time`: monotonic time, timer/ticker ownership, Go 1.23 timer semantics, and retry-loop design
- `sync` + `sync/atomic`: `RWMutex`, `WaitGroup`, `Once`, `sync.Map`, `sync.Cond`, and read-mostly snapshot publication
- `os/signal` + `runtime/metrics` + `net/http/pprof`: process shutdown, signal-aware cancellation, runtime telemetry, and incident-time profiling
- `os/exec`: subprocess lifecycle, `CommandContext`, copy goroutines, `WaitDelay`, `ErrDot`, and output capture
- `io` + `bufio` + `bytes`: interface-driven streaming, buffering, `io.Copy` fast paths, scanner limits, and byte-slice aliasing
- `encoding/json`: stream decoding, strictness knobs, number handling, compatibility traps, and NDJSON-friendly encoding

## Included playbooks

- `net/http`: transport reuse, timeout boundaries, response body ownership, and streaming caveats
- `grpc-go`: `NewClient`, long-lived channels, RPC deadlines, keepalive caution, and streaming ownership
- `go-redis`: pooling, protocol selection, RESP2/RESP3 tradeoffs, instrumentation, and timeout policy
- `Kafka with IBM Sarama`: producer durability, consumer-group session lifecycle, offsets, and rebalance discipline

## Included advanced topics

- Structured concurrency with `errgroup`
- Weighted semaphore admission control
- Singleflight request coalescing
- Actor pattern
- CSP theory in Go
- Backpressure and load shedding

## Included production topics

- Large-scale Go systems: admission control, goroutine ownership, memory budgets, and shutdown policy
- Temporal and durable execution: history shards, task queues, worker polling, and why this kind of workflow engine fits Go well
- Open-source case studies from Kubernetes, etcd/raft, Prometheus, NATS, gRPC-Go, CockroachDB, and go-redis
- Go open-source histories: why Prometheus, NATS, etcd, Kubernetes, CockroachDB, Caddy, and Temporal landed in Go

## Included testing topics

- AI-assisted Go safety: what compile-time checks do and do not prove, plus an early-detection verification stack
- Race detector strategy
- Deterministic tests with `testing/synctest`
- Integration testing with `testcontainers-go`, teardown boundaries, clean Postgres/Redis state, and typed fixture builders
- Leak, shutdown, and timeout testing
- Tracing, contention, and profiling workflow

## Included extras

- Go CSP vs Rust Tokio comparison
- Go vs Rust decision guide for when to keep services in Go and when to move a subsystem to Rust
