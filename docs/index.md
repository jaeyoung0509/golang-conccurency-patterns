---
layout: home

hero:
  name: Go Concurrency Patterns
  text: From runtime internals to production-safe patterns
  tagline: Learn Go concurrency through deep fundamentals, tested examples, decision guides, and bilingual English/Korean docs.
  actions:
    - theme: brand
      text: Start With Fundamentals
      link: /fundamentals/
    - theme: alt
      text: Browse Patterns
      link: /patterns/
    - theme: alt
      text: 한국어 보기
      link: /ko/

features:
  - title: Clear learning path
    details: "The site now has section overviews, reading tracks, and pattern selection guides instead of throwing readers into dense pages."
  - title: Runtime-first depth
    details: "The fundamentals section explains the scheduler, memory model, channel internals, and mutex/runtime semaphore behavior."
  - title: Practical examples
    details: "Examples use realistic backend domains such as shipping, fraud analysis, inventory coordination, and cache-miss suppression."
  - title: Tested behavior
    details: "Every example is backed by `go test` so ordering, cancellation, limits, and failure policy are verified."
  - title: Advanced production topics
    details: "Structured concurrency, weighted semaphores, singleflight, actors, and load shedding are documented alongside the basics."
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
    <h3>2. Practical Patterns</h3>
    <p>Move into worker pools, pipelines, fan-out/fan-in, and context cancellation when you are mapping code to real workloads.</p>
    <p><a href="/patterns/">Browse patterns</a></p>
  </div>
  <div class="path-card">
    <h3>3. Advanced Topics</h3>
    <p>Study resource budgeting, structured lifetimes, duplicate suppression, ownership models, and overload behavior.</p>
    <p><a href="/advanced/">Go deeper</a></p>
  </div>
</div>

## Choose The Right Starting Point

| If you need to understand... | Start with |
| --- | --- |
| Why goroutines are cheap and how the scheduler actually runs them | [Go Runtime and Scheduler](/fundamentals/go-runtime-scheduler) |
| Why channels synchronize memory visibility | [Channels, Select, and the Memory Model](/fundamentals/channels-memory-model) |
| How to cap parallelism across many independent tasks | [Worker Pool](/patterns/worker-pool) |
| How to structure one request with several sibling tasks | [Structured Concurrency](/advanced/structured-concurrency) |
| How to stop duplicate cache-miss fetches | [Singleflight](/advanced/singleflight) |
| How to survive overload instead of just failing later | [Backpressure and Load Shedding](/advanced/backpressure-load-shedding) |

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
    <span>The fundamentals section goes down to runtime source concepts such as `hchan`, `sudog`, run queues, and starvation mode.</span>
  </div>
</div>

## Recommended Reading Flow

1. Read [Getting Started](/guide/getting-started) to understand the repo layout and validation commands.
2. Read [How to Read the Examples](/guide/how-to-read) to set the review lens.
3. Work through [Fundamentals Overview](/fundamentals/) before jumping into implementation patterns.
4. Pick the practical pattern that matches your workload in [Patterns Overview](/patterns/).
5. Finish with [Advanced Overview](/advanced/) when you need stronger production control.
