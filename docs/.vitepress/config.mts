import { defineConfig } from "vitepress";

const enSidebar = [
  {
    text: "Guide",
    items: [
      { text: "Getting Started", link: "/guide/getting-started" },
      { text: "How to Read the Examples", link: "/guide/how-to-read" },
    ],
  },
  {
    text: "Fundamentals",
    items: [
      { text: "Go Runtime and Scheduler", link: "/fundamentals/go-runtime-scheduler" },
      { text: "Channels, Select, and the Memory Model", link: "/fundamentals/channels-memory-model" },
      { text: "Channel Internals", link: "/fundamentals/channel-internals" },
      { text: "Mutex and Runtime Semaphore Internals", link: "/fundamentals/mutex-semaphore-internals" },
    ],
  },
  {
    text: "Patterns",
    items: [
      { text: "Worker Pool", link: "/patterns/worker-pool" },
      { text: "Pipeline", link: "/patterns/pipeline" },
      { text: "Fan-Out / Fan-In", link: "/patterns/fan-out-fan-in" },
      { text: "Context Cancellation", link: "/patterns/context-cancellation" },
    ],
  },
  {
    text: "Advanced",
    items: [
      { text: "Structured Concurrency", link: "/advanced/structured-concurrency" },
      { text: "Weighted Semaphore", link: "/advanced/weighted-semaphore" },
      { text: "Singleflight", link: "/advanced/singleflight" },
      { text: "Backpressure and Load Shedding", link: "/advanced/backpressure-load-shedding" },
      { text: "Actor Pattern", link: "/advanced/actor-pattern" },
      { text: "CSP Theory in Go", link: "/advanced/csp-theory" },
    ],
  },
] as const;

const koSidebar = [
  {
    text: "가이드",
    items: [
      { text: "시작하기", link: "/ko/guide/getting-started" },
      { text: "예제 읽는 법", link: "/ko/guide/how-to-read" },
    ],
  },
  {
    text: "기초 원리",
    items: [
      { text: "Go 런타임과 스케줄러", link: "/ko/fundamentals/go-runtime-scheduler" },
      { text: "Channels, Select, 그리고 Memory Model", link: "/ko/fundamentals/channels-memory-model" },
      { text: "채널 내부 동작", link: "/ko/fundamentals/channel-internals" },
      { text: "Mutex와 런타임 세마포어 내부", link: "/ko/fundamentals/mutex-semaphore-internals" },
    ],
  },
  {
    text: "패턴",
    items: [
      { text: "워커 풀", link: "/ko/patterns/worker-pool" },
      { text: "파이프라인", link: "/ko/patterns/pipeline" },
      { text: "팬아웃 / 팬인", link: "/ko/patterns/fan-out-fan-in" },
      { text: "컨텍스트 취소", link: "/ko/patterns/context-cancellation" },
    ],
  },
  {
    text: "고급 주제",
    items: [
      { text: "구조화된 동시성", link: "/ko/advanced/structured-concurrency" },
      { text: "가중 세마포어", link: "/ko/advanced/weighted-semaphore" },
      { text: "Singleflight", link: "/ko/advanced/singleflight" },
      { text: "역압력과 로드 셰딩", link: "/ko/advanced/backpressure-load-shedding" },
      { text: "액터 패턴", link: "/ko/advanced/actor-pattern" },
      { text: "Go에서의 CSP 이론", link: "/ko/advanced/csp-theory" },
    ],
  },
] as const;

export default defineConfig({
  title: "Go Concurrency Patterns",
  description: "Detailed, practical Go concurrency patterns with tests and Mermaid diagrams.",
  base: "/golang-conccurency-patterns/",
  cleanUrls: true,
  lastUpdated: true,
  head: [
    ["meta", { name: "theme-color", content: "#115e59" }],
  ],
  markdown: {
    config(md) {
      const fence = md.renderer.rules.fence;

      md.renderer.rules.fence = (tokens, idx, options, env, self) => {
        const token = tokens[idx];

        if (token.info.trim() === "mermaid") {
          return `<MermaidDiagram :chart='${JSON.stringify(token.content)}' />`;
        }

        return fence ? fence(tokens, idx, options, env, self) : self.renderToken(tokens, idx, options);
      };
    },
  },
  themeConfig: {
    search: {
      provider: "local",
    },
    socialLinks: [
      {
        icon: "github",
        link: "https://github.com/jaeyoung0509/golang-conccurency-patterns",
      },
    ],
    footer: {
      message: "Built with VitePress and real Go examples.",
      copyright: "MIT",
    },
  },
  locales: {
    root: {
      label: "English",
      lang: "en-US",
      title: "Go Concurrency Patterns",
      description: "Detailed, practical Go concurrency patterns with tests and Mermaid diagrams.",
      themeConfig: {
        nav: [
          { text: "Guide", link: "/guide/getting-started" },
          { text: "Fundamentals", link: "/fundamentals/go-runtime-scheduler" },
          { text: "Patterns", link: "/patterns/worker-pool" },
          { text: "Advanced", link: "/advanced/actor-pattern" },
          {
            text: "GitHub",
            link: "https://github.com/jaeyoung0509/golang-conccurency-patterns",
          },
        ],
        sidebar: enSidebar,
        outline: {
          level: [2, 3],
        },
        outlineTitle: "On this page",
        docFooter: {
          prev: "Previous page",
          next: "Next page",
        },
        darkModeSwitchLabel: "Appearance",
        sidebarMenuLabel: "Menu",
        returnToTopLabel: "Back to top",
        langMenuLabel: "Languages",
      },
    },
    ko: {
      label: "한국어",
      lang: "ko-KR",
      link: "/ko/",
      title: "Go Concurrency Patterns",
      description: "테스트와 Mermaid 다이어그램까지 포함한 실전 Go 동시성 패턴 문서.",
      themeConfig: {
        nav: [
          { text: "가이드", link: "/ko/guide/getting-started" },
          { text: "기초 원리", link: "/ko/fundamentals/go-runtime-scheduler" },
          { text: "패턴", link: "/ko/patterns/worker-pool" },
          { text: "고급 주제", link: "/ko/advanced/actor-pattern" },
          {
            text: "GitHub",
            link: "https://github.com/jaeyoung0509/golang-conccurency-patterns",
          },
        ],
        sidebar: koSidebar,
        outline: {
          level: [2, 3],
        },
        outlineTitle: "이 페이지에서",
        docFooter: {
          prev: "이전 페이지",
          next: "다음 페이지",
        },
        darkModeSwitchLabel: "테마",
        sidebarMenuLabel: "메뉴",
        returnToTopLabel: "맨 위로",
        langMenuLabel: "언어",
      },
    },
  },
});
