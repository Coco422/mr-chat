<template>
  <section class="chat-page">
    <header class="page-header">
      <h1>Chat</h1>
      <div class="header-actions">
        <button type="button" class="action-btn" @click="reloadAll" :disabled="loading">刷新</button>
        <button type="button" class="action-btn warning" @click="stopStreaming" :disabled="!sending">停止生成</button>
      </div>
    </header>

    <p v-if="errorMessage" class="error-banner">{{ errorMessage }}</p>

    <div class="meta-card">
      <label class="meta-item">
        当前会话
        <span class="meta-value">{{ currentConversationId || 'new' }}</span>
      </label>
      <label class="meta-item">
        模型
        <select v-model="selectedModelID" class="model-select">
          <option value="">请选择模型</option>
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.display_name }} ({{ model.model_key }})
          </option>
        </select>
      </label>
    </div>

    <section class="messages-card">
      <h2>消息列表</h2>
      <p v-if="loadingMessages" class="empty">消息加载中...</p>
      <ul v-else-if="messages.length > 0" class="message-list">
        <li v-for="message in messages" :key="message.id" class="message-item" :class="`role-${message.role}`">
          <div class="message-meta">{{ message.role }} / {{ message.status }} / {{ message.created_at }}</div>
          <div class="message-content">{{ message.content }}</div>
          <div v-if="message.reasoning_content" class="message-extra">reasoning: {{ message.reasoning_content }}</div>
          <div v-if="message.finish_reason" class="message-extra">finish_reason: {{ message.finish_reason }}</div>
        </li>
      </ul>
      <p v-else class="empty">暂无消息</p>
    </section>

    <section class="composer-card">
      <h2>发送消息</h2>
      <form @submit.prevent="sendMessage" class="composer-form">
        <div class="input-wrap">
          <textarea v-model="inputMessage" rows="6" placeholder="输入消息..." class="message-input" />
        </div>
        <button type="submit" class="send-btn" :disabled="sending || !inputMessage.trim()">发送</button>
      </form>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError, apiBaseUrl, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

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
  reasoning_content: string
  status: string
  finish_reason?: string | null
  usage?: Record<string, unknown>
  created_at: string
}

interface StreamStartEvent {
  type: 'response.start'
  request_id: string
  conversation_id: string
  assistant_message_id: string
}

interface StreamDeltaEvent {
  type: 'response.delta'
  delta: {
    content?: string
  }
}

interface StreamReasoningEvent {
  type: 'reasoning.delta'
  delta: {
    reasoning_content?: string
  }
}

interface StreamCompletedEvent {
  type: 'response.completed'
  request_id?: string
  conversation_id?: string
  assistant_message_id?: string
  finish_reason?: string
  usage?: {
    prompt_tokens: number
    completion_tokens: number
    total_tokens: number
  }
}

interface StreamFailedEvent {
  type: 'response.failed'
  request_id?: string
  conversation_id?: string
  assistant_message_id?: string
  error?: {
    code?: string
    message?: string
  }
}

type StreamEvent =
  | StreamStartEvent
  | StreamDeltaEvent
  | StreamReasoningEvent
  | StreamCompletedEvent
  | StreamFailedEvent

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const loading = ref(false)
const loadingMessages = ref(false)
const sending = ref(false)
const errorMessage = ref('')
const models = ref<UserModel[]>([])
const conversations = ref<ConversationSummary[]>([])
const messages = ref<MessageItem[]>([])
const selectedModelID = ref('')
const streamAbortController = ref<AbortController | null>(null)

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)

onMounted(async () => {
  await reloadAll()
})

watch(currentConversationId, async () => {
  if (sending.value) {
    return
  }
  await reloadAll()
})

const inputMessage = ref('')

async function reloadAll() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [modelsResponse, conversationsResponse] = await Promise.all([
      apiRequest<UserModel[]>('/models', {
        accessToken: auth.accessToken
      }),
      apiRequest<ConversationSummary[]>('/conversations', {
        accessToken: auth.accessToken
      })
    ])

    models.value = modelsResponse.data
    conversations.value = conversationsResponse.data

    const currentConversation = conversations.value.find((item) => item.id === currentConversationId.value) ?? null
    if (currentConversation?.model_id) {
      selectedModelID.value = currentConversation.model_id
    } else if (!selectedModelID.value && models.value.length > 0) {
      selectedModelID.value = models.value[0].id
    }

    if (currentConversationId.value) {
      await loadMessages(currentConversationId.value)
    } else {
      messages.value = []
    }
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function loadMessages(conversationID: string) {
  loadingMessages.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<MessageItem[]>(`/conversations/${conversationID}/messages`, {
      accessToken: auth.accessToken
    })
    messages.value = data
  } catch (error) {
    messages.value = []
    errorMessage.value = toErrorMessage(error)
  } finally {
    loadingMessages.value = false
  }
}

