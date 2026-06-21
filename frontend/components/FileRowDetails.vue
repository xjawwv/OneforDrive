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
