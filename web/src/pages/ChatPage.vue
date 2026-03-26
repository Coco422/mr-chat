<template>
  <section class="chat-page">
    <header class="page-toolbar">
      <div class="model-panel">
        <el-select
          v-model="selectedModelID"
          class="model-select"
          :style="modelSelectStyle"
          popper-class="chat-model-popper"
          placeholder="请选择模型"
          size="large"
        >
          <el-option
            v-for="model in models"
            :key="model.id"
            :label="model.display_name"
            :value="model.id"
          >
            <div class="model-option">
              <span class="model-option-name">{{ model.display_name }}</span>
            </div>
          </el-option>
        </el-select>
      </div>

      <!-- <div class="toolbar-actions">
        <button type="button" class="action-btn" @click="reloadAll" :disabled="loading">刷新</button>
      </div> -->
    </header>

    <div class="chat-shell">

      <div v-if="loadingMessages" class="chat-body chat-body-loading">
        <p class="state-copy">消息加载中...</p>
      </div>

      <div v-else class="chat-body" :class="{ 'chat-body-empty': !hasMessages }">
        <div v-if="hasMessages" ref="messagesViewport" class="messages-viewport" @scroll="handleMessagesScroll">
          <ul class="message-list">
            <li
              v-for="(message, index) in messages"
              :key="message.id"
              class="message-row"
              :class="`role-${message.role}`"
            >
              <article class="message-bubble">
                <div class="message-meta">
                  <span class="message-role">{{ roleLabel(message.role) }}</span>
                  <time>{{ formatMessageTime(message.created_at) }}</time>
                </div>

                <template v-if="message.role === 'assistant'">
                  <section v-if="message.reasoning_content" class="message-panel thinking-panel">
                    <div class="panel-label">思考模式</div>
                    <div class="markdown-body" v-html="renderMarkdown(message.reasoning_content)"></div>
                  </section>

                  <section class="message-panel answer-panel" :class="{ streaming: message.status === 'streaming' }">
                    <div v-if="message.reasoning_content" class="panel-label">正式回答</div>
                    <div
                      v-if="assistantHasContent(message)"
                      class="markdown-body"
                      v-html="renderMarkdown(assistantDisplayContent(message))"
                    ></div>
                    <div
                      v-if="message.status === 'streaming'"
                      class="streaming-indicator"
                      :class="{ 'streaming-indicator-empty': !assistantHasContent(message) }"
                      aria-label="生成中"
                    >
                      <span class="streaming-dot"></span>
                      <span class="streaming-dot"></span>
                      <span class="streaming-dot"></span>
                    </div>
                  </section>
                </template>

                <template v-else-if="editingMessageID === message.id">
                  <div class="edit-box">
                    <textarea
                      ref="editInput"
                      v-model="editingMessageDraft"
                      rows="1"
                      class="message-edit-input"
                      @input="resizeEditInput"
                      @keydown.enter.exact.prevent="confirmEdit(index)"
                      @compositionstart="isEditingComposing = true"
                      @compositionend="handleEditingCompositionEnd"
                    ></textarea>
                  </div>
                </template>

                <div v-else class="user-content">{{ message.content }}</div>

                <div v-if="shouldShowStatus(message.status)" class="message-status">{{ statusLabel(message.status) }}</div>

                <div class="message-tools">
                  <template v-if="message.role === 'assistant'">
                    <button
                      type="button"
                      class="tool-btn"
                      title="复制"
                      :disabled="!canCopyMessage(message)"
                      @click="copyMessage(message)"
                    >
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <rect x="9" y="9" width="11" height="11" rx="2"></rect>
                        <path d="M6 15H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v1"></path>
                      </svg>
                    </button>
                    <button type="button" class="tool-btn" title="重新生成" :disabled="sending" @click="regenerateFromAssistant(index)">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M3 12a9 9 0 0 1 15.3-6.36L21 8"></path>
                        <path d="M21 3v5h-5"></path>
                        <path d="M21 12a9 9 0 0 1-15.3 6.36L3 16"></path>
                        <path d="M8 16H3v5"></path>
                      </svg>
                    </button>
                  </template>

                  <template v-else-if="editingMessageID === message.id">
                    <button type="button" class="tool-btn confirm-btn" title="确认修改" :disabled="sending" @click="confirmEdit(index)">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M20 6 9 17l-5-5"></path>
                      </svg>
                    </button>
                    <button type="button" class="tool-btn" title="取消" @click="cancelEditing">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M18 6 6 18"></path>
                        <path d="m6 6 12 12"></path>
                      </svg>
                    </button>
                  </template>

                  <template v-else>
                    <button
                      type="button"
                      class="tool-btn"
                      title="复制"
                      :disabled="!canCopyMessage(message)"
                      @click="copyMessage(message)"
                    >
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <rect x="9" y="9" width="11" height="11" rx="2"></rect>
                        <path d="M6 15H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v1"></path>
                      </svg>
                    </button>
                    <button type="button" class="tool-btn" title="编辑" @click="startEditingMessage(message)">
                      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8">
                        <path d="M12 20h9"></path>
                        <path d="M16.5 3.5a2.12 2.12 0 1 1 3 3L7 19l-4 1 1-4 12.5-12.5z"></path>
                      </svg>
                    </button>
                  </template>
                </div>
              </article>
            </li>
          </ul>
        </div>

        <div v-else class="empty-state">
          <h1>开始新的对话</h1>
          <p class="empty-state-typewriter">
            <span>{{ emptyStateTypewriterText }}</span>
            <span class="typewriter-caret" aria-hidden="true"></span>
          </p>
        </div>

        <form
          @submit.prevent="sendMessage"
          class="composer-form"
          :class="{
            'composer-form-centered': !hasMessages,
            'composer-form-docked': hasMessages
          }"
        >
          <div class="composer-row">
            <div class="composer-input-wrap">
              <textarea
                ref="composerInput"
                v-model="inputMessage"
                rows="1"
                placeholder="输入消息..."
                class="message-input"
                @input="resizeComposerInput"
                @keydown.enter.exact.prevent="handleComposerEnter"
                @compositionstart="isComposing = true"
                @compositionend="handleComposerCompositionEnd"
              ></textarea>
            </div>
            <button
              type="button"
              class="send-btn"
              :class="{ 'send-btn-stop': sending }"
              :disabled="!sending && !inputMessage.trim()"
              :title="sending ? '停止生成' : '发送'"
              @click="handlePrimaryAction"
            >
              <svg v-if="sending" viewBox="0 0 24 24" aria-hidden="true">
                <rect x="7" y="7" width="10" height="10" rx="2.2" fill="#ffffff"></rect>
              </svg>
              <svg v-else viewBox="0 0 24 24" aria-hidden="true">
                <path fill="#ffffff" d="M12 4.5c.4 0 .78.16 1.06.44l5 5a1.5 1.5 0 0 1-2.12 2.12l-2.44-2.44V18a1.5 1.5 0 0 1-3 0V9.62l-2.44 2.44a1.5 1.5 0 0 1-2.12-2.12l5-5A1.5 1.5 0 0 1 12 4.5Z"></path>
              </svg>
            </button>
          </div>
        </form>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify'
