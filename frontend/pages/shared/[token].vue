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
        <div class="file-card-icon-large" :class="child.is_folder ? 'file-icon-folder' : `file-type-${getChildType(child)}`" @click="openLightbox(child)">
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

    <Transition name="modal">
      <div v-if="lightboxFile" class="lightbox-overlay">
        <div class="lightbox-header">
          <button class="lightbox-close" @click="closeLightbox"><X :size="20" /></button>
          <div class="lightbox-file-info">
            <div class="lightbox-file-icon">
              <Image :size="16" />
            </div>
            <span class="lightbox-filename">{{ lightboxFile.name }}</span>
          </div>
          <div class="lightbox-actions">
            <a :href="`${useRuntimeConfig().public.apiBase}/shared/${token}/download?child_id=${lightboxFile.id}`" class="lightbox-btn" title="Download"><Download :size="18" /></a>
          </div>
        </div>
        <div class="lightbox-body" @click="closeLightbox" @mousedown="startPan" @mousemove="onPan" @mouseup="stopPan" @mouseleave="stopPan">
          <button v-if="hasMultipleImages" class="lightbox-nav lightbox-nav-prev" @click.stop="prevImage"><ChevronLeft :size="24" /></button>
          <img :src="lightboxThumbnailUrl" class="lightbox-img" :style="lightboxImageStyle" @click.stop @wheel.prevent="handleZoom" draggable="false" />
          <button v-if="hasMultipleImages" class="lightbox-nav lightbox-nav-next" @click.stop="nextImage"><ChevronRight :size="24" /></button>
        </div>
        <div class="lightbox-footer">
          <button class="lightbox-zoom-btn" @click="lightboxZoom = Math.max(0.25, lightboxZoom - 0.25)"><Minus :size="18" /></button>
          <button class="lightbox-zoom-btn" @click="resetView"><Search :size="18" /></button>
          <button class="lightbox-zoom-btn" @click="lightboxZoom = Math.min(4, lightboxZoom + 0.25)"><Plus :size="18" /></button>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { Folder, File, FolderOpen, Download, Loader2, AlertTriangle, Clock, Film, Music, FileText, Home, X, Minus, Plus, Search, ChevronLeft, ChevronRight } from 'lucide-vue-next'

definePageMeta({ layout: false })

const route = useRoute()

const loading = ref(true)
const error = ref('')
const fileInfo = ref<any>(null)
const children = ref<any[]>([])
const expiresAt = ref('')
const token = ref('')
const lightboxFile = ref<any>(null)
const lightboxZoom = ref(1)
const lightboxPanX = ref(0)
const lightboxPanY = ref(0)
const lightboxIndex = ref(0)
const isPanning = ref(false)
const panStart = ref({ x: 0, y: 0 })

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

