<template>
  <div class="chat-container">
    <!-- Sidebar -->
    <aside class="sidebar" :class="{ collapsed: sidebarCollapsed }">
      <div class="sidebar-header">
        <h1 class="logo">MrChat</h1>
        <button @click="toggleSidebar" class="collapse-btn">
          <ChevronLeft v-if="!sidebarCollapsed" :size="18" />
          <ChevronRight v-else :size="18" />
        </button>
      </div>

      <button @click="startNewChat" class="new-chat-btn">
        <Plus :size="18" />
        <span v-if="!sidebarCollapsed">New Conversation</span>
      </button>

      <div v-if="!sidebarCollapsed" class="conversations-list">
        <div
          v-for="conv in conversations"
          :key="conv.id"
          @click="selectConversation(conv.id)"
          class="conversation-item"
          :class="{ active: conv.id === currentConversationId }"
        >
          <div class="conv-title">{{ conv.title }}</div>
          <div class="conv-time">{{ formatTime(conv.updated_at) }}</div>
        </div>
      </div>
    </aside>

    <!-- Main Chat Area -->
    <main class="chat-main">
      <header class="chat-header">
        <div class="model-selector">
          <label>Model</label>
          <select v-model="selectedModel" @change="onModelChange">
            <option v-for="model in availableModels" :key="model.id" :value="model.id">
              {{ model.name }}
            </option>
          </select>
        </div>
      </header>

      <div class="messages-container" ref="messagesContainer">
        <div
          v-for="message in messages"
          :key="message.id"
          class="message-wrapper"
          :class="message.role"
        >
          <div class="message-bubble">
            <div class="message-header">
              <span class="role-label">{{ message.role === 'user' ? 'You' : 'Assistant' }}</span>
            </div>
            <div class="message-content" v-html="renderMarkdown(message.content)"></div>
          </div>
        </div>

        <div v-if="isStreaming" class="message-wrapper assistant streaming">
          <div class="message-bubble">
            <div class="message-header">
              <span class="role-label">Assistant</span>
              <span class="streaming-indicator">
                <span class="dot"></span>
                <span class="dot"></span>
                <span class="dot"></span>
              </span>
            </div>
            <div class="message-content" v-html="renderMarkdown(streamingContent)"></div>
          </div>
        </div>
      </div>

      <div class="input-area">
        <textarea
          v-model="inputMessage"
          @keydown.enter.exact.prevent="sendMessage"
          placeholder="Type your message..."
          rows="1"
          ref="textarea"
          :disabled="isStreaming"
        ></textarea>
        <button
          v-if="!isStreaming"
          @click="sendMessage"
          :disabled="!inputMessage.trim()"
          class="send-btn"
        >
          <Send :size="20" />
        </button>
        <button
          v-else
          @click="stopGeneration"
          class="send-btn stop-btn"
        >
          <Square :size="20" />
        </button>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick, watch } from 'vue'
import { marked } from 'marked'
import { Send, Square, ChevronLeft, ChevronRight, Plus } from 'lucide-vue-next'

const sidebarCollapsed = ref(false)
const conversations = ref([])
const currentConversationId = ref(null)
const messages = ref([])
const inputMessage = ref('')
const isStreaming = ref(false)
const streamingContent = ref('')
const streamingMessageId = ref(null)
const selectedModel = ref('gpt-4')
const availableModels = ref([
  { id: 'gpt-4', name: 'GPT-4' },
  { id: 'claude-3', name: 'Claude 3' },
  { id: 'gemini-pro', name: 'Gemini Pro' }
])

const messagesContainer = ref(null)
const textarea = ref(null)

const toggleSidebar = () => {
  sidebarCollapsed.value = !sidebarCollapsed.value
}

const startNewChat = () => {
  currentConversationId.value = null
  messages.value = []
}

const selectConversation = (id) => {
  currentConversationId.value = id
  loadMessages(id)
}

