<script setup lang="ts">
import { computed, onMounted, ref, watch } from "vue";
import { useData } from "vitepress";

const storageKey = "vp-sidebar-collapsed";
const sidebarCollapsedClass = "sidebar-collapsed";

const { lang, frontmatter } = useData();
const isCollapsed = ref(false);
const isMounted = ref(false);

const buttonLabel = computed(() =>
  lang.value.startsWith("ko")
    ? isCollapsed.value
      ? "사이드바 펼치기"
      : "사이드바 접기"
    : isCollapsed.value
      ? "Show sidebar"
      : "Hide sidebar"
);
const icon = computed(() => (isCollapsed.value ? ">" : "<"));

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
  <div v-if="showToggle" class="sidebar-toggle-shell">
    <button
      type="button"
      class="sidebar-toggle-button"
      :aria-label="buttonLabel"
      :aria-pressed="isCollapsed"
      :title="buttonLabel"
      @click="toggleSidebar"
    >
      <span class="sidebar-toggle-icon" aria-hidden="true">{{ icon }}</span>
      <span class="sidebar-toggle-text">{{ buttonLabel }}</span>
    </button>
  </div>
</template>
