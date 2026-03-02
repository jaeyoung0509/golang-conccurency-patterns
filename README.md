# Go Concurrency Patterns

Practical Go concurrency patterns documented with VitePress.

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
- Structured concurrency (`errgroup`)
- Weighted semaphore
- Singleflight request coalescing
- Actor pattern

## Included foundations

- Go runtime and scheduler
- Channels, `select`, and the memory model
- Channel internals
- Mutex and runtime semaphore internals

## Included advanced topics

- Structured concurrency with `errgroup`
- Weighted semaphore admission control
- Singleflight request coalescing
- Actor pattern
- CSP theory in Go
- Backpressure and load shedding
