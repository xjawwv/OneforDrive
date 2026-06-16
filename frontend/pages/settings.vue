<template>
  <div class="app-layout">
    <AppSidebar current="settings" />
    <div class="app-main">
      <header class="page-header">
        <div>
          <h1 class="page-title">Drive Accounts</h1>
          <p class="page-subtitle">Manage your connected Google Drive accounts</p>
        </div>
        <div class="header-actions">
          <div class="user-menu-wrapper">
            <button class="user-avatar-btn" @click="showUserMenu = !showUserMenu">
              <div class="user-avatar-circle">{{ userInitial }}</div>
            </button>
            <Transition name="menu">
              <div v-if="showUserMenu" class="user-menu">
                <NuxtLink to="/explorer" class="user-menu-item" @click="showUserMenu = false">
                  <FolderOpen :size="14" />
                  <span>Explorer</span>
                </NuxtLink>
                <div class="user-menu-divider"></div>
                <button class="user-menu-item danger" @click="logout">
                  <LogOut :size="14" />
                  <span>Logout</span>
                </button>
              </div>
            </Transition>
          </div>
          <button class="btn-primary" @click="connectAccount">
            <Plus :size="16" />
            <span>Connect Drive</span>
          </button>
        </div>
      </header>

      <div class="card" style="margin-bottom: 1.5rem;">
        <h2 class="section-title">Storage Overview</h2>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ stats.total_files }}</div>
            <div class="stat-label">Files</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatSize(stats.total_used_bytes) }}</div>
            <div class="stat-label">Total Size</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatSize(stats.total_capacity_bytes) }}</div>
            <div class="stat-label">Capacity</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ stats.active_drive_accounts }}</div>
            <div class="stat-label">Drive Accounts</div>
          </div>
        </div>
      </div>

      <div v-if="accounts.length === 0" class="card" style="text-align: center; padding: 3rem 1.5rem;">
        <CloudOff :size="40" style="color: var(--color-text-muted); margin: 0 auto 1rem auto; display: block;" />
        <h3 style="font-size: 1rem; font-weight: 600; color: var(--color-text-primary); margin: 0 0 0.375rem 0;">No drive accounts connected</h3>
        <p style="font-size: 0.8125rem; color: var(--color-text-muted); margin: 0 0 1.5rem 0;">Connect a Google Drive account to start storing files.</p>
        <button class="btn-primary" @click="connectAccount">
          <Plus :size="16" />
          <span>Connect Google Drive</span>
        </button>
      </div>

      <div v-else style="display: flex; flex-direction: column; gap: 0.75rem;">
        <div v-for="account in accounts" :key="account.id" class="card account-card">
          <div class="account-info">
            <div class="account-avatar">
              <User :size="20" style="color: var(--color-brand-600);" />
            </div>
            <div>
              <div style="font-weight: 500; color: var(--color-text-primary); font-size: 0.875rem;">{{ account.email }}</div>
              <div style="font-size: 0.75rem; color: var(--color-text-muted); margin-top: 0.125rem;">{{ formatSize(account.capacity_used) }} / {{ formatSize(account.capacity_total) }}</div>
              <div class="progress-bar-wrapper">
                <div class="progress-bar-track">
                  <div class="progress-bar-fill" :style="{ width: drivePercent(account) + '%' }"></div>
                </div>
              </div>
            </div>
          </div>
          <div style="display: flex; align-items: center; gap: 0.5rem;">
            <span class="status-badge" :class="account.is_active ? 'status-active' : 'status-inactive'">
              {{ account.is_active ? 'Active' : 'Inactive' }}
            </span>
            <button class="sync-btn" @click="syncAccount(account.id)" :disabled="account._syncing" title="Sync Drive">
              <RefreshCw :size="14" :class="{ spin: account._syncing }" />
            </button>
            <button class="delete-btn" @click="removeAccount(account.id)" title="Remove account">
              <Trash2 :size="16" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Plus, CloudOff, User, Trash2, RefreshCw, LogOut, Menu } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()

const accounts = ref<any[]>([])
const stats = ref({ total_files: 0, total_size_bytes: 0, total_drive_accounts: 0, active_drive_accounts: 0, total_capacity_bytes: 0, total_used_bytes: 0 })
const showUserMenu = ref(false)

const userName = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).name } catch { return 'U' }
    }
  }
  return 'U'
})

const userInitial = computed(() => userName.value.charAt(0).toUpperCase())

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  navigateTo('/login')
}

