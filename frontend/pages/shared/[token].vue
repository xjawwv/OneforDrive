<template>
  <div class="shared-page">
    <div v-if="loading" class="shared-loading">
      <Loader2 :size="32" class="spin" style="color: var(--color-brand-500);" />
      <p>Loading shared file...</p>
    </div>

    <div v-else-if="error" class="shared-error">
      <AlertTriangle :size="48" style="color: var(--color-danger); margin-bottom: 1rem;" />
      <h2>{{ error }}</h2>
      <p>The link may have expired or been revoked.</p>
      <button class="btn-primary" @click="navigateTo('/login')" style="margin-top: 1.5rem;">
        Go to RouteStorage
      </button>
    </div>

    <div v-else-if="fileInfo" class="shared-content">
      <div class="shared-header">
        <div class="shared-logo">
          <HardDrive :size="20" color="white" />
        </div>
        <span class="shared-brand">RouteStorage</span>
      </div>

      <div class="shared-card" :class="{ 'shared-card-wide': fileInfo.is_folder }">
        <div class="shared-preview">
          <template v-if="isImage">
            <img :src="thumbnailUrl" class="shared-image" />
          </template>
          <template v-else-if="fileInfo.is_folder">
            <div class="shared-icon-large folder-icon">
              <Folder :size="48" />
            </div>
          </template>
          <template v-else>
            <div class="shared-icon-large file-icon">
              <File :size="48" />
            </div>
          </template>
        </div>

        <div class="shared-info">
          <h1 class="shared-name">{{ fileInfo.name }}</h1>
          <div class="shared-meta">
            <span v-if="!fileInfo.is_folder">{{ formatSize(fileInfo.size) }}</span>
            <span v-if="!fileInfo.is_folder && fileInfo.size"> · </span>
            <span>{{ fileInfo.is_folder ? `${children.length} item${children.length !== 1 ? 's' : ''}` : fileInfo.mime_type }}</span>
            <span v-if="expiresAt"> · Expires {{ formatDate(expiresAt) }}</span>
          </div>
        </div>

        <div v-if="fileInfo.is_folder && children.length" class="shared-folder-list">
          <div v-for="child in children" :key="child.id" class="shared-folder-item">
            <div class="shared-folder-icon" :class="child.is_folder ? 'folder-icon' : 'file-icon'">
              <Folder v-if="child.is_folder" :size="18" />
              <File v-else :size="18" />
            </div>
            <span class="shared-folder-name">{{ child.name }}</span>
            <span class="shared-folder-size">{{ child.is_folder ? '--' : formatSize(child.size) }}</span>
          </div>
        </div>

        <div class="shared-actions">
          <button v-if="!fileInfo.is_folder" class="btn-primary shared-download-btn" @click="downloadFile">
            <Download :size="16" />
            <span>Download</span>
          </button>
          <button v-else class="btn-primary shared-download-btn" @click="navigateTo('/login')">
            <FolderOpen :size="16" />
            <span>Open in RouteStorage</span>
          </button>
        </div>
      </div>

      <div class="shared-footer">
        <p>Shared via <strong>RouteStorage</strong></p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Loader2, AlertTriangle, HardDrive, Download, FolderOpen, Folder, File } from 'lucide-vue-next'

definePageMeta({ layout: false })

const route = useRoute()
const router = useRouter()

const loading = ref(true)
const error = ref('')
const fileInfo = ref<any>(null)
const children = ref<any[]>([])
const expiresAt = ref('')
const token = ref('')

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const isImage = computed(() => {
  if (!fileInfo.value || fileInfo.value.is_folder || !fileInfo.value.name) return false
  const ext = fileInfo.value.name.split('.').pop()?.toLowerCase()
  return imageExtensions.includes(ext || '')
})

const thumbnailUrl = computed(() => {
  return `${useRuntimeConfig().public.apiBase}/shared/${token.value}/thumbnail`
})

const formatSize = (bytes: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric', hour: '2-digit', minute: '2-digit' })
}

const downloadFile = () => {
  window.location.href = `${useRuntimeConfig().public.apiBase}/shared/${token.value}/download`
}

onMounted(async () => {
  token.value = route.params.token as string
  if (!token.value) {
    error.value = 'Invalid link'
    loading.value = false
    return
  }

  try {
    const resp = await $fetch<any>(`${useRuntimeConfig().public.apiBase}/shared/${token.value}`)
    fileInfo.value = resp.file
    children.value = resp.children || []
    expiresAt.value = resp.expires_at
  } catch (e: any) {
    error.value = e.data?.error || 'Failed to load shared file'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.shared-page {
  min-height: 100vh;
  background-color: var(--color-surface-1);
  display: flex;
  flex-direction: column;
}

.shared-loading {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 1rem;
}

.shared-loading p {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin: 0;
}

.shared-error {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 2rem;
}

.shared-error h2 {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.375rem 0;
}

.shared-error p {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin: 0;
}

.shared-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 1rem 2rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.shared-logo {
  width: 2rem;
  height: 2rem;
  background-color: var(--color-brand-600);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.shared-brand {
  font-size: 1rem;
  font-weight: 700;
  color: var(--color-text-primary);
}

.shared-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 3rem 1.5rem;
}

.shared-card {
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 1rem;
  padding: 2rem;
  width: 100%;
  max-width: 480px;
  text-align: center;
}

.shared-card-wide {
  max-width: 640px;
  text-align: left;
}

.shared-preview {
  margin-bottom: 1.5rem;
  display: flex;
  justify-content: center;
}

.shared-image {
  max-width: 100%;
  max-height: 300px;
  object-fit: contain;
  border-radius: 0.5rem;
}

.shared-icon-large {
  width: 6rem;
  height: 6rem;
  border-radius: 1rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.folder-icon {
  background-color: rgba(76, 110, 245, 0.1);
  color: var(--color-brand-600);
}

.file-icon {
  background-color: var(--color-surface-2);
  color: var(--color-text-muted);
}

.shared-info {
  margin-bottom: 1.5rem;
}

.shared-name {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
  word-break: break-word;
}

.shared-meta {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.shared-download-btn {
  width: 100%;
  padding: 0.75rem 1.5rem;
}

.shared-folder-list {
  margin: 1.25rem 0;
  border: 1px solid var(--color-surface-2);
  border-radius: 0.5rem;
  overflow: hidden;
}

.shared-folder-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.shared-folder-item:last-child {
  border-bottom: none;
}

.shared-folder-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.shared-folder-name {
  flex: 1;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.shared-folder-size {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.shared-footer {
  margin-top: 3rem;
  text-align: center;
}

.shared-footer p {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  margin: 0;
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
  justify-content: center;
  gap: 0.5rem;
  transition: opacity 0.12s ease;
}

.btn-primary:hover {
  opacity: 0.9;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
