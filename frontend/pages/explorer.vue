<template>
  <div class="app-layout" @dragover.prevent @drop.prevent="handleDrop" @dragenter.prevent="dragEnter" @dragleave.prevent="dragLeave">
    <AppSidebar current="explorer" />
    <div class="app-main">
      <header class="page-header">
        <div class="breadcrumb" v-if="breadcrumbs.length">
          <button class="breadcrumb-item" @click="navigateToFolder(null)">
            <Home :size="14" />
            <span>My Drive</span>
          </button>
          <template v-for="(crumb, i) in breadcrumbs" :key="crumb.id">
            <ChevronRight :size="14" class="breadcrumb-sep" />
            <button class="breadcrumb-item" :class="{ active: i === breadcrumbs.length - 1 }" @click="navigateToFolder(crumb.id)">
              {{ crumb.name }}
            </button>
          </template>
        </div>
        <div v-else>
          <h1 class="page-title">My Drive</h1>
          <p class="page-subtitle">Browse and manage your files</p>
        </div>
        <div class="header-actions">
          <div class="view-toggle-wrapper">
            <button class="btn-icon" @click="showViewMenu = !showViewMenu" title="Change view">
              <component :is="currentViewIcon" :size="16" />
            </button>
            <Transition name="menu">
              <div v-if="showViewMenu" class="view-menu">
                <button v-for="v in viewModes" :key="v.id" class="view-menu-item" :class="{ active: viewMode === v.id }" @click="setViewMode(v.id)">
                  <component :is="v.icon" :size="14" />
                  <span>{{ v.label }}</span>
                  <span v-if="viewMode === v.id" class="view-check">&#10003;</span>
                </button>
              </div>
            </Transition>
          </div>
          <button class="btn-secondary" @click="showNewFolder = true">
            <FolderPlus :size="16" />
            <span>New Folder</span>
          </button>
          <label class="btn-primary upload-btn">
            <Upload :size="16" />
            <span>Upload</span>
            <input type="file" multiple @change="handleUpload" style="display: none;" />
          </label>
        </div>
      </header>

      <div v-if="showNewFolder" class="card" style="margin-bottom: 1rem; padding: 1rem 1.25rem;">
        <div style="display: flex; gap: 0.75rem; align-items: flex-end;">
          <div style="flex: 1;">
            <label style="display: block; font-size: 0.75rem; font-weight: 500; color: var(--color-text-secondary); margin-bottom: 0.25rem;">Folder name</label>
            <input v-model="newFolderName" class="input-field" placeholder="Untitled folder" @keyup.enter="createFolder" autofocus />
          </div>
          <button class="btn-primary" @click="createFolder" :disabled="!newFolderName.trim()" style="height: 2.25rem;">Create</button>
          <button class="btn-secondary" @click="showNewFolder = false; newFolderName = ''" style="height: 2.25rem;">Cancel</button>
        </div>
      </div>

      <div class="drop-zone" :class="{ 'drop-zone-active': isDragging }">
        <div v-if="isDragging" class="drop-overlay">
          <Upload :size="40" style="color: var(--color-brand-500);" />
          <p style="font-size: 1rem; font-weight: 600; color: var(--color-text-primary); margin: 0.75rem 0 0.25rem 0;">Drop files here</p>
          <p style="font-size: 0.8125rem; color: var(--color-text-muted); margin: 0;">Files will upload to the current folder</p>
        </div>

        <template v-else>
          <div v-if="loading" class="empty-state">
            <Loader2 :size="24" class="spin" style="color: var(--color-text-muted);" />
            <p style="margin-top: 0.75rem; font-size: 0.8125rem; color: var(--color-text-muted);">Loading files...</p>
          </div>

          <div v-else-if="files.length === 0" class="empty-state">
            <FolderOpen :size="48" style="color: var(--color-surface-3); margin-bottom: 1rem;" />
            <h3 style="font-size: 1rem; font-weight: 600; color: var(--color-text-primary); margin: 0 0 0.375rem 0;">No files here</h3>
            <p style="font-size: 0.8125rem; color: var(--color-text-muted);">Upload files or create a folder to get started.</p>
          </div>

          <div v-else>
            <!-- Details view -->
            <div v-if="viewMode === 'details'">
              <div class="file-list-header">
                <span class="file-col-name">Name</span>
                <span class="file-col-size">Size</span>
                <span class="file-col-date">Date Modified</span>
                <span class="file-col-actions"></span>
              </div>
              <div class="file-list">
                <div v-for="file in files" :key="file.id" class="file-row" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                  <div class="file-col-name">
                    <div class="file-icon" :class="file.is_folder ? 'file-icon-folder' : 'file-icon-file'">
                      <Folder v-if="file.is_folder" :size="16" />
                      <File v-else :size="16" />
                    </div>
                    <span class="file-name" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
                  </div>
                  <span class="file-col-size">{{ file.is_folder ? '--' : formatSize(file.size_total) }}</span>
                  <span class="file-col-date">{{ formatDate(file.updated_at) }}</span>
                  <div class="file-col-actions">
                    <button class="icon-btn" @click="confirmDelete(file)" title="Delete"><Trash2 :size="14" /></button>
                    <button v-if="!file.is_folder" class="icon-btn" @click="downloadFile(file)" title="Download"><Download :size="14" /></button>
                  </div>
                </div>
              </div>
            </div>

            <!-- List view -->
            <div v-else-if="viewMode === 'list'" class="file-list-simple">
              <div v-for="file in files" :key="file.id" class="file-row-simple" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-icon-sm" :class="file.is_folder ? 'file-icon-folder' : 'file-icon-file'">
                  <Folder v-if="file.is_folder" :size="14" />
                  <File v-else :size="14" />
                </div>
                <span class="file-name" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
                <div class="file-row-actions">
                  <button class="icon-btn" @click="confirmDelete(file)" title="Delete"><Trash2 :size="14" /></button>
                  <button v-if="!file.is_folder" class="icon-btn" @click="downloadFile(file)" title="Download"><Download :size="14" /></button>
                </div>
              </div>
            </div>

            <!-- Large icons -->
            <div v-else-if="viewMode === 'large'" class="file-grid-large">
              <div v-for="file in files" :key="file.id" class="file-card-large" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-card-icon-large" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : isVideo(file) ? 'video' : isAudio(file) ? 'audio' : isDoc(file) ? 'doc' : 'file'}`">
                  <Folder v-if="file.is_folder" :size="36" />
                  <Image v-else-if="isImage(file)" :size="36" />
                  <Film v-else-if="isVideo(file)" :size="36" />
                  <Music v-else-if="isAudio(file)" :size="36" />
                  <FileText v-else-if="isDoc(file)" :size="36" />
                  <File v-else :size="36" />
                </div>
                <span class="file-card-name" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
                <div class="file-card-actions">
                  <button class="icon-btn" @click="confirmDelete(file)" title="Delete"><Trash2 :size="14" /></button>
                  <button v-if="!file.is_folder" class="icon-btn" @click="downloadFile(file)" title="Download"><Download :size="14" /></button>
                </div>
              </div>
            </div>

            <!-- Medium icons -->
            <div v-else-if="viewMode === 'medium'" class="file-grid-medium">
              <div v-for="file in files" :key="file.id" class="file-card-medium" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-card-icon-medium" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : isVideo(file) ? 'video' : isAudio(file) ? 'audio' : isDoc(file) ? 'doc' : 'file'}`">
                  <Folder v-if="file.is_folder" :size="24" />
                  <Image v-else-if="isImage(file)" :size="24" />
                  <Film v-else-if="isVideo(file)" :size="24" />
                  <Music v-else-if="isAudio(file)" :size="24" />
                  <FileText v-else-if="isDoc(file)" :size="24" />
                  <File v-else :size="24" />
                </div>
                <span class="file-card-name-sm" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
              </div>
            </div>

            <!-- Small icons -->
            <div v-else-if="viewMode === 'small'" class="file-grid-small">
              <div v-for="file in files" :key="file.id" class="file-card-small" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-card-icon-sm" :class="file.is_folder ? 'file-icon-folder' : 'file-icon-file'">
                  <Folder v-if="file.is_folder" :size="16" />
                  <File v-else :size="16" />
                </div>
                <span class="file-card-name-sm" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>

    <Transition name="modal">
      <div v-if="showDeleteConfirm && deleteTarget" class="modal-overlay" @click.self="cancelDelete">
        <div class="modal-card">
          <div class="modal-icon">
            <AlertTriangle :size="24" style="color: var(--color-danger);" />
          </div>
          <h3 class="modal-title">Delete {{ deleteTarget.isFolder ? 'folder' : 'file' }}?</h3>
          <p class="modal-desc">"{{ deleteTarget.name }}" will be permanently deleted{{ deleteTarget.isFolder ? ' along with all its contents' : '' }}. This action cannot be undone.</p>
          <div class="modal-actions">
            <button class="btn-secondary" @click="cancelDelete" style="height: 2.25rem;">Cancel</button>
            <button class="btn-danger" @click="executeDelete" style="height: 2.25rem;">Delete</button>
          </div>
        </div>
      </div>
    </Transition>

    <div class="panels-container">
      <Transition name="upload-panel">
        <div v-if="uploads.length" class="upload-panel">
          <div class="upload-panel-header">
            <div style="display: flex; align-items: center; gap: 0.5rem;">
              <Upload :size="14" style="color: var(--color-brand-500);" />
              <span style="font-size: 0.8125rem; font-weight: 600; color: var(--color-text-primary);">
                {{ uploadingCount ? `Uploading ${uploadingCount} file${uploadingCount > 1 ? 's' : ''}...` : `Uploaded ${completedCount} file${completedCount > 1 ? 's' : ''}` }}
              </span>
            </div>
            <button class="upload-panel-close" @click="clearCompletedUploads" v-if="!uploadingCount">
              <X :size="14" />
            </button>
          </div>
          <div class="upload-panel-body">
            <div v-for="upload in uploads" :key="upload.id" class="upload-item">
              <div class="upload-item-icon">
                <File :size="14" />
              </div>
              <div class="upload-item-info">
                <div class="upload-item-name">{{ upload.name }}</div>
                <div class="upload-item-meta">
                  <span v-if="upload.status === 'uploading'">{{ upload.percent }}% - {{ formatSize(upload.loaded) }} of {{ formatSize(upload.total) }}</span>
                  <span v-else-if="upload.status === 'done'" style="color: var(--color-success);">Done</span>
                  <span v-else-if="upload.status === 'error'" style="color: var(--color-danger);">Failed</span>
                  <span v-else-if="upload.status === 'cancelled'" style="color: var(--color-text-muted);">Cancelled</span>
                  <span v-else-if="upload.status === 'queued'" style="color: var(--color-text-muted);">Waiting...</span>
                </div>
                <div class="upload-item-progress">
                  <div class="upload-item-progress-track">
                    <div
                      class="upload-item-progress-fill"
                      :class="{ 'fill-error': upload.status === 'error' || upload.status === 'cancelled', 'fill-done': upload.status === 'done' }"
                      :style="{ width: upload.percent + '%' }"
                    ></div>
                  </div>
                </div>
              </div>
              <button v-if="upload.status === 'uploading'" class="cancel-btn" @click="cancelUpload(upload)" title="Cancel">
                <X :size="14" />
              </button>
            </div>
          </div>
        </div>
      </Transition>

      <Transition name="upload-panel">
        <div v-if="activeDownloads.length" class="download-panel">
          <div class="upload-panel-header">
            <div style="display: flex; align-items: center; gap: 0.5rem;">
              <Download :size="14" style="color: var(--color-brand-500);" />
              <span style="font-size: 0.8125rem; font-weight: 600; color: var(--color-text-primary);">
                {{ activeDownloads.some(d => d.status === 'downloading') ? `Downloading...` : `Downloaded ${activeDownloads.length} file${activeDownloads.length > 1 ? 's' : ''}` }}
              </span>
            </div>
            <button class="upload-panel-close" @click="clearCompletedDownloads" v-if="!activeDownloads.some(d => d.status === 'downloading')">
              <X :size="14" />
            </button>
          </div>
          <div class="upload-panel-body">
            <div v-for="dl in activeDownloads" :key="dl.id" class="upload-item">
              <div class="upload-item-icon">
                <Download :size="14" />
              </div>
              <div class="upload-item-info">
                <div class="upload-item-name">{{ dl.fileName }}</div>
                <div class="upload-item-meta">
                  <span v-if="dl.status === 'downloading'">{{ dl.progress }}% - {{ dl.chunksDone }} / {{ dl.chunksTotal }} chunks</span>
                  <span v-else-if="dl.status === 'done'" style="color: var(--color-success);">Done</span>
                  <span v-else-if="dl.status === 'error'" style="color: var(--color-danger);">Failed</span>
                  <span v-else-if="dl.status === 'cancelled'" style="color: var(--color-text-muted);">Cancelled</span>
                </div>
                <div class="upload-item-progress">
                  <div class="upload-item-progress-track">
                    <div
                      class="upload-item-progress-fill"
                      :class="{ 'fill-error': dl.status === 'error' || dl.status === 'cancelled', 'fill-done': dl.status === 'done' }"
                      :style="{ width: dl.progress + '%' }"
                    ></div>
                  </div>
                </div>
              </div>
              <button v-if="dl.status === 'downloading'" class="cancel-btn" @click="cancelDownload(dl)" title="Cancel">
                <X :size="14" />
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </div>
  </div>