const loadMessages = async (conversationId) => {
  messages.value = []
}

const sendMessage = async () => {
  if (!inputMessage.value.trim() || isStreaming.value) return

  const userMessage = {
    id: Date.now(),
    role: 'user',
    content: inputMessage.value,
    created_at: new Date()
  }

  messages.value.push(userMessage)
  inputMessage.value = ''

  await nextTick()
  scrollToBottom()

  isStreaming.value = true
  streamingContent.value = ''
  streamingMessageId.value = Date.now()

  await streamResponse(userMessage.content)
}

const stopGeneration = () => {
  if (!isStreaming.value) return

  // Save partial content
  if (streamingContent.value.trim()) {
    messages.value.push({
      id: streamingMessageId.value,
      role: 'assistant',
      content: streamingContent.value,
      created_at: new Date()
    })
  }

  isStreaming.value = false
  streamingContent.value = ''
  streamingMessageId.value = null
}

const streamResponse = async (userInput) => {
  const response = 'This is a simulated streaming response. It will continue for a while to demonstrate the streaming effect and the stop button functionality. '

  for (let i = 0; i < response.length; i++) {
    if (!isStreaming.value) break // Check if stopped

    streamingContent.value += response[i]
    await new Promise(resolve => setTimeout(resolve, 30))
    scrollToBottom()
  }

  if (isStreaming.value) {
    messages.value.push({
      id: streamingMessageId.value,
      role: 'assistant',
      content: streamingContent.value,
      created_at: new Date()
    })

    isStreaming.value = false
    streamingContent.value = ''
    streamingMessageId.value = null
  }
}

const onModelChange = () => {
  console.log('Model changed to:', selectedModel.value)
}

const renderMarkdown = (content) => {
  return marked(content || '')
}

