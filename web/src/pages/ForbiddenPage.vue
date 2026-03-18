<template>
  <div class="status-page">
    <div class="bg-orb orb-a"></div>
    <div class="bg-orb orb-b"></div>

    <button class="theme-toggle" @click="toggleTheme" aria-label="切换主题">
      <svg v-if="isDark()" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <circle cx="12" cy="12" r="5" stroke-width="2" />
        <line x1="12" y1="1" x2="12" y2="3" stroke-width="2" />
        <line x1="12" y1="21" x2="12" y2="23" stroke-width="2" />
      </svg>
      <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke-width="2" />
      </svg>
    </button>

    <section class="status-card">
      <p class="code">403</p>
      <h1>没有访问权限</h1>
      <p class="desc">当前账号无法访问这个页面，请联系管理员开通权限，或返回可访问模块。</p>

      <div class="actions">
        <button class="ghost-btn" @click="goBack">返回上一页</button>
        <RouterLink class="primary-btn" to="/chat">回到应用</RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, useRouter } from 'vue-router'

import { useTheme } from '@/composables/useTheme'

const router = useRouter()
const { toggleTheme, isDark } = useTheme()

function goBack() {
  if (window.history.length > 1) {
    router.back()
    return
  }
  router.push('/chat')
}
</script>

<style scoped>
.status-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
  background: radial-gradient(circle at 10% 15%, color-mix(in srgb, var(--accent-primary) 18%, transparent) 0%, transparent 48%),
    radial-gradient(circle at 90% 80%, color-mix(in srgb, var(--accent-secondary) 20%, transparent) 0%, transparent 45%),
    var(--bg-primary);
  overflow: hidden;
}

.bg-orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(50px);
  opacity: 0.45;
  pointer-events: none;
}

.orb-a {
  width: 260px;
  height: 260px;
  top: -70px;
  left: -70px;
  background: var(--accent-primary);
}

.orb-b {
  width: 300px;
  height: 300px;
  right: -90px;
  bottom: -90px;
  background: var(--accent-secondary);
}

.theme-toggle {
  position: fixed;
  top: 1rem;
  right: 1rem;
  width: 40px;
  height: 40px;
  border-radius: 10px;
  border: 1px solid var(--glass-border);
  background: var(--glass-bg);
  color: var(--text-primary);
  display: grid;
  place-items: center;
  cursor: pointer;
  backdrop-filter: blur(12px);
}

.status-card {
  position: relative;
  width: min(560px, 100%);
  padding: 2rem;
  border: 1px solid var(--glass-border);
  border-radius: 18px;
  background: var(--glass-bg);
  backdrop-filter: blur(14px);
  text-align: center;
  z-index: 1;
}

.code {
  margin: 0;
  font-size: 0.85rem;
  letter-spacing: 0.2em;
  color: var(--text-secondary);
}

h1 {
  margin: 0.7rem 0;
  font-size: 2rem;
  color: var(--text-primary);
}

.desc {
  margin: 0;
  color: var(--text-secondary);
  line-height: 1.7;
}

.actions {
  margin-top: 1.5rem;
  display: flex;
  gap: 0.75rem;
  justify-content: center;
  flex-wrap: wrap;
}

.primary-btn,
.ghost-btn {
  min-width: 132px;
  padding: 0.7rem 1rem;
  border-radius: 10px;
  font-size: 0.9rem;
  text-decoration: none;
  border: 1px solid transparent;
  cursor: pointer;
}

.primary-btn {
  background: var(--accent-primary);
  color: #fff;
}

.primary-btn:hover {
  background: var(--accent-secondary);
}

.ghost-btn {
  background: transparent;
  border-color: var(--input-border);
  color: var(--text-primary);
}

.ghost-btn:hover {
  border-color: var(--accent-primary);
}

@media (max-width: 640px) {
  .status-card {
    padding: 1.4rem;
  }

  h1 {
    font-size: 1.6rem;
  }

  .actions {
    flex-direction: column;
  }
}
</style>
