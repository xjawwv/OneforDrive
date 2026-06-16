<template>
  <div class="app-layout">
    <AppSidebar current="settings" />
    <div class="app-main">
      <header class="top-bar">
        <button class="hamburger-btn" @click="sidebarOpen = true">
          <Menu :size="20" />
        </button>
        <div class="top-bar-title">
          <h1 class="page-title">Drive Accounts</h1>
          <p class="page-subtitle">Manage your connected Google Drive accounts</p>
        </div>
        <div class="user-menu-wrapper">
          <button class="avatar-btn" @click="showUserMenu = !showUserMenu">
            <div class="avatar-circle">{{ userInitial }}</div>
            <div class="notification-dot"></div>
          </button>
          <Transition name="menu">
            <div v-if="showUserMenu" class="user-dropdown">
              <div class="dropdown-user-info">
                <div class="dropdown-user-name">{{ userName }}</div>
                <div class="dropdown-user-email">{{ userEmail }}</div>
              </div>
              <div class="dropdown-divider"></div>
              <NuxtLink to="/explorer" class="dropdown-item" @click="showUserMenu = false">
                <FolderOpen :size="14" />
                <span>Explorer</span>
              </NuxtLink>
              <div class="dropdown-divider"></div>
              <button class="dropdown-item danger" @click="logout">
                <LogOut :size="14" />
                <span>Log out</span>
              </button>
            </div>
          </Transition>
        </div>
      </header>
      <div class="header-divider"></div>
      <div class="action-toolbar">
        <button class="btn-primary" @click="connectAccount">
          <Plus :size="16" />
          <span>Connect Drive</span>
        </button>
      </div>

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
const sidebarOpen = ref(false)

provide('sidebarOpen', sidebarOpen)

const userName = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).name } catch { return 'U' }
    }
  }
  return 'U'
})

const userEmail = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).email } catch { return '' }
    }
  }
  return ''
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

.top-bar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1rem;
  background-color: var(--color-surface-0);
  border-radius: 0.75rem;
}

.hamburger-btn {
  width: 36px;
  height: 36px;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  cursor: pointer;
  display: none;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  flex-shrink: 0;
  transition: background-color 0.12s ease;
}

.hamburger-btn:hover {
  background-color: var(--color-surface-1);
}

@media (max-width: 768px) {
  .hamburger-btn {
    display: flex;
  }
}

.top-bar-title {
  flex: 1;
  min-width: 0;
}

.page-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.025em;
}

.page-subtitle {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.header-divider {
  height: 1px;
  background-color: #E4E4E7;
  margin: 0.75rem 0;
}

@media (max-width: 768px) {
  .header-divider {
    margin: 0.5rem 0;
  }
}

.avatar-btn {
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
}

.avatar-circle {
  width: 38px;
  height: 38px;
  border-radius: 9999px;
  background-color: #F43F5E;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.notification-dot {
  position: absolute;
  top: 0;
  right: 0;
  width: 9px;
  height: 9px;
  background-color: #EF4444;
  border: 2px solid var(--color-surface-0);
  border-radius: 9999px;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 200px;
  z-index: 50;
  padding: 0.25rem 0;
}

.dropdown-user-info {
  padding: 0.625rem 0.75rem;
}

.dropdown-user-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.dropdown-user-email {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--color-surface-2);
  margin: 0.25rem 0;
}

.dropdown-item {
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

.dropdown-item:hover {
  background-color: var(--color-surface-1);
}

.dropdown-item.danger {
  color: #F43F5E;
}

.dropdown-item.danger:hover {
  background-color: rgba(244, 63, 94, 0.08);
}

.action-toolbar {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  margin-bottom: 1.25rem;
}

@media (max-width: 768px) {
  .action-toolbar {
    flex-wrap: wrap;
  }

  .action-toolbar .btn-primary {
    flex: 1;
  }
}

.menu-enter-active { transition: opacity 0.1s ease, transform 0.1s ease; }
.menu-leave-active { transition: opacity 0.08s ease, transform 0.08s ease; }
.menu-enter-from, .menu-leave-to { opacity: 0; transform: translateY(-4px); }

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
