---
title: Extras Overview
description: Comparative and ecosystem-oriented pages that widen the mental model beyond idiomatic Go examples.
---

# Extras Overview

The main site is Go-first.

This extra section exists for two reasons:

- to compare Go's concurrency model with adjacent systems,
- to help experienced engineers transfer mental models across runtimes.

## Current extra topics

| Topic | Why it matters |
| --- | --- |
| [Go CSP vs Rust Tokio](/extras/go-csp-vs-rust-tokio) | Shows how similar high-level goals lead to very different runtime and language designs |
| [Go vs Rust Decision Guide](/extras/go-vs-rust-decision-guide) | Turns language comparison into a practical call about when to keep a service in Go and when to move a subsystem to Rust |

## Practical takeaway

You do not master Go by pretending other ecosystems do not exist.

You master it by understanding exactly what Go chose to integrate into the language and runtime, and what systems like Rust plus Tokio choose to express with libraries and async state machines.
