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
        <button class="icon-btn" @click="handleSignOut" title="退出登录">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" stroke-width="2"/><polyline points="16 17 21 12 16 7" stroke-width="2"/>
            <line x1="21" y1="12" x2="9" y2="12" stroke-width="2"/>
          </svg>
        </button>
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
import { apiRequest } from '@/lib/api'

interface ConversationSummary {
  id: string
  title: string
  model_id: string | null
  message_count: number
  status: string
}

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
    const { data } = await apiRequest<ConversationSummary[]>('/conversations', {
      accessToken: auth.accessToken
    })
    conversations.value = data
  } catch (error) {
    console.error('Failed to load conversations:', error)
  }
}

async function createConversation() {
  try {
    const { data } = await apiRequest<ConversationSummary>('/conversations', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: { title: '新对话', model_id: null }
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
  background: var(--bg-primary);
}

.sidebar {
  width: 280px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 1.5rem 1rem;
  border-bottom: 1px solid var(--glass-border);
}

.logo {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 1rem;
}

.new-chat-btn {
  width: 100%;
  padding: 0.75rem 1rem;
  background: var(--accent-primary);
  color: white;
  border: none;
  border-radius: 8px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 0.5rem;
  transition: all 0.2s ease;
}

.new-chat-btn:hover {
  background: var(--accent-secondary);
  transform: translateY(-1px);
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
  padding: 0.5rem;
}

.conversation-item {
  display: block;
  padding: 0.875rem 1rem;
  margin-bottom: 0.25rem;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-secondary);
  transition: all 0.2s ease;
  cursor: pointer;
}

.conversation-item:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.conversation-item.active {
  background: var(--input-bg);
  color: var(--text-primary);
  border-left: 3px solid var(--accent-primary);
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
  padding: 0.5rem;
  border-top: 1px solid var(--glass-border);
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  margin-bottom: 0.25rem;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background: var(--input-bg);
  color: var(--accent-primary);
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--glass-border);
  display: flex;
  align-items: center;
  gap: 0.5rem;
}

.user-info {
  flex: 1;
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.user-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--accent-primary);
  color: white;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.9rem;
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
  border-radius: 8px;
  background: transparent;
  border: 1px solid var(--glass-border);
  color: var(--text-secondary);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
}

.icon-btn:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.main-content {
  flex: 1;
  overflow: hidden;
}

@media (max-width: 768px) {
  .sidebar {
    width: 100%;
    max-width: 280px;
    position: fixed;
    left: -280px;
    z-index: 100;
    transition: left 0.3s ease;
  }

  .sidebar.open {
    left: 0;
  }
}
</style>
