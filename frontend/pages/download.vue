<template>
  <div class="download-page">
    <div class="download-card">
      <div v-if="status === 'initializing'" class="download-state">
        <Loader2 :size="32" class="spin" style="color: var(--color-brand-500);" />
        <h2>Preparing download...</h2>
        <p>Setting up chunk retrieval from Drive</p>
      </div>

      <div v-else-if="status === 'downloading'" class="download-state">
        <Download :size="32" style="color: var(--color-brand-500);" />
        <h2>Downloading from Drive</h2>
        <p>{{ fileName }} ({{ formatSize(fileSize) }})</p>
        <div class="download-progress">
          <div class="download-progress-track">
            <div class="download-progress-fill" :style="{ width: progress + '%' }"></div>
          </div>
          <div class="download-progress-info">
            <span>{{ progress }}%</span>
            <span>{{ chunksDone }} / {{ chunksTotal }} chunks</span>
          </div>
        </div>
      </div>

      <div v-else-if="status === 'ready'" class="download-state">
        <CheckCircle :size="32" style="color: var(--color-success);" />
        <h2>Download ready</h2>
        <p>Your file is being saved</p>
      </div>

      <div v-else-if="status === 'error'" class="download-state">
        <AlertTriangle :size="32" style="color: var(--color-danger);" />
        <h2>Download failed</h2>
        <p>{{ errorMsg }}</p>
      </div>

      <button class="btn-secondary" @click="goBack" style="margin-top: 1.5rem;">
        <ArrowLeft :size="16" />
        <span>Back to files</span>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2, Download, CheckCircle, AlertTriangle, ArrowLeft } from 'lucide-vue-next'

definePageMeta({ layout: false })

const route = useRoute()
const router = useRouter()
const { apiFetch } = useApi()

const status = ref('initializing')
const progress = ref(0)
const fileName = ref('')
const fileSize = ref(0)
const chunksDone = ref(0)
const chunksTotal = ref(0)
const errorMsg = ref('')
const sessionId = ref('')

const formatSize = (bytes: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

const goBack = () => {
  router.back()
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
  }

  const fileId = route.params.id as string
  try {
    const resp = await apiFetch(`/api/files/${fileId}/download`, { method: 'POST' }) as any
    sessionId.value = resp.session_id
    fileName.value = resp.file_name
    fileSize.value = resp.file_size
    chunksTotal.value = resp.chunks
    status.value = 'downloading'
    pollProgress()
  } catch (e: any) {
    status.value = 'error'
    errorMsg.value = e?.data?.error || 'Failed to start download'
  }
})

const pollProgress = async () => {
  const token = localStorage.getItem('token')
  const apiBase = useRuntimeConfig().public.apiBase

  const poll = async () => {
    try {
      const resp = await fetch(`${apiBase}/api/files/${route.params.id}/download-progress?session=${sessionId.value}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!resp.ok) return
      const data = await resp.json()

      progress.value = data.progress || 0
      chunksDone.value = data.chunks_done || 0
      chunksTotal.value = data.chunks_total || 1

      if (data.status === 'ready') {
        status.value = 'ready'
        const fileResp = await fetch(`${apiBase}/api/files/${route.params.id}/download`, {
          headers: { Authorization: `Bearer ${token}` }
        })
        if (fileResp.ok) {
          const blob = await fileResp.blob()
          const disposition = fileResp.headers.get('Content-Disposition') || ''
          const nameMatch = disposition.match(/filename="?([^"]+)"?/)
          const filename = nameMatch ? nameMatch[1] : fileName.value || 'download'
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = filename
          a.click()
          URL.revokeObjectURL(url)
        }
        return
      }

      if (data.status === 'error') {
        status.value = 'error'
        errorMsg.value = data.error || 'Download failed'
        return
      }

      setTimeout(poll, 500)
    } catch {
      setTimeout(poll, 1000)
    }
  }
  poll()
}
</script>

<style scoped>
.download-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--color-surface-1);
}

.download-card {
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.75rem;
  padding: 2.5rem;
  width: 420px;
  max-width: 90vw;
  text-align: center;
}

.download-state h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 1rem 0 0.375rem 0;
}

.download-state p {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin: 0;
}

.download-progress {
  margin-top: 1.5rem;
}

.download-progress-track {
  width: 100%;
  height: 6px;
  background-color: var(--color-surface-2);
  border-radius: 9999px;
  overflow: hidden;
}

.download-progress-fill {
  height: 100%;
  background-color: var(--color-brand-500);
  border-radius: 9999px;
  transition: width 0.3s ease;
}

.download-progress-info {
  display: flex;
  justify-content: space-between;
  margin-top: 0.5rem;
  font-size: 0.75rem;
  color: var(--color-text-muted);
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
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: background-color 0.12s ease;
}

.btn-secondary:hover {
  background-color: var(--color-surface-1);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