const openLightbox = (child: any) => {
  if (child.is_folder || !isChildImage(child)) return
  lightboxIndex.value = children.value.findIndex((c: any) => c.id === child.id)
  lightboxFile.value = children.value[lightboxIndex.value]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const closeLightbox = () => {
  lightboxFile.value = null
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const handleZoom = (e: WheelEvent) => {
  if (e.deltaY < 0) lightboxZoom.value = Math.min(4, lightboxZoom.value + 0.1)
  else lightboxZoom.value = Math.max(0.25, lightboxZoom.value - 0.1)
}

const resetView = () => {
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const startPan = (e: MouseEvent) => {
  if (lightboxZoom.value <= 1) return
  isPanning.value = true
  panStart.value = { x: e.clientX - lightboxPanX.value, y: e.clientY - lightboxPanY.value }
}

const onPan = (e: MouseEvent) => {
  if (!isPanning.value) return
  lightboxPanX.value = e.clientX - panStart.value.x
  lightboxPanY.value = e.clientY - panStart.value.y
}

const stopPan = () => {
  isPanning.value = false
}

const lightboxImageStyle = computed(() => {
  return {
    transform: `translate(${lightboxPanX.value}px, ${lightboxPanY.value}px) scale(${lightboxZoom.value})`,
    cursor: lightboxZoom.value > 1 ? (isPanning.value ? 'grabbing' : 'grab') : 'default'
  }
})

const nextImage = () => {
  const images = children.value.filter((c: any) => !c.is_folder && isChildImage(c))
  if (images.length <= 1) return
  const currentIdx = images.findIndex((c: any) => c.id === lightboxFile.value?.id)
  const nextIdx = (currentIdx + 1) % images.length
  lightboxIndex.value = children.value.indexOf(images[nextIdx])
  lightboxFile.value = images[nextIdx]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const prevImage = () => {
  const images = children.value.filter((c: any) => !c.is_folder && isChildImage(c))
  if (images.length <= 1) return
  const currentIdx = images.findIndex((c: any) => c.id === lightboxFile.value?.id)
  const prevIdx = (currentIdx - 1 + images.length) % images.length
  lightboxIndex.value = children.value.indexOf(images[prevIdx])
  lightboxFile.value = images[prevIdx]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const hasMultipleImages = computed(() => {
  return children.value.filter((c: any) => !c.is_folder && isChildImage(c)).length > 1
})

const lightboxThumbnailUrl = computed(() => {
  if (!lightboxFile.value) return ''
  return `${useRuntimeConfig().public.apiBase}/shared/${token.value}/thumbnail?child_id=${lightboxFile.value.id}`
})

const handleKeydown = (e: KeyboardEvent) => {
  if (!lightboxFile.value) return
  if (e.key === 'Escape') closeLightbox()
  else if (e.key === 'ArrowRight') nextImage()
  else if (e.key === 'ArrowLeft') prevImage()
}

onMounted(async () => {
  if (import.meta.client) {
    window.addEventListener('keydown', handleKeydown)
  }
  // ... existing fetch code
})

onUnmounted(() => {
  if (import.meta.client) {
    window.removeEventListener('keydown', handleKeydown)
  }
})

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

.lightbox-overlay {
  position: fixed;
  inset: 0;
  background-color: #1a1a2e;
  display: flex;
  flex-direction: column;
  z-index: 300;
}

.lightbox-header {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  background-color: #1a1a2e;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  z-index: 301;
}

.lightbox-close {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.375rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s ease;
}

.lightbox-close:hover { background-color: rgba(255, 255, 255, 0.1); }

.lightbox-file-info {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  flex: 1;
  min-width: 0;
}

.lightbox-file-icon {
  width: 1.75rem;
  height: 1.75rem;
  background-color: rgba(59, 130, 246, 0.15);
  color: #3b82f6;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.lightbox-filename {
  color: rgba(255, 255, 255, 0.9);
  font-size: 0.875rem;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.lightbox-actions { display: flex; gap: 0.25rem; }

.lightbox-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s ease;
  text-decoration: none;
}

.lightbox-btn:hover { background-color: rgba(255, 255, 255, 0.1); }

.lightbox-body {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  cursor: pointer;
  position: relative;
  user-select: none;
}

.lightbox-img {
  max-width: 85vw;
  max-height: 80vh;
  object-fit: contain;
  transition: transform 0.15s ease;
  pointer-events: none;
}

.lightbox-nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background-color: rgba(0, 0, 0, 0.4);
  border: none;
  color: white;
  width: 3rem;
  height: 3rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  z-index: 10;
  transition: background-color 0.15s ease;
}

.lightbox-nav:hover {
  background-color: rgba(0, 0, 0, 0.7);
}

.lightbox-nav-prev { left: 1rem; }
.lightbox-nav-next { right: 1rem; }

.lightbox-footer {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.25rem;
  padding: 0.75rem;
  background-color: #1a1a2e;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  z-index: 301;
}

.lightbox-zoom-btn {
  background: none;
  border: none;
  color: rgba(255, 255, 255, 0.7);
  cursor: pointer;
  padding: 0.5rem;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background-color 0.15s ease;
}

.lightbox-zoom-btn:hover { background-color: rgba(255, 255, 255, 0.1); }

.modal-enter-active { transition: opacity 0.15s ease; }
.modal-leave-active { transition: opacity 0.1s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }

.spin { animation: spin 1s linear infinite; }
@keyframes spin { from { transform: rotate(0deg); } to { transform: rotate(360deg); } }
</style>