import { ElMessage } from 'element-plus'
import MarkdownIt from 'markdown-it'
import { computed, nextTick, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import { ApiError } from '@/lib/api'
import {
  listConversations,
  listConversationMessages,
  listModels,
  streamChatCompletion,
  type ConversationSummary,
  type MessageItem,
  type UserModel
} from '@/api/chat'
import { useAuthStore } from '@/stores/auth'

type RequestMessage = {
  role: string
  content: string
}

const markdown = new MarkdownIt({
  html: false,
  breaks: true,
  linkify: true
})

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const loading = ref(false)
const loadingMessages = ref(false)
const sending = ref(false)
const models = ref<UserModel[]>([])
const conversations = ref<ConversationSummary[]>([])
const messages = ref<MessageItem[]>([])
const selectedModelID = ref('')
const streamAbortController = ref<AbortController | null>(null)
const inputMessage = ref('')
const messagesViewport = ref<HTMLElement | null>(null)
const composerInput = ref<HTMLTextAreaElement | null>(null)
const editInput = ref<HTMLTextAreaElement | null>(null)
const isComposing = ref(false)
const isEditingComposing = ref(false)
const editingMessageID = ref('')
const editingMessageDraft = ref('')
const emptyStateTypewriterText = ref('')
const typewriterTimer = ref<number | null>(null)
const typewriterPromptIndex = ref(0)
const typewriterCharIndex = ref(0)
const typewriterDeleting = ref(false)
const autoScrollPinned = ref(true)
let pendingScrollBehavior: ScrollBehavior = 'auto'
let scrollScheduled = false
const scrollBottomThreshold = 72

const emptyStatePrompts = [
  '选择模型后，在这里直接输入你的问题。',
  '试试让它帮你整理方案、写代码或者润色文案。',
  '也可以直接丢一个需求，让它一步步往下做。'
]

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)
const hasMessages = computed(() => messages.value.length > 0)
const selectedModelLabel = computed(
  () => models.value.find((item) => item.id === selectedModelID.value)?.display_name || '请选择模型'
)
const modelSelectStyle = computed(() => {
  const widthCh = Math.min(Math.max(selectedModelLabel.value.length + 4, 11), 28)
  return {
    width: `${widthCh}ch`
  }
})

