<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useData } from "vitepress";

const storageKey = "vp-sidebar-collapsed";
const sidebarCollapsedClass = "sidebar-collapsed";

const { lang, frontmatter } = useData();
const isCollapsed = ref(false);
const isMounted = ref(false);

const isKorean = computed(() => lang.value.startsWith("ko"));
const buttonLabel = computed(() =>
  isKorean.value
    ? isCollapsed.value
      ? "사이드바 펼치기"
      : "사이드바 접기"
    : isCollapsed.value
      ? "Show sidebar"
      : "Hide sidebar"
);

const helperText = computed(() =>
  isKorean.value ? "읽기 집중 모드" : "Focus mode"
);

const showToggle = computed(() => frontmatter.value.layout !== "home");

function applyCollapsedState(value: boolean) {
  if (typeof document === "undefined") {
    return;
  }

  document.documentElement.classList.toggle(sidebarCollapsedClass, value);
}

function toggleSidebar() {
  isCollapsed.value = !isCollapsed.value;
}

onMounted(() => {
  const stored = window.localStorage.getItem(storageKey);
  isCollapsed.value = stored === "1";
  applyCollapsedState(isCollapsed.value);
  isMounted.value = true;
});

watch(isCollapsed, (value) => {
  applyCollapsedState(value);

  if (!isMounted.value) {
    return;
  }

  window.localStorage.setItem(storageKey, value ? "1" : "0");
});
</script>

<template>
  <div
    v-if="showToggle"
    class="sidebar-toggle-shell"
    :class="{ 'is-collapsed': isCollapsed }"
  >
    <button
      type="button"
      class="sidebar-toggle-button"
      :aria-label="buttonLabel"
      :aria-pressed="isCollapsed"
      @click="toggleSidebar"
    >
      <span class="sidebar-toggle-kicker">{{ helperText }}</span>
      <span class="sidebar-toggle-label">{{ buttonLabel }}</span>
    </button>
  </div>
</template>
