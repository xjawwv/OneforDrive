<template>
  <div class="app-layout">
    <AppSidebar />
    <div class="app-main">
      <AppTopBar :title="topbar.title" :subtitle="topbar.subtitle" :current-page="topbar.currentPage">
        <template v-if="topbarTitleFn" #title>
          <component :is="topbarTitleFn" />
        </template>
        <template v-if="topbarActionsFn" #actions>
          <component :is="topbarActionsFn" />
        </template>
      </AppTopBar>
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
const sidebarOpen = ref(false)
provide('sidebarOpen', sidebarOpen)

const topbar = useState('topbar', () => ({
  title: '',
  subtitle: '',
  currentPage: ''
}))

const topbarTitleFn = inject<Ref<(() => any) | null>>('topbar:title', ref(null))
const topbarActionsFn = inject<Ref<(() => any) | null>>('topbar:actions', ref(null))

provide('topbar:config', topbar)
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
