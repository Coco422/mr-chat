<template>
  <div class="chat-page">
    <div class="chat-header">
      <div class="chat-title">
        <h2>{{ currentConversation?.title || '新对话' }}</h2>
        <span class="model-badge" v-if="currentConversation?.model_id">
          {{ getModelName(currentConversation.model_id) }}
        </span>
      </div>
      <button class="icon-btn" @click="reloadAll" :disabled="loading" title="刷新">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/>
          <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
        </svg>
      </button>
    </div>

    <div class="messages-container" ref="messagesContainer">
      <div v-if="!currentConversationId" class="welcome-screen">
        <div class="welcome-icon">
          <svg width="64" height="64" viewBox="0 0 64 64" fill="none">
            <circle cx="32" cy="32" r="28" stroke="currentColor" stroke-width="2" opacity="0.3"/>
            <path d="M20 32 L28 24 L36 32 L44 24" stroke="currentColor" stroke-width="2" stroke-linecap="round" opacity="0.5"/>
          </svg>
        </div>
        <h3>开始新对话</h3>
        <p>选择一个模型开始聊天</p>
        <div class="model-grid">
          <button
            v-for="model in models"
            :key="model.id"
            @click="startWithModel(model.id)"
            class="model-card"
          >
            <div class="model-name">{{ model.display_name }}</div>
            <div class="model-key">{{ model.model_key }}</div>
          </button>
        </div>
      </div>

      <div v-else class="messages-list">
        <div v-if="loadingMessages" class="loading-state">加载中...</div>
        <div v-else-if="messages.length === 0" class="empty-messages">暂无消息</div>
        <div v-else>
          <div
            v-for="message in messages"
            :key="message.id"
            class="message"
            :class="{ user: message.role === 'user', assistant: message.role === 'assistant' }"
          >
            <div class="message-avatar">
              {{ message.role === 'user' ? 'U' : 'A' }}
            </div>
            <div class="message-content">
              <div class="message-text">{{ message.content }}</div>
              <div class="message-time">{{ formatTime(message.created_at) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="input-area" v-if="currentConversationId">
      <form @submit.prevent="sendMessage" class="input-form">
        <textarea
          v-model="inputMessage"
          placeholder="输入消息..."
          rows="1"
          @keydown.enter.exact.prevent="sendMessage"
        ></textarea>
        <button type="submit" :disabled="!inputMessage.trim() || sending" class="send-btn">
          <svg v-if="!sending" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="22" y1="2" x2="11" y2="13"/><polygon points="22 2 15 22 11 13 2 9 22 2"/>
          </svg>
          <span v-else class="spinner"></span>
        </button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { apiRequest, ApiError } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

interface UserModel {
  id: string
  model_key: string
  display_name: string
  provider_type: string
}

interface ConversationSummary {
  id: string
  title: string
  model_id: string | null
  message_count: number
  status: string
}

interface MessageItem {
  id: string
  role: string
  content: string
  status: string
  created_at: string
}

const loading = ref(false)
const loadingMessages = ref(false)
const sending = ref(false)
const models = ref<UserModel[]>([])
const messages = ref<MessageItem[]>([])
const currentConversation = ref<ConversationSummary | null>(null)
const inputMessage = ref('')
const messagesContainer = ref<HTMLElement>()

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)

onMounted(async () => {
  await reloadAll()
})

watch(currentConversationId, async (id) => {
  if (!id) {
    messages.value = []
    currentConversation.value = null
    return
  }
  await loadMessages(id)
}, { immediate: true })

async function reloadAll() {
  loading.value = true
  try {
    const { data } = await apiRequest<UserModel[]>('/models', {
      accessToken: auth.accessToken
    })
    models.value = data
  } catch (error) {
    console.error('Failed to load models:', error)
  } finally {
    loading.value = false
  }
}

async function loadMessages(conversationId: string) {
  loadingMessages.value = true
  try {
    const { data } = await apiRequest<MessageItem[]>(`/conversations/${conversationId}/messages`, {
      accessToken: auth.accessToken
    })
    messages.value = data
    await nextTick()
    scrollToBottom()
  } catch (error) {
    console.error('Failed to load messages:', error)
  } finally {
    loadingMessages.value = false
  }
}

async function startWithModel(modelId: string) {
  try {
    const { data } = await apiRequest<ConversationSummary>('/conversations', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: { title: '新对话', model_id: modelId }
    })
    router.push(`/chat/${data.id}`)
  } catch (error) {
    console.error('Failed to create conversation:', error)
  }
}