const formatTime = (date) => {
  const d = new Date(date)
  const now = new Date()
  const diff = now - d

  if (diff < 60000) return 'Just now'
  if (diff < 3600000) return `${Math.floor(diff / 60000)}m ago`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)}h ago`
  return d.toLocaleDateString()
}

const scrollToBottom = () => {
  if (messagesContainer.value) {
    messagesContainer.value.scrollTop = messagesContainer.value.scrollHeight
  }
}

watch(inputMessage, () => {
  if (textarea.value) {
    textarea.value.style.height = 'auto'
    textarea.value.style.height = textarea.value.scrollHeight + 'px'
  }
})

onMounted(() => {
  conversations.value = [
    { id: '1', title: 'Previous conversation', updated_at: new Date(Date.now() - 3600000) },
    { id: '2', title: 'Another chat', updated_at: new Date(Date.now() - 86400000) }
  ]
})
</script>

<style scoped>
.chat-container {
  display: flex;
  height: 100vh;
  background: #faf8f6;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif;
}

/* Sidebar */
.sidebar {
  width: 280px;
  background: #2c2420;
  color: #e8e3df;
  display: flex;
  flex-direction: column;
  transition: width 0.3s ease;
  border-right: 1px solid #3d3530;
}

.sidebar.collapsed {
  width: 60px;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px 16px;
  border-bottom: 1px solid #3d3530;
}

.logo {
  font-size: 20px;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: #f4ede4;
  margin: 0;
}

.sidebar.collapsed .logo {
  display: none;
}

.collapse-btn {
  background: none;
  border: none;
  color: #a89984;
  cursor: pointer;
  padding: 6px;
  border-radius: 6px;
  transition: all 0.2s;
}

.collapse-btn:hover {
  background: #3d3530;
  color: #e8e3df;
}

.new-chat-btn {
  margin: 16px;
  padding: 12px 16px;
  background: #4a3f35;
  border: none;
  border-radius: 8px;
  color: #f4ede4;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  transition: all 0.2s;
}

.new-chat-btn:hover {
  background: #5a4f45;
}

.sidebar.collapsed .new-chat-btn {
  padding: 12px;
  justify-content: center;
}

.sidebar.collapsed .new-chat-btn span:not(.icon) {
  display: none;
}

.conversations-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
}

.conversation-item {
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 4px;
  transition: all 0.2s;
}

.conversation-item:hover {
  background: #3d3530;
}

.conversation-item.active {
  background: #4a3f35;
}

.conv-title {
  font-size: 14px;
  font-weight: 500;
  color: #e8e3df;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.conv-time {
  font-size: 12px;
  color: #a89984;
}

/* Main Chat */
.chat-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  max-width: 900px;
  margin: 0 auto;
  width: 100%;
}

.chat-header {
  padding: 20px 24px;
  border-bottom: 1px solid #e8e3df;
  background: #faf8f6;
}

.model-selector {
  display: flex;
  align-items: center;
  gap: 12px;
}

.model-selector label {
  font-size: 14px;
  font-weight: 500;
  color: #5a4f45;
}

.model-selector select {
  padding: 8px 12px;
  border: 1px solid #d4cdc4;
  border-radius: 8px;
  background: white;
  color: #2c2420;
  font-size: 14px;
  cursor: pointer;
  transition: all 0.2s;
}

.model-selector select:hover {
  border-color: #a89984;
}

.model-selector select:focus {
  outline: none;
  border-color: #7a6f5d;
  box-shadow: 0 0 0 3px rgba(122, 111, 93, 0.1);
}

.messages-container {
  flex: 1;
  overflow-y: auto;
  overflow-anchor: none;
  padding: 32px 24px;
  scroll-behavior: smooth;
}

.message-wrapper {
  margin-bottom: 24px;
  display: flex;
}

.message-wrapper.user {
  justify-content: flex-end;
}

.message-wrapper.assistant {
  justify-content: flex-start;
}

.message-bubble {
  max-width: 85%;
  padding: 16px 20px;
  border-radius: 12px;
}

.user .message-bubble {
  background: #2c2420;
  color: #f4ede4;
}

.assistant .message-bubble {
  background: white;
  color: #2c2420;
  border: 1px solid #e8e3df;
}

.message-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.role-label {
  font-size: 12px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.user .role-label {
  color: #d4cdc4;
}

.assistant .role-label {
  color: #7a6f5d;
}

.message-content {
  font-size: 15px;
  line-height: 1.6;
}

.streaming-indicator {
  display: flex;
  gap: 4px;
}

.dot {
  width: 6px;
  height: 6px;
  background: #a89984;
  border-radius: 50%;
  animation: pulse 1.4s infinite;
}

.dot:nth-child(2) {
  animation-delay: 0.2s;
}

.dot:nth-child(3) {
  animation-delay: 0.4s;
}

@keyframes pulse {
  0%, 60%, 100% {
    opacity: 0.3;
  }
  30% {
    opacity: 1;
  }
}

.input-area {
  padding: 24px;
  background: #faf8f6;
  border-top: 1px solid #e8e3df;
  display: flex;
  gap: 12px;
  align-items: flex-end;
}

.input-area textarea {
  flex: 1;
  padding: 14px 16px;
  border: 1px solid #d4cdc4;
  border-radius: 12px;
  font-size: 15px;
  font-family: inherit;
  resize: none;
  max-height: 200px;
  background: white;
  color: #2c2420;
  transition: all 0.2s;
}

.input-area textarea:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.input-area textarea:focus {
  outline: none;
  border-color: #7a6f5d;
  box-shadow: 0 0 0 3px rgba(122, 111, 93, 0.1);
}

.send-btn {
  width: 44px;
  height: 44px;
  border: none;
  border-radius: 10px;
  background: #2c2420;
  color: white;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
  flex-shrink: 0;
}

.send-btn:hover:not(:disabled) {
  background: #3d3530;
  transform: translateY(-1px);
}

.send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.send-btn.stop-btn {
  background: #c44536;
}

.send-btn.stop-btn:hover {
  background: #d4554a;
}
</style>
