<template>
  <div class="app-layout">
    <aside class="sidebar" :class="{ collapsed: isSidebarCollapsed }">
      <div class="sidebar-header">
        <div class="header-top">
          <h1 class="logo">{{ isSidebarCollapsed ? 'M' : 'MrChat' }}</h1>
          <div class="header-actions">
            <button
              class="icon-btn collapse-btn"
              @click="toggleSidebar"
              :title="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
              :aria-label="isSidebarCollapsed ? '展开侧边栏' : '收起侧边栏'"
            >
              <ElIcon class="header-icon">
                <Expand v-if="isSidebarCollapsed" />
                <Fold v-else />
              </ElIcon>
            </button>
            <button
              class="icon-btn theme-btn"
              @click="toggleTheme"
              :title="isDark() ? '切换浅色主题' : '切换深色主题'"
              :aria-label="isDark() ? '切换浅色主题' : '切换深色主题'"
            >
              <ElIcon class="header-icon">
                <Sunny v-if="isDark()" />
                <Moon v-else />
              </ElIcon>
            </button>
          </div>
        </div>
        <button
          class="new-chat-btn"
          :class="{ compact: isSidebarCollapsed }"
          @click="createConversation"
          :title="isSidebarCollapsed ? '新对话' : undefined"
          aria-label="新对话"
        >
          <ElIcon class="button-icon">
            <Plus />
          </ElIcon>
          <span v-if="!isSidebarCollapsed">新对话</span>
        </button>
      </div>

      <div class="conversations-list">
        <RouterLink
          v-for="conv in conversations"
          :key="conv.id"
          :to="`/chat/${conv.id}`"
          class="conversation-item"
          :class="{ active: currentConversationId === conv.id }"
          :title="conv.title || '新对话'"
        >
          <div class="conv-title">{{ conv.title || '新对话' }}</div>
          <!-- <div class="conv-meta">{{ conv.message_count }} 条消息</div> -->
        </RouterLink>
        <div v-if="conversations.length === 0" class="empty-state">暂无对话</div>
      </div>

      <nav class="nav-menu">
        <RouterLink to="/chat" class="nav-item" :title="isSidebarCollapsed ? '对话' : undefined">
          <ElIcon class="nav-icon">
            <Message />
          </ElIcon>
          <span class="nav-label">对话</span>
        </RouterLink>
        <RouterLink to="/usage" class="nav-item" :title="isSidebarCollapsed ? '用量' : undefined">
          <ElIcon class="nav-icon">
            <Histogram />
          </ElIcon>
          <span class="nav-label">用量</span>
        </RouterLink>
        <RouterLink to="/settings/profile" class="nav-item" :title="isSidebarCollapsed ? '设置' : undefined">
          <ElIcon class="nav-icon">
            <Setting />
          </ElIcon>
          <span class="nav-label">设置</span>
        </RouterLink>
        <RouterLink to="/admin/upstreams" class="nav-item" :title="isSidebarCollapsed ? '管理' : undefined">
          <ElIcon class="nav-icon">
            <Operation />
          </ElIcon>
          <span class="nav-label">管理</span>
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div class="footer-row" v-if="auth.user">
          <div class="user-info">
            <div class="user-avatar">{{ auth.user.username[0].toUpperCase() }}</div>
            <div class="user-details">
              <div class="user-name">{{ auth.user.username }}</div>
              <div class="user-role">{{ auth.user.role }}</div>
            </div>
          </div>
          <button class="logout-btn" @click="handleSignOut" title="退出登录">
            <ElIcon class="logout-icon">
              <SwitchButton />
            </ElIcon>
            <span v-if="!isSidebarCollapsed">退出</span>
          </button>
        </div>
      </div>
    </aside>

    <main class="main-content">
      <RouterView />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ElIcon } from 'element-plus'
import { Expand, Fold, Histogram, Message, Moon, Operation, Plus, Setting, Sunny, SwitchButton } from '@element-plus/icons-vue'
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { RouterLink, RouterView, useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useTheme } from '@/composables/useTheme'
import { createConversation as createConversationRequest, listConversations, type ConversationSummary } from '@/api/chat'

const auth = useAuthStore()
const router = useRouter()
const route = useRoute()
const { toggleTheme, isDark } = useTheme()
const conversations = ref<ConversationSummary[]>([])
const isSidebarCollapsed = ref(false)
const SIDEBAR_COLLAPSED_STORAGE_KEY = 'mrchat:sidebar:collapsed'

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)

onMounted(async () => {
  isSidebarCollapsed.value = readSidebarCollapsedState()
  if (auth.isAuthenticated && !auth.user) {
    await auth.fetchMe()
  }
  window.addEventListener('mrchat:conversations:refresh', loadConversations)
  await loadConversations()
})

onUnmounted(() => {
  window.removeEventListener('mrchat:conversations:refresh', loadConversations)
})

async function loadConversations() {
  try {
    conversations.value = await listConversations(auth.accessToken)
  } catch (error) {
    console.error('Failed to load conversations:', error)
  }
}

async function createConversation() {
  try {
    const data = await createConversationRequest(auth.accessToken, {
      title: '新对话',
      model_id: null
    })
    await loadConversations()
    router.push(`/chat/${data.id}`)
  } catch (error) {
    console.error('Failed to create conversation:', error)
  }
}