const formatSize = (bytes: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

const drivePercent = (account: any) => {
  if (!account.capacity_total) return 0
  return Math.min(100, Math.round((account.capacity_used / account.capacity_total) * 100))
}

const loadAccounts = async () => {
  try { accounts.value = (await apiFetch('/api/accounts')) as any[] } catch {}
}

const loadStats = async () => {
  try { stats.value = (await apiFetch('/api/storage/stats')) as any } catch {}
}

const connectAccount = async () => {
  try {
    const data = await apiFetch('/api/accounts/connect') as any
    if (data.url) window.location.href = data.url
  } catch {}
}

const removeAccount = async (id: number) => {
  try { await apiFetch(`/api/accounts/${id}`, { method: 'DELETE' }); await loadAccounts(); await loadStats() } catch {}
}

const syncAccount = async (id: number) => {
  const account = accounts.value.find(a => a.id === id)
  if (!account) return
  account._syncing = true
  try {
    const resp = await apiFetch(`/api/accounts/${id}/sync`, { method: 'POST' }) as any
    if (resp.deleted > 0) {
      await loadAccounts()
      await loadStats()
    } else {
      account.capacity_total = resp.capacity_total
      account.capacity_used = resp.capacity_used
    }
  } catch {}
  account._syncing = false
}

onMounted(() => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
    document.addEventListener('click', (e) => {
      if (showUserMenu.value) {
        const target = e.target as HTMLElement
        if (!target.closest('.user-menu-wrapper')) {
          showUserMenu.value = false
        }
      }
    })
  }
  loadAccounts()
  loadStats()
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
    padding: 1rem;
    padding-top: 3.5rem;
  }
}

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1.75rem;
}

@media (max-width: 768px) {
  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .header-actions {
    width: 100%;
    justify-content: flex-end;
  }
}

.header-actions {
  display: flex;
  gap: 0.5rem;
  align-items: center;
}

.user-menu-wrapper {
  position: relative;
}

.user-avatar-btn {
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.user-avatar-circle {
  width: 2rem;
  height: 2rem;
  border-radius: 9999px;
  background-color: var(--color-brand-600);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 600;
}

.user-menu {
  position: absolute;
  top: calc(100% + 0.25rem);
  right: 0;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 160px;
  z-index: 50;
  padding: 0.25rem 0;
}

.user-menu-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  width: 100%;
  padding: 0.5rem 0.75rem;
  border: none;
  background: none;
  color: var(--color-text-secondary);
  font-size: 0.8125rem;
  cursor: pointer;
  text-align: left;
  text-decoration: none;
  transition: background-color 0.1s ease;
}

.user-menu-item:hover {
  background-color: var(--color-surface-1);
}

.user-menu-item.danger {
  color: var(--color-danger);
}

.user-menu-item.danger:hover {
  background-color: rgba(250, 82, 82, 0.08);
}

.user-menu-divider {
  height: 1px;
  background-color: var(--color-surface-2);
  margin: 0.25rem 0;
}

.menu-enter-active { transition: opacity 0.1s ease, transform 0.1s ease; }
.menu-leave-active { transition: opacity 0.08s ease, transform 0.08s ease; }
.menu-enter-from, .menu-leave-to { opacity: 0; transform: translateY(-4px); }

.page-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.025em;
}

.page-subtitle {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin-top: 0.25rem;
}

.section-title {
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin-bottom: 1.25rem;
}

.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 1rem;
}

@media (max-width: 768px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}

.stat-item {
  text-align: center;
  padding: 0.75rem 0.5rem;
  border-radius: 0.5rem;
  background-color: var(--color-surface-1);
}

.stat-value {
  font-size: 1.375rem;
  font-weight: 700;
  color: var(--color-brand-600);
}

.stat-label {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  font-weight: 500;
  margin-top: 0.25rem;
}

.account-card {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

@media (max-width: 768px) {
  .account-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }
}

.account-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.account-avatar {
  width: 2.5rem;
  height: 2.5rem;
  background-color: var(--color-brand-50);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.status-badge {
  padding: 0.25rem 0.625rem;
  border-radius: 9999px;
  font-size: 0.6875rem;
  font-weight: 600;
}

.status-active {
  background-color: rgba(64, 192, 87, 0.1);
  color: var(--color-success);
}

.status-inactive {
  background-color: var(--color-surface-2);
  color: var(--color-text-muted);
}

.delete-btn {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.375rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.12s ease, background-color 0.12s ease;
}

.delete-btn:hover {
  color: var(--color-danger);
  background-color: rgba(250, 82, 82, 0.08);
}

.sync-btn {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.375rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.12s ease, background-color 0.12s ease;
}

.sync-btn:hover {
  color: var(--color-brand-600);
  background-color: rgba(76, 110, 245, 0.08);
}

.sync-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.progress-bar-wrapper {
  margin-top: 0.5rem;
}

.progress-bar-track {
  width: 100%;
  height: 6px;
  background-color: var(--color-surface-2);
  border-radius: 9999px;
  overflow: hidden;
}

.progress-bar-fill {
  height: 100%;
  background-color: var(--color-brand-500);
  border-radius: 9999px;
  transition: width 0.4s ease;
}
</style>