onMounted(async () => {
  await reloadAll()
  resizeComposerInput()
})

onBeforeUnmount(() => {
  stopEmptyStateTypewriter()
})

watch(currentConversationId, async () => {
  cancelEditing()
  if (sending.value) {
    return
  }
  await reloadAll()
})

watch(inputMessage, () => {
  resizeComposerInput()
})

watch(
  [hasMessages, loadingMessages],
  ([nextHasMessages, nextLoadingMessages]) => {
    if (nextHasMessages || nextLoadingMessages) {
      stopEmptyStateTypewriter()
      return
    }
    startEmptyStateTypewriter()
  },
  { immediate: true }
)

async function reloadAll() {
  loading.value = true

  try {
    const [modelsResponse, conversationsResponse] = await Promise.all([
      listModels(auth.accessToken),
      listConversations(auth.accessToken)
    ])

    models.value = modelsResponse
    conversations.value = conversationsResponse

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
    ElMessage.error(toErrorMessage(error))
  } finally {
    loading.value = false
  }
}

async function loadMessages(conversationID: string) {
  loadingMessages.value = true

  try {
    messages.value = await listConversationMessages(auth.accessToken, conversationID)
    scrollMessagesToBottom('auto', true)
  } catch (error) {
    messages.value = []
    ElMessage.error(toErrorMessage(error))
  } finally {
    loadingMessages.value = false
  }
}

async function sendMessage() {
  const content = inputMessage.value.trim()
  if (!content) {
    return
  }

  inputMessage.value = ''

  await runCompletion({
    conversationID: currentConversationId.value || null,
    requestMessages: [
      {
        role: 'user',
        content
      }
    ],
    optimisticBaseMessages: messages.value,
    optimisticUserContent: content
  })
}

async function runCompletion(options: {
  conversationID: string | null
  requestMessages: RequestMessage[]
  optimisticBaseMessages: MessageItem[]
  optimisticUserContent?: string
}) {
  if (sending.value) {
    return
  }
  if (!selectedModelID.value) {
    ElMessage.warning('请先选择模型')
    return
  }

  sending.value = true
  cancelEditing()

  const controller = new AbortController()
  streamAbortController.value = controller
  const createdAt = new Date().toISOString()
  const userTempID = `local-user-${Date.now()}`
  let assistantMessageID = `local-assistant-${Date.now()}`
  let nextConversationID = options.conversationID || ''

  const nextMessages = options.optimisticBaseMessages.map((item) => ({ ...item }))
  if (options.optimisticUserContent) {
    nextMessages.push({
      id: userTempID,
      role: 'user',
      content: options.optimisticUserContent,
      reasoning_content: '',
      status: 'completed',
      created_at: createdAt
    })
  }
  nextMessages.push({
    id: assistantMessageID,
    role: 'assistant',
    content: '',
    reasoning_content: '',
    status: 'streaming',
    finish_reason: null,
    usage: {},
    created_at: createdAt
  })
  messages.value = nextMessages
  scrollMessagesToBottom('smooth', true)

  try {
    await streamChatCompletion(
      {
        conversation_id: options.conversationID,
        model_id: selectedModelID.value,
        stream: true,
        messages: options.requestMessages
      },
      auth.accessToken,
      controller.signal,
      (event) => {
        switch (event.type) {
          case 'response.start':
            nextConversationID = event.conversation_id
            replaceMessageID(assistantMessageID, event.assistant_message_id)
            assistantMessageID = event.assistant_message_id
            window.dispatchEvent(new Event('mrchat:conversations:refresh'))
            if (nextConversationID && nextConversationID !== currentConversationId.value) {
              void router.replace(`/chat/${nextConversationID}`)
            }
            break
          case 'response.delta':
            appendMessageValue(assistantMessageID, 'content', event.delta.content ?? '')
            break
          case 'reasoning.delta':
            appendMessageValue(assistantMessageID, 'reasoning_content', event.delta.reasoning_content ?? '')
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
            ElMessage.error(`${event.error?.code ?? 'CHAT_STREAM_FAILED'}: ${event.error?.message ?? 'Streaming failed'}`)
            break
        }
      }
    )

  } catch (error) {
    if (isAbortError(error)) {
      patchMessage(assistantMessageID, { status: 'cancelled' })
      ElMessage.warning('已停止生成')
    } else {
      patchMessage(assistantMessageID, { status: 'failed' })
      ElMessage.error(toErrorMessage(error))
    }

    if (nextConversationID && nextConversationID !== currentConversationId.value) {
      await router.push(`/chat/${nextConversationID}`)
    }
  } finally {
    window.dispatchEvent(new Event('mrchat:conversations:refresh'))
    streamAbortController.value = null
    sending.value = false
    focusComposer()
  }
}

