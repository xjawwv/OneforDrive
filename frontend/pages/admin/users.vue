<template>
  <div>
      <AppTopBar title="User Management" subtitle="View and manage user roles" current-page="users" />

      <div v-if="loading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else class="table-card">
        <table class="data-table">
          <thead>
            <tr>
              <th class="col-id">ID</th>
              <th class="col-name">Name</th>
              <th class="col-email">Email</th>
              <th class="col-role">Role</th>
              <th class="col-actions">Actions</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="user in users" :key="user.id">
              <td class="col-id">{{ user.id }}</td>
              <td class="col-name">
                <div class="user-cell">
                  <div class="user-avatar-sm">{{ user.name?.charAt(0)?.toUpperCase() || 'U' }}</div>
                  <span>{{ user.name }}</span>
                </div>
              </td>
              <td class="col-email">{{ user.email }}</td>
              <td class="col-role">
                <span v-if="user.roles?.length" class="role-tag">{{ user.roles[0].name }}</span>
                <span v-else class="role-tag none">No role</span>
              </td>
              <td class="col-actions">
                <select class="role-select" :value="user.roles?.[0]?.id || ''" @change="assignRole(user.id, Number($event.target.value))">
                  <option value="">No role</option>
                  <option v-for="role in allRoles" :key="role.id" :value="role.id">{{ role.name }}</option>
                </select>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2 } from 'lucide-vue-next'

const { apiFetch } = useApi()
const { can, fetchPermissions } = usePermissions()
const users = ref<any[]>([])
const allRoles = ref<any[]>([])
const loading = ref(true)

const loadUsers = async () => {
  loading.value = true
  try {
    const data = await apiFetch('/api/rbac/users') as any[]
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

const assignRole = async (userId: number, roleId: number) => {
  try {
    const user = users.value.find((u: any) => u.id === userId)
    if (user?.roles?.[0]?.id) {
      await apiFetch(`/api/rbac/users/${userId}/roles/${user.roles[0].id}`, { method: 'DELETE' })
    }
    await apiFetch(`/api/rbac/users/${userId}/roles`, {
      method: 'POST',
      body: { role_id: roleId }
    })
    await loadUsers()
  } catch {}
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  await fetchPermissions()
  if (!can('nav.admin')) { navigateTo('/'); return }
  loadUsers()
})
</script>

<style scoped>
.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.table-card {
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 0.75rem;
  overflow-x: auto;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.8125rem;
}

.data-table thead {
  border-bottom: 1px solid var(--color-surface-2);
}

.data-table th {
  text-align: left;
  padding: 0.75rem 1rem;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.data-table td {
  padding: 0.75rem 1rem;
  color: var(--color-text-secondary);
  border-bottom: 1px solid var(--color-surface-2);
  vertical-align: middle;
}

.data-table tr:last-child td {
  border-bottom: none;
}

.data-table tr:hover td {
  background-color: var(--color-surface-1);
}

.col-id {
  width: 60px;
}

.col-name {
  min-width: 150px;
}

.col-email {
  min-width: 200px;
}

.col-role {
  min-width: 120px;
}

.col-actions {
  width: 150px;
}

.user-cell {
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-avatar-sm {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 9999px;
  background-color: #F43F5E;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.6875rem;
  font-weight: 600;
  flex-shrink: 0;
}

.role-tag {
  display: inline-block;
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-brand-600);
  background-color: var(--color-brand-50);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  margin-right: 0.25rem;
}

.role-tag.none {
  color: var(--color-text-muted);
  background-color: var(--color-surface-2);
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

@media (max-width: 768px) {
  .col-id, .col-email {
    display: none;
  }
}
</style>
