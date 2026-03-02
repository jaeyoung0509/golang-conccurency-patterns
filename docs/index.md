---
layout: home

hero:
  name: Go Concurrency Patterns
  text: Practical, bilingual, and production-minded
  tagline: Learn Go concurrency from runtime fundamentals to advanced patterns through detailed explanations, tested code, and Mermaid diagrams.
  actions:
    - theme: brand
      text: Start Reading
      link: /guide/getting-started
    - theme: alt
      text: 한국어 보기
      link: /ko/
    - theme: alt
      text: GitHub Repository
      link: https://github.com/jaeyoung0509/golang-conccurency-patterns

features:
  - title: English and Korean
    details: "The site ships with mirrored English and Korean navigation so teams can study in the language they prefer."
  - title: Fundamentals included
    details: "The site now explains the G-M-P scheduler, goroutine cost model, channels, select, and memory-visibility rules."
  - title: Real examples
    details: "Every section is backed by practical Go packages such as shipment quoting, checkout risk pipelines, dashboard aggregation, and inventory actors."
  - title: Tests included
    details: "Patterns are verified with `go test`, including worker limits, cancellation behavior, and deadline propagation."
  - title: Mermaid diagrams
    details: "Each pattern has a visual explanation so channel ownership and control flow are obvious before you copy code."
  - title: Advanced topics
    details: "Actor pattern and CSP theory are documented alongside practical Go tradeoffs, not as detached theory notes."
  - title: Readability first
    details: "Content is organized around tradeoffs, failure modes, and implementation boundaries instead of dumping long code listings."
---

## What this site is for

Most concurrency tutorials stop at toy examples. This project takes the opposite approach:

- the examples solve operational problems that look like real backend work,
- the tests prove the concurrency guarantees instead of assuming them,
- the documents explain why a pattern is safe, not just how to type it.

<div class="custom-card-grid">
  <div class="custom-card">
    <h3>Fundamentals</h3>
    <p>Understand the scheduler, channels, and memory model before copying any concurrency pattern into production.</p>
  </div>
  <div class="custom-card">
    <h3>Worker Pool</h3>
    <p>Bound parallelism while preserving order and cancelling on the first failure.</p>
  </div>
  <div class="custom-card">
    <h3>Pipeline</h3>
    <p>Compose stages so ingestion, scoring, and alert generation stay readable under load.</p>
  </div>
  <div class="custom-card">
    <h3>Fan-Out / Fan-In</h3>
    <p>Query multiple backends concurrently and combine partial results without losing useful signal.</p>
  </div>
  <div class="custom-card">
    <h3>Context Cancellation</h3>
    <p>Build request-scoped workflows that stop fast when deadlines or sibling failures happen.</p>
  </div>
  <div class="custom-card">
    <h3>Advanced Topics</h3>
    <p>Go beyond basics with structured concurrency, weighted semaphores, singleflight, actor-style ownership, and load shedding.</p>
  </div>
</div>

## Reading order

1. Start with [Getting Started](/guide/getting-started) for repository layout and commands.
2. Read [How to Read the Examples](/guide/how-to-read) to understand the review checklist.
3. Build the runtime mental model in [Fundamentals](/fundamentals/go-runtime-scheduler).
4. Move into the practical pattern pages based on the problem you are solving.
5. Finish with [Advanced](/advanced/actor-pattern) when you want to compare communication-first and ownership-first designs.
