# Go Concurrency Patterns

Expert-level Go concurrency fundamentals, practical patterns, testing guidance, and bilingual English/Korean documentation built with VitePress.

## Where to view it

- GitHub repository: `https://github.com/jaeyoung0509/golang-conccurency-patterns`
- Expected GitHub Pages site: `https://jaeyoung0509.github.io/golang-conccurency-patterns/`
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
- GitHub Actions page: `https://github.com/jaeyoung0509/golang-conccurency-patterns/actions`

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

## Included advanced topics

- Structured concurrency with `errgroup`
- Weighted semaphore admission control
- Singleflight request coalescing
- Actor pattern
- CSP theory in Go
- Backpressure and load shedding

## Included testing topics

- Race detector strategy
- Deterministic tests with `testing/synctest`
- Leak, shutdown, and timeout testing
- Tracing, contention, and profiling workflow

## Included extras

- Go CSP vs Rust Tokio comparison