</template>

<script setup lang="ts">
import { FolderOpen, FolderPlus, Upload, Folder, File, Trash2, Download, ChevronRight, Home, Loader2, X, AlertTriangle, LayoutGrid, List, LayoutList, Grip, Image, Film, Music, FileText } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()
const route = useRoute()
const router = useRouter()

const files = ref<any[]>([])
const loading = ref(true)
const currentFolder = ref<number | null>(null)
const breadcrumbs = ref<{ id: number; name: string }[]>([])
const showNewFolder = ref(false)
const newFolderName = ref('')
const isDragging = ref(false)
const viewMode = ref('details')
const showViewMenu = ref(false)

const viewModes = [
  { id: 'details', label: 'Details', icon: LayoutList },
  { id: 'list', label: 'List', icon: List },
  { id: 'large', label: 'Large icons', icon: LayoutGrid },
  { id: 'medium', label: 'Medium icons', icon: Grip },
  { id: 'small', label: 'Small icons', icon: LayoutGrid },
]

onMounted(() => {
  if (import.meta.client) {
    const saved = localStorage.getItem('viewMode')
    if (saved && viewModes.find(v => v.id === saved)) {
      viewMode.value = saved
    }
  }
})

const setViewMode = (mode: string) => {
  viewMode.value = mode
  localStorage.setItem('viewMode', mode)
  showViewMenu.value = false
}