async function sendMessage() {
  if (sending.value) {
    return
  }

  const content = inputMessage.value.trim()
  if (!content) {
    return
  }
  if (!selectedModelID.value) {
    errorMessage.value = '请先选择模型'
    return
  }

  sending.value = true
  errorMessage.value = ''
  const controller = new AbortController()
  streamAbortController.value = controller
  const createdAt = new Date().toISOString()
  const userTempID = `local-user-${Date.now()}`
  let assistantMessageID = `local-assistant-${Date.now()}`
  let nextConversationID = currentConversationId.value

  messages.value = [
    ...messages.value,
    {
      id: userTempID,
      role: 'user',
      content,
      reasoning_content: '',
      status: 'completed',
      created_at: createdAt
    },
    {
      id: assistantMessageID,
      role: 'assistant',
      content: '',
      reasoning_content: '',
      status: 'streaming',
      finish_reason: null,
      usage: {},
      created_at: createdAt
    }
  ]
  inputMessage.value = ''

  try {
    await streamChatCompletion(
      {
        conversation_id: currentConversationId.value || null,
        model_id: selectedModelID.value,
        stream: true,
        messages: [
          {
            role: 'user',
            content
          }
        ]
      },
      auth.accessToken,
      controller.signal,
      (event) => {
        switch (event.type) {
          case 'response.start':
            nextConversationID = event.conversation_id
            replaceMessageID(assistantMessageID, event.assistant_message_id)
            assistantMessageID = event.assistant_message_id
            break
          case 'response.delta':
            patchMessage(assistantMessageID, {
              content: currentMessageValue(assistantMessageID, 'content') + (event.delta.content ?? ''),
              status: 'streaming'
            })
            break
          case 'reasoning.delta':
            patchMessage(assistantMessageID, {
              reasoning_content:
                currentMessageValue(assistantMessageID, 'reasoning_content') + (event.delta.reasoning_content ?? ''),
              status: 'streaming'
            })
            break
          case 'response.completed':
            patchMessage(assistantMessageID, {
              status: 'completed',
              finish_reason: event.finish_reason ?? null,
              usage: event.usage ?? {}
            })
            break
          case 'response.failed':
            patchMessage(assistantMessageID, {
              status: 'failed'
            })
            errorMessage.value = `${event.error?.code ?? 'CHAT_STREAM_FAILED'}: ${event.error?.message ?? 'Streaming failed'}`
            break
        }
      }
    )

    window.dispatchEvent(new Event('mrchat:conversations:refresh'))
    if (nextConversationID && nextConversationID !== currentConversationId.value) {
      await router.push(`/chat/${nextConversationID}`)
    }
    await reloadAll()
  } catch (error) {
    if (isAbortError(error)) {
      patchMessage(assistantMessageID, { status: 'cancelled' })
      errorMessage.value = '已停止生成'
    } else {
      patchMessage(assistantMessageID, { status: 'failed' })
      errorMessage.value = toErrorMessage(error)
    }

    if (nextConversationID && nextConversationID !== currentConversationId.value) {
      await router.push(`/chat/${nextConversationID}`)
    }
    await reloadAll()
  } finally {
    streamAbortController.value = null
    sending.value = false
  }
}

function stopStreaming() {
  streamAbortController.value?.abort()
}

function patchMessage(messageID: string, patch: Partial<MessageItem>) {
  messages.value = messages.value.map((item) => (item.id === messageID ? { ...item, ...patch } : item))
}

function replaceMessageID(currentID: string, nextID: string) {
  messages.value = messages.value.map((item) => (item.id === currentID ? { ...item, id: nextID } : item))
}

function currentMessageValue(messageID: string, field: 'content' | 'reasoning_content') {
  return messages.value.find((item) => item.id === messageID)?.[field] ?? ''
}

async function streamChatCompletion(
  body: Record<string, unknown>,
  accessToken: string,
  signal: AbortSignal,
  onEvent: (event: StreamEvent) => void
) {
  const response = await fetch(`${apiBaseUrl}/api/v1/chat/completions`, {
    method: 'POST',
    headers: {
      Accept: 'text/event-stream',
      'Content-Type': 'application/json',
      Authorization: `Bearer ${accessToken}`
    },
    credentials: 'include',
    body: JSON.stringify(body),
    signal
  })

  if (!response.ok || !response.body) {
    throw await toStreamError(response)
  }

  const reader = response.body.getReader()
  const decoder = new TextDecoder()
  let buffer = ''

  while (true) {
    const { value, done } = await reader.read()
    if (done) {
      break
    }

    buffer += decoder.decode(value, { stream: true })
    const chunks = buffer.split('\n\n')
    buffer = chunks.pop() ?? ''

    for (const chunk of chunks) {
      const event = parseSSEChunk(chunk)
      if (event === '[DONE]') {
        return
      }
      if (event) {
        onEvent(event)
      }
    }
  }

  buffer += decoder.decode()
  const lastEvent = parseSSEChunk(buffer)
  if (lastEvent && lastEvent !== '[DONE]') {
    onEvent(lastEvent)
  }
}

