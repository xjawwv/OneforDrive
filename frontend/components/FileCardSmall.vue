<script setup lang="ts">
import { Folder, File, CheckSquare, Square } from 'lucide-vue-next'

defineProps<{
  file: any
  selected: boolean
  selectMode: boolean
}>()

const emit = defineEmits<{
  toggleSelect: [fileId: number]
  open: [file: any]
}>()

const imageExtensions = ['jpg','jpeg','png','gif','webp','bmp','svg','ico']
const getFileExt = (name: string) => name.split('.').pop()?.toLowerCase() || ''
const isImage = (f: any) => !f.is_folder && imageExtensions.includes(getFileExt(f.name))

const thumbnailUrl = (fileId: number) => {
  const token = import.meta.client ? localStorage.getItem('token') || '' : ''
  const base = useRuntimeConfig().public.apiBase
  return `${base}/api/files/${fileId}/thumbnail?token=${token}`
}
</script>

<template>
  <div class="file-card-small" :class="{ 'file-selected': selected }"
    @click="selectMode ? emit('toggleSelect', file.id) : emit('open', file)">
    <div v-if="selectMode" class="file-card-check-sm" @click.stop="emit('toggleSelect', file.id)">
      <component :is="selected ? CheckSquare : Square" :size="12" class="select-icon" />
    </div>
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
</template>