const currentViewIcon = computed(() => viewModes.find(v => v.id === viewMode.value)?.icon || LayoutList)

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const videoExtensions = ['mp4','avi','mkv','mov','wmv','flv','webm']
const audioExtensions = ['mp3','wav','ogg','flac','aac','m4a']
const docExtensions = ['pdf','doc','docx','xls','xlsx','ppt','pptx','txt','csv']

const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''

const isImage = (file: any) => !file.is_folder && imageExtensions.includes(getFileExt(file.name))
const isVideo = (file: any) => !file.is_folder && videoExtensions.includes(getFileExt(file.name))
const isAudio = (file: any) => !file.is_folder && audioExtensions.includes(getFileExt(file.name))
const isDoc = (file: any) => !file.is_folder && docExtensions.includes(getFileExt(file.name))
let dragCounter = 0

const showDeleteConfirm = ref(false)
const deleteTarget = ref<{ id: number; name: string; isFolder: boolean } | null>(null)

const dragEnter = () => {
  dragCounter++
  isDragging.value = true
}

const dragLeave = () => {
  dragCounter--
  if (dragCounter <= 0) {
    dragCounter = 0
    isDragging.value = false
  }
}

interface UploadItem {
  id: string
  name: string
  status: 'queued' | 'uploading' | 'done' | 'error'
  percent: number
  loaded: number
  total: number
}

