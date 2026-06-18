<template>
  <div>
      <div v-if="loading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else class="routes-grid">
        <div v-for="route in routes" :key="route.id" class="route-card">
          <div class="route-icon">
            <component :is="getIcon(route.icon)" :size="20" />
          </div>
          <div class="route-info">
            <div class="route-name">{{ route.name }}</div>
            <div class="route-desc">{{ route.description }}</div>
            <div class="route-meta">
              <span class="route-path">{{ route.path }}</span>
              <span class="route-category">{{ route.category }}</span>
            </div>
          </div>
          <div class="route-actions">
            <button class="toggle" :class="{ active: route.enabled }" @click="toggleRoute(route)">
              <span class="toggle-knob"></span>
            </button>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2, FolderOpen, Settings, ShieldCheck, Users, Circle } from 'lucide-vue-next'

const { apiFetch } = useApi()
const { can, fetchPermissions } = usePermissions()
const routes = ref<any[]>([])
const loading = ref(true)

const topbar = useState('topbar')
topbar.value = { title: 'Route Management', subtitle: 'Enable or disable features', currentPage: 'routes' }

const iconMap: Record<string, any> = {
  FolderOpen,
  Settings,
  ShieldCheck,
  Users,
  Circle,
}

const getIcon = (name: string) => iconMap[name] || Circle

const loadRoutes = async () => {
  loading.value = true
  try {
    routes.value = (await apiFetch('/api/routes')) as any[]
  } catch {}
  loading.value = false
}

const toggleRoute = async (route: any) => {
  try {
    await apiFetch(`/api/routes/${route.id}`, {
      method: 'PUT',
      body: { enabled: !route.enabled }
    })
    route.enabled = !route.enabled
  } catch {}
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  await fetchPermissions()
  if (!can('nav.route_management')) { navigateTo('/'); return }
  loadRoutes()
})
</script>

<style scoped>
.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.routes-grid {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.route-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 0.5rem;
  transition: background-color 0.12s ease;
}

.route-card:hover {
  background-color: var(--color-surface-1);
}

.route-icon {
  width: 2rem;
  height: 2rem;
  background-color: var(--color-brand-50);
  color: var(--color-brand-600);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.route-info {
  flex: 1;
  min-width: 0;
}

.route-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.route-desc {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.route-meta {
  display: flex;
  gap: 0.5rem;
  margin-top: 0.375rem;
}

.route-path {
  font-size: 0.625rem;
  font-family: monospace;
  color: var(--color-text-muted);
  background-color: var(--color-surface-2);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

.route-category {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-brand-600);
  background-color: var(--color-brand-50);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  text-transform: uppercase;
}

.route-actions {
  flex-shrink: 0;
}

.toggle {
  position: relative;
  width: 40px;
  height: 22px;
  background-color: var(--color-surface-3);
  border: none;
  border-radius: 9999px;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.toggle.active {
  background-color: var(--color-brand-600);
}

.toggle-knob {
  position: absolute;
  top: 3px;
  left: 3px;
  width: 16px;
  height: 16px;
  background-color: white;
  border-radius: 9999px;
  transition: transform 0.2s ease;
}

.toggle.active .toggle-knob {
  transform: translateX(18px);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
