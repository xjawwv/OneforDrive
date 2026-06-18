<template>
  <div>
      <div v-if="routeLoading" class="empty-state">
        <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      </div>

      <div v-else-if="!routeEnabled" class="maintenance-state">
        <div class="maintenance-icon">
          <AlertTriangle :size="48" />
        </div>
        <h2>Feature Under Maintenance</h2>
        <p>{{ routeDesc || 'This feature is temporarily unavailable. Please check back later.' }}</p>
      </div>

      <div v-else class="settings-content">
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
              <img src="/assets/Google_Drive_Logo_128px.png" alt="Google Drive" style="width:20px;height:20px;" />
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
import { Plus, Trash2, RefreshCw, ShieldCheck, Users, Loader2, AlertTriangle } from 'lucide-vue-next'
import { h } from 'vue'

const topbar = useState('topbar')
topbar.value = { title: 'Drive Accounts', subtitle: 'Manage your connected Google Drive accounts', currentPage: 'settings' }
const topbarActionsFn = inject<Ref<(() => any) | null>>('topbar:actions', ref(null))

const { apiFetch } = useApi()
const { can, fetchPermissions } = usePermissions()
const { enabled: routeEnabled, loading: routeLoading, description: routeDesc, checkRoute: checkFeatureRoute } = useFeatureRoute('/settings')
topbarActionsFn.value = {
  setup() {
    return () => routeEnabled.value ? h('button', { class: 'btn-primary', onClick: connectAccount }, [h(Plus, { size: 16 }), h('span', null, 'Connect Drive')]) : null
  }
}

const accounts = ref<any[]>([])
const loading = ref(true)
const stats = ref({ total_files: 0, total_size_bytes: 0, total_drive_accounts: 0, active_drive_accounts: 0, total_capacity_bytes: 0, total_used_bytes: 0 })
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

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }
  await fetchPermissions()
  if (!can('nav.settings')) { navigateTo('/'); return }
  await checkFeatureRoute()
  if (!routeEnabled.value) return
  loadAccounts()
  loadStats()
})
</script>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 5rem 1.5rem;
}

.maintenance-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 5rem 1.5rem;
}

.maintenance-state .maintenance-icon {
  color: var(--color-text-muted);
  margin-bottom: 1rem;
}

.maintenance-state h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.maintenance-state p {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin: 0;
  max-width: 400px;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
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