function stopStreaming() {
  streamAbortController.value?.abort()
}

function findMessageIndex(messageID: string) {
  return messages.value.findIndex((item) => item.id === messageID)
}

function patchMessage(messageID: string, patch: Partial<MessageItem>) {
  const messageIndex = findMessageIndex(messageID)
  if (messageIndex === -1) {
    return
  }

  Object.assign(messages.value[messageIndex], patch)
  scrollMessagesToBottom()
}

function replaceMessageID(currentID: string, nextID: string) {
  const messageIndex = findMessageIndex(currentID)
  if (messageIndex === -1) {
    return
  }

  messages.value[messageIndex].id = nextID
  scrollMessagesToBottom()
}

function appendMessageValue(messageID: string, field: 'content' | 'reasoning_content', delta: string) {
  if (!delta) {
    return
  }

  const messageIndex = findMessageIndex(messageID)
  if (messageIndex === -1) {
    return
  }

  const message = messages.value[messageIndex]
  message[field] += delta
  message.status = 'streaming'
  scrollMessagesToBottom()
}

function handleMessagesScroll() {
  const element = messagesViewport.value
  if (!element) {
    return
  }

  autoScrollPinned.value = isScrolledNearBottom(element)
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

function scrollMessagesToBottom(behavior: ScrollBehavior = 'auto', force = false) {
  if (force) {
    autoScrollPinned.value = true
  } else if (!autoScrollPinned.value) {
    return
  }

  pendingScrollBehavior = behavior === 'smooth' ? 'smooth' : pendingScrollBehavior
  if (scrollScheduled) {
    return
  }

  if (pendingScrollBehavior !== 'smooth') {
    pendingScrollBehavior = behavior
  }

  scrollScheduled = true

  nextTick(() => {
    window.requestAnimationFrame(() => {
      scrollScheduled = false

      const element = messagesViewport.value
      const nextBehavior = pendingScrollBehavior
      pendingScrollBehavior = 'auto'
      if (!element) {
        return
      }

      element.scrollTo({
        top: element.scrollHeight,
        behavior: nextBehavior
      })
    })
  })
}

function isScrolledNearBottom(element: HTMLElement) {
  const remainingDistance = element.scrollHeight - element.scrollTop - element.clientHeight
  return remainingDistance <= scrollBottomThreshold
}

function resizeTextarea(element: HTMLTextAreaElement | null, maxHeight: number) {
  if (!element) {
    return
  }

  element.style.height = '0px'
  const nextHeight = Math.min(element.scrollHeight, maxHeight)
  element.style.height = `${Math.max(nextHeight, 22)}px`
  element.style.overflowY = element.scrollHeight > maxHeight ? 'auto' : 'hidden'
}

function resizeComposerInput() {
  nextTick(() => {
    resizeTextarea(composerInput.value, 96)
  })
}

function resizeEditInput() {
  nextTick(() => {
    resizeTextarea(editInput.value, 120)
  })
}

function focusComposer() {
  nextTick(() => {
    composerInput.value?.focus()
  })
}

function handleComposerEnter() {
  if (isComposing.value) {
    return
  }
  if (sending.value) {
    stopStreaming()
    return
  }
  void sendMessage()
}

function handlePrimaryAction() {
  if (sending.value) {
    stopStreaming()
    return
  }
  void sendMessage()
}

function handleComposerCompositionEnd() {
  isComposing.value = false
}

function handleEditingCompositionEnd() {
  isEditingComposing.value = false
}

function formatMessageTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return new Intl.DateTimeFormat('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  }).format(date)
}

function roleLabel(role: string) {
  if (role === 'user') {
    return ''
  }
  if (role === 'assistant') {
    return 'AI'
  }
  return role
}

function statusLabel(status: string) {
  if (status === 'streaming') {
    return '生成中'
  }
  if (status === 'failed') {
    return '生成失败'
  }
  if (status === 'cancelled') {
    return '已停止'
  }
  return status
}

function shouldShowStatus(status: string) {
  return status !== 'completed' && status !== 'streaming'
}

function assistantDisplayContent(message: MessageItem) {
  return message.content
}

function assistantHasContent(message: MessageItem) {
  return Boolean(message.content.trim())
}

function renderMarkdown(content: string) {
  const source = content.trim()
  if (!source) {
    return ''
  }
  return DOMPurify.sanitize(markdown.render(source))
}

function canCopyMessage(message: MessageItem) {
  return Boolean(message.content.trim() || message.reasoning_content.trim())
}

async function copyMessage(message: MessageItem) {
  const segments = [message.content.trim()]
  if (message.role === 'assistant' && message.reasoning_content.trim()) {
    segments.push(`思考过程\n${message.reasoning_content.trim()}`)
  }

  try {
    await navigator.clipboard.writeText(segments.filter(Boolean).join('\n\n'))
    ElMessage.success('已复制')
  } catch (error) {
    console.error('Failed to copy message:', error)
    ElMessage.error('复制失败')
  }
}

function startEditingMessage(message: MessageItem) {
  editingMessageID.value = message.id
  editingMessageDraft.value = message.content
  nextTick(() => {
    resizeEditInput()
    editInput.value?.focus()
  })
}

function cancelEditing() {
  editingMessageID.value = ''
  editingMessageDraft.value = ''
  isEditingComposing.value = false
}

async function confirmEdit(messageIndex: number) {
  if (isEditingComposing.value) {
    return
  }

  const content = editingMessageDraft.value.trim()
  if (!content) {
    ElMessage.warning('消息不能为空')
    return
  }

  const prefixMessages = messages.value.slice(0, messageIndex)
  cancelEditing()

  await runCompletion({
    conversationID: null,
    requestMessages: [
      ...toRequestMessages(prefixMessages),
      {
        role: 'user',
        content
      }
    ],
    optimisticBaseMessages: prefixMessages,
    optimisticUserContent: content
  })
}

async function regenerateFromAssistant(messageIndex: number) {
  const relatedUserIndex = findNearestUserBefore(messageIndex)
  if (relatedUserIndex === -1) {
    ElMessage.warning('没有找到可重新生成的问题')
    return
  }

  const prefixMessages = messages.value.slice(0, relatedUserIndex + 1)
  await runCompletion({
    conversationID: null,
    requestMessages: toRequestMessages(prefixMessages),
    optimisticBaseMessages: prefixMessages
  })
}

function findNearestUserBefore(messageIndex: number) {
  for (let index = messageIndex - 1; index >= 0; index -= 1) {
    if (messages.value[index]?.role === 'user') {
      return index
    }
  }
  return -1
}

function toRequestMessages(items: MessageItem[]): RequestMessage[] {
  return items
    .filter((item) => item.role === 'user' || item.role === 'assistant')
    .filter((item) => item.content.trim())
    .map((item) => ({
      role: item.role,
      content: item.content
    }))
}

function startEmptyStateTypewriter() {
  if (typewriterTimer.value !== null) {
    return
  }
  scheduleEmptyStateTypewriterTick(180)
}

function stopEmptyStateTypewriter() {
  if (typewriterTimer.value !== null) {
    window.clearTimeout(typewriterTimer.value)
    typewriterTimer.value = null
  }
  emptyStateTypewriterText.value = ''
  typewriterPromptIndex.value = 0
  typewriterCharIndex.value = 0
  typewriterDeleting.value = false
}

function scheduleEmptyStateTypewriterTick(delay: number) {
  if (typewriterTimer.value !== null) {
    window.clearTimeout(typewriterTimer.value)
  }
  typewriterTimer.value = window.setTimeout(runEmptyStateTypewriterTick, delay)
}

function runEmptyStateTypewriterTick() {
  typewriterTimer.value = null

  if (hasMessages.value || loadingMessages.value) {
    stopEmptyStateTypewriter()
    return
  }

  const prompt = emptyStatePrompts[typewriterPromptIndex.value] ?? ''
  if (!prompt) {
    return
  }

  if (!typewriterDeleting.value) {
    typewriterCharIndex.value += 1
    emptyStateTypewriterText.value = prompt.slice(0, typewriterCharIndex.value)

    if (typewriterCharIndex.value >= prompt.length) {
      typewriterDeleting.value = true
      scheduleEmptyStateTypewriterTick(1800)
      return
    }

    scheduleEmptyStateTypewriterTick(70)
    return
  }

  typewriterCharIndex.value = Math.max(typewriterCharIndex.value - 1, 0)
  emptyStateTypewriterText.value = prompt.slice(0, typewriterCharIndex.value)

  if (typewriterCharIndex.value === 0) {
    typewriterDeleting.value = false
    typewriterPromptIndex.value = (typewriterPromptIndex.value + 1) % emptyStatePrompts.length
    scheduleEmptyStateTypewriterTick(260)
    return
  }

  scheduleEmptyStateTypewriterTick(32)
}

</script>

<style scoped>
.chat-page {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 0.85rem;
  padding: 1rem;
  background: var(--layout-content-bg);
  color: var(--text-primary);
}

.page-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 1rem;
  width: 100%;
  padding: 0 0.5rem;
}

