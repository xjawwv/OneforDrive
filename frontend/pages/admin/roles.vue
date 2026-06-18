<template>
  <div>
      <AppTopBar title="Role Management" subtitle="Manage roles and permissions" current-page="settings">
        <template #actions>
          <button class="btn-primary" @click="showCreateModal = true">
            <Plus :size="16" />
            <span>New Role</span>
          </button>
        </template>
      </AppTopBar>
      <div class="header-divider"></div>

      <div v-if="loading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else class="roles-grid">
        <div v-for="role in roles" :key="role.id" class="role-card" :class="{ locked: role.is_system }" @click="selectRole(role)">
          <div class="role-icon">
            <ShieldCheck :size="20" />
          </div>
          <div class="role-info">
            <div class="role-name">{{ role.name }}</div>
            <div class="role-desc">{{ role.description }}</div>
            <div v-if="role.is_system" class="role-badge">System</div>
          </div>
          <Lock v-if="role.is_system" :size="14" style="color: var(--color-text-muted);" />
          <ChevronRight v-else :size="16" style="color: var(--color-text-muted);" />
        </div>
      </div>

      <Transition name="modal">
        <div v-if="selectedRole" class="modal-overlay" @click.self="selectedRole = null">
          <div class="modal-card">
            <div class="modal-header">
              <h3>Permissions for "{{ selectedRole.name }}"</h3>
              <button class="icon-btn" @click="selectedRole = null"><X :size="16" /></button>
            </div>
            <div class="modal-body">
              <div v-for="(perms, category) in groupedPermissions" :key="category" class="perm-group">
                <div class="perm-category">{{ category }}</div>
                <div v-for="perm in perms" :key="perm.id" class="perm-item">
                  <div class="perm-row">
                    <span class="perm-text">{{ perm.description }}</span>
                    <button class="toggle" :class="{ active: selectedPermIds.has(perm.id) }" @click="togglePerm(perm.id)">
                      <span class="toggle-knob"></span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn-secondary" @click="selectedRole = null">Cancel</button>
              <button class="btn-primary" @click="savePermissions">Save</button>
            </div>
          </div>
        </div>
      </Transition>

      <Transition name="modal">
        <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
          <div class="modal-card">
            <div class="modal-header">
              <h3>Create Role</h3>
              <button class="icon-btn" @click="showCreateModal = false"><X :size="16" /></button>
            </div>
            <div class="modal-body">
              <div class="form-group">
                <label>Name</label>
                <input v-model="newRoleName" class="input-field" placeholder="e.g. editor" />
              </div>
              <div class="form-group">
                <label>Description</label>
                <input v-model="newRoleDesc" class="input-field" placeholder="What can this role do?" />
              </div>
            </div>
            <div class="modal-footer">
              <button class="btn-secondary" @click="showCreateModal = false">Cancel</button>
              <button class="btn-primary" @click="createRole" :disabled="!newRoleName.trim()">Create</button>
            </div>
          </div>
        </div>
      </Transition>
  </div>
</template>

<script setup lang="ts">
import { Plus, ShieldCheck, ChevronRight, X, Loader2, Lock } from 'lucide-vue-next'

const { apiFetch } = useApi()
const { can, fetchPermissions } = usePermissions()
const route = useRoute()
const router = useRouter()
const roles = ref<any[]>([])
const permissions = ref<any[]>([])
const loading = ref(true)
const selectedRole = ref<any>(null)
const selectedPermIds = ref(new Set<number>())
const showCreateModal = ref(false)
const newRoleName = ref('')
const newRoleDesc = ref('')

const groupedPermissions = computed(() => {
  const groups: Record<string, any[]> = {}
  permissions.value.forEach((p: any) => {
    if (!groups[p.category]) groups[p.category] = []
    groups[p.category].push(p)
  })
  return groups
})

const loadData = async () => {
  loading.value = true
  try {
    roles.value = (await apiFetch('/api/rbac/roles')) as any[]
    permissions.value = (await apiFetch('/api/rbac/permissions')) as any[]
  } catch {}
  loading.value = false
}

const selectRole = async (role: any) => {
  if (role.is_system) return
  selectedRole.value = role
  try {
    const data = await apiFetch(`/api/rbac/roles/${role.id}/permissions`) as any
    selectedPermIds.value = new Set(data.permission_ids || [])
  } catch {}
}

