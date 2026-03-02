---
title: Getting Started
description: Repository structure, commands, and study approach for this VitePress-based Go concurrency guide.
---

# Getting Started

This repository is a documentation site and a Go example repository at the same time.

The documentation lives in `docs/`, while the runnable examples and tests live in `examples/`.
That split matters because the site explains the patterns, but the Go packages prove they actually work.

## Repository layout

```text
.
├── docs/
│   ├── .vitepress/
│   ├── fundamentals/
│   ├── advanced/
│   ├── guide/
│   ├── patterns/
│   └── ko/
├── examples/
│   ├── actor/
│   ├── contexttimeout/
│   ├── errgroupbatch/
│   ├── fanoutfanin/
│   ├── pipeline/
│   ├── singleflightcache/
│   ├── weightedsemaphore/
│   └── workerpool/
├── .github/workflows/
├── go.mod
└── package.json
```

## Commands you will use

```bash
npm install
npm run docs:dev
```

Use the commands above while writing or revising the site.

```bash
npm run docs:build
go test ./...
```

Use the commands above before pushing changes. The first validates the static site build. The second validates the Go examples.

## What is inside each example

Each package in `examples/` follows the same structure:

- a practical domain model instead of a toy `hello world` payload,
- one main exported function that shows the pattern boundary clearly,
- tests that verify concurrency guarantees such as bounded parallelism or cancellation.

The actor example adds one more angle: serialized state ownership without external mutex sharing.
The advanced examples add three more angles: structured cancellation with `errgroup`, weighted concurrency with semaphores, and duplicate suppression with `singleflight`.

:::tip Why this is useful
If you cannot explain what a test is proving about goroutine behavior, you probably do not understand the pattern well enough to use it in production.
:::

## Documentation philosophy

The pattern pages are written to answer four questions:

1. What operational problem is this pattern solving?
2. Where are the ownership boundaries for channels, goroutines, and cancellation?
3. What failure mode does the example optimize for?
4. What should you test before copying this structure into a service?

The new fundamentals pages answer a fifth question: why can Go support these patterns efficiently in the first place?

## Deployment model

The repository includes a GitHub Pages workflow that builds VitePress on every push to `main`.
The configured base path is `/golang-conccurency-patterns/`, which matches the current repository name.

If you rename the repository later, update the `base` value in `docs/.vitepress/config.mts`.
