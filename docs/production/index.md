---
title: Production Overview
description: Learn the concurrency concerns that matter once Go code is serving real traffic at scale.
---

# Production Overview

The earlier sections explain how Go concurrency works and how to apply specific patterns.

This section is about what changes when the code is no longer a neat example and starts carrying real production load.

## What this section covers

| Topic | Why it matters |
| --- | --- |
| [Large-Scale Go Systems](/production/large-scale-go-systems) | Turns runtime knowledge into operating rules for latency, memory, shutdown, and overload |
| [Open-Source Case Studies](/production/open-source-case-studies) | Shows how Kubernetes, etcd, Prometheus, NATS, gRPC-Go, CockroachDB, and go-redis make concurrency policy visible in real source code |

## The production shift

In examples, the main question is often "is this pattern correct?"

In production, the questions become:

- who owns goroutine lifetime,
- where do we reject load,
- what happens during deploy and shutdown,
- what does the runtime do under heap pressure,
- which hot loops or hot locks dominate tail latency,
- how do we observe all of that without guessing.

## Practical takeaway

If fundamentals teach you why Go concurrency works, the production section teaches you how not to lose that advantage at scale.
