<script setup lang="ts">
import { Folder, File, MoreVertical, CheckSquare, Square } from 'lucide-vue-next'

defineProps<{
  file: any
  selected: boolean
  selectMode: boolean
}>()

const emit = defineEmits<{
  toggleSelect: [fileId: number]
  open: [file: any]
  contextMenu: [event: MouseEvent, file: any]
}>()

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''
const isImage = (f: any) => !f.is_folder && imageExtensions.includes(getFileExt(f.name))

const thumbnailUrl = (fileId: number) => {
  const token = import.meta.client ? localStorage.getItem('token') || '' : ''
  const base = useRuntimeConfig().public.apiBase
  return `${base}/api/files/${fileId}/thumbnail?token=${token}`
}

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
  return d.toLocaleDateString()
}
</script>

<template>
  <div class="file-row" :class="{ 'file-selected': selected }"
    @click="selectMode ? emit('toggleSelect', file.id) : emit('open', file)">
    <div v-if="selectMode" class="file-col-check" @click.stop="emit('toggleSelect', file.id)">
      <component :is="selected ? CheckSquare : Square" :size="16" class="select-icon" />
    </div>
    <div class="file-col-name">
      <div class="file-icon" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : 'file'}`">
        <template v-if="file.is_folder"><Folder :size="16" /></template>
        <template v-else-if="isImage(file)"><img :src="thumbnailUrl(file.id)" class="file-thumb-xs" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" /></template>
        <template v-else><File :size="16" /></template>
      </div>
      <div class="file-name-block">
        <span class="file-name" :class="{ 'folder-name': file.is_folder }">{{ file.name }}</span>
        <span class="file-date">{{ formatDate(file.updated_at) }}</span>
      </div>
    </div>
    <span class="file-col-size">{{ file.is_folder ? '--' : formatSize(file.size_total) }}</span>
    <div class="file-col-actions">
      <button class="icon-btn" @click.stop="emit('contextMenu', $event, file)" title="More">
        <MoreVertical :size="14" />
      </button>
    </div>
  </div>
</template>

<style scoped>
.file-row {
  display: flex;
  align-items: center;
  padding: 0.625rem 0.75rem;
  border-bottom: 1px solid var(--color-surface-2);
  gap: 0.5rem;
  min-width: 0;
  overflow: hidden;
  transition: background-color 0.1s ease;
}

.file-row:hover {
  background-color: var(--color-surface-1);
}

.file-row.file-selected {
  background-color: color-mix(in srgb, var(--color-brand-500) 10%, transparent);
}

.file-col-check {
  flex-shrink: 0;
  width: 1.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.select-icon {
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color 0.1s ease;
}

.select-icon:hover {
  color: var(--color-brand-500);
}

.file-col-name {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  display: flex;
  align-items: center;
  gap: 0.625rem;
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

.file-thumb-xs {
  width: 32px;
  height: 32px;
  object-fit: cover;
  border-radius: 4px;
}

.file-name-block {
  display: flex;
  flex-direction: column;
  min-width: 0;
  flex: 1;
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

.file-date {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
  white-space: nowrap;
}

.file-col-size {
  font-size: 0.75rem;
  color: var(--color-text-muted);
  white-space: nowrap;
  flex-shrink: 0;
  text-align: right;
  min-width: 56px;
}

.file-col-actions {
  opacity: 0;
  display: flex;
  align-items: center;
  flex-shrink: 0;
  width: auto;
  padding-left: 0.25rem;
  transition: opacity 0.1s ease;
}

.file-row:hover .file-col-actions {
  opacity: 1;
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

@media (max-width: 768px) {
  .file-row {
    padding: 0.875rem 0.75rem;
  }

  .file-col-actions {
    opacity: 1;
  }
}
</style>