async function sendMessage() {
  if (!inputMessage.value.trim() || sending.value) return

  sending.value = true
  const content = inputMessage.value
  inputMessage.value = ''

  try {
    await apiRequest(`/conversations/${currentConversationId.value}/messages`, {
      method: 'POST',
      accessToken: auth.accessToken,
      body: { content }
    })
    await loadMessages(currentConversationId.value)
  } catch (error) {
    console.error('Failed to send message:', error)
    inputMessage.value = content
  } finally {
    sending.value = false
  }
}

function getModelName(modelId: string) {
  return models.value.find(m => m.id === modelId)?.display_name || 'Unknown'
}

function formatTime(timestamp: string) {
  return new Date(timestamp).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
}

function scrollToBottom() {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}
</script>

<style scoped>
.chat-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-primary);
}

.chat-header {
  padding: 1rem 1.5rem;
  border-bottom: 1px solid var(--glass-border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: var(--bg-secondary);
}

.chat-title {
  display: flex;
  align-items: center;
  gap: 0.75rem;
}

.chat-title h2 {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.model-badge {
  padding: 0.25rem 0.75rem;
  background: var(--input-bg);
  border-radius: 12px;
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

.icon-btn:hover:not(:disabled) {
  background: var(--input-bg);
  color: var(--text-primary);
}

.icon-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  padding: 1.5rem;
}

.welcome-screen {
  max-width: 600px;
  margin: 0 auto;
  text-align: center;
  padding: 3rem 1rem;
}

.welcome-icon {
  color: var(--text-secondary);
  margin-bottom: 1.5rem;
}

.welcome-screen h3 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.5rem;
}

.welcome-screen p {
  color: var(--text-secondary);
  margin: 0 0 2rem;
}

.model-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 1rem;
}

.model-card {
  padding: 1.25rem;
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.2s ease;
  text-align: left;
}

.model-card:hover {
  border-color: var(--accent-primary);
  transform: translateY(-2px);
}

.model-name {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 0.25rem;
}

.model-key {
  font-size: 0.75rem;
  color: var(--text-secondary);
}

.messages-list {
  max-width: 800px;
  margin: 0 auto;
  width: 100%;
}

.loading-state,
.empty-messages {
  text-align: center;
  padding: 2rem;
  color: var(--text-secondary);
}

.message {
  display: flex;
  gap: 0.75rem;
  margin-bottom: 1.5rem;
}

.message-avatar {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-weight: 600;
  font-size: 0.875rem;
  flex-shrink: 0;
}

.message.user .message-avatar {
  background: var(--accent-primary);
  color: white;
}

.message.assistant .message-avatar {
  background: var(--input-bg);
  color: var(--text-primary);
}

.message-content {
  flex: 1;
  min-width: 0;
}

.message-text {
  padding: 0.875rem 1rem;
  border-radius: 12px;
  font-size: 0.9rem;
  line-height: 1.6;
  color: var(--text-primary);
  word-wrap: break-word;
}

.message.user .message-text {
  background: var(--accent-primary);
  color: white;
}

.message.assistant .message-text {
  background: var(--input-bg);
}

.message-time {
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-top: 0.5rem;
  padding: 0 0.25rem;
}

.input-area {
  padding: 1rem 1.5rem;
  border-top: 1px solid var(--glass-border);
  background: var(--bg-secondary);
}

.input-form {
  max-width: 800px;
  margin: 0 auto;
  display: flex;
  gap: 0.75rem;
  align-items: flex-end;
}

.input-form textarea {
  flex: 1;
  padding: 0.875rem 1rem;
  background: var(--input-bg);
  border: 2px solid var(--input-border);
  border-radius: 12px;
  color: var(--text-primary);
  font-size: 0.9rem;
  font-family: inherit;
  resize: none;
  max-height: 200px;
  transition: all 0.2s ease;
  outline: none;
}

.input-form textarea:focus {
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px var(--accent-glow);
}

.send-btn {
  width: 44px;
  height: 44px;
  border-radius: 12px;
  background: var(--accent-primary);
  border: none;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s ease;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
  transform: translateY(-1px);
}

.send-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spinner {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255,255,255,0.3);
  border-top-color: white;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
</style>