.model-panel {
  min-width: 0;
  flex: none;
}

.chat-shell {
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1.25rem;
  width: min(100%, 960px);
  align-self: center;
}

.toolbar-actions {
  display: flex;
  gap: 0.65rem;
  flex-wrap: wrap;
}

.action-btn {
  border: 1px solid var(--input-border);
  background: var(--bg-secondary);
  color: var(--text-primary);
  border-radius: 999px;
  padding: 0.55rem 1rem;
  cursor: pointer;
  transition: border-color 0.2s ease, background 0.2s ease;
}

.action-btn:hover:not(:disabled) {
  background: var(--surface-muted);
  border-color: var(--accent-primary);
}

.action-btn.warning {
  border-color: color-mix(in srgb, var(--error-color) 45%, var(--input-border));
}

.action-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.chat-body {
  min-height: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 1rem;
}

.chat-body-loading {
  align-items: center;
  justify-content: center;
}

.chat-body-empty {
  justify-content: center;
}

.state-copy {
  margin: 0;
  color: var(--text-secondary);
}

.messages-viewport {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-anchor: none;
  padding-right: 0.35rem;
  scrollbar-width: none;
  -ms-overflow-style: none;
}

.messages-viewport::-webkit-scrollbar {
  width: 0;
  height: 0;
}