const togglePerm = (permId: number) => {
  if (selectedPermIds.value.has(permId)) {
    selectedPermIds.value.delete(permId)
  } else {
    selectedPermIds.value.add(permId)
  }
}

const savePermissions = async () => {
  if (!selectedRole.value) return
  try {
    await apiFetch(`/api/rbac/roles/${selectedRole.value.id}/permissions`, {
      method: 'PUT',
      body: { permission_ids: Array.from(selectedPermIds.value) }
    })
    selectedRole.value = null
  } catch {}
}

const createRole = async () => {
  if (!newRoleName.value.trim()) return
  try {
    await apiFetch('/api/rbac/roles', {
      method: 'POST',
      body: { name: newRoleName.value.trim(), description: newRoleDesc.value.trim() }
    })
    showCreateModal.value = false
    newRoleName.value = ''
    newRoleDesc.value = ''
    await loadData()
  } catch {}
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  await fetchPermissions()
  if (!can('nav.admin')) { navigateTo('/'); return }
  loadData()
})
</script>

<style scoped>
.header-divider {
  height: 1px;
  background-color: #E4E4E7;
  margin: 0.75rem 0;
}

.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.roles-grid {
  display: flex;
  flex-direction: column;
  gap: 0.5rem;
}

.role-card {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 0.5rem;
  cursor: pointer;
  transition: background-color 0.12s ease;
}

.role-card:hover {
  background-color: var(--color-surface-1);
}

.role-card.locked {
  cursor: default;
  opacity: 0.7;
}

.role-card.locked:hover {
  background-color: var(--color-surface-0);
}

.role-icon {
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

.role-info {
  flex: 1;
  min-width: 0;
}

.role-name {
  font-size: 0.875rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.role-desc {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.role-badge {
  display: inline-block;
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-brand-600);
  background-color: var(--color-brand-50);
  padding: 0.125rem 0.375rem;
  border-radius: 0.25rem;
  margin-top: 0.25rem;
}

.modal-overlay {
  position: fixed;
  inset: 0;
  background-color: rgba(0, 0, 0, 0.4);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 200;
}

.modal-card {
  background-color: var(--color-surface-0);
  border-radius: 0.75rem;
  width: 500px;
  max-width: 90vw;
  max-height: 80vh;
  display: flex;
  flex-direction: column;
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.modal-header h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
}

.modal-body {
  padding: 1rem 1.25rem;
  overflow-y: auto;
  flex: 1;
}

.modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--color-surface-2);
}

.perm-group {
  margin-bottom: 1rem;
}

.perm-category {
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 0.375rem;
}

.perm-item {
  padding: 0.375rem 0;
}

.perm-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.75rem;
}

.perm-text {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.toggle {
  position: relative;
  width: 36px;
  height: 20px;
  background-color: var(--color-surface-3);
  border: none;
  border-radius: 9999px;
  cursor: pointer;
  flex-shrink: 0;
  transition: background-color 0.2s ease;
}

.toggle.active {
  background-color: var(--color-brand-600);
}

.toggle-knob {
  position: absolute;
  top: 2px;
  left: 2px;
  width: 16px;
  height: 16px;
  background-color: white;
  border-radius: 9999px;
  transition: transform 0.2s ease;
}

.toggle.active .toggle-knob {
  transform: translateX(16px);
}

.form-group {
  margin-bottom: 1rem;
}

.form-group label {
  display: block;
  font-size: 0.75rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  margin-bottom: 0.25rem;
}

.input-field {
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  padding: 0.5rem 0.75rem;
  font-size: 0.875rem;
  color: var(--color-text-primary);
  width: 100%;
  outline: none;
}

.input-field:focus {
  border-color: var(--color-brand-500);
}

.btn-primary {
  background-color: var(--color-brand-600);
  color: white;
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  font-weight: 500;
  font-size: 0.8125rem;
  border: none;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.5rem;
}

.btn-primary:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-secondary {
  background-color: var(--color-surface-0);
  color: var(--color-text-secondary);
  padding: 0.5rem 1rem;
  border-radius: 0.5rem;
  font-weight: 500;
  font-size: 0.8125rem;
  border: 1px solid var(--color-surface-3);
  cursor: pointer;
}

.icon-btn {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.icon-btn:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface-2);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.modal-enter-active { transition: opacity 0.15s ease; }
.modal-leave-active { transition: opacity 0.1s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
</style>
