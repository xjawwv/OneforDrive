<template>
  <aside class="sidebar">
    <div class="sidebar-header">
      <div class="sidebar-logo">
        <HardDrive :size="20" color="white" :stroke-width="2" />
      </div>
      <span class="sidebar-brand">RouteStorage</span>
    </div>

    <nav class="sidebar-nav">
      <NuxtLink to="/explorer" class="sidebar-link" :class="{ active: current === 'explorer' }">
        <FolderOpen :size="18" />
        <span>Explorer</span>
      </NuxtLink>
      <NuxtLink to="/settings" class="sidebar-link" :class="{ active: current === 'settings' }">
        <Settings :size="18" />
        <span>Settings</span>
      </NuxtLink>
    </nav>

    <div class="sidebar-footer">
      <div class="sidebar-user">
        <div class="sidebar-avatar">
          <User :size="16" />
        </div>
        <span class="sidebar-user-name">{{ userName }}</span>
      </div>
      <button class="sidebar-logout" @click="logout" title="Sign out">
        <LogOut :size="16" />
      </button>
    </div>
  </aside>
</template>

<script setup lang="ts">
import { HardDrive, FolderOpen, Settings, LogOut, User } from 'lucide-vue-next'

const props = defineProps<{ current: string }>()

const userName = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).name } catch { return 'User' }
    }
  }
  return 'User'
})

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  navigateTo('/login')
}
</script>

<style scoped>
.sidebar {
  width: 240px;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  background-color: var(--color-surface-0);
  border-right: 1px solid var(--color-surface-2);
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 1.25rem 1.25rem 1rem 1.25rem;
}

.sidebar-logo {
  width: 2rem;
  height: 2rem;
  background-color: var(--color-brand-600);
  border-radius: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sidebar-brand {
  font-weight: 700;
  font-size: 1rem;
  color: var(--color-text-primary);
  letter-spacing: -0.015em;
}

.sidebar-nav {
  flex: 1;
  padding: 0.5rem 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.25rem;
}

.sidebar-link {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
  text-decoration: none;
  transition: background-color 0.12s ease, color 0.12s ease;
}

.sidebar-link:hover {
  background-color: var(--color-surface-1);
  color: var(--color-text-primary);
}

.sidebar-link.active {
  background-color: var(--color-brand-50);
  color: var(--color-brand-700);
}

.sidebar-footer {
  padding: 0.75rem 1.25rem;
  border-top: 1px solid var(--color-surface-2);
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.sidebar-user {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  min-width: 0;
}

.sidebar-avatar {
  width: 1.75rem;
  height: 1.75rem;
  border-radius: 9999px;
  background-color: var(--color-surface-2);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--color-text-muted);
  flex-shrink: 0;
}

.sidebar-user-name {
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sidebar-logout {
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.375rem;
  border-radius: 0.375rem;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  transition: color 0.12s ease, background-color 0.12s ease;
}

.sidebar-logout:hover {
  color: var(--color-danger);
  background-color: rgba(250, 82, 82, 0.08);
}
</style>