function parseSSEChunk(chunk: string): StreamEvent | '[DONE]' | null {
  const lines = chunk
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line.startsWith('data:'))

  if (lines.length === 0) {
    return null
  }

  const payload = lines.map((line) => line.slice(5).trim()).join('\n')
  if (!payload) {
    return null
  }
  if (payload === '[DONE]') {
    return '[DONE]'
  }

  return JSON.parse(payload) as StreamEvent
}

async function toStreamError(response: Response) {
  const contentType = response.headers.get('content-type') ?? ''
  if (contentType.includes('application/json')) {
    const payload = (await response.json()) as {
      error?: {
        code?: string
        message?: string
        details?: unknown
      }
    }
    return new ApiError(
      payload.error?.message ?? `Request failed with status ${response.status}`,
      payload.error?.code ?? 'HTTP_ERROR',
      response.status,
      payload.error?.details
    )
  }

  return new ApiError(`Request failed with status ${response.status}`, 'HTTP_ERROR', response.status)
}

function isAbortError(error: unknown) {
  return error instanceof DOMException && error.name === 'AbortError'
}

function toErrorMessage(error: unknown) {
  if (isAbortError(error)) {
    return '已停止生成'
  }
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>

<style scoped>
.chat-page {
  height: 100%;
  display: grid;
  grid-template-rows: auto auto 1fr auto;
  gap: 1rem;
  padding: 1.25rem;
  background: var(--bg-primary);
  color: var(--text-primary);
}

.page-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
}

.page-header h1 {
  margin: 0;
  font-size: 1.4rem;
}

.header-actions {
  display: flex;
  gap: 0.6rem;
}

.action-btn {
  border: 1px solid var(--input-border);
  background: var(--input-bg);
  color: var(--text-primary);
  border-radius: 8px;
  padding: 0.45rem 0.8rem;
  cursor: pointer;
}

.action-btn.warning {
  border-color: color-mix(in srgb, var(--error-color) 45%, var(--input-border));
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.error-banner {
  margin: 0;
  padding: 0.7rem 0.85rem;
  border-radius: 8px;
  border: 1px solid color-mix(in srgb, var(--error-color) 45%, transparent);
  background: color-mix(in srgb, var(--error-color) 10%, transparent);
  color: var(--error-color);
}

.meta-card,
.messages-card,
.composer-card {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 0.9rem;
}

.meta-card {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 0.9rem;
}

.meta-item {
  display: flex;
  flex-direction: column;
  gap: 0.45rem;
  color: var(--text-secondary);
  font-size: 0.88rem;
}

.meta-value {
  color: var(--text-primary);
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'Courier New', monospace;
}

.model-select,
.message-input {
  width: 100%;
  border: 1px solid var(--input-border);
  background: var(--bg-primary);
  color: var(--text-primary);
  border-radius: 8px;
  padding: 0.6rem 0.7rem;
}

.messages-card {
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.messages-card h2,
.composer-card h2 {
  margin: 0 0 0.8rem;
  font-size: 1rem;
}

.message-list {
  list-style: none;
  margin: 0;
  padding: 0;
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
  overflow: auto;
}

.message-item {
  border: 1px solid var(--input-border);
  border-radius: 10px;
  padding: 0.75rem;
  background: var(--bg-primary);
}

.message-item.role-user {
  border-left: 3px solid var(--accent-primary);
}

.message-item.role-assistant {
  border-left: 3px solid var(--accent-secondary);
}

.message-meta {
  color: var(--text-secondary);
  font-size: 0.78rem;
  margin-bottom: 0.35rem;
}

.message-content {
  white-space: pre-wrap;
  line-height: 1.55;
}

.message-extra {
  margin-top: 0.35rem;
  color: var(--text-secondary);
  font-size: 0.83rem;
  white-space: pre-wrap;
}

.empty {
  margin: 0;
  color: var(--text-secondary);
}

.composer-form {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

.message-input {
  resize: vertical;
  min-height: 124px;
}

.send-btn {
  align-self: flex-end;
  border: none;
  background: var(--accent-primary);
  color: #fff;
  border-radius: 8px;
  padding: 0.6rem 1.2rem;
  cursor: pointer;
}

.send-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

@media (max-width: 900px) {
  .chat-page {
    padding: 0.9rem;
  }

  .meta-card {
    grid-template-columns: 1fr;
  }

  .page-header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