const uploads = ref<UploadItem[]>([])
let uploadCounter = 0
const pendingFiles = new Set<string>()

const uploadingCount = computed(() => uploads.value.filter(u => u.status === 'uploading' || u.status === 'queued').length)
const completedCount = computed(() => uploads.value.filter(u => u.status === 'done').length)

const formatSize = (bytes: number) => {
  if (!bytes) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return `${(bytes / Math.pow(1024, i)).toFixed(1)} ${units[i]}`
}

const formatDate = (dateStr: string) => {
  if (!dateStr) return '--'
  const d = new Date(dateStr)
  if (isNaN(d.getTime())) return '--'
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  const mins = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  if (mins < 1) return 'Just now'
  if (mins < 60) return `${mins}m ago`
  if (hours < 24) return `${hours}h ago`
  if (days < 7) return `${days}d ago`
  return d.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: d.getFullYear() !== now.getFullYear() ? 'numeric' : undefined })
}

const loadFiles = async () => {
  loading.value = true
  try {
    const params = currentFolder.value ? `?parent_id=${currentFolder.value}` : ''
    files.value = (await apiFetch(`/api/files${params}`)) as any[]
  } catch {
    files.value = []
  } finally {
    loading.value = false
  }
}

const loadBreadcrumbs = async (folderId: number | null) => {
  if (!folderId) { breadcrumbs.value = []; return }
  try {
    const data = await apiFetch(`/api/files/breadcrumb?folder_id=${folderId}`) as any[]
    breadcrumbs.value = data || []
  } catch {
    breadcrumbs.value = []
  }
}

const navigateToFolder = async (id: number | null) => {
  currentFolder.value = id
  if (id) {
    router.replace({ query: { folder: String(id) } })
  } else {
    router.replace({ query: {} })
  }
  await loadBreadcrumbs(id)
  await loadFiles()
}

