<template>
  <div class="app-layout">
    <AppSidebar current="users" />
    <div class="app-main">
      <AppTopBar title="User Management" subtitle="View and manage user roles" current-page="settings" @hamburger-click="sidebarOpen = true" />

      <div v-if="loading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else class="users-list">
        <div v-for="user in users" :key="user.id" class="user-card">
          <div class="user-avatar">
            {{ user.name?.charAt(0)?.toUpperCase() || 'U' }}
          </div>
          <div class="user-info">
            <div class="user-name">{{ user.name }}</div>
            <div class="user-email">{{ user.email }}</div>
            <div class="user-roles">
              <span v-for="role in user.roles" :key="role.id" class="role-tag">{{ role.name }}</span>
              <span v-if="!user.roles?.length" class="role-tag none">No role</span>
            </div>
          </div>
          <div class="user-actions">
            <select class="role-select" :value="user.roles?.[0]?.id || ''" @change="assignRole(user.id, $event)">
              <option value="">No role</option>
              <option v-for="role in allRoles" :key="role.id" :value="role.id">{{ role.name }}</option>
            </select>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()
const sidebarOpen = ref(false)

provide('sidebarOpen', sidebarOpen)

const users = ref<any[]>([])
const allRoles = ref<any[]>([])
const loading = ref(true)

const loadUsers = async () => {
  loading.value = true
  try {
    const data = await apiFetch('/api/accounts') as any[]
    for (const user of data) {
      try {
        const roles = await apiFetch(`/api/rbac/users/${user.id}/roles`) as any[]
        user.roles = roles
      } catch {
        user.roles = []
      }
    }
    users.value = data
    allRoles.value = (await apiFetch('/api/rbac/roles')) as any[]
  } catch {}
  loading.value = false
}

const assignRole = async (userId: number, roleId: string) => {
  try {
    if (roleId) {
      await apiFetch(`/api/rbac/users/${userId}/roles`, {
        method: 'POST',
        body: { role_id: Number(roleId) }
      })
    } else {
      const user = users.value.find((u: any) => u.id === userId)
      if (user?.roles?.[0]?.id) {
        await apiFetch(`/api/rbac/users/${userId}/roles/${user.roles[0].id}`, { method: 'DELETE' })
      }
    }
    await loadUsers()
  } catch {}
}

onMounted(() => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  loadUsers()
})
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
  padding: 2rem 2.5rem;
}

@media (max-width: 768px) {
  .app-main {
    margin-left: 0;
    padding: 0.75rem;
  }
}

.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.users-list {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 0.5rem;
}

.user-avatar {
  width: 2.25rem;
  height: 2.25rem;
  border-radius: 9999px;
  background-color: var(--color-brand-600);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.875rem;
  font-weight: 600;
  flex-shrink: 0;
}

.user-info {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.user-email {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.user-roles {
  display: flex;
  gap: 0.25rem;
  margin-top: 0.25rem;
  flex-wrap: wrap;
}

.role-tag {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-brand-600);
  background-color: var(--color-brand-50);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
}

.role-tag.none {
  color: var(--color-text-muted);
  background-color: var(--color-surface-2);
}

.user-actions {
  flex-shrink: 0;
}

.role-select {
  padding: 0.375rem 0.5rem;
  border: 1px solid var(--color-surface-3);
  border-radius: 0.375rem;
  font-size: 0.75rem;
  color: var(--color-text-primary);
  background-color: var(--color-surface-0);
  cursor: pointer;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