.message-list {
  list-style: none;
  margin: 0;
  padding: 0.25rem 0 0;
  display: flex;
  flex-direction: column;
  gap: 0.95rem;
}

.message-row {
  display: flex;
  width: 100%;
}

.message-row.role-user {
  justify-content: flex-end;
}

.message-row.role-assistant {
  justify-content: flex-start;
}

.message-bubble {
  width: fit-content;
  max-width: min(88%, 58rem);
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 0.65rem;
  padding: 0.5rem;
}

.message-row.role-user .message-bubble {
  max-width: min(72%, 34rem);
  margin-left: auto;
  align-items: flex-end;
}

.message-row.role-user .user-content {
  padding: 0.5rem 1rem;
  border-radius: 50px;
  background: color-mix(in srgb, var(--accent-primary) 25%, var(--bg-secondary));
  width: fit-content;
  margin-left: auto;
}

.message-meta {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 0.9rem;
  color: var(--text-secondary);
  font-size: 0.78rem;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.message-bubble:hover .message-meta {
  opacity: 1;
}

.message-role {
  font-weight: 600;
}

.message-panel {
  display: flex;
  flex-direction: column;
  gap: 0.7rem;
}

.thinking-panel {
  padding: 0.9rem 1rem;
  border-radius: 14px;
  background: var(--surface-subtle);
  border: 1px dashed var(--input-border);
}

.answer-panel {
  padding-top: 0.1rem;
}

.panel-label {
  color: var(--text-secondary);
  font-size: 0.78rem;
  letter-spacing: 0.03em;
}

.streaming-indicator {
  display: inline-flex;
  align-items: center;
  gap: 0.28rem;
  color: var(--text-secondary);
  min-height: 1.4rem;
  padding: 0.1rem 0;
}

.streaming-indicator-empty {
  padding-top: 0.2rem;
}

.streaming-dot {
  width: 0.42rem;
  height: 0.42rem;
  border-radius: 999px;
  background: currentColor;
  opacity: 0.28;
  animation: streaming-dot-bounce 1.05s ease-in-out infinite;
}

.streaming-dot:nth-child(2) {
  animation-delay: 0.16s;
}

.streaming-dot:nth-child(3) {
  animation-delay: 0.32s;
}

@keyframes streaming-dot-bounce {
  0%,
  80%,
  100% {
    transform: translateY(0);
    opacity: 0.28;
  }

  40% {
    transform: translateY(-3px);
    opacity: 0.9;
  }
}

.user-content {
  white-space: pre-wrap;
  word-break: break-word;
  line-height: 1.7;
}

.edit-box {
  min-width: min(26rem, 60vw);
}

.message-edit-input {
  width: 100%;
  min-height: 22px;
  max-height: 120px;
  border: 1px solid var(--input-border);
  outline: none;
  resize: none;
  background: var(--surface-subtle);
  color: var(--text-primary);
  padding: 0.55rem 0.7rem;
  border-radius: 12px;
  font: inherit;
  line-height: 1.6;
}

.message-status {
  color: var(--text-secondary);
  font-size: 0.84rem;
}

.message-tools {
  display: flex;
  align-items: center;
  gap: 0.45rem;
  padding-top: 0.2rem;
  opacity: 0;
  transition: opacity 0.2s ease;
}

.message-bubble:hover .message-tools {
  opacity: 1;
}

.tool-btn {
  width: 2.1rem;
  height: 2rem;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  transition: background 0.2s ease, color 0.2s ease;
}

.tool-btn svg {
  width: 0.95rem;
  height: 0.95rem;
}

.tool-btn:hover:not(:disabled) {
  background: var(--surface-muted);
  color: var(--text-primary);
}

.tool-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.confirm-btn {
  color: var(--accent-primary);
}

.empty-state {
  text-align: center;
  display: flex;
  flex-direction: column;
  gap: 0.75rem;
  align-items: center;
}

.empty-state h1 {
  margin: 0;
  font-size: clamp(1.6rem, 2vw, 2.2rem);
}

.empty-state p {
  margin: 0;
  color: var(--text-secondary);
}

.empty-state-typewriter {
  min-height: 1.7em;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 0.15rem;
}

.typewriter-caret {
  width: 1px;
  height: 1.05em;
  background: currentColor;
  animation: caret-blink 0.9s steps(1) infinite;
}

@keyframes caret-blink {
  0%,
  49% {
    opacity: 1;
  }

  50%,
  100% {
    opacity: 0;
  }
}

.composer-form {
  width: 100%;
}

.composer-form-centered {
  max-width: 920px;
  align-self: center;
}

.composer-form-docked {
  padding-top: 0.2rem;
}

.composer-row {
  width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) 52px;
  align-items: center;
  gap: 0.8rem;
  flex-wrap: nowrap;
}

