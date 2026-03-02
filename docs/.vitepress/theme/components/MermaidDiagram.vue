<script setup lang="ts">
import { onMounted, ref, watch } from "vue";

const props = defineProps<{
  chart: string;
}>();

const svg = ref("");
const error = ref("");
let renderCount = 0;

async function renderChart() {
  const currentRender = ++renderCount;

  try {
    error.value = "";

    const mermaid = (await import("mermaid")).default;
    mermaid.initialize({
      startOnLoad: false,
      securityLevel: "loose",
      theme: "neutral",
      fontFamily: "IBM Plex Sans, sans-serif",
    });

    const { svg: rendered } = await mermaid.render(`mermaid-${currentRender}`, props.chart);

    if (currentRender === renderCount) {
      svg.value = rendered;
    }
  } catch (cause) {
    if (currentRender === renderCount) {
      svg.value = "";
      error.value = cause instanceof Error ? cause.message : "Unknown Mermaid error";
    }
  }
}

onMounted(() => {
  void renderChart();
});

watch(
  () => props.chart,
  () => {
    void renderChart();
  },
);
</script>

<template>
  <div class="mermaid-shell">
    <div v-if="error" class="mermaid-error">
      Mermaid render failed: {{ error }}
    </div>
    <div v-else class="mermaid-diagram" v-html="svg" />
  </div>
</template>
