<template>
  <div>
    <header class="top-bar">
      <button v-if="showHamburger" class="hamburger-btn" @click="sidebarOpen = true">
        <Menu :size="20" />
      </button>
      <div class="top-bar-title">
        <slot name="title">
          <h1 class="page-title">{{ title }}</h1>
          <p class="page-subtitle">{{ subtitle }}</p>
        </slot>
      </div>
      <slot name="actions"></slot>
      <div class="user-menu-wrapper">
        <button class="avatar-btn" @click="showUserMenu = !showUserMenu">
          <div class="avatar-circle">{{ userInitial }}</div>
          <div class="notification-dot"></div>
        </button>
        <Transition name="menu">
          <div v-if="showUserMenu" class="user-dropdown">
            <div class="dropdown-user-info">
              <div class="dropdown-user-name">{{ userName }}</div>
              <div class="dropdown-user-email">{{ userEmail }}</div>
            </div>
            <div class="dropdown-divider"></div>
            <NuxtLink v-if="currentPage !== 'settings'" to="/settings" class="dropdown-item" @click="showUserMenu = false">
              <Settings :size="14" />
              <span>Settings</span>
            </NuxtLink>
            <NuxtLink v-if="currentPage !== 'explorer'" to="/explorer" class="dropdown-item" @click="showUserMenu = false">
              <FolderOpen :size="14" />
              <span>Explorer</span>
            </NuxtLink>
            <div class="dropdown-divider"></div>
            <button class="dropdown-item danger" @click="logout">
              <LogOut :size="14" />
              <span>Log out</span>
            </button>
          </div>
        </Transition>
      </div>
    </header>
    <div class="header-divider"></div>
  </div>
</template>

<script setup lang="ts">
import { Menu, Settings, LogOut, FolderOpen } from 'lucide-vue-next'

const props = withDefaults(defineProps<{
  title?: string
  subtitle?: string
  showHamburger?: boolean
  currentPage?: string
}>(), {
  title: '',
  subtitle: '',
  showHamburger: true,
  currentPage: ''
})

const showUserMenu = ref(false)
const sidebarOpen = inject<Ref<boolean>>('sidebarOpen', ref(false))

const userName = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).name } catch { return 'User' }
    }
  }
  return 'User'
})

const userEmail = computed(() => {
  if (import.meta.client) {
    const user = localStorage.getItem('user')
    if (user) {
      try { return JSON.parse(user).email } catch { return '' }
    }
  }
  return ''
})

const userInitial = computed(() => userName.value.charAt(0).toUpperCase())

const logout = () => {
  localStorage.removeItem('token')
  localStorage.removeItem('user')
  localStorage.removeItem('permissions')
  navigateTo('/login')
}

const handleDocumentClick = (e: MouseEvent) => {
  if (showUserMenu.value) {
    const target = e.target as HTMLElement
    if (!target.closest('.user-menu-wrapper')) {
      showUserMenu.value = false
    }
  }
}

onMounted(() => {
  if (import.meta.client) {
    document.addEventListener('click', handleDocumentClick)
  }
})

onUnmounted(() => {
  if (import.meta.client) {
    document.removeEventListener('click', handleDocumentClick)
  }
})
</script>

<style scoped>
.top-bar {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  background-color: var(--color-surface-0);
  border-radius: 0.75rem;
  flex-wrap: nowrap;
}

.hamburger-btn {
  width: 36px;
  height: 36px;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  cursor: pointer;
  display: none;
  align-items: center;
  justify-content: center;
  color: var(--color-text-secondary);
  flex-shrink: 0;
  transition: background-color 0.12s ease;
}

.hamburger-btn:hover {
  background-color: var(--color-surface-1);
}

@media (max-width: 480px) {
  .top-bar {
    padding: 0.5rem 0.75rem;
    flex-wrap: wrap;
  }

  .page-title {
    font-size: 14px;
  }

  .page-subtitle {
    display: none;
  }

  .avatar-circle {
    width: 34px;
    height: 34px;
    font-size: 13px;
  }

  .notification-dot {
    width: 8px;
    height: 8px;
  }
}

@media (max-width: 768px) {
  .hamburger-btn {
    display: flex;
  }
}

.top-bar-title {
  flex: 1;
  min-width: 0;
  overflow: hidden;
}

.page-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.025em;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.page-subtitle {
  font-size: 11px;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-divider {
  height: 1px;
  background-color: #E4E4E7;
  margin: 0.75rem 0;
}

@media (max-width: 768px) {
  .header-divider {
    margin: 0.5rem 0;
  }
}

.avatar-btn {
  position: relative;
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
  flex-shrink: 0;
}

.avatar-circle {
  width: 38px;
  height: 38px;
  border-radius: 9999px;
  background-color: #F43F5E;
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  font-weight: 600;
}

.notification-dot {
  position: absolute;
  top: 0;
  right: 0;
  width: 9px;
  height: 9px;
  background-color: #EF4444;
  border: 2px solid var(--color-surface-0);
  border-radius: 9999px;
}

.user-menu-wrapper {
  position: relative;
}

.user-dropdown {
  position: absolute;
  top: calc(100% + 0.5rem);
  right: 0;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-3);
  border-radius: 0.5rem;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  min-width: 200px;
  z-index: 50;
  padding: 0.25rem 0;
}

.dropdown-user-info {
  padding: 0.625rem 0.75rem;
}

.dropdown-user-name {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--color-text-primary);
}

.dropdown-user-email {
  font-size: 0.6875rem;
  color: var(--color-text-muted);
  margin-top: 0.125rem;
}

.dropdown-divider {
  height: 1px;
  background-color: var(--color-surface-2);
  margin: 0.25rem 0;
}

.dropdown-item {
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
  text-decoration: none;
  transition: background-color 0.1s ease;
}

.dropdown-item:hover {
  background-color: var(--color-surface-1);
}

.dropdown-item.danger {
  color: #F43F5E;
}

.dropdown-item.danger:hover {
  background-color: rgba(244, 63, 94, 0.08);
}

.menu-enter-active { transition: opacity 0.1s ease, transform 0.1s ease; }
.menu-leave-active { transition: opacity 0.08s ease, transform 0.08s ease; }
.menu-enter-from, .menu-leave-to { opacity: 0; transform: translateY(-4px); }
</style>