.composer-input-wrap {
  flex: 1;
  min-width: 0;
  min-height: 52px;
  display: flex;
  align-items: center;
  padding: 0.6rem 1rem;
  border: 1px solid var(--input-border);
  border-radius: 50px;
  background: var(--bg-secondary);
}

.message-input {
  flex: 1;
  min-width: 0;
  min-height: 22px;
  max-height: 96px;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  color: var(--text-primary);
  padding: 0;
  font: inherit;
  line-height: 1.55;
}

.message-input::placeholder {
  color: color-mix(in srgb, var(--text-secondary) 72%, transparent);
}

.send-btn {
  flex: none;
  width: 52px;
  height: 52px;
  border: none;
  border-radius: 50px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: var(--accent-primary);
  color: #fff;
  cursor: pointer;
  transition: background 0.2s ease, opacity 0.2s ease;
}

.send-btn-stop {
  background: #c44536;
}

.send-btn svg {
  width: 1.45rem;
  height: 1.45rem;
  display: block;
  flex: none;
}

.send-btn:hover:not(:disabled) {
  background: var(--accent-secondary);
}

.send-btn-stop:hover:not(:disabled) {
  background: #a83a2e;
}

.send-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

:deep(.model-select) {
  display: inline-block;
  width: auto;
  min-width: 11ch;
  max-width: min(70vw, 22rem);
}

:deep(.model-select .el-select__wrapper) {
  min-height: 2.25rem;
  min-width: 100%;
  display: grid;
  grid-template-columns: minmax(0, 1fr) auto;
  align-items: center;
  column-gap: 0.35rem;
  padding: 0 0.1rem;
  border: none;
  background: transparent;
  box-shadow: none;
}

:deep(.model-select .el-select__selection) {
  min-width: 0;
  width: auto;
}

:deep(.model-select .el-select__selected-item) {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  color: var(--text-primary);
  font-weight: 600;
  white-space: nowrap;
}

