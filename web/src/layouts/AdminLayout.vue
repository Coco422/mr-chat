<template>
  <div class="admin-layout">
    <aside class="admin-sidebar" :class="{ collapsed: isSidebarCollapsed }">
      <div class="sidebar-header">
        <div class="title-wrap">
          <h1>{{ isSidebarCollapsed ? 'MA' : 'MrChat Admin' }}</h1>
          <p v-if="!isSidebarCollapsed">管理控制台</p>
        </div>
        <div class="header-actions">
          <button
            class="icon-btn"
            @click="toggleSidebar"
            :title="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            :aria-label="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
          >
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path
                v-if="isSidebarCollapsed"
                d="m9 6 6 6-6 6"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
              <path
                v-else
                d="m15 6-6 6 6 6"
                stroke-width="2"
                stroke-linecap="round"
                stroke-linejoin="round"
              />
            </svg>
          </button>
          <button
            class="icon-btn"
            @click="toggleTheme"
            :title="isDark() ? '切换浅色主题' : '切换深色主题'"
            :aria-label="isDark() ? '切换浅色主题' : '切换深色主题'"
          >
            {{ isDark() ? '☀' : '☾' }}
          </button>
        </div>
      </div>

      <nav class="admin-nav">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ active: isActive(item.to) }"
          :title="isSidebarCollapsed ? item.label : undefined"
        >
          <span class="nav-dot" v-if="!isSidebarCollapsed"></span>
          <span class="nav-label">{{ isSidebarCollapsed ? item.shortLabel : item.label }}</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div v-if="auth.user" class="user-card">
          <div class="avatar">{{ auth.user.username.slice(0, 1).toUpperCase() }}</div>
          <div class="user-meta" v-if="!isSidebarCollapsed">
            <div class="name">{{ auth.user.username }}</div>
            <div class="role">{{ auth.user.role }}</div>
          </div>
        </div>

        <div class="footer-actions">
          <RouterLink to="/chat" class="action-link" :title="isSidebarCollapsed ? '返回聊天' : undefined">
            {{ isSidebarCollapsed ? '聊天' : '返回聊天' }}
          </RouterLink>
          <button class="action-btn" @click="handleSignOut" :title="isSidebarCollapsed ? '退出登录' : undefined">
            {{ isSidebarCollapsed ? '退出' : '退出登录' }}
          </button>
        </div>
      </div>
    </aside>

    <main class="admin-main">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'

import { useTheme } from '@/composables/useTheme'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()
const { toggleTheme, isDark } = useTheme()
const isSidebarCollapsed = ref(false)
const ADMIN_SIDEBAR_COLLAPSED_STORAGE_KEY = 'mrchat:admin-sidebar:collapsed'

const navItems = [
  { to: '/admin/upstreams', label: '上游配置', shortLabel: '上游' },
  { to: '/admin/channels', label: '渠道管理', shortLabel: '渠道' },
  { to: '/admin/models', label: '模型管理', shortLabel: '模型' },
  { to: '/admin/user-groups', label: '用户组', shortLabel: '分组' },
  { to: '/admin/users', label: '用户管理', shortLabel: '用户' },
  { to: '/admin/redeem-codes', label: '兑换码', shortLabel: '兑换' },
  { to: '/admin/audit-logs', label: '审计日志', shortLabel: '日志' }
]

onMounted(() => {
  isSidebarCollapsed.value = readSidebarCollapsedState()
})

function isActive(path: string) {
  return route.path === path
}

function toggleSidebar() {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
  persistSidebarCollapsedState(isSidebarCollapsed.value)
}

async function handleSignOut() {
  await auth.signOut()
  router.push({ name: 'login' })
}

