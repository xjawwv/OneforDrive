<template>
  <div v-if="loading" class="feature-gate-loading">
    <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
  </div>
  <MaintenancePage v-else-if="!enabled" :description="description" />
  <slot v-else />
</template>

<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

const props = defineProps<{ routePath: string }>()

const { enabled, loading, description, checkRoute } = useFeatureRoute(props.routePath)

checkRoute()
</script>

<style scoped>
.feature-gate-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 5rem 1.5rem;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