:deep(.model-select .el-select__placeholder) {
  display: block;
  overflow: hidden;
  text-overflow: ellipsis;
  color: color-mix(in srgb, var(--text-secondary) 100%, transparent);
}

:deep(.model-select .el-select__caret) {
  color: var(--text-secondary);
  margin-left: 0.35rem;
}

.model-option {
  display: flex;
  align-items: center;
}

.model-option-name {
  color: var(--text-primary);
  font-weight: 600;
  line-height: 1.25;
}

:global(.chat-model-popper.el-popper) {
  --el-bg-color-overlay: var(--bg-secondary);
  --el-fill-color-blank: var(--bg-secondary);
  --el-fill-color-light: var(--surface-muted);
  --el-fill-color-lightest: var(--surface-subtle);
  --el-border-color-light: var(--glass-border);
  --el-text-color-regular: var(--text-primary);
  --el-text-color-placeholder: var(--text-secondary);
  --el-color-primary: var(--accent-primary);
  border-radius: 14px;
  overflow: hidden;
  border: 1px solid var(--glass-border);
  background: var(--bg-secondary);
  box-shadow: var(--shadow-md);
}

:global(.chat-model-popper .el-select-dropdown) {
  background: var(--bg-secondary);
}

:global(.chat-model-popper .el-popper__arrow::before) {
  background: var(--bg-secondary);
  border-color: var(--glass-border);
}

:global(.chat-model-popper .el-select-dropdown__wrap) {
  padding: 0.22rem;
  background: var(--bg-secondary);
}

:global(.chat-model-popper .el-select-dropdown__item) {
  height: auto;
  min-height: 2.5rem;
  line-height: 1.3;
  padding: 0.55rem 0.7rem;
  border-radius: 10px;
  color: var(--text-primary);
}

:global(.chat-model-popper .el-select-dropdown__item.is-hovering) {
  background: var(--surface-muted);
}

:global(.chat-model-popper .el-select-dropdown__item.is-selected) {
  background: color-mix(in srgb, var(--accent-primary) 12%, var(--bg-secondary));
  color: var(--text-primary);
}

:deep(.markdown-body) {
  color: var(--text-primary);
  line-height: 1.75;
  word-break: break-word;
}

:deep(.markdown-body > :first-child) {
  margin-top: 0;
}

:deep(.markdown-body > :last-child) {
  margin-bottom: 0;
}

:deep(.markdown-body p),
:deep(.markdown-body ul),
:deep(.markdown-body ol),
:deep(.markdown-body pre),
:deep(.markdown-body blockquote),
:deep(.markdown-body table) {
  margin: 0 0 0.85rem;
}

:deep(.markdown-body ul),
:deep(.markdown-body ol) {
  padding-left: 1.35rem;
}

:deep(.markdown-body li + li) {
  margin-top: 0.28rem;
}

:deep(.markdown-body pre) {
  overflow-x: auto;
  padding: 0.9rem 1rem;
  border-radius: 14px;
  background: var(--surface-subtle);
}

:deep(.markdown-body code) {
  padding: 0.15rem 0.35rem;
  border-radius: 6px;
  background: var(--surface-muted);
  font-size: 0.92em;
}

:deep(.markdown-body pre code) {
  padding: 0;
  background: transparent;
}

:deep(.markdown-body blockquote) {
  padding-left: 0.9rem;
  border-left: 3px solid color-mix(in srgb, var(--accent-primary) 42%, var(--input-border));
  color: var(--text-secondary);
}

:deep(.markdown-body a) {
  color: var(--accent-primary);
}

:deep(.markdown-body hr) {
  border: none;
  border-top: 1px solid color-mix(in srgb, var(--input-border) 90%, transparent);
  margin: 1rem 0;
}

@media (max-width: 900px) {
  .chat-page {
    padding: 0.9rem;
  }

  .chat-toolbar {
    flex-direction: column;
    align-items: flex-start;
  }

  .toolbar-actions {
    width: 100%;
  }

  .message-row.role-user .message-bubble,
  .message-bubble {
    max-width: 100%;
  }
}

@media (max-width: 640px) {
  .chat-page {
    padding: 0.65rem;
  }

  .chat-shell {
    padding: 0.8rem;
    border-radius: 18px;
  }

  .composer-row {
    gap: 0.6rem;
    grid-template-columns: minmax(0, 1fr) 48px;
  }

  .composer-input-wrap {
    min-height: 48px;
    padding: 0.55rem 0.85rem;
  }

  .send-btn {
    width: 48px;
    height: 48px;
  }
}
</style>