function readSidebarCollapsedState() {
  try {
    return window.localStorage.getItem(ADMIN_SIDEBAR_COLLAPSED_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function persistSidebarCollapsedState(collapsed: boolean) {
  try {
    window.localStorage.setItem(ADMIN_SIDEBAR_COLLAPSED_STORAGE_KEY, String(collapsed))
  } catch {
    // Ignore storage write failures and keep the in-memory state.
  }
}
</script>

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  overflow: hidden;
  background: var(--layout-content-bg);
}

.admin-sidebar {
  width: 272px;
  height: 100vh;
  flex-shrink: 0;
  background: var(--layout-sidebar-bg);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  transition: width 0.24s ease;
}

.admin-sidebar.collapsed {
  width: 92px;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid var(--glass-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.title-wrap h1 {
  margin: 0;
  font-size: 1.1rem;
  line-height: 1.2;
  color: var(--text-primary);
}

.title-wrap p {
  margin: 0.4rem 0 0;
  font-size: 0.8rem;
  color: var(--text-secondary);
}

.icon-btn {
  width: 34px;
  height: 34px;
  border-radius: 10px;
  border: 1px solid var(--glass-border);
  background: var(--layout-sidebar-bg);
  color: var(--text-primary);
  cursor: pointer;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.icon-btn:hover {
  background: var(--surface-muted);
}

.admin-nav {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 0.75rem 0.5rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.65rem;
  padding: 0.7rem 0.85rem;
  margin-bottom: 0.35rem;
  border-radius: 12px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
  border: 1px solid transparent;
}

.nav-item:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.nav-item.active {
  color: var(--text-primary);
  border-color: var(--glass-border);
  background: var(--surface-muted);
}

.nav-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.nav-dot {
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: var(--accent-primary);
  opacity: 0.8;
}

.sidebar-footer {
  border-top: 1px solid var(--glass-border);
  padding: 0.9rem;
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.user-card {
  display: flex;
  align-items: center;
  gap: 0.65rem;
}

.avatar {
  width: 32px;
  height: 32px;
  border-radius: 999px;
  background: var(--surface-muted);
  color: var(--text-primary);
  display: grid;
  place-items: center;
  font-size: 0.85rem;
  font-weight: 600;
  border: 1px solid var(--glass-border);
}

.user-meta {
  min-width: 0;
}

.name {
  color: var(--text-primary);
  font-size: 0.84rem;
  font-weight: 600;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
}

.role {
  color: var(--text-secondary);
  font-size: 0.78rem;
}

.footer-actions {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 0.5rem;
}

.action-link,
.action-btn {
  padding: 0.55rem 0.6rem;
  border-radius: 10px;
  font-size: 0.82rem;
  text-align: center;
  border: 1px solid var(--glass-border);
  background: var(--layout-sidebar-bg);
  color: var(--text-primary);
  text-decoration: none;
  transition: background 0.2s ease, border-color 0.2s ease;
}

.action-btn {
  cursor: pointer;
}

.action-link:hover,
.action-btn:hover {
  border-color: var(--accent-primary);
  background: var(--surface-muted);
}

.admin-main {
  flex: 1;
  height: 100vh;
  min-height: 0;
  min-width: 0;
  overflow-y: auto;
  overflow-x: hidden;
  background: var(--layout-content-bg);
}

.admin-sidebar.collapsed .sidebar-header {
  padding: 1rem 0.75rem;
  flex-direction: column;
}

.admin-sidebar.collapsed .header-actions {
  flex-direction: column;
}

.admin-sidebar.collapsed .title-wrap {
  text-align: center;
}

.admin-sidebar.collapsed .admin-nav {
  padding: 0.75rem 0.45rem;
}

.admin-sidebar.collapsed .nav-item {
  justify-content: center;
  padding: 0.7rem 0.45rem;
}

.admin-sidebar.collapsed .nav-label {
  font-size: 0.78rem;
  text-align: center;
}

.admin-sidebar.collapsed .user-card {
  justify-content: center;
}

.admin-sidebar.collapsed .footer-actions {
  grid-template-columns: 1fr;
}

@media (max-width: 960px) {
  .admin-layout {
    flex-direction: column;
    height: auto;
    overflow: visible;
  }

  .admin-sidebar {
    width: 100%;
    height: auto;
    border-right: none;
    border-bottom: 1px solid var(--glass-border);
  }

  .admin-sidebar.collapsed {
    width: 100%;
  }

  .admin-nav {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 0.35rem;
  }

  .admin-main {
    height: auto;
    overflow: visible;
  }

  .admin-sidebar.collapsed .sidebar-header {
    flex-direction: row;
  }

  .admin-sidebar.collapsed .header-actions {
    flex-direction: row;
  }

  .admin-sidebar.collapsed .nav-item {
    justify-content: flex-start;
    padding: 0.7rem 0.85rem;
  }

  .admin-sidebar.collapsed .nav-label {
    font-size: 0.9rem;
    text-align: left;
  }

  .admin-sidebar.collapsed .footer-actions {
    grid-template-columns: 1fr 1fr;
  }
}

@media (max-width: 600px) {
  .admin-nav {
    grid-template-columns: 1fr;
  }

  .footer-actions {
    grid-template-columns: 1fr;
  }
}
</style>
