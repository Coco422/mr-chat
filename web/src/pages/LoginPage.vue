<template>
  <div class="login-container">
    <canvas ref="canvasRef" class="particle-bg"></canvas>

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

    <div class="login-card">
      <div class="card-left">
        <div class="logo-section">
          <div class="logo-icon">
            <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
              <circle cx="32" cy="32" r="28" stroke="url(#grad1)" stroke-width="3"/>
              <path d="M20 32 L28 24 L36 32 L44 24" stroke="url(#grad1)" stroke-width="3" stroke-linecap="round"/>
              <circle cx="32" cy="40" r="3" fill="url(#grad1)"/>
              <defs>
                <linearGradient id="grad1" x1="0%" y1="0%" x2="100%" y2="100%">
                  <stop offset="0%" style="stop-color:var(--accent-primary);stop-opacity:1" />
                  <stop offset="100%" style="stop-color:var(--accent-secondary);stop-opacity:1" />
                </linearGradient>
              </defs>
            </svg>
          </div>
          <h1 class="brand-title">MrChat</h1>
          <p class="brand-tagline">聚合智能 · 对话未来</p>
        </div>
      </div>

      <div class="card-right">
        <h2 class="form-title">欢迎回来</h2>

        <form @submit.prevent="submit" class="login-form">
          <div class="form-group" :class="{ focused: identifierFocused, error: errorMessage }">
            <label for="identifier">邮箱或用户名</label>
            <input
              id="identifier"
              v-model="identifier"
              autocomplete="username"
              @focus="identifierFocused = true"
              @blur="identifierFocused = false"
            />
          </div>

          <div class="form-group" :class="{ focused: passwordFocused, error: errorMessage }">
            <label for="password">密码</label>
            <input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              @focus="passwordFocused = true"
              @blur="passwordFocused = false"
            />
          </div>

          <p v-if="errorMessage" class="error-message">{{ errorMessage }}</p>

          <button type="submit" :disabled="submitting" class="submit-btn">
            <span v-if="!submitting">登录</span>
            <span v-else class="spinner"></span>
          </button>
        </form>

        <RouterLink to="/signup" class="signup-link">没有账号？立即注册</RouterLink>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { ApiError } from '@/lib/api'
import { signIn } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'

const { toggleTheme, isDark } = useTheme()
const identifier = ref('')
const password = ref('')
const errorMessage = ref('')
const submitting = ref(false)
const identifierFocused = ref(false)
const passwordFocused = ref(false)
const auth = useAuthStore()
const router = useRouter()
const canvasRef = ref<HTMLCanvasElement>()

async function submit() {
  submitting.value = true
  errorMessage.value = ''

  try {
    const data = await signIn({
      identifier: identifier.value,
      password: password.value
    })
    auth.setSession(data.access_token, data.user)
    await auth.fetchMe()
    router.push(data.user.role === 'admin' || data.user.role === 'root' ? '/admin/upstreams' : '/chat')
  } catch (error) {
    errorMessage.value = error instanceof ApiError ? error.message : '登录失败'
  } finally {
    submitting.value = false
  }
}

// Particle animation
let animationId: number
onMounted(() => {
  const canvas = canvasRef.value
  if (!canvas) return

  const ctx = canvas.getContext('2d')
  if (!ctx) return

  canvas.width = window.innerWidth
  canvas.height = window.innerHeight

  const particles: Array<{ x: number; y: number; vx: number; vy: number; size: number }> = []
  for (let i = 0; i < 50; i++) {
    particles.push({
      x: Math.random() * canvas.width,
      y: Math.random() * canvas.height,
      vx: (Math.random() - 0.5) * 0.5,
      vy: (Math.random() - 0.5) * 0.5,
      size: Math.random() * 2 + 1
    })
  }

  function animate() {
    if (!canvas || !ctx) return
    ctx.clearRect(0, 0, canvas.width, canvas.height)

    const isDarkMode = document.documentElement.getAttribute('data-theme') === 'dark'
    ctx.fillStyle = isDarkMode ? 'rgba(45, 155, 210, 0.5)' : 'rgba(45, 155, 210, 0.4)'
    ctx.strokeStyle = isDarkMode ? 'rgba(129, 195, 228, 0.3)' : 'rgba(129, 195, 228, 0.25)'

    particles.forEach((p, i) => {
      p.x += p.vx
      p.y += p.vy
      if (p.x < 0 || p.x > canvas.width) p.vx *= -1
      if (p.y < 0 || p.y > canvas.height) p.vy *= -1

      ctx.beginPath()
      ctx.arc(p.x, p.y, p.size, 0, Math.PI * 2)
      ctx.fill()

      particles.slice(i + 1).forEach(p2 => {
        const dx = p.x - p2.x
        const dy = p.y - p2.y
        const dist = Math.sqrt(dx * dx + dy * dy)
        if (dist < 120) {
          ctx.beginPath()
          ctx.moveTo(p.x, p.y)
          ctx.lineTo(p2.x, p2.y)
          ctx.stroke()
        }
      })
    })

    animationId = requestAnimationFrame(animate)
  }

  animate()

  const handleResize = () => {
    canvas.width = window.innerWidth
    canvas.height = window.innerHeight
  }
  window.addEventListener('resize', handleResize)

  onUnmounted(() => {
    cancelAnimationFrame(animationId)
    window.removeEventListener('resize', handleResize)
  })
})
</script>

