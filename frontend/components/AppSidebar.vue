<template>
  <div>
    <div v-if="sidebarOpen" class="sidebar-backdrop" @click="sidebarOpen = false"></div>
    <aside class="sidebar" :class="{ 'sidebar-open': sidebarOpen }">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <HardDrive :size="20" color="white" :stroke-width="2" />
        </div>
        <span class="sidebar-brand">RouteStorage</span>
        <button class="sidebar-close" @click="sidebarOpen = false"><X :size="18" /></button>
      </div>

      <nav class="sidebar-nav">
        <template v-for="(group, category) in groupedRoutes" :key="category">
          <div v-if="canViewCategory(category)" class="nav-section">
            <div class="nav-label">{{ category }}</div>
            <template v-for="r in group" :key="r.id">
              <NuxtLink
                v-if="canViewRoute(r)"
                :to="r.enabled ? r.path : '#'"
                class="sidebar-link"
                :class="{ active: route.path === r.path, disabled: !r.enabled }"
                @click="handleClick(r)"
              >
                <component :is="getIcon(r.icon)" :size="18" />
                <span>{{ r.name }}</span>
                <Wrench v-if="!r.enabled" :size="12" class="maintenance-icon" />
              </NuxtLink>
            </template>
          </div>
        </template>
      </nav>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { HardDrive, FolderOpen, Settings, ShieldCheck, Users, X, Wrench, Circle, Map } from 'lucide-vue-next'

const props = defineProps<{ current?: string }>()

const { apiFetch } = useApi()
const sidebarOpen = inject<Ref<boolean>>('sidebarOpen', ref(false))
const { can, fetchPermissions } = usePermissions()
const route = useRoute()

const featureRoutes = ref<any[]>([])

const iconMap: Record<string, any> = {
  FolderOpen,
  Settings,
  ShieldCheck,
  Users,
  Circle,
  Map,
}

const getIcon = (name: string) => iconMap[name] || Circle

const groupedRoutes = computed(() => {
  const groups: Record<string, any[]> = {}
  for (const r of featureRoutes.value) {
    if (!groups[r.category]) groups[r.category] = []
    groups[r.category].push(r)
  }
  return groups
})

const canViewCategory = (category: string) => {
  if (category === 'admin') return can('nav.admin')
  if (category === 'files') return can('nav.explorer')
  if (category === 'account') return can('nav.settings')
  return true
}

const canViewRoute = (r: any) => {
  if (r.path === '/explorer') return can('nav.explorer')
  if (r.path === '/settings') return can('nav.settings')
  if (r.path === '/admin/roles') return can('nav.admin')
  if (r.path === '/admin/users') return can('nav.admin')
  if (r.path === '/admin/routes') return can('nav.route_management')
  return true
}

const handleClick = (r: any) => {
  if (!r.enabled) return
  sidebarOpen.value = false
}

const loadRoutes = async () => {
  try {
    featureRoutes.value = (await apiFetch('/api/routes')) as any[]
  } catch {}
}

onMounted(() => {
  if (import.meta.client && localStorage.getItem('token')) {
    fetchPermissions()
  }
  loadRoutes()
})
</script>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  background-color: var(--color-surface-0);
  border-right: 1px solid var(--color-surface-2);
}

@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    transition: transform 0.25s ease;
    z-index: 101;
  }

  .sidebar.sidebar-open {
    transform: translateX(0);
    box-shadow: 4px 0 16px rgba(0, 0, 0, 0.15);
  }
}

.sidebar-backdrop {
  display: none;
}

@media (max-width: 768px) {
  .sidebar-backdrop {
    display: block;
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.4);
    z-index: 100;
  }
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 1.25rem 1.25rem 1rem 1.25rem;
}

.sidebar-close {
  display: none;
  margin-left: auto;
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
}

@media (max-width: 768px) {
  .sidebar-close {
    display: flex;
  }
}

.sidebar-logo {
  width: 2rem;
  height: 2rem;
  background-color: var(--color-brand-600);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sidebar-brand {
  font-weight: 700;
  font-size: 1rem;
  color: var(--color-text-primary);
  letter-spacing: -0.015em;
}

.sidebar-nav {
  flex: 1;
  padding: 0.5rem 0.75rem;
  display: flex;
  flex-direction: column;
}

.nav-section {
  margin-bottom: 0.75rem;
}

.nav-label {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.5rem 0.75rem 0.25rem 0.75rem;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: background-color 0.12s ease, color 0.12s ease;
}

.sidebar-link:hover {
  background-color: var(--color-surface-1);
  color: var(--color-text-primary);
}

.sidebar-link.active {
  background-color: var(--color-brand-50);
  color: var(--color-brand-700);
}

.sidebar-link.disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.sidebar-link.disabled:hover {
  background-color: transparent;
  color: var(--color-text-secondary);
}

.maintenance-icon {
  color: var(--color-warning, #f59e0b);
  margin-left: auto;
}
</style>
