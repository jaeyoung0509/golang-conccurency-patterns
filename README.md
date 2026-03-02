# Go Concurrency Patterns

Practical Go concurrency patterns documented with VitePress.

- English docs: `docs/`
- Korean docs: `docs/ko/`
- Runnable examples and tests: `examples/`
- Fundamentals: scheduler, channels, memory model
- Advanced topics: actor pattern, CSP theory

## Run locally

```bash
npm install
npm run docs:dev
go test ./...
```

## Included patterns

- Worker pool
- Pipeline
- Fan-out / fan-in
- Context cancellation
- Structured concurrency (`errgroup`)
- Weighted semaphore
- Singleflight request coalescing
- Actor pattern
