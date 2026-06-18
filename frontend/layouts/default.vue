<template>
  <div class="app-layout">
    <AppSidebar />
    <div class="app-main">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
const sidebarOpen = ref(false)
provide('sidebarOpen', sidebarOpen)
</script>

<style scoped>
.app-layout {
  display: flex;
  min-height: 100vh;
  background-color: var(--color-surface-1);
}

.app-main {
  flex: 1;
  margin-left: 240px;
}

@media (max-width: 768px) {
  .app-main {
    margin-left: 0;
  }
}
</style>
