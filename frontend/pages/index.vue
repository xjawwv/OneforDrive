<template>
  <div />
</template>

<script setup lang="ts">
definePageMeta({ layout: false })

if (import.meta.client) {
  const token = localStorage.getItem('token')
  if (token) {
    navigateTo('/explorer')
  } else {
    navigateTo('/login')
  }
} else {
  navigateTo('/login')
}
</script>
