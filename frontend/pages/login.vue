<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <div class="login-logo">
          <HardDrive :size="28" color="white" :stroke-width="2" />
        </div>
        <h1 class="login-title">RouteStorage</h1>
        <p class="login-subtitle">Sign in to your account</p>
      </div>

      <div v-if="error" class="login-error">
        {{ error }}
      </div>

      <form v-if="!showRegister" @submit.prevent="handleLogin" class="login-form">
        <div class="form-group">
          <label for="email" class="form-label">Email address</label>
          <input
            id="email"
            v-model="email"
            type="email"
            class="input-field"
            placeholder="you@example.com"
            required
            autocomplete="email"
          />
        </div>

        <div class="form-group">
          <label for="password" class="form-label">Password</label>
          <div class="password-wrapper">
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              class="input-field"
              placeholder="Enter your password"
              required
              autocomplete="current-password"
            />
            <button type="button" class="password-toggle" @click="showPassword = !showPassword" tabindex="-1">
              <EyeOff v-if="showPassword" :size="16" />
              <Eye v-else :size="16" />
            </button>
          </div>
        </div>

        <button type="submit" class="btn-primary login-btn" :disabled="loading">
          <Loader2 v-if="loading" :size="16" class="spin" />
          <LogIn v-else :size="16" />
          <span>{{ loading ? 'Signing in...' : 'Sign in' }}</span>
        </button>
      </form>

      <form v-else @submit.prevent="handleRegister" class="login-form">
        <div class="form-group">
          <label for="reg-name" class="form-label">Name</label>
          <input
            id="reg-name"
            v-model="name"
            type="text"
            class="input-field"
            placeholder="Your name"
            required
            autocomplete="name"
          />
        </div>

        <div class="form-group">
          <label for="reg-email" class="form-label">Email address</label>
          <input
            id="reg-email"
            v-model="regEmail"
            type="email"
            class="input-field"
            placeholder="you@example.com"
            required
            autocomplete="email"
          />
        </div>

        <div class="form-group">
          <label for="reg-password" class="form-label">Password</label>
          <div class="password-wrapper">
            <input
              id="reg-password"
              v-model="regPassword"
              :type="showRegPassword ? 'text' : 'password'"
              class="input-field"
              placeholder="Min 6 characters"
              required
              minlength="6"
              autocomplete="new-password"
            />
            <button type="button" class="password-toggle" @click="showRegPassword = !showRegPassword" tabindex="-1">
              <EyeOff v-if="showRegPassword" :size="16" />
              <Eye v-else :size="16" />
            </button>
          </div>
        </div>

        <button type="submit" class="btn-primary login-btn" :disabled="loading">
          <Loader2 v-if="loading" :size="16" class="spin" />
          <UserPlus v-else :size="16" />
          <span>{{ loading ? 'Creating account...' : 'Create account' }}</span>
        </button>
      </form>

      <div class="login-toggle">
        <template v-if="!showRegister">
          Don't have an account?
          <button type="button" class="toggle-btn" @click="showRegister = true">Create one</button>
        </template>
        <template v-else>
          Already have an account?
          <button type="button" class="toggle-btn" @click="showRegister = false">Sign in</button>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { HardDrive, LogIn, UserPlus, Loader2, Eye, EyeOff } from 'lucide-vue-next'

definePageMeta({ layout: false })

const { apiFetch } = useApi()

const email = ref('')
const password = ref('')
const name = ref('')
const regEmail = ref('')
const regPassword = ref('')
const error = ref('')
const loading = ref(false)
const showRegister = ref(false)
const showPassword = ref(false)
const showRegPassword = ref(false)

const handleLogin = async () => {
  loading.value = true
  error.value = ''
  try {
    const data = await $fetch<{ token: string; user: any }>(`${useRuntimeConfig().public.apiBase}/api/auth/login`, {
      method: 'POST',
      body: { email: email.value, password: password.value }
    })
    localStorage.setItem('token', data.token)
    localStorage.setItem('user', JSON.stringify(data.user))
    navigateTo('/settings')
  } catch (e: any) {
    error.value = e.data?.error || 'Login failed'
  } finally {
    loading.value = false
  }
}

const handleRegister = async () => {
  loading.value = true
  error.value = ''
  try {
    await $fetch(`${useRuntimeConfig().public.apiBase}/api/auth/register`, {
      method: 'POST',
      body: { email: regEmail.value, password: regPassword.value, name: name.value }
    })
    email.value = regEmail.value
    password.value = regPassword.value
    showRegister.value = false
    await handleLogin()
  } catch (e: any) {
    error.value = e.data?.error || 'Registration failed'
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background-color: var(--color-surface-1);
}

.login-card {
  width: 100%;
  max-width: 24rem;
  background-color: var(--color-surface-0);
  border: 1px solid var(--color-surface-2);
  border-radius: 1rem;
  padding: 2.5rem 2rem;
}

.login-header {
  text-align: center;
  margin-bottom: 2rem;
}

.login-logo {
  width: 3.5rem;
  height: 3.5rem;
  background-color: var(--color-brand-600);
  border-radius: 0.75rem;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 1.25rem auto;
}

.login-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--color-text-primary);
  letter-spacing: -0.025em;
}

.login-subtitle {
  font-size: 0.875rem;
  color: var(--color-text-muted);
  margin-top: 0.375rem;
}

.login-error {
  background-color: rgba(250, 82, 82, 0.1);
  color: var(--color-danger);
  padding: 0.75rem 1rem;
  border-radius: 0.5rem;
  font-size: 0.875rem;
  margin-bottom: 1.25rem;
  border: 1px solid rgba(250, 82, 82, 0.2);
}

.login-form {
  display: flex;
  flex-direction: column;
  gap: 1.125rem;
}

.form-group {
  display: flex;
  flex-direction: column;
  gap: 0.375rem;
}

.form-label {
  font-size: 0.8125rem;
  font-weight: 500;
  color: var(--color-text-secondary);
}

.login-btn {
  width: 100%;
  margin-top: 0.5rem;
  padding: 0.75rem 1.25rem;
}

.login-toggle {
  text-align: center;
  margin-top: 1.5rem;
  padding-top: 1.25rem;
  border-top: 1px solid var(--color-surface-2);
  font-size: 0.8125rem;
  color: var(--color-text-muted);
}

.toggle-btn {
  background: none;
  border: none;
  color: var(--color-brand-600);
  cursor: pointer;
  font-weight: 600;
  font-size: inherit;
  padding: 0;
  transition: color 0.15s ease;
}

.toggle-btn:hover {
  color: var(--color-brand-700);
}

.password-wrapper {
  position: relative;
}

.password-wrapper .input-field {
  padding-right: 2.75rem;
}

.password-toggle {
  position: absolute;
  right: 0.75rem;
  top: 50%;
  transform: translateY(-50%);
  background: none;
  border: none;
  color: var(--color-text-muted);
  cursor: pointer;
  padding: 0.25rem;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: color 0.15s ease;
}

.password-toggle:hover {
  color: var(--color-text-secondary);
}
</style>
