<template>
  <div class="app-layout">
    <aside class="sidebar">
      <div class="sidebar-header">
        <h1 class="logo">MrChat</h1>
        <button class="new-chat-btn" @click="createConversation">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/>
          </svg>
          新对话
        </button>
      </div>

      <div class="conversations-list">
        <RouterLink
          v-for="conv in conversations"
          :key="conv.id"
          :to="`/chat/${conv.id}`"
          class="conversation-item"
          :class="{ active: currentConversationId === conv.id }"
        >
          <div class="conv-title">{{ conv.title || '新对话' }}</div>
          <div class="conv-meta">{{ conv.message_count }} 条消息</div>
        </RouterLink>
        <div v-if="conversations.length === 0" class="empty-state">暂无对话</div>
      </div>

      <nav class="nav-menu">
        <RouterLink to="/chat" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" stroke-width="2"/>
          </svg>
          对话
        </RouterLink>
        <RouterLink to="/usage" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <line x1="12" y1="1" x2="12" y2="23" stroke-width="2"/><path d="M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" stroke-width="2"/>
          </svg>
          用量
        </RouterLink>
        <RouterLink to="/settings/profile" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <circle cx="12" cy="12" r="3" stroke-width="2"/><path d="M12 1v6m0 6v6" stroke-width="2"/>
            <path d="m4.93 4.93 4.24 4.24m5.66 5.66 4.24 4.24m0-16.97-4.24 4.24m-5.66 5.66L4.93 19.07" stroke-width="2"/>
          </svg>
          设置
        </RouterLink>
        <RouterLink to="/admin/upstreams" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2" stroke-width="2"/><line x1="9" y1="9" x2="15" y2="9" stroke-width="2"/>
            <line x1="9" y1="15" x2="15" y2="15" stroke-width="2"/>
          </svg>
          管理
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <div class="user-info" v-if="auth.user">
          <div class="user-avatar">{{ auth.user.username[0].toUpperCase() }}</div>
          <div class="user-details">
            <div class="user-name">{{ auth.user.username }}</div>
            <div class="user-role">{{ auth.user.role }}</div>
          </div>
        </div>
        <div class="footer-actions">
          <button class="icon-btn" @click="toggleTheme" title="切换主题">
            <svg v-if="isDark()" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <circle cx="12" cy="12" r="5" stroke-width="2"/><line x1="12" y1="1" x2="12" y2="3" stroke-width="2"/>
              <line x1="12" y1="21" x2="12" y2="23" stroke-width="2"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64" stroke-width="2"/>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" stroke-width="2"/><line x1="1" y1="12" x2="3" y2="12" stroke-width="2"/>
              <line x1="21" y1="12" x2="23" y2="12" stroke-width="2"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36" stroke-width="2"/>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" stroke-width="2"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" stroke-width="2"/>
            </svg>
          </button>
          <button class="logout-btn" @click="handleSignOut" title="退出登录">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
              <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" stroke-width="2"/><polyline points="16 17 21 12 16 7" stroke-width="2"/>
              <line x1="21" y1="12" x2="9" y2="12" stroke-width="2"/>
            </svg>
            <span>退出登录</span>
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

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)

onMounted(async () => {
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
</script>

<style scoped>
.app-layout {
  display: flex;
  height: 100vh;
  background: var(--layout-content-bg);
}

.sidebar {
  width: 272px;
  background: var(--layout-sidebar-bg);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 1.25rem 1rem;
  border-bottom: 1px solid var(--glass-border);
}

.logo {
  font-size: 1.25rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.9rem;
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

.new-chat-btn:hover {
  background: var(--accent-secondary);
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.75rem;
  display: flex;
  flex-direction: column;
  gap: 0.35rem;
}

.conversation-item {
  display: block;
  padding: 0.85rem 0.9rem;
  border-radius: 12px;
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
  background: var(--surface-muted);
  color: var(--text-primary);
  border-color: var(--glass-border);
  box-shadow: inset 3px 0 0 var(--accent-primary);
}

.conv-title {
  font-size: 0.9rem;
  font-weight: 500;
  margin-bottom: 0.25rem;
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
  padding: 0.5rem 0.75rem 0.75rem;
  border-top: 1px solid var(--glass-border);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.8rem 0.9rem;
  margin-bottom: 0.25rem;
  border-radius: 12px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  transition: background 0.2s ease, color 0.2s ease;
}

.nav-item:hover {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background: var(--surface-muted);
  color: var(--accent-primary);
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  width: 100%;
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

.footer-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.55rem;
  width: 100%;
}

.logout-btn {
  height: 36px;
  padding: 0 0.85rem;
  border-radius: 10px;
  border: 1px solid var(--glass-border);
  background: var(--layout-sidebar-bg);
  color: var(--text-secondary);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  gap: 0.45rem;
  transition: background 0.2s ease, color 0.2s ease;
  white-space: nowrap;
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
}
</style>
