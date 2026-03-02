---
title: Race Detector
description: Go race detector가 무엇을 잡고 무엇을 보장하지 않는지 설명합니다.
---

# Race Detector

nontrivial한 동시성 코드라면 가장 먼저 돌려야 하는 도구가 race detector입니다.

## 무엇을 잡나

같은 메모리 위치에 대한 concurrent access 중 하나 이상이 write이고, 그 사이에 동기화 경계가 없을 때 발생하는 data race를 잡습니다.

대표 예시는:

- 한 goroutine이 map을 읽고 다른 goroutine이 동시에 쓰는 경우,
- channel protocol이 보호해야 할 상태를 바깥에서 따로 읽거나 쓰는 경우,
- cache update에 lock을 빠뜨린 경우.

## 무엇을 보장하지 않나

`-race`가 깨끗하다고 해서 다음이 보장되지는 않습니다.

- deadlock 부재,
- goroutine leak 부재,
- fairness,
- 좋은 shutdown behavior,
- 좋은 timeout behavior.

또한 실제로 실행된 경로만 관측할 수 있습니다.

## 사용법

```bash
go test -race ./...
```

느리고 무거운 건 정상입니다. 동시성 코드에선 그 비용을 감수할 가치가 충분합니다.

## 좋은 사용 방식

이 도구는 이런 질문에 답합니다.

- mutable memory를 실수로 공유했는가
- synchronization boundary를 우회하는 경로가 생겼는가
- 리팩터링이 ownership contract를 깨뜨렸는가

하지만 이것만으로 validation을 끝내면 안 됩니다.

## 공식 자료

- [Data Race Detector](https://go.dev/doc/articles/race_detector)
- [Go Memory Model](https://go.dev/ref/mem)

## Practical takeaway

`-race`는 최소 기준이지 종착점이 아닙니다.

초기에 자주 돌리고, deterministic test와 shutdown test와 함께 써야 합니다.