const createFolder = async () => {
  if (!newFolderName.value.trim()) return
  try {
    await apiFetch('/api/files/folder', {
      method: 'POST',
      body: { name: newFolderName.value.trim(), parent_id: currentFolder.value }
    })
    newFolderName.value = ''
    showNewFolder.value = false
    await loadFiles()
  } catch {}
}

const confirmDelete = (file: any) => {
  deleteTarget.value = { id: file.id, name: file.name, isFolder: file.is_folder }
  showDeleteConfirm.value = true
}

const cancelDelete = () => {
  showDeleteConfirm.value = false
  deleteTarget.value = null
}

const executeDelete = async () => {
  if (!deleteTarget.value) return
  try {
    await apiFetch(`/api/files/${deleteTarget.value.id}`, { method: 'DELETE' })
    await loadFiles()
  } catch {}
  showDeleteConfirm.value = false
  deleteTarget.value = null
}

const downloadFile = async (file: any) => {
  if (activeDownloads.value.find(d => d.fileName === file.name && d.status === 'downloading')) return
  const id = `dl-${++downloadCounter}`
  activeDownloads.value.push({
    id,
    fileName: file.name,
    fileSize: file.size_total,
    status: 'downloading',
    progress: 0,
    chunksDone: 0,
    chunksTotal: 0,
    sessionId: ''
  })
  const idx = activeDownloads.value.length - 1
  try {
    const resp = await apiFetch('/api/files/download-by-name', {
      method: 'POST',
      body: { name: file.name }
    }) as any
    activeDownloads.value[idx].sessionId = resp.session_id
    activeDownloads.value[idx].chunksTotal = resp.chunks
    activeDownloads.value[idx].fileSize = resp.file_size
    pollDownloadProgress(idx, file.name)
  } catch (e: any) {
    activeDownloads.value[idx].status = 'error'
  }
}

const pollDownloadProgress = async (idx: number, fileName: string) => {
  const token = localStorage.getItem('token')
  const apiBase = useRuntimeConfig().public.apiBase

  const poll = async () => {
    try {
      if (activeDownloads.value[idx]?._aborted) return
      const sessId = activeDownloads.value[idx]?.sessionId
      if (!sessId) return
      const resp = await fetch(`${apiBase}/api/files/0/download-progress?session=${sessId}`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!resp.ok) return
      const data = await resp.json()

      activeDownloads.value[idx].progress = data.progress || 0
      activeDownloads.value[idx].chunksDone = data.chunks_done || 0
      activeDownloads.value[idx].chunksTotal = data.chunks_total || 1

      if (data.status === 'ready') {
        activeDownloads.value[idx].status = 'done'
        activeDownloads.value[idx].progress = 100
        const fileResp = await fetch(`${apiBase}/api/files/download-by-name`, {
          method: 'POST',
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({ name: fileName })
        })
        if (fileResp.ok) {
          const blob = await fileResp.blob()
          const disposition = fileResp.headers.get('Content-Disposition') || ''
          const nameMatch = disposition.match(/filename="?([^"]+)"?/)
          const fname = nameMatch ? nameMatch[1] : fileName
          const url = URL.createObjectURL(blob)
          const a = document.createElement('a')
          a.href = url
          a.download = fname
          a.click()
          URL.revokeObjectURL(url)
        }
        return
      }
      if (data.status === 'error') {
        activeDownloads.value[idx].status = 'error'
        return
      }
      setTimeout(poll, 500)
    } catch {
      if (!activeDownloads.value[idx]?._aborted) {
        setTimeout(poll, 1000)
      }
    }
  }
  poll()
}

const activeDownloads = ref<any[]>([])
let downloadCounter = 0

const clearCompletedDownloads = () => {
  activeDownloads.value = activeDownloads.value.filter(d => d.status === 'downloading')
}

const cancelUpload = async (upload: any) => {
  upload._aborted = true
  if (upload._pollTimer) clearTimeout(upload._pollTimer)
  if (upload._xhr) {
    upload._xhr.abort()
  }
  if (upload._fileId) {
    try {
      await apiFetch(`/api/files/${upload._fileId}`, { method: 'DELETE' })
    } catch {}
  }
  upload.status = 'cancelled'
}

const cancelDownload = async (dl: any) => {
  dl._aborted = true
  dl.status = 'cancelled'
  if (dl.sessionId) {
    try {
      const token = localStorage.getItem('token')
      const apiBase = useRuntimeConfig().public.apiBase
      await fetch(`${apiBase}/api/files/download-cancel?session=${dl.sessionId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${token}` }
      })
    } catch {}
  }
}

