<template>
  <div class="maintenance-state">
    <div class="maintenance-icon">
      <Wrench :size="48" />
    </div>
    <h2>Feature Under Maintenance</h2>
    <p>{{ description || 'This feature is temporarily unavailable. Please check back later.' }}</p>
  </div>
</template>

<script setup lang="ts">
import { Wrench } from 'lucide-vue-next'

defineProps<{ description?: string }>()
</script>

<style scoped>
.maintenance-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 5rem 1.5rem;
  gap: 1rem;
}

.maintenance-icon {
  width: 5rem;
  height: 5rem;
  border-radius: 9999px;
  background-color: var(--color-warning, #f59e0b);
  background-color: color-mix(in srgb, var(--color-warning, #f59e0b) 12%, transparent);
  color: var(--color-warning, #f59e0b);
  display: flex;
  align-items: center;
  justify-content: center;
}

.maintenance-state h2 {
  font-size: 1.25rem;
  font-weight: 700;
  color: var(--color-text-primary);
  margin: 0;
}

.maintenance-state p {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  max-width: 24rem;
  margin: 0;
  line-height: 1.5;
}
</style>
