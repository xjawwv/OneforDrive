<template>
  <div class="shared-layout">
    <header class="page-header">
      <div class="breadcrumb">
        <Home :size="14" class="bc-icon" />
        <button class="breadcrumb-item active">
          <Folder v-if="fileInfo?.is_folder" :size="14" />
          <File v-else :size="14" />
          <span>{{ fileInfo?.name || 'Shared file' }}</span>
        </button>
      </div>
      <div class="header-meta">
        <span v-if="expiresAt" class="expiry-badge">
          <Clock :size="12" />
          Expires {{ formatDate(expiresAt) }}
        </span>
      </div>
    </header>

    <div v-if="loading" class="empty-state">
      <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
      <p style="margin-top: 0.75rem; font-size: 0.8125rem; color: var(--color-text-muted);">Loading shared file...</p>
    </div>

    <div v-else-if="error" class="empty-state">
      <AlertTriangle :size="48" style="color: var(--color-danger); margin-bottom: 1rem;" />
      <h3>{{ error }}</h3>
      <p>The link may have expired or been revoked.</p>
    </div>

    <div v-else-if="fileInfo && !fileInfo.is_folder" class="file-grid-large">
      <div class="file-card-large">
        <div class="file-card-icon-large" :class="isImage ? 'file-type-image' : 'file-type-file'">
          <template v-if="isImage">
            <img :src="thumbnailUrl" class="file-thumb" />
          </template>
          <template v-else>
            <File :size="48" />
          </template>
        </div>
        <span class="file-card-name">{{ fileInfo.name }}</span>
        <div class="file-card-meta">{{ formatSize(fileInfo.size) }} · {{ fileInfo.mime_type }}</div>
        <button class="btn-primary shared-download-btn" @click="downloadFile">
          <Download :size="16" />
          <span>Download</span>
        </button>
      </div>
    </div>

    <div v-else-if="fileInfo && fileInfo.is_folder" class="file-grid-large">
      <div v-if="children.length === 0" class="empty-state">
        <FolderOpen :size="48" style="color: var(--color-surface-3); margin-bottom: 1rem;" />
        <h3>This folder is empty</h3>
      </div>
      <div v-for="child in children" :key="child.id" class="file-card-large">
        <div class="file-card-icon-large" :class="child.is_folder ? 'file-icon-folder' : `file-type-${getChildType(child)}`">
          <template v-if="child.is_folder">
            <Folder :size="48" />
          </template>
          <template v-else-if="isChildImage(child)">
            <img :src="childThumbnailUrl(child.id)" class="file-thumb" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
          </template>
          <template v-else>
            <Film v-if="getChildType(child) === 'video'" :size="48" />
            <Music v-else-if="getChildType(child) === 'audio'" :size="48" />
            <FileText v-else-if="getChildType(child) === 'doc'" :size="48" />
            <File v-else :size="48" />
          </template>
        </div>
        <span class="file-card-name">{{ child.name }}</span>
        <div class="file-card-meta">{{ child.is_folder ? 'Folder' : formatSize(child.size) }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { Folder, File, FolderOpen, Download, Loader2, AlertTriangle, Clock, Film, Music, FileText, Home } from 'lucide-vue-next'

definePageMeta({ layout: false })

const route = useRoute()

const loading = ref(true)
const error = ref('')
const fileInfo = ref<any>(null)
const children = ref<any[]>([])
const expiresAt = ref('')
const token = ref('')

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const videoExtensions = ['mp4','avi','mkv','mov','wmv','flv','webm']
const audioExtensions = ['mp3','wav','ogg','flac','aac','m4a']
const docExtensions = ['pdf','doc','docx','xls','xlsx','ppt','pptx','txt','csv']

const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''

const isImage = computed(() => {
  if (!fileInfo.value || fileInfo.value.is_folder || !fileInfo.value.name) return false
  return imageExtensions.includes(getFileExt(fileInfo.value.name))
})

const isChildImage = (child: any) => !child.is_folder && imageExtensions.includes(getFileExt(child.name))
const getChildType = (child: any) => {
  const ext = getFileExt(child.name)
  if (videoExtensions.includes(ext)) return 'video'
  if (audioExtensions.includes(ext)) return 'audio'
  if (docExtensions.includes(ext)) return 'doc'
  return 'file'
}

const thumbnailUrl = computed(() => {
  if (!fileInfo.value) return ''
  return `${useRuntimeConfig().public.apiBase}/shared/${token.value}/thumbnail`
})

const childThumbnailUrl = (childId: number) => {
  return `${useRuntimeConfig().public.apiBase}/shared/${token.value}/thumbnail?child_id=${childId}`
}

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
.shared-layout {
  min-height: 100vh;
  background-color: var(--color-surface-1);
  padding: 2rem 2.5rem;
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 1.75rem;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.bc-icon {
  color: var(--color-text-muted);
}

.breadcrumb-item {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: default;
  padding: 0.25rem 0.375rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
}

.breadcrumb-item.active {
  color: var(--color-text-primary);
  font-weight: 600;
}

.header-meta {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.expiry-badge {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  font-size: 0.75rem;
  color: var(--color-danger);
  background-color: rgba(250, 82, 82, 0.1);
  padding: 0.375rem 0.625rem;
  border-radius: 9999px;
}

.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.empty-state h3 {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.375rem 0;
}

.empty-state p {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin: 0;
}

.file-grid-large {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 1rem;
}

.file-card-large {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  border-radius: 0.5rem;
  cursor: default;
  transition: background-color 0.1s ease;
}

.file-card-large:hover {
  background-color: var(--color-surface-0);
}

.file-card-icon-large {
  width: 8rem;
  height: 8rem;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.625rem;
  overflow: hidden;
}

.file-type-image { background-color: rgba(59, 130, 246, 0.1); color: #3b82f6; }
.file-type-video { background-color: rgba(168, 85, 247, 0.1); color: #a855f7; }
.file-type-audio { background-color: rgba(249, 115, 22, 0.1); color: #f97316; }
.file-type-doc { background-color: rgba(34, 197, 94, 0.1); color: #22c55e; }
.file-type-file { background-color: var(--color-surface-2); color: var(--color-text-muted); }
.file-icon-folder { background-color: rgba(76, 110, 245, 0.1); color: var(--color-brand-600); }

.file-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.file-card-name {
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  text-align: center;
  word-break: break-word;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.file-card-meta {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.25rem;
  margin-bottom: 0.5rem;
}

.shared-download-btn {
  width: 100%;
  margin-top: 0.5rem;
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

.btn-primary:hover { opacity: 0.9; }

.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
