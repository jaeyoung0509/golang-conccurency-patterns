---
title: Getting Started
description: Repository structure, commands, and study approach for this VitePress-based Go concurrency guide.
---

# Getting Started

<div class="lead-panel">
  <p>
    If the site feels large, do not start by clicking random deep pages.
    Start here, understand the repository shape, then use the overview pages to enter the right section.
  </p>
</div>

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
│   ├── production/
│   ├── guide/
│   ├── patterns/
│   ├── testing/
│   ├── extras/
│   └── ko/
├── examples/
│   ├── actor/
│   ├── contexttimeout/
│   ├── errgroupbatch/
│   ├── fanoutfanin/
│   ├── gracefulshutdown/
│   ├── pipeline/
│   ├── requestreply/
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

## Fast path for first-time readers

1. Read [Fundamentals Overview](/fundamentals/).
2. Move through [Patterns Overview](/patterns/) with the matching package under `examples/`.
3. Read [Testing Overview](/testing/) before trusting any timeout or shutdown path.
4. Read [Production Overview](/production/) once you care about operating rules and large-scale system tradeoffs.
5. Only then move into [Advanced Overview](/advanced/) and [Extras Overview](/extras/).

## What is inside each example

Each package in `examples/` follows the same structure:

- a practical domain model instead of a toy `hello world` payload,
- one main exported function that shows the pattern boundary clearly,
- tests that verify concurrency guarantees such as bounded parallelism or cancellation.

The actor example adds one more angle: serialized state ownership without external mutex sharing.
The advanced examples add three more angles: structured cancellation with `errgroup`, weighted concurrency with semaphores, and duplicate suppression with `singleflight`.
The newer patterns add request/reply with embedded reply channels and graceful draining on shutdown.

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
The testing pages answer a sixth: how do you prove the behavior rather than merely describe it?

## Deployment model

The repository includes a GitHub Pages workflow that builds VitePress on pushes to `develop`.
The configured base path is `/golang-handbook/`, which matches the current repository name.

If you rename the repository later, update the `base` value in `docs/.vitepress/config.mts`.
