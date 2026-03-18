<template>
  <div class="signup-container">
    <button class="theme-toggle" @click="toggleTheme" aria-label="切换主题">
      <svg v-if="isDark()" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <circle cx="12" cy="12" r="5" stroke-width="2"/><line x1="12" y1="1" x2="12" y2="3" stroke-width="2"/><line x1="12" y1="21" x2="12" y2="23" stroke-width="2"/>
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" stroke-width="2"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78" stroke-width="2"/>
        <line x1="1" y1="12" x2="3" y2="12" stroke-width="2"/><line x1="21" y1="12" x2="23" y2="12" stroke-width="2"/>
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" stroke-width="2"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22" stroke-width="2"/>
      </svg>
      <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke-width="2"/>
      </svg>
    </button>

    <div class="signup-card">
      <div class="card-left">
        <div class="brand-panel">
          <div class="brand-mark">M</div>
          <h1 class="brand-title">MrChat</h1>
          <p class="brand-tagline">创建账号后进入统一的亮色或暗色工作区</p>
          <p class="brand-copy">注册页和登录页现在共用一套更克制的视觉规则，避免和主界面割裂。</p>
        </div>
      </div>

      <div class="card-right">
        <h2 class="form-title">创建账号</h2>

        <form @submit.prevent="submit" class="signup-form">
          <div class="form-group" :class="{ focused: usernameFocused, error: errorMessage }">
            <label for="username">用户名</label>
            <input
              id="username"
              v-model="username"
              autocomplete="username"
              @focus="usernameFocused = true"
              @blur="usernameFocused = false"
            />
          </div>

          <div class="form-group" :class="{ focused: emailFocused, error: errorMessage }">
            <label for="email">邮箱</label>
            <input
              id="email"
              v-model="email"
              type="email"
              autocomplete="email"
              @focus="emailFocused = true"
              @blur="emailFocused = false"
            />
          </div>

          <div class="form-group" :class="{ focused: passwordFocused, error: errorMessage }">
            <label for="password">密码</label>
            <input
              id="password"
              v-model="password"
              type="password"
              autocomplete="new-password"
              @focus="passwordFocused = true"
              @blur="passwordFocused = false"
            />
          </div>

          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

          <button type="submit" :disabled="submitting" class="submit-btn">
            <span v-if="!submitting">注册</span>
            <span v-else class="spinner"></span>
          </button>
        </form>

        <RouterLink to="/login" class="login-link">已有账号？立即登录</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ApiError } from '@/lib/api'
import { signUp } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const { toggleTheme, isDark } = useTheme()
const username = ref('')
const email = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const usernameFocused = ref(false)
const emailFocused = ref(false)
const passwordFocused = ref(false)
const auth = useAuthStore()
const router = useRouter()

async function submit() {
  submitting.value = true
  errorMessage.value = ''

  try {
    const data = await signUp({
      username: username.value,
      email: email.value,
      password: password.value
    })

    auth.setSession(data.access_token, data.user)
    await auth.fetchMe()
    router.push('/chat')
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '注册失败'
  } finally {
    submitting.value = false
  }
}
</script>

<style scoped>
.signup-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  padding: 2rem;
}

.theme-toggle {
  position: fixed;
  top: 2rem;
  right: 2rem;
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--bg-secondary);
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease, border-color 0.2s ease;
  z-index: 10;
}

.theme-toggle:hover {
  background: var(--surface-muted);
}

.signup-card {
  display: grid;
  grid-template-columns: minmax(260px, 0.9fr) minmax(320px, 1fr);
  width: min(100%, 880px);
  min-height: 560px;
  background: var(--bg-secondary);
  border-radius: 24px;
  border: 1px solid var(--glass-border);
  box-shadow: var(--shadow-md);
  overflow: hidden;
}

.card-left {
  display: flex;
  align-items: center;
  padding: 3rem 2.5rem;
  background: var(--surface-subtle);
  border-right: 1px solid var(--glass-border);
}

.brand-panel {
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.brand-mark {
  width: 3rem;
  height: 3rem;
  border-radius: 14px;
  display: grid;
  place-items: center;
  background: var(--accent-primary);
  color: #fff;
  font-size: 1.25rem;
  font-weight: 700;
}

.brand-title {
  font-size: 2.25rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.brand-tagline {
  font-size: 1rem;
  color: var(--text-secondary);
  margin: 0;
}

.brand-copy {
  margin: 0;
  max-width: 28rem;
  line-height: 1.7;
  color: var(--text-secondary);
}

.card-right {
  padding: 3rem;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.form-title {
  font-size: 2rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 2rem;
}

.signup-form {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.form-group {
  position: relative;
}

.form-group label {
  display: block;
  font-size: 0.9rem;
  color: var(--text-secondary);
  margin-bottom: 0.5rem;
  font-weight: 500;
}

.form-group input {
  width: 100%;
  padding: 0.875rem 1rem;
  background: var(--bg-secondary);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 1rem;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
  outline: none;
}

.form-group.focused input {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.form-group.error input {
  border-color: var(--error-color);
}

.error-message {
  color: var(--error-color);
  font-size: 0.875rem;
  margin: -0.25rem 0 0;
}

.submit-btn {
  width: 100%;
  padding: 1rem;
  background: var(--accent-primary);
  border: none;
  border-radius: 12px;
  color: white;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s ease, opacity 0.2s ease;
}

.submit-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.submit-btn span {
  position: relative;
}

.spinner {
  display: inline-block;
  width: 20px;
  height: 20px;
  border: 3px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.login-link {
  text-align: center;
  margin-top: 1.5rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.3s ease;
}

.login-link:hover {
  color: var(--accent-primary);
}

@media (max-width: 768px) {
  .signup-card {
    grid-template-columns: 1fr;
  }

  .card-left {
    padding: 2rem;
    border-right: none;
    border-bottom: 1px solid var(--glass-border);
  }

  .card-right {
    padding: 2rem;
  }
}
</style>
