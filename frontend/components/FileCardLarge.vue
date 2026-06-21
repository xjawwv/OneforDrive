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
        <img :src="thumbnailUrl(file.id)" class="file-thumb" @error="(e: any) => e.target.style.display='none'" crossorigin="anonymous" />
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
