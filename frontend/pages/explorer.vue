<template>
  <div class="app-layout" @dragover.prevent @drop.prevent="handleDrop" @dragenter.prevent="dragEnter" @dragleave.prevent="dragLeave">
    <AppSidebar current="explorer" />
    <div class="app-main">
      <AppTopBar title="My Drive" subtitle="Browse and manage your files" current-page="explorer" @hamburger-click="sidebarOpen = true">
        <template #title>
          <div v-if="breadcrumbs.length" class="breadcrumb">
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
          <template v-else>
            <h1 class="page-title">My Drive</h1>
            <p class="page-subtitle">Browse and manage your files</p>
          </template>
        </template>
      </AppTopBar>
      <div class="action-toolbar">
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

      <div class="drop-zone" :class="{ 'drop-zone-active': isDragging }" @dragstart.prevent>
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
                <div v-for="file in files" :key="file.id" class="file-row" @click="isImage(file) ? openLightbox(file) : (file.is_folder ? navigateToFolder(file.id) : null)">
                  <div class="file-col-name">
                    <div class="file-icon" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : 'file'}`">
                      <template v-if="file.is_folder"><Folder :size="16" /></template>
                      <template v-else-if="isImage(file)"><img :src="thumbnailUrl(file.id)" class="file-thumb-detail" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" /></template>
                      <template v-else><File :size="16" /></template>
                    </div>
                    <span class="file-name" :class="{ 'folder-name': file.is_folder }">{{ file.name }}</span>
                  </div>
                  <span class="file-col-size">{{ file.is_folder ? '--' : formatSize(file.size_total) }}</span>
                  <span class="file-col-date">{{ formatDate(file.updated_at) }}</span>
                  <div class="file-col-actions">
                    <button class="icon-btn" @click.stop="openContextMenu($event, file)" title="More">
                      <MoreVertical :size="14" />
                    </button>
                  </div>
                </div>
              </div>
            </div>

            <!-- List view -->
            <div v-else-if="viewMode === 'list'" class="file-list-simple">
              <div v-for="file in files" :key="file.id" class="file-row-simple" @click="isImage(file) ? openLightbox(file) : (file.is_folder ? navigateToFolder(file.id) : null)">
                <div class="file-icon-sm" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : 'file'}`">
                  <template v-if="file.is_folder"><Folder :size="14" /></template>
                  <template v-else-if="isImage(file)"><img :src="thumbnailUrl(file.id)" class="file-thumb-sm" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" /></template>
                  <template v-else><File :size="14" /></template>
                </div>
                <span class="file-name" :class="{ 'folder-name': file.is_folder }">{{ file.name }}</span>
                <div class="file-row-actions">
                  <button class="icon-btn" @click="openContextMenu($event, file)" title="More">
                    <MoreVertical :size="14" />
                  </button>
                </div>
              </div>
            </div>

            <!-- Large icons -->
            <div v-else-if="viewMode === 'large'" class="file-grid-large">
              <div v-for="file in files" :key="file.id" class="file-card-large" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-card-icon-large" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : isVideo(file) ? 'video' : isAudio(file) ? 'audio' : isDoc(file) ? 'doc' : 'file'}`" @click="!file.is_folder && isImage(file) ? openLightbox(file) : null">
                  <template v-if="file.is_folder">
                    <Folder :size="48" />
                  </template>
                  <template v-else-if="isImage(file)">
                    <img :src="thumbnailUrl(file.id)" class="file-thumb" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
                  </template>
                  <template v-else>
                    <Film v-if="isVideo(file)" :size="48" />
                    <Music v-else-if="isAudio(file)" :size="48" />
                    <FileText v-else-if="isDoc(file)" :size="48" />
                    <File v-else :size="48" />
                  </template>
                  <button class="card-menu-btn" @click.stop="openContextMenu($event, file)" title="More">
                    <MoreVertical :size="14" />
                  </button>
                </div>
                <span class="file-card-name" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
              </div>
            </div>

            <!-- Medium icons -->
            <div v-else-if="viewMode === 'medium'" class="file-grid-medium">
              <div v-for="file in files" :key="file.id" class="file-card-medium" @dblclick="file.is_folder ? navigateToFolder(file.id) : null">
                <div class="file-card-icon-medium" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : isVideo(file) ? 'video' : isAudio(file) ? 'audio' : isDoc(file) ? 'doc' : 'file'}`" @click="!file.is_folder && isImage(file) ? openLightbox(file) : null">
                  <template v-if="file.is_folder">
                    <Folder :size="32" />
                  </template>
                  <template v-else-if="isImage(file)">
                    <img :src="thumbnailUrl(file.id)" class="file-thumb-sm" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
                  </template>
                  <template v-else>
                    <Film v-if="isVideo(file)" :size="32" />
                    <Music v-else-if="isAudio(file)" :size="32" />
                    <FileText v-else-if="isDoc(file)" :size="32" />
                    <File v-else :size="32" />
                  </template>
                  <button class="card-menu-btn" @click.stop="openContextMenu($event, file)" title="More">
                    <MoreVertical :size="12" />
                  </button>
                </div>
                <span class="file-card-name-sm" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? navigateToFolder(file.id) : null">{{ file.name }}</span>
              </div>
            </div>

            <!-- Small icons -->
            <div v-else-if="viewMode === 'small'" class="file-grid-small">
              <div v-for="file in files" :key="file.id" class="file-card-small" @click="isImage(file) ? openLightbox(file) : (file.is_folder ? navigateToFolder(file.id) : null)">
                <div class="file-card-icon-sm" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : 'file'}`">
                  <template v-if="file.is_folder">
                    <Folder :size="14" />
                  </template>
                  <template v-else-if="isImage(file)">
                    <img :src="thumbnailUrl(file.id)" class="file-thumb-sm" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
                  </template>
                  <template v-else>
                    <File :size="14" />
                  </template>
                </div>
                <span class="file-card-name-sm" :class="{ 'folder-name': file.is_folder }">{{ file.name }}</span>
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

    <Transition name="modal">
      <div v-if="showShareDialog && shareTarget" class="modal-overlay" @click.self="closeShareDialog">
        <div class="share-dialog">
          <div class="share-dialog-header">
            <Share2 :size="18" style="color: var(--color-brand-500);" />
            <h3>Share "{{ shareTarget.name }}"</h3>
            <button class="icon-btn" @click="closeShareDialog"><X :size="16" /></button>
          </div>

          <div class="share-create">
            <select v-model="shareExpiry" class="share-select">
              <option value="1h">Expires in 1 hour</option>
              <option value="24h">Expires in 24 hours</option>
              <option value="7d">Expires in 7 days</option>
              <option value="30d">Expires in 30 days</option>
              <option value="never">Never expires</option>
            </select>
            <button class="btn-primary" @click="createShareLink" :disabled="shareLoading" style="height: 2.25rem;">
              <Loader2 v-if="shareLoading" :size="14" class="spin" />
              <span v-else>Create Link</span>
            </button>
          </div>

          <div v-if="shareLinks.length" class="share-links">
            <div v-for="link in shareLinks" :key="link.id" class="share-link-item">
              <div class="share-link-info">
                <div class="share-link-url">{{ link.url }}</div>
                <div class="share-link-meta">
                  {{ link.expires_at ? `Expires ${formatDate(link.expires_at)}` : 'Never expires' }}
                  <span v-if="!link.is_valid" style="color: var(--color-danger);"> · Expired</span>
                </div>
              </div>
              <div class="share-link-actions">
                <button class="icon-btn" @click="copyShareLink(link.url, link.id)" :title="copiedLinkId === link.id ? 'Copied!' : 'Copy link'">
                  <Check v-if="copiedLinkId === link.id" :size="14" style="color: var(--color-success);" />
                  <Copy v-else :size="14" />
                </button>
                <button class="icon-btn" @click="revokeShareLink(link.id)" title="Revoke">
                  <Trash2 :size="14" />
                </button>
              </div>
            </div>
          </div>
          <div v-else class="share-empty">
            <p>No share links yet. Create one above.</p>
          </div>
        </div>
      </div>
    </Transition>

    <Teleport to="body">
      <div v-if="contextMenu.show" class="context-menu-overlay" @click="closeContextMenu" @contextmenu.prevent="closeContextMenu">
        <div class="context-menu" :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }">
          <button v-if="!contextMenu.file?.is_folder" class="context-menu-item" @click="contextAction('download')">
            <Download :size="14" /> <span>Download</span>
          </button>
          <button class="context-menu-item" @click="contextAction('share')">
            <Share2 :size="14" /> <span>Share</span>
          </button>
          <div class="context-menu-divider"></div>
          <button class="context-menu-item danger" @click="contextAction('delete')">
            <Trash2 :size="14" /> <span>Delete</span>
          </button>
        </div>
      </div>
    </Teleport>

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
            <button class="lightbox-btn" @click="downloadFile(lightboxFile)" title="Download"><Download :size="18" /></button>
          </div>
        </div>
        <div class="lightbox-body" @click="lightboxBodyClick" @mousedown="startPan" @mousemove="onPan" @mouseup="stopPan" @mouseleave="stopPan" @touchstart.passive="startPan" @touchmove="onPan" @touchend="stopPan">
          <button v-if="hasMultipleImages" class="lightbox-nav lightbox-nav-prev" @click.stop="prevImage"><ChevronLeft :size="24" /></button>
          <img :src="thumbnailUrl(lightboxFile.id)" class="lightbox-img" :style="lightboxImageStyle" @click.stop @wheel.prevent="handleZoom" draggable="false" />
          <button v-if="hasMultipleImages" class="lightbox-nav lightbox-nav-next" @click.stop="nextImage"><ChevronRight :size="24" /></button>
        </div>
        <div class="lightbox-footer">
          <button class="lightbox-zoom-btn" @click="lightboxZoom = Math.max(0.25, lightboxZoom - 0.25)"><Minus :size="18" /></button>
          <button class="lightbox-zoom-btn" @click="resetView"><Search :size="18" /></button>
          <button class="lightbox-zoom-btn" @click="lightboxZoom = Math.min(4, lightboxZoom + 0.25)"><Plus :size="18" /></button>
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
import { FolderOpen, FolderPlus, Upload, Folder, File, Trash2, Download, ChevronRight, Home, Loader2, X, AlertTriangle, LayoutGrid, List, LayoutList, Grip, Image, Film, Music, FileText, Minus, Plus, Search, Share2, Copy, Check, MoreVertical } from 'lucide-vue-next'

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
const lightboxFile = ref<any>(null)
const lightboxZoom = ref(1)
const lightboxPanX = ref(0)
const lightboxPanY = ref(0)
const lightboxIndex = ref(0)
const isPanning = ref(false)
const didPan = ref(false)
const panStart = ref({ x: 0, y: 0 })
const showShareDialog = ref(false)
const shareTarget = ref<any>(null)
const shareLinks = ref<any[]>([])
const shareExpiry = ref('24h')
const shareLoading = ref(false)
const copiedLinkId = ref<number | null>(null)
const contextMenu = ref<{ show: boolean; file: any; x: number; y: number }>({ show: false, file: null, x: 0, y: 0 })
const sidebarOpen = ref(false)