async function handleSignOut() {
  await auth.signOut()
  router.push({ name: 'login' })
}

function toggleSidebar() {
  isSidebarCollapsed.value = !isSidebarCollapsed.value
  persistSidebarCollapsedState(isSidebarCollapsed.value)
}

function readSidebarCollapsedState() {
  try {
    return window.localStorage.getItem(SIDEBAR_COLLAPSED_STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

function persistSidebarCollapsedState(collapsed: boolean) {
  try {
    window.localStorage.setItem(SIDEBAR_COLLAPSED_STORAGE_KEY, String(collapsed))
  } catch {
    // Ignore storage write failures and keep the in-memory state.
  }
}
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  background: var(--layout-content-bg);
}

.sidebar {
  width: 254px;
  flex-shrink: 0;
  background: var(--layout-sidebar-bg);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  transition: width 0.24s ease;
  overflow: hidden;
}

.sidebar.collapsed {
  width: 84px;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid var(--glass-border);
}

.header-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 0.9rem;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 0.45rem;
}

.logo {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.theme-btn {
  width: 32px;
  height: 32px;
}

.header-icon {
  font-size: 1rem;
}

.new-chat-btn {
  width: 100%;
  padding: 0.4rem 0.5rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: background 0.2s ease, opacity 0.2s ease;
}

.button-icon {
  font-size: 1rem;
}

.new-chat-btn:hover {
  background: var(--accent-secondary);
}

.new-chat-btn.compact {
  padding: 0.55rem;
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.3rem;
}

.conversation-item {
  display: block;
  padding: 0.4rem 0.4rem;
  border-radius: 6px;
  text-decoration: none;
  color: var(--text-secondary);
  transition: background 0.2s ease, border-color 0.2s ease, color 0.2s ease;
  cursor: pointer;
  border: 1px solid transparent;
  background: transparent;
}

.conversation-item:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.conversation-item.active {
  background: color-mix(in srgb, var(--accent-primary) 15%, var(--surface-muted));
  color: var(--text-primary);
}

.conv-title {
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-meta {
  font-size: 0.75rem;
  opacity: 0.7;
}

.empty-state {
  text-align: center;
  padding: 2rem 1rem;
  color: var(--text-secondary);
  font-size: 0.875rem;
}

.nav-menu {
  padding: 0.5rem 0.75rem;
  /* border-top: 1px solid var(--glass-border); */
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.6rem 0.6rem;
  margin-bottom: 0.25rem;
  border-radius: 12px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  transition: background 0.2s ease, color 0.2s ease;
}

.nav-icon {
  flex: none;
  font-size: 1.05rem;
}

.nav-item:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background: var(--surface-muted);
  color: var(--accent-primary);
}

.nav-label {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--glass-border);
}

.footer-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.5rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  min-width: 0;
  flex: 1;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--surface-muted);
  color: var(--text-primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.9rem;
  border: 1px solid var(--glass-border);
}

.user-details {
  flex: 1;
  min-width: 0;
}

.user-name {
  font-size: 0.875rem;
  font-weight: 500;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.user-role {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.icon-btn {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: var(--layout-sidebar-bg);
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: background 0.2s ease, color 0.2s ease, border-color 0.2s ease;
}

.icon-btn:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.logout-btn {
  height: 32px;
  padding: 0 0.65rem;
  border-radius: 8px;
  border: 1px solid var(--glass-border);
  background: var(--layout-sidebar-bg);
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.35rem;
  transition: background 0.2s ease, color 0.2s ease;
  white-space: nowrap;
  font-size: 0.8rem;
  flex-shrink: 0;
}

.logout-icon {
  flex: none;
  font-size: 0.95rem;
}

.logout-btn:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.main-content {
  flex: 1;
  overflow: hidden;
  background: var(--layout-content-bg);
}

.sidebar.collapsed .sidebar-header {
  padding: 1rem 0.75rem;
}

.sidebar.collapsed .header-top {
  flex-direction: column;
  align-items: center;
  gap: 0.65rem;
}

.sidebar.collapsed .header-actions {
  flex-direction: column;
}

.sidebar.collapsed .conversations-list {
  display: none;
}

.sidebar.collapsed .new-chat-btn.compact {
  width: 36px;
  height: 36px;
  padding: 0;
  margin: 0 auto;
  border-radius: 10px;
}

.sidebar.collapsed .nav-menu {
  padding: 0.75rem 0.55rem;
}

.sidebar.collapsed .nav-item {
  justify-content: center;
  padding: 0.7rem 0;
}

.sidebar.collapsed .nav-label {
  display: none;
}

.sidebar.collapsed .footer-row {
  flex-direction: column;
}

.sidebar.collapsed .user-info {
  justify-content: center;
}

.sidebar.collapsed .user-details {
  display: none;
}

.sidebar.collapsed .logout-btn {
  width: 100%;
  justify-content: center;
  padding: 0;
}

@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    max-width: 272px;
    position: fixed;
    left: -272px;
    z-index: 100;
    transition: left 0.3s ease;
  }

  .sidebar.open {
    left: 0;
  }

  .sidebar.collapsed {
    width: 100%;
    max-width: 272px;
  }
}
</style>
