<template>
  <div class="app-layout">
    <AppSidebar current="settings" />
    <div class="app-main">
      <AppTopBar title="Drive Accounts" subtitle="Manage your connected Google Drive accounts" current-page="settings" @hamburger-click="sidebarOpen = true" />
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

      <div v-if="!loading && accounts.length === 0" class="card" style="text-align: center; padding: 3rem 1.5rem;">
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
              <svg width="20" height="19" viewBox="0 0 87.3 78" fill="none" xmlns="http://www.w3.org/2000/svg">
                <mask id="a" width="168" height="154" x="12" y="18" maskUnits="userSpaceOnUse">
                  <path fill="#fff" d="M63.09 37c14.626-25.333 51.193-25.334 65.819 0l45.033 78c14.626 25.334-3.657 57.001-32.91 57.001H50.967c-29.253 0-47.536-31.667-32.91-57.001Z"/>
                </mask>
                <g mask="url(#a)" transform="matrix(4.8140532,0,0,4.8140532,-62.146701,-86.652356)">
                  <path fill="url(#b)" d="M206.905 172.02h-91.888l-19.015-32.934 45.944-79.578Z"/>
                  <path fill="url(#c)" d="M-14.919 172.006 50.04 59.494v.002L31.032 92.422h38.02L115 172.004l-129.918.001Z"/>
                  <path fill="url(#d)" d="M96.007-20.085 141.954 59.5l-19.011 32.928H31.048Z"/>
                </g>
                <defs>
                  <linearGradient id="b" x1="193.6" x2="103.09" y1="165.6" y2="111.21" gradientUnits="userSpaceOnUse">
                    <stop offset=".09" stop-color="#ffe921"/>
                    <stop offset="1" stop-color="#fec700"/>
                  </linearGradient>
                  <linearGradient id="c" x1="114.4" x2="15.53" y1="181.61" y2="121.8" gradientUnits="userSpaceOnUse">
                    <stop offset=".15" stop-color="#a9a8ff"/>
                    <stop offset=".33" stop-color="#6d97ff"/>
                    <stop offset=".48" stop-color="#3186ff"/>
                  </linearGradient>
                  <linearGradient id="d" x1="128.88" x2="28.7" y1="37.88" y2="84.64" gradientUnits="userSpaceOnUse">
                    <stop offset=".55" stop-color="#0ebc5f"/>
                    <stop offset=".85" stop-color="#78c9ff"/>
                  </linearGradient>
                </defs>
              </svg>
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
import { Plus, Trash2, RefreshCw, HardDrive } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()

const accounts = ref<any[]>([])
const loading = ref(true)
const stats = ref({ total_files: 0, total_size_bytes: 0, total_drive_accounts: 0, active_drive_accounts: 0, total_capacity_bytes: 0, total_used_bytes: 0 })
const sidebarOpen = ref(false)

provide('sidebarOpen', sidebarOpen)

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
  loading.value = true
  try { accounts.value = (await apiFetch('/api/accounts')) as any[] } catch {}
  loading.value = false
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
    padding: 0.75rem;
  }
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
}

@media (max-width: 480px) {
  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    gap: 0.5rem;
  }

  .stat-value {
    font-size: 1.125rem;
  }

  .stat-label {
    font-size: 0.625rem;
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

@media (max-width: 480px) {
  .account-card {
    flex-direction: column;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .account-card > div:last-child {
    width: 100%;
    justify-content: flex-end;
  }

  .progress-bar-wrapper {
    width: 100%;
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
