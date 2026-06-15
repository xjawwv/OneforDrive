<template>
  <div class="app-layout">
    <AppSidebar current="settings" />
    <div class="app-main">
      <header class="page-header">
        <div>
          <h1 class="page-title">Drive Accounts</h1>
          <p class="page-subtitle">Manage your connected Google Drive accounts</p>
        </div>
        <button class="btn-primary" @click="connectAccount">
          <Plus :size="16" />
          <span>Connect Drive</span>
        </button>
      </header>

      <div class="card" style="margin-bottom: 1.5rem;">
        <h2 class="section-title">Storage Overview</h2>
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">{{ stats.total_users }}</div>
            <div class="stat-label">Users</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ stats.total_files }}</div>
            <div class="stat-label">Files</div>
          </div>
          <div class="stat-item">
            <div class="stat-value">{{ formatSize(stats.total_used_bytes) }}</div>
            <div class="stat-label">Total Size</div>
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
import { Plus, CloudOff, User, Trash2 } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()

const accounts = ref<any[]>([])
const stats = ref({ total_users: 0, total_files: 0, total_size_bytes: 0, total_drive_accounts: 0, active_drive_accounts: 0, total_capacity_bytes: 0, total_used_bytes: 0 })

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

onMounted(() => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
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

.page-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 1.75rem;
}

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

@media (max-width: 640px) {
  .stats-grid { grid-template-columns: repeat(2, 1fr); }
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