const uploadFile = (file: File) => {
  const key = `${file.name}-${file.size}-${file.lastModified}`
  if (pendingFiles.has(key)) return
  pendingFiles.add(key)

  const id = `upload-${++uploadCounter}-${Date.now()}`
  const abortController = new AbortController()
  uploads.value.push({
    id,
    name: file.name,
    status: 'uploading',
    percent: 0,
    loaded: 0,
    total: file.size,
    _xhr: null as XMLHttpRequest | null,
    _aborted: false,
    _pollTimer: null as ReturnType<typeof setTimeout> | null,
    _fileId: null as number | null
  })

  const idx = uploads.value.length - 1

  const formData = new FormData()
  formData.append('file', file)
  if (currentFolder.value) formData.append('parent_id', String(currentFolder.value))

  const xhr = new XMLHttpRequest()
  uploads.value[idx]._xhr = xhr

  xhr.upload.addEventListener('progress', (e) => {
    if (e.lengthComputable && e.total > 0) {
      const pct = Math.round((e.loaded / e.total) * 100)
      uploads.value[idx].percent = Math.min(pct, 99)
      uploads.value[idx].loaded = e.loaded
      uploads.value[idx].total = e.total
    }
  })

  xhr.addEventListener('load', () => {
    if (xhr.status >= 200 && xhr.status < 300) {
      const resp = JSON.parse(xhr.responseText)
      const fileId = resp.id
      uploads.value[idx]._fileId = fileId
      uploads.value[idx].loaded = uploads.value[idx].total
      uploads.value[idx].percent = 99
      pollUploadProgress(idx, fileId)
    } else {
      pendingFiles.delete(key)
      uploads.value[idx].status = 'error'
    }
  })

  xhr.addEventListener('error', () => {
    pendingFiles.delete(key)
    uploads.value[idx].status = 'error'
  })

  xhr.open('POST', `${useRuntimeConfig().public.apiBase}/api/files/upload`)
  xhr.setRequestHeader('Authorization', `Bearer ${localStorage.getItem('token')}`)
  xhr.send(formData)
}

const pollUploadProgress = async (idx: number, fileId: number) => {
  const token = localStorage.getItem('token')
  const apiBase = useRuntimeConfig().public.apiBase
  let lastServerPercent = 0

  const poll = async (): Promise<boolean> => {
    try {
      const resp = await fetch(`${apiBase}/api/files/${fileId}/progress`, {
        headers: { Authorization: `Bearer ${token}` }
      })
      if (!resp.ok) return false
      const data = await resp.json()

      if (data.status === 'active') {
        uploads.value[idx].percent = 100
        uploads.value[idx].loaded = uploads.value[idx].total
        uploads.value[idx].status = 'done'
        loadFiles()
        return true
      }

      if (data.status === 'error') {
        uploads.value[idx].status = 'error'
        return true
      }

      const serverPercent = data.progress || 0
      if (serverPercent > lastServerPercent) {
        lastServerPercent = serverPercent
        uploads.value[idx].percent = Math.min(serverPercent, 99)
        uploads.value[idx].loaded = Math.round((serverPercent / 100) * uploads.value[idx].total)
      }

      return false
    } catch {
      return false
    }
  }

  const loop = async () => {
    if (uploads.value[idx]?._aborted) return
    const done = await poll()
    if (!done && !uploads.value[idx]?._aborted) {
      uploads.value[idx]._pollTimer = setTimeout(loop, 500)
    }
  }
  uploads.value[idx]._pollTimer = setTimeout(loop, 500)
}

const handleUpload = (e: Event) => {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  const fileList = Array.from(input.files)
  input.value = ''
  for (const file of fileList) {
    uploadFile(file)
  }
}

const handleDrop = (e: DragEvent) => {
  isDragging.value = false
  dragCounter = 0
  const droppedFiles = e.dataTransfer?.files
  if (!droppedFiles?.length) return
  for (const file of Array.from(droppedFiles)) {
    uploadFile(file)
  }
}

const clearCompletedUploads = () => {
  uploads.value = uploads.value.filter(u => u.status === 'uploading' || u.status === 'queued')
}

const handlePopState = () => {
  const folderParam = route.query.folder
  const folderId = folderParam ? Number(folderParam) : null
  currentFolder.value = folderId
  loadBreadcrumbs(folderId)
  loadFiles()
}

