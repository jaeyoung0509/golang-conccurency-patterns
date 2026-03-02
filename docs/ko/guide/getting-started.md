---
title: 시작하기
description: VitePress 기반 Go 동시성 문서 저장소의 구조와 실행 방법을 설명합니다.
---

# 시작하기

<div class="lead-panel">
  <p>
    사이트가 커 보인다면 처음부터 깊은 페이지를 무작정 읽지 않는 편이 낫습니다.
    먼저 저장소 구조를 이해하고, 섹션 개요 페이지를 통해 들어가면 훨씬 덜 거칠게 느껴집니다.
  </p>
</div>

이 저장소는 문서 사이트이면서 동시에 실제 Go 예제 저장소입니다.

문서는 `docs/` 아래에 있고, 실행 가능한 예제와 테스트는 `examples/` 아래에 있습니다.
이 분리는 중요합니다. 문서는 패턴을 설명하고, Go 패키지는 그 설명이 실제로 맞는지 증명합니다.

## 저장소 구조

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

## 자주 쓰는 명령어

```bash
npm install
npm run docs:dev
```

위 명령어는 문서를 작성하거나 디자인을 다듬을 때 사용합니다.

```bash
npm run docs:build
go test ./...
```

위 명령어는 푸시 전에 반드시 확인해야 하는 검증입니다.
첫 번째는 정적 사이트 빌드 검증, 두 번째는 Go 예제의 동작 검증입니다.

## 처음 읽는 사람을 위한 빠른 경로

1. [기초 원리 개요](/ko/fundamentals/)부터 읽습니다.
2. [패턴 개요](/ko/patterns/)에서 자기 문제와 가장 가까운 패턴 하나를 고릅니다.
3. 구현 디테일에 들어가기 전에 `examples/`의 테스트 파일을 먼저 읽습니다.
4. 그 다음에 [고급 주제 개요](/ko/advanced/)로 넘어갑니다.

## 예제 패키지 구성 원칙

`examples/` 아래 패키지는 모두 같은 기준을 따릅니다.

- 단순 `hello world` 대신 실제 서비스 상황에 가까운 도메인 사용
- 패턴 경계가 명확한 대표 함수 하나 중심으로 설계
- 동시성 제한, 취소 전파, 부분 실패 처리 같은 계약을 테스트로 검증

액터 예제는 여기에 한 가지를 더 보여줍니다. 외부 mutex 공유 없이 상태 소유권을 직렬화하는 방식입니다.
새 고급 예제들은 `errgroup` 기반 구조화된 동시성, 가중 세마포어, `singleflight` 기반 중복 억제를 같이 보여줍니다.

:::tip 중요한 기준
테스트가 "어떤 동시성 보장을 증명하는지" 설명할 수 없으면, 그 패턴을 아직 충분히 이해한 것이 아닙니다.
:::

## 문서 작성 철학

각 패턴 문서는 다음 네 가지 질문에 답하도록 구성했습니다.

1. 이 패턴은 어떤 운영 문제를 해결하는가?
2. 채널, 고루틴, 취소의 소유권 경계는 어디인가?
3. 어떤 실패 모드를 기준으로 설계했는가?
4. 실제 서비스에 넣기 전에 무엇을 테스트해야 하는가?

여기에 새로 추가된 기초 원리 문서는 다섯 번째 질문을 다룹니다. 왜 Go 런타임이 이런 패턴을 현실적으로 가능하게 하는가입니다.

## 배포 방식

GitHub Pages 워크플로가 포함되어 있어서 `develop` 브랜치에 push 하면 VitePress 사이트가 빌드됩니다.

현재 `docs/.vitepress/config.mts`의 `base` 값은 저장소 이름에 맞춰 `/golang-conccurency-patterns/`로 설정되어 있습니다.
저장소 이름을 바꾸면 이 값도 함께 수정해야 합니다.
