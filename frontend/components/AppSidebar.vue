<template>
  <div>
    <div v-if="sidebarOpen" class="sidebar-backdrop" @click="sidebarOpen = false"></div>
    <aside class="sidebar" :class="{ 'sidebar-open': sidebarOpen }">
      <div class="sidebar-header">
        <div class="sidebar-logo">
          <HardDrive :size="20" color="white" :stroke-width="2" />
        </div>
        <span class="sidebar-brand">RouteStorage</span>
        <button class="sidebar-close" @click="sidebarOpen = false"><X :size="18" /></button>
      </div>

      <nav class="sidebar-nav">
        <div class="nav-section">
          <div class="nav-label">FILES</div>
          <NuxtLink to="/explorer" class="sidebar-link" :class="{ active: current === 'explorer' }" @click="sidebarOpen = false">
            <FolderOpen :size="18" />
            <span>Explorer</span>
          </NuxtLink>
        </div>
        <div class="nav-section">
          <div class="nav-label">ACCOUNT</div>
          <NuxtLink to="/settings" class="sidebar-link" :class="{ active: current === 'settings' }" @click="sidebarOpen = false">
            <Settings :size="18" />
            <span>Settings</span>
          </NuxtLink>
        </div>
        <div v-if="hasAdminAccess" class="nav-section">
          <div class="nav-label">ADMIN</div>
          <NuxtLink to="/admin/roles" class="sidebar-link" :class="{ active: current === 'roles' }" @click="sidebarOpen = false">
            <ShieldCheck :size="18" />
            <span>Role Management</span>
          </NuxtLink>
          <NuxtLink to="/admin/users" class="sidebar-link" :class="{ active: current === 'users' }" @click="sidebarOpen = false">
            <Users :size="18" />
            <span>User Management</span>
          </NuxtLink>
        </div>
      </nav>
    </aside>
  </div>
</template>

<script setup lang="ts">
import { HardDrive, FolderOpen, Settings, ShieldCheck, Users, Menu, X } from 'lucide-vue-next'

const props = defineProps<{ current: string }>()

const sidebarOpen = inject<Ref<boolean>>('sidebarOpen', ref(false))
const { can, fetchPermissions } = usePermissions()
const hasAdminAccess = computed(() => can('users.manage'))

onMounted(() => {
  if (import.meta.client && localStorage.getItem('token')) {
    fetchPermissions()
  }
})
</script>

<style scoped>
.mobile-menu-btn {
  display: none;
  position: fixed;
  top: 1rem;
  left: 1rem;
  z-index: 30;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  padding: 0.5rem;
  cursor: pointer;
  color: var(--color-text-secondary);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

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

@media (max-width: 768px) {
  .sidebar {
    transform: translateX(-100%);
    transition: transform 0.25s ease;
    z-index: 101;
  }

  .sidebar.sidebar-open {
    transform: translateX(0);
    box-shadow: 4px 0 16px rgba(0, 0, 0, 0.15);
  }
}

.sidebar-backdrop {
  display: none;
}

@media (max-width: 768px) {
  .sidebar-backdrop {
    display: block;
    position: fixed;
    inset: 0;
    background-color: rgba(0, 0, 0, 0.4);
    z-index: 100;
  }
}

.sidebar-header {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 1.25rem 1.25rem 1rem 1.25rem;
}

.sidebar-close {
  display: none;
  margin-left: auto;
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  border-radius: 0.25rem;
}

@media (max-width: 768px) {
  .sidebar-close {
    display: flex;
  }
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
}

.nav-section {
  margin-bottom: 0.75rem;
}

.nav-label {
  font-size: 0.625rem;
  font-weight: 600;
  color: var(--color-text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 0.5rem 0.75rem 0.25rem 0.75rem;
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

.sidebar-enter-active { transition: opacity 0.2s ease; }
.sidebar-leave-active { transition: opacity 0.15s ease; }
.sidebar-enter-from, .sidebar-leave-to { opacity: 0; }
.sidebar-slide-enter-active { transition: transform 0.25s ease; }
.sidebar-slide-leave-active { transition: transform 0.2s ease; }
.sidebar-slide-enter-from, .sidebar-slide-leave-to { transform: translateX(-100%); }
</style>
