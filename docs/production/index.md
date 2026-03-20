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
| [Docker, containerd, and Kubernetes](/production/docker-containerd-kubernetes) | Explains how Go daemons, runtime cores, and controller loops form a modern container platform |
| [Debugging Go Network Services in Production](/production/debugging-go-network-services) | Gives you a phase-by-phase debugging workflow for DNS, connect, TLS, request, body, and socket-state failures |
| [Kubernetes and Service Networking for Go Engineers](/production/kubernetes-service-networking) | Explains how pod DNS, Services, draining, proxies, and cluster networking reshape Go transport behavior |
| [Regulated Go Systems](/production/regulated-systems) | Explains how event sourcing, cryptographic erase, retention, and audit constraints change Go system design |
| [Temporal and Durable Execution](/production/temporal-durable-execution) | Shows how a modern workflow engine becomes a Go system of history shards, task queues, and worker polling |
| [Open-Source Case Studies](/production/open-source-case-studies) | Shows how Kubernetes, etcd, Prometheus, NATS, gRPC-Go, CockroachDB, and go-redis make concurrency policy visible in real source code |
| [Go Open-Source Histories](/production/go-open-source-histories) | Explains why Go became such a strong fit for infrastructure software and where to read the story in major projects |

## The production shift

In examples, the main question is often "is this pattern correct?"

In production, the questions become:

- who owns goroutine lifetime,
- where do we reject load,
- what happens during deploy and shutdown,
- how to prove whether latency is DNS, connect, TLS, handler, or body ownership,
- how cluster networking and proxy layers change connection semantics,
- what does the runtime do under heap pressure,
- which hot loops or hot locks dominate tail latency,
- how do we observe all of that without guessing.

## Practical takeaway

If fundamentals teach you why Go concurrency works, the production section teaches you how not to lose that advantage at scale.