onMounted(async () => {
  if (import.meta.client) {
    if (!localStorage.getItem('token')) { navigateTo('/login'); return }
    document.addEventListener('click', (e) => {
      if (showViewMenu.value) {
        const target = e.target as HTMLElement
        if (!target.closest('.view-toggle-wrapper')) {
          showViewMenu.value = false
        }
      }
    })
  }
  const folderParam = route.query.folder
  if (folderParam) {
    currentFolder.value = Number(folderParam)
    await loadBreadcrumbs(currentFolder.value)
  }
  await loadFiles()
  window.addEventListener('popstate', handlePopState)
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
  align-items: center;
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

.header-actions {
  display: flex;
  gap: 0.5rem;
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

.upload-btn {
  cursor: pointer;
}

.breadcrumb {
  display: flex;
  align-items: center;
  gap: 0.25rem;
}

.breadcrumb-item {
  background: none;
  border: none;
  color: var(--color-text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
  cursor: pointer;
  padding: 0.25rem 0.375rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  gap: 0.375rem;
  transition: color 0.12s ease, background-color 0.12s ease;
}

.breadcrumb-item:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface-2);
}

.breadcrumb-item.active {
  color: var(--color-text-primary);
  font-weight: 600;
}

.breadcrumb-sep {
  color: var(--color-text-muted);
}

.empty-state {
  text-align: center;
  padding: 5rem 1.5rem;
}

.file-list-header {
  display: flex;
  align-items: center;
  padding: 0.5rem 0.75rem;
  font-size: 0.6875rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  border-bottom: 1px solid var(--color-surface-2);
}

.file-list {
  display: flex;
  flex-direction: column;
}

.file-row {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-surface-2);
  transition: background-color 0.1s ease;
}

.file-row:hover {
  background-color: var(--color-surface-1);
}

.file-col-name {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.625rem;
  min-width: 0;
}

.file-col-size {
  width: 100px;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  text-align: right;
}

.file-col-date {
  width: 120px;
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  text-align: right;
}

.file-col-actions {
  width: 80px;
  display: flex;
  justify-content: flex-end;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.1s ease;
}

.file-row:hover .file-col-actions {
  opacity: 1;
}

.file-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-icon-folder {
  background-color: rgba(76, 110, 245, 0.1);
  color: var(--color-brand-600);
}

.file-icon-file {
  background-color: var(--color-surface-2);
  color: var(--color-text-muted);
}

.file-type-image {
  background-color: rgba(59, 130, 246, 0.1);
  color: #3b82f6;
}

.file-type-video {
  background-color: rgba(168, 85, 247, 0.1);
  color: #a855f7;
}

.file-type-audio {
  background-color: rgba(249, 115, 22, 0.1);
  color: #f97316;
}

.file-type-doc {
  background-color: rgba(34, 197, 94, 0.1);
  color: #22c55e;
}

.file-name {
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.folder-name {
  cursor: pointer;
  font-weight: 500;
}

.folder-name:hover {
  color: var(--color-brand-600);
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
  transition: color 0.1s ease, background-color 0.1s ease;
}

.icon-btn:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface-2);
}

.btn-icon {
  background: none;
  border: 1px solid var(--color-surface-3);
  color: var(--color-text-secondary);
  padding: 0.5rem;
  border-radius: 0.5rem;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.12s ease;
}

.btn-icon:hover {
  background-color: var(--color-surface-1);
}

.view-toggle-wrapper {
  position: relative;
}

.view-menu {
  position: absolute;
  top: calc(100% + 0.25rem);
  right: 0;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 180px;
  z-index: 50;
  padding: 0.25rem 0;
}

.view-menu-item {
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
  transition: background-color 0.1s ease;
}

.view-menu-item:hover {
  background-color: var(--color-surface-1);
}

.view-menu-item.active {
  color: var(--color-brand-600);
  font-weight: 500;
}

.view-check {
  margin-left: auto;
  font-size: 0.75rem;
  color: var(--color-brand-600);
}

.menu-enter-active { transition: opacity 0.1s ease, transform 0.1s ease; }
.menu-leave-active { transition: opacity 0.08s ease, transform 0.08s ease; }
.menu-enter-from, .menu-leave-to { opacity: 0; transform: translateY(-4px); }

/* List view */
.file-list-simple {
  display: flex;
  flex-direction: column;
}

.file-row-simple {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.375rem 0.5rem;
  border-radius: 0.25rem;
  transition: background-color 0.1s ease;
}

.file-row-simple:hover {
  background-color: var(--color-surface-1);
}