provide('sidebarOpen', sidebarOpen)

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

const openLightbox = (file: any) => {
  if (!isImage(file)) return
  lightboxIndex.value = files.value.findIndex((f: any) => f.id === file.id)
  lightboxFile.value = files.value[lightboxIndex.value]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const closeLightbox = () => {
  if (isPanning.value) return
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

const startPan = (e: MouseEvent | TouchEvent) => {
  if (lightboxZoom.value <= 1) return
  isPanning.value = true
  didPan.value = false
  const pt = 'touches' in e ? e.touches[0] : e
  panStart.value = { x: pt.clientX - lightboxPanX.value, y: pt.clientY - lightboxPanY.value }
}

const onPan = (e: MouseEvent | TouchEvent) => {
  if (!isPanning.value) return
  e.preventDefault()
  const pt = 'touches' in e ? e.touches[0] : e
  lightboxPanX.value = pt.clientX - panStart.value.x
  lightboxPanY.value = pt.clientY - panStart.value.y
  didPan.value = true
}

const stopPan = () => {
  isPanning.value = false
  setTimeout(() => { didPan.value = false }, 100)
}

const lightboxBodyClick = () => {
  if (!didPan.value) closeLightbox()
}

const lightboxImageStyle = computed(() => {
  return {
    transform: `translate(${lightboxPanX.value}px, ${lightboxPanY.value}px) scale(${lightboxZoom.value})`,
    cursor: lightboxZoom.value > 1 ? (isPanning.value ? 'grabbing' : 'grab') : 'default'
  }
})

const nextImage = () => {
  const images = files.value.filter((f: any) => isImage(f))
  if (images.length <= 1) return
  const currentIdx = images.findIndex((f: any) => f.id === lightboxFile.value?.id)
  const nextIdx = (currentIdx + 1) % images.length
  lightboxIndex.value = files.value.indexOf(images[nextIdx])
  lightboxFile.value = images[nextIdx]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const prevImage = () => {
  const images = files.value.filter((f: any) => isImage(f))
  if (images.length <= 1) return
  const currentIdx = images.findIndex((f: any) => f.id === lightboxFile.value?.id)
  const prevIdx = (currentIdx - 1 + images.length) % images.length
  lightboxIndex.value = files.value.indexOf(images[prevIdx])
  lightboxFile.value = images[prevIdx]
  lightboxZoom.value = 1
  lightboxPanX.value = 0
  lightboxPanY.value = 0
}

const hasMultipleImages = computed(() => {
  return files.value.filter((f: any) => isImage(f)).length > 1
})

const handleLightboxKeydown = (e: KeyboardEvent) => {
  if (!lightboxFile.value) return
  if (e.key === 'Escape') closeLightbox()
  else if (e.key === 'ArrowRight') nextImage()
  else if (e.key === 'ArrowLeft') prevImage()
}

const openShareDialog = async (file: any) => {
  shareTarget.value = file
  shareExpiry.value = '24h'
  shareLinks.value = []
  showShareDialog.value = true
  try {
    const links = await apiFetch(`/api/files/${file.id}/shares`) as any[]
    shareLinks.value = links
  } catch {}
}

const createShareLink = async () => {
  if (!shareTarget.value) return
  shareLoading.value = true
  try {
    const resp = await apiFetch(`/api/files/${shareTarget.value.id}/share`, {
      method: 'POST',
      body: { expires_in: shareExpiry.value }
    }) as any
    shareLinks.value.unshift(resp)
  } catch {}
  shareLoading.value = false
}

const copyShareLink = async (url: string, linkId: number) => {
  try {
    await navigator.clipboard.writeText(url)
    copiedLinkId.value = linkId
    setTimeout(() => { copiedLinkId.value = null }, 2000)
  } catch {}
}

const revokeShareLink = async (linkId: number) => {
  if (!shareTarget.value) return
  try {
    await apiFetch(`/api/files/${shareTarget.value.id}/share/${linkId}`, { method: 'DELETE' })
    shareLinks.value = shareLinks.value.filter(l => l.id !== linkId)
  } catch {}
}

const closeShareDialog = () => {
  showShareDialog.value = false
  shareTarget.value = null
  shareLinks.value = []
}

const openContextMenu = (e: MouseEvent, file: any) => {
  e.preventDefault()
  e.stopPropagation()
  if (window.innerWidth <= 768) {
    contextMenu.value = { show: true, file, x: 0, y: 0 }
  } else {
    const menuWidth = 170
    const menuHeight = 140
    let x = e.clientX
    let y = e.clientY
    if (x + menuWidth > window.innerWidth) x = window.innerWidth - menuWidth - 8
    if (y + menuHeight > window.innerHeight) y = window.innerHeight - menuHeight - 8
    contextMenu.value = { show: true, file, x, y }
  }
}

const closeContextMenu = () => {
  contextMenu.value.show = false
  contextMenu.value.file = null
}

const contextAction = (action: string) => {
  const file = contextMenu.value.file
  closeContextMenu()
  if (!file) return
  if (action === 'download' && !file.is_folder) downloadFile(file)
  else if (action === 'delete') confirmDelete(file)
  else if (action === 'share') openShareDialog(file)
}

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const videoExtensions = ['mp4','avi','mkv','mov','wmv','flv','webm']
const audioExtensions = ['mp3','wav','ogg','flac','aac','m4a']
const docExtensions = ['pdf','doc','docx','xls','xlsx','ppt','pptx','txt','csv']

const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''

const thumbnailUrl = (fileId: number) => {
  const token = import.meta.client ? localStorage.getItem('token') || '' : ''
  const base = useRuntimeConfig().public.apiBase
  return `${base}/api/files/${fileId}/thumbnail?token=${token}`
}

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
  if (import.meta.client) {
    window.addEventListener('keydown', handleLightboxKeydown)
  }
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
  max-width: 100vw;
  overflow-x: hidden;
}

@media (max-width: 768px) {
  .app-main {
    margin-left: 0;
    padding: 0.75rem;
    overflow-x: hidden;
  }
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

  .action-toolbar .btn-secondary,
  .action-toolbar .upload-btn {
    flex: 1;
  }
}

@media (max-width: 480px) {
  .action-toolbar {
    flex-wrap: nowrap;
    gap: 0.375rem;
  }

  .action-toolbar .view-toggle-wrapper {
    flex: 0 0 auto;
  }

  .action-toolbar .btn-secondary {
    flex: 1;
    height: 2.25rem;
    font-size: 0.8125rem;
  }

  .action-toolbar .upload-btn {
    flex: 1;
    height: 2.25rem;
    font-size: 0.8125rem;
  }
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
  min-width: 0;
  overflow: hidden;
  flex: 1;
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
  max-width: 120px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.breadcrumb-sep {
  color: var(--color-text-muted);
}

@media (max-width: 480px) {
  .breadcrumb {
    font-size: 0.8125rem;
  }

  .breadcrumb-item span {
    max-width: 100px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
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
  width: 100%;
  overflow-x: hidden;
}

.file-row {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-surface-2);
  transition: background-color 0.1s ease;
  max-width: 100%;
  overflow: hidden;
}

.file-row:hover {
  background-color: var(--color-surface-1);
}

.file-col-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 0.625rem;
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

@media (max-width: 768px) {
  .file-list-header {
    display: none;
  }

  .file-row {
    flex-wrap: wrap;
    gap: 0.125rem 0.5rem;
    padding: 0.75rem;
  }

  .file-col-name {
    flex: 1 1 50%;
    min-width: 0;
  }

  .file-col-size {
    flex: 0 0 auto;
    font-size: 0.75rem;
    margin-left: auto;
  }

  .file-col-date {
    width: 100%;
    text-align: left;
    font-size: 0.6875rem;
    margin-top: -0.125rem;
  }

  .file-col-actions {
    opacity: 1;
  }
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

@media (max-width: 768px) {
  .file-row {
    padding: 0.875rem 0.75rem;
  }

  .file-col-actions {
    opacity: 1;
  }
}

.file-icon {
  width: 2rem;
  height: 2rem;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  overflow: hidden;
}

.file-thumb-detail {
  width: 100%;
  height: 100%;
  object-fit: cover;
  border-radius: 0.375rem;
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
  max-width: 100%;
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

@media (max-width: 768px) {
  .view-menu {
    right: auto;
    left: 0;
  }
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
  grid-template-columns: repeat(auto-fill, minmax(160px, 1fr));
  gap: 1rem;
}

@media (max-width: 768px) {
  .file-grid-large {
    grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
    gap: 0.75rem;
  }
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
  width: 8rem;
  height: 8rem;
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.625rem;
  overflow: hidden;
  cursor: pointer;
  position: relative;
}

.card-menu-btn {
  position: absolute;
  top: 0.25rem;
  right: 0.25rem;
  background-color: rgba(255, 255, 255, 0.85);
  border: none;
  color: var(--color-text-secondary);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0;
  transition: opacity 0.1s ease;
  z-index: 2;
}

.file-card-large:hover .card-menu-btn,
.file-card-medium:hover .card-menu-btn {
  opacity: 1;
}

.card-menu-btn:hover {
  background-color: rgba(255, 255, 255, 1);
  color: var(--color-text-primary);
}

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
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 0.75rem;
}

@media (max-width: 768px) {
  .file-grid-medium {
    grid-template-columns: repeat(auto-fill, minmax(90px, 1fr));
  }
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
  width: 5rem;
  height: 5rem;
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 0.5rem;
  overflow: hidden;
  cursor: pointer;
  position: relative;
}

.file-thumb-sm {
  width: 100%;
  height: 100%;
  object-fit: cover;
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
  cursor: pointer;
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
  overflow: hidden;
}

.file-card-icon-sm img {
  width: 100%;
  height: 100%;
  object-fit: cover;
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

.share-dialog {
  background-color: var(--color-surface-0);
  border-radius: 0.75rem;
  padding: 0;
  width: 480px;
  max-width: 90vw;
  box-shadow: 0 16px 48px rgba(0, 0, 0, 0.16);
}

.share-dialog-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.share-dialog-header h3 {
  flex: 1;
  font-size: 0.9375rem;
  font-weight: 600;
  color: var(--color-text-primary);
  margin: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.share-create {
  display: flex;
  gap: 0.5rem;
  padding: 1rem 1.25rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.share-select {
  flex: 1;
  padding: 0.5rem 0.75rem;
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  color: var(--color-text-primary);
  background-color: var(--color-surface-0);
  cursor: pointer;
}

.share-links {
  max-height: 240px;
  overflow-y: auto;
}

.share-link-item {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--color-surface-2);
}

.share-link-item:last-child {
  border-bottom: none;
}

.share-link-info {
  flex: 1;
  min-width: 0;
}

.share-link-url {
  font-size: 0.75rem;
  color: var(--color-brand-600);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  word-break: break-all;
}

.share-link-meta {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.share-link-actions {
  display: flex;
  gap: 0.25rem;
  flex-shrink: 0;
}

.share-empty {
  padding: 1.5rem 1.25rem;
  text-align: center;
}

.share-empty p {
  font-size: 0.8125rem;
  color: var(--color-text-muted);
  margin: 0;
}

.context-menu-overlay {
  position: fixed;
  inset: 0;
  z-index: 500;
}

.context-menu {
  position: fixed;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 160px;
  padding: 0.25rem 0;
  z-index: 501;
}

.context-menu-item {
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

.context-menu-item:hover {
  background-color: var(--color-surface-1);
}

.context-menu-item.danger {
  color: var(--color-danger);
}

.context-menu-item.danger:hover {
  background-color: rgba(250, 82, 82, 0.08);
}

.context-menu-divider {
  height: 1px;
  background-color: var(--color-surface-2);
  margin: 0.25rem 0;
}

@media (max-width: 768px) {
  .context-menu-overlay {
    display: flex;
    align-items: flex-end;
    justify-content: stretch;
  }

  .context-menu {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    top: auto;
    border-radius: 1rem 1rem 0 0;
    min-width: unset;
    width: 100%;
    padding-bottom: env(safe-area-inset-bottom, 1rem);
  }

  .context-menu-item {
    padding: 0.75rem 1rem;
    min-height: 3rem;
    font-size: 0.9375rem;
  }
}

@media (max-width: 480px) {
  .modal-overlay {
    align-items: flex-end;
  }

  .modal-card {
    width: calc(100vw - 2rem);
    border-radius: 1rem 1rem 0 0;
    padding-bottom: env(safe-area-inset-bottom, 1rem);
    animation: slideUp 0.25s ease;
  }

  .share-dialog {
    width: calc(100vw - 2rem);
    border-radius: 1rem 1rem 0 0;
    padding-bottom: env(safe-area-inset-bottom, 1rem);
    animation: slideUp 0.25s ease;
  }
}

@keyframes slideUp {
  from { transform: translateY(100%); }
  to { transform: translateY(0); }
}

.drop-zone {
  min-height: calc(100vh - 8rem);
  border: 2px dashed transparent;
  border-radius: 0.75rem;
  transition: border-color 0.15s ease, background-color 0.15s ease;
  position: relative;
  overflow-x: hidden;
  max-width: 100%;
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

@media (max-width: 768px) {
  .upload-panel,
  .download-panel {
    width: calc(100vw - 2rem);
    right: 1rem;
  }

  .panels-container {
    right: 0;
    left: 0;
    justify-content: center;
  }
}

@media (max-width: 480px) {
  .upload-panel,
  .download-panel {
    width: calc(100vw - 1.5rem);
    left: 0.75rem;
    right: 0.75rem;
    bottom: 5rem;
  }
}

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

.lightbox-close:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

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
  background-color: rgba(239, 68, 68, 0.15);
  color: #ef4444;
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

.lightbox-actions {
  display: flex;
  gap: 0.25rem;
}

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
}

.lightbox-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
}

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

@media (max-width: 768px) {
  .lightbox-img {
    max-width: 95vw;
    max-height: 75vh;
  }

  .lightbox-nav {
    width: 2.5rem;
    height: 2.5rem;
  }

  .lightbox-nav-prev { left: 0.25rem; }
  .lightbox-nav-next { right: 0.25rem; }
}

@media (max-width: 480px) {
  .lightbox-img {
    max-width: 100vw;
    max-height: 70vh;
  }

  .lightbox-nav {
    width: 2.5rem;
    height: 2.5rem;
  }

  .lightbox-zoom-btn {
    min-width: 44px;
    min-height: 44px;
    padding: 0.75rem;
  }
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

.lightbox-zoom-btn:hover {
  background-color: rgba(255, 255, 255, 0.1);
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