<style scoped>
.login-container {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-primary);
  overflow: hidden;
}

.particle-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
}

.theme-toggle {
  position: fixed;
  top: 2rem;
  right: 2rem;
  width: 48px;
  height: 48px;
  border-radius: 50%;
  background: var(--glass-bg);
  backdrop-filter: blur(12px);
  border: 1px solid var(--glass-border);
  color: var(--text-primary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.3s ease;
  z-index: 10;
}

.theme-toggle:hover {
  transform: scale(1.1) rotate(15deg);
  box-shadow: 0 0 20px var(--accent-primary);
}

.login-card {
  position: relative;
  display: flex;
  width: 90%;
  max-width: 1000px;
  min-height: 600px;
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  border-radius: 24px;
  border: 1px solid var(--glass-border);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.2);
  overflow: hidden;
  z-index: 1;
}

.card-left {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #ffffff;
  padding: 3rem;
  position: relative;
}

.card-left::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(circle at 30% 50%, rgba(45, 155, 210, 0.03) 0%, transparent 60%);
}

.logo-section {
  text-align: center;
  position: relative;
  z-index: 1;
}

.logo-icon {
  animation: float 3s ease-in-out infinite;
}

@keyframes float {
  0%, 100% { transform: translateY(0); }
  50% { transform: translateY(-10px); }
}

.brand-title {
  font-size: 3rem;
  font-weight: 700;
  color: #000000;
  margin: 1.5rem 0 0.5rem;
  text-shadow: none;
}

.brand-tagline {
  font-size: 1.1rem;
  color: #666666;
  font-weight: 300;
}

.card-right {
  flex: 1;
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

.login-form {
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
  background: var(--input-bg);
  border: 2px solid var(--input-border);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 1rem;
  transition: all 0.3s ease;
  outline: none;
}

.form-group.focused input {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 4px var(--accent-glow);
  transform: scale(1.02);
}

.form-group.error input {
  border-color: var(--error-color);
  animation: shake 0.4s ease;
}

@keyframes shake {
  0%, 100% { transform: translateX(0); }
  25% { transform: translateX(-8px); }
  75% { transform: translateX(8px); }
}

.error-message {
  color: var(--error-color);
  font-size: 0.875rem;
  margin: -0.5rem 0 0;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(-5px); }
  to { opacity: 1; transform: translateY(0); }
}

.submit-btn {
  width: 100%;
  padding: 1rem;
  background: linear-gradient(135deg, var(--accent-primary) 0%, var(--accent-secondary) 100%);
  border: none;
  border-radius: 12px;
  color: white;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.3s ease;
  position: relative;
  overflow: hidden;
}

.submit-btn::before {
  content: '';
  position: absolute;
  inset: 0;
  background: linear-gradient(135deg, var(--accent-secondary) 0%, var(--accent-primary) 100%);
  opacity: 0;
  transition: opacity 0.3s ease;
}

.submit-btn:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 8px 24px var(--accent-glow);
}

.submit-btn:hover:not(:disabled)::before {
  opacity: 1;
}

.submit-btn:disabled {
  opacity: 0.7;
  cursor: not-allowed;
}

.submit-btn span {
  position: relative;
  z-index: 1;
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

.signup-link {
  text-align: center;
  margin-top: 1.5rem;
  color: var(--text-secondary);
  text-decoration: none;
  font-size: 0.9rem;
  transition: color 0.3s ease;
}

.signup-link:hover {
  color: var(--accent-primary);
}

@media (max-width: 768px) {
  .login-card {
    flex-direction: column;
    max-width: 90%;
  }

  .card-left {
    padding: 2rem;
  }

  .brand-title {
    font-size: 2rem;
  }

  .card-right {
    padding: 2rem;
  }
}
</style>
