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
      <p class="code">404</p>
      <h1>页面未找到</h1>
      <p class="desc">你访问的地址不存在，可能已被移动或删除。</p>
      <p class="path">当前路径：{{ route.fullPath }}</p>

      <div class="actions">
        <button class="ghost-btn" @click="goBack">返回上一页</button>
        <RouterLink class="primary-btn" to="/chat">回到应用</RouterLink>
      </div>
    </section>
  </div>
</template>

<script setup lang="ts">
import { RouterLink, useRoute, useRouter } from 'vue-router'

import { useTheme } from '@/composables/useTheme'

const route = useRoute()
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
  background: radial-gradient(circle at 8% 18%, color-mix(in srgb, var(--accent-primary) 16%, transparent) 0%, transparent 46%),
    radial-gradient(circle at 90% 78%, color-mix(in srgb, var(--accent-secondary) 20%, transparent) 0%, transparent 45%),
    var(--bg-primary);
  overflow: hidden;
}

.bg-orb {
  position: absolute;
  border-radius: 999px;
  filter: blur(50px);
  opacity: 0.42;
  pointer-events: none;
}

.orb-a {
  width: 230px;
  height: 230px;
  top: -70px;
  left: -60px;
  background: var(--accent-primary);
}

.orb-b {
  width: 280px;
  height: 280px;
  right: -85px;
  bottom: -85px;
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
  width: min(580px, 100%);
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

.path {
  margin: 0.8rem 0 0;
  color: var(--text-secondary);
  font-size: 0.82rem;
  word-break: break-all;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
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
