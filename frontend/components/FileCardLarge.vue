<script setup lang="ts">
import { Folder, File, Film, Music, FileText, MoreVertical, CheckSquare, Square } from 'lucide-vue-next'

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
const videoExtensions = ['mp4','avi','mkv','mov','wmv','flv','webm']
const audioExtensions = ['mp3','wav','ogg','flac','aac','m4a']
const docExtensions = ['pdf','doc','docx','xls','xlsx','ppt','pptx','txt','csv']
const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''
const isImage = (f: any) => !f.is_folder && imageExtensions.includes(getFileExt(f.name))
const isVideo = (f: any) => !f.is_folder && videoExtensions.includes(getFileExt(f.name))
const isAudio = (f: any) => !f.is_folder && audioExtensions.includes(getFileExt(f.name))
const isDoc = (f: any) => !f.is_folder && docExtensions.includes(getFileExt(f.name))

const thumbnailUrl = (fileId: number) => {
  const token = import.meta.client ? localStorage.getItem('token') || '' : ''
  const base = useRuntimeConfig().public.apiBase
  return `${base}/api/files/${fileId}/thumbnail?token=${token}`
}
</script>

<template>
  <div class="file-card-large" :class="{ 'file-selected': selected }"
    @dblclick="selectMode ? emit('toggleSelect', file.id) : emit('open', file)">
    <div v-if="selectMode" class="file-card-check" @click.stop="emit('toggleSelect', file.id)">
      <component :is="selected ? CheckSquare : Square" :size="16" class="select-icon" />
    </div>
    <div class="file-card-icon-large" :class="file.is_folder ? 'file-icon-folder' : `file-type-${isImage(file) ? 'image' : isVideo(file) ? 'video' : isAudio(file) ? 'audio' : isDoc(file) ? 'doc' : 'file'}`"
      @click="selectMode ? emit('toggleSelect', file.id) : (!file.is_folder && isImage(file) ? emit('open', file) : null)">
      <template v-if="file.is_folder">
        <Folder :size="48" />
      </template>
      <template v-else-if="isImage(file)">
        <img :src="thumbnailUrl(file.id)" loading="lazy" class="file-thumb" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
      </template>
      <template v-else>
        <Film v-if="isVideo(file)" :size="48" />
        <Music v-else-if="isAudio(file)" :size="48" />
        <FileText v-else-if="isDoc(file)" :size="48" />
        <File v-else :size="48" />
      </template>
      <button class="card-menu-btn" @click.stop="emit('contextMenu', $event, file)" title="More">
        <MoreVertical :size="14" />
      </button>
    </div>
    <span class="file-card-name" :class="{ 'folder-name': file.is_folder }" @click="file.is_folder ? emit('open', file) : null">{{ file.name }}</span>
  </div>
</template>

<style scoped>
.file-card-large {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 1rem;
  border-radius: 0.5rem;
  cursor: default;
  transition: background-color 0.1s ease;
  position: relative;
}

.file-card-large:hover {
  background-color: var(--color-surface-1);
}

.file-card-large.file-selected {
  outline: 2px solid var(--color-brand-500);
  outline-offset: -2px;
}

.file-card-check {
  position: absolute;
  top: 0.5rem;
  left: 0.5rem;
  z-index: 2;
  cursor: pointer;
  background: var(--color-surface-0);
  border-radius: 0.25rem;
  padding: 0.125rem;
}

.select-icon {
  color: var(--color-text-muted);
  cursor: pointer;
  transition: color 0.1s ease;
}

.select-icon:hover {
  color: var(--color-brand-500);
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

.file-icon-folder {
  background-color: rgba(76, 110, 245, 0.1);
  color: var(--color-brand-600);
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

.file-type-file {
  background-color: var(--color-surface-2);
  color: var(--color-text-muted);
}

.file-thumb {
  width: 100%;
  height: 100%;
  object-fit: cover;
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

.file-card-large:hover .card-menu-btn {
  opacity: 1;
}

.card-menu-btn:hover {
  background-color: rgba(255, 255, 255, 1);
  color: var(--color-text-primary);
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

.folder-name {
  cursor: pointer;
  font-weight: 500;
}

.folder-name:hover {
  color: var(--color-brand-600);
}
</style>
