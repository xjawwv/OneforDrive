<template>
  <div>
      <div v-if="loading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else class="routes-grid">
        <div v-for="route in routes" :key="route.id">
          <div class="route-card" :class="{ disabled: !route.enabled, locked: isLocked(route) }">
            <div class="route-icon">
              <component :is="getIcon(route.icon)" :size="20" />
            </div>
            <div class="route-info">
              <div class="route-name">{{ route.name }}</div>
              <div class="route-desc">{{ route.description }}</div>
              <div class="route-meta">
                <span class="route-path">{{ route.path }}</span>
                <span class="route-category">{{ route.category }}</span>
                <span v-if="isLocked(route)" class="route-locked-badge">
                  <Lock :size="10" /> Always Active
                </span>
              </div>
            </div>
            <div class="route-actions">
              <Lock v-if="isLocked(route)" :size="16" class="lock-icon" />
              <button v-else class="toggle" :class="{ active: route.enabled }" @click="toggleRoute(route)">
                <span class="toggle-knob"></span>
              </button>
            </div>
          </div>

          <div v-if="!isLocked(route) && !route.enabled" class="exempt-panel">
            <div class="exempt-header">
              <span class="exempt-label">Bypass maintenance for:</span>
            </div>
            <div class="exempt-roles">
              <label v-for="role in allRoles" :key="role.id" class="exempt-role">
                <input
                  type="checkbox"
                  :checked="isExempt(route, role.id)"
                  @change="toggleExempt(route, role.id)"
                />
                <span class="role-check-label">{{ role.name }}</span>
              </label>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2, FolderOpen, Settings, ShieldCheck, Users, Circle, Lock } from 'lucide-vue-next'

const { apiFetch } = useApi()
const { can, fetchPermissions } = usePermissions()
const routes = ref<any[]>([])
const allRoles = ref<any[]>([])
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

const isLocked = (route: any) => route.path === '/admin/routes'

const isExempt = (route: any, roleId: number) => {
  return route.exempt_role_ids?.includes(roleId) || false
}

const loadData = async () => {
  loading.value = true
  try {
    const [routesData, rolesData] = await Promise.all([
      apiFetch('/api/routes') as any[],
      apiFetch('/api/rbac/roles') as any[]
    ])
    routes.value = routesData
    allRoles.value = rolesData
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

const toggleExempt = async (route: any, roleId: number) => {
  const current = route.exempt_role_ids || []
  const updated = current.includes(roleId)
    ? current.filter((id: number) => id !== roleId)
    : [...current, roleId]

  try {
    await apiFetch(`/api/routes/${route.id}/exempt-roles`, {
      method: 'PUT',
      body: { role_ids: updated }
    })
    route.exempt_role_ids = updated
  } catch {}
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  await fetchPermissions()
  if (!can('nav.route_management')) { navigateTo('/'); return }
  loadData()
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

.route-card.disabled {
  opacity: 0.7;
}

.route-card.locked {
  opacity: 0.6;
  cursor: default;
}

.route-card:hover {
  background-color: var(--color-surface-1);
}

.route-card.locked:hover {
  background-color: var(--color-surface-0);
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
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.lock-icon {
  color: var(--color-text-muted);
}

.route-locked-badge {
  display: inline-flex;
  align-items: center;
  gap: 0.25rem;
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-success);
  background-color: rgba(64, 192, 87, 0.1);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

.exempt-panel {
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-top: none;
  border-radius: 0 0 0.5rem 0.5rem;
  padding: 0.75rem 1rem;
  margin-top: -0.5rem;
  padding-top: 0;
}

.exempt-header {
  padding: 0.5rem 0;
}

.exempt-label {
  font-size: 0.75rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.exempt-roles {
  display: flex;
  flex-wrap: wrap;
  gap: 0.5rem;
}

.exempt-role {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  cursor: pointer;
  padding: 0.25rem 0.5rem;
  border-radius: 0.375rem;
  transition: background-color 0.12s ease;
}

.exempt-role:hover {
  background-color: var(--color-surface-1);
}

.exempt-role input[type="checkbox"] {
  width: 14px;
  height: 14px;
  accent-color: var(--color-brand-600);
  cursor: pointer;
}

.role-check-label {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
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