.file-icon-sm {
  width: 1.5rem;
  height: 1.5rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.file-row-actions {
  margin-left: auto;
  display: flex;
  gap: 0.25rem;
  opacity: 0;
  transition: opacity 0.1s ease;
}

.file-row-simple:hover .file-row-actions {
  opacity: 1;
}

/* Large icons */
.file-grid-large {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
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
  background-color: var(--color-surface-1);
}

.file-card-icon-large {
  width: 5rem;
  height: 5rem;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.625rem;
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

.file-card-actions {
  display: flex;
  gap: 0.25rem;
  margin-top: 0.375rem;
  opacity: 0;
  transition: opacity 0.1s ease;
}

.file-card-large:hover .file-card-actions {
  opacity: 1;
}

/* Medium icons */
.file-grid-medium {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr));
  gap: 0.75rem;
}

.file-card-medium {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 0.75rem 0.5rem;
  border-radius: 0.5rem;
  cursor: default;
  transition: background-color 0.1s ease;
}

.file-card-medium:hover {
  background-color: var(--color-surface-1);
}

.file-card-icon-medium {
  width: 3.5rem;
  height: 3.5rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5rem;
}

.file-card-name-sm {
  font-size: 0.75rem;
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

/* Small icons */
.file-grid-small {
  display: flex;
  flex-wrap: wrap;
  gap: 0.375rem;
}

.file-card-small {
  display: flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0.25rem 0.5rem;
  border-radius: 0.25rem;
  cursor: default;
  transition: background-color 0.1s ease;
  min-width: 120px;
}

.file-card-small:hover {
  background-color: var(--color-surface-1);
}

.file-card-icon-sm {
  width: 1.25rem;
  height: 1.25rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
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
  padding: 1.5rem;
  width: 400px;
  max-width: 90vw;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.16);
}

.modal-icon {
  margin-bottom: 0.75rem;
}

.modal-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0 0 0.5rem 0;
}

.modal-desc {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin: 0 0 1.25rem 0;
  line-height: 1.5;
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 0.5rem;
}

.btn-danger {
  background-color: var(--color-danger);
  color: #fff;
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

.btn-danger:hover {
  opacity: 0.9;
}

.modal-enter-active { transition: opacity 0.15s ease; }
.modal-leave-active { transition: opacity 0.1s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }

.drop-zone {
  min-height: calc(100vh - 8rem);
  border: 2px dashed transparent;
  border-radius: 0.75rem;
  transition: border-color 0.15s ease, background-color 0.15s ease;
  position: relative;
}

.drop-zone-active {
  border-color: var(--color-brand-400);
  background-color: rgba(76, 110, 245, 0.04);
}

.drop-overlay {
  position: absolute;
  inset: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  z-index: 10;
  border-radius: 0.75rem;
  background-color: rgba(76, 110, 245, 0.06);
}

.panels-container {
  position: fixed;
  bottom: 1.5rem;
  right: 1.5rem;
  display: flex;
  gap: 0.75rem;
  z-index: 100;
}

.upload-panel {
  width: 360px;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.75rem;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  overflow: hidden;
}

.download-panel {
  width: 360px;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.75rem;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
  overflow: hidden;
}

.upload-panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.upload-panel-close {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.1s ease, background-color 0.1s ease;
}

.upload-panel-close:hover {
  color: var(--color-text-primary);
  background-color: var(--color-surface-2);
}

.upload-panel-body {
  max-height: 240px;
  overflow-y: auto;
  padding: 0.5rem 0;
}

.upload-item {
  display: flex;
  align-items: flex-start;
  gap: 0.625rem;
  padding: 0.5rem 1rem;
}

.upload-item-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  background-color: var(--color-surface-2);
  color: var(--color-text-muted);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  margin-top: 0.125rem;
}

.upload-item-info {
  flex: 1;
  min-width: 0;
}

.upload-item-name {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.upload-item-meta {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.upload-item-progress {
  margin-top: 0.375rem;
}

.cancel-btn {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: color 0.1s ease, background-color 0.1s ease;
}

.cancel-btn:hover {
  color: var(--color-danger);
  background-color: rgba(250, 82, 82, 0.08);
}

.upload-item-progress-track {
  width: 100%;
  height: 4px;
  background-color: var(--color-surface-2);
  border-radius: 9999px;
  overflow: hidden;
}

.upload-item-progress-fill {
  height: 100%;
  background-color: var(--color-brand-500);
  border-radius: 9999px;
  transition: width 0.2s ease;
}

.upload-item-progress-fill.fill-error {
  background-color: var(--color-danger);
}

.upload-item-progress-fill.fill-done {
  background-color: var(--color-success);
}

.upload-panel-enter-active {
  transition: all 0.25s ease;
}

.upload-panel-leave-active {
  transition: all 0.2s ease;
}

.upload-panel-enter-from {
  opacity: 0;
  transform: translateY(1rem);
}

.upload-panel-leave-to {
  opacity: 0;
  transform: translateY(0.5rem);
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
