<template>
  <section>
    <h1>Chat</h1>
    <p>当前用户：{{ auth.user?.username ?? 'unknown' }}</p>
    <p>当前会话：{{ currentConversationId || 'new' }}</p>

    <p v-if="errorMessage">{{ errorMessage }}</p>

    <div>
      <button type="button" @click="reloadAll" :disabled="loading">刷新数据</button>
    </div>

    <hr />

    <section>
      <h2>创建会话</h2>
      <form @submit.prevent="createConversation">
        <div>
          <label>
            标题
            <input v-model.trim="createForm.title" type="text" placeholder="New conversation" />
          </label>
        </div>
        <div>
          <label>
            模型
            <select v-model="createForm.modelId">
              <option value="">不指定</option>
              <option v-for="model in models" :key="model.id" :value="model.id">
                {{ model.display_name }} ({{ model.model_key }})
              </option>
            </select>
          </label>
        </div>
        <button type="submit" :disabled="submittingConversation">创建会话</button>
      </form>
    </section>

    <hr />

    <section>
      <h2>可用模型</h2>
      <p v-if="loading">加载中...</p>
      <ul v-else>
        <li v-for="model in models" :key="model.id">
          {{ model.display_name }} / {{ model.model_key }} / {{ model.provider_type }}
        </li>
        <li v-if="models.length === 0">暂无可用模型</li>
      </ul>
    </section>

    <hr />

    <section>
      <h2>会话列表</h2>
      <ul v-if="conversations.length > 0">
        <li v-for="conversation in conversations" :key="conversation.id">
          <RouterLink :to="`/chat/${conversation.id}`">
            {{ conversation.title }}
          </RouterLink>
          <span> / {{ conversation.status }} / {{ conversation.message_count }} 条消息</span>
        </li>
      </ul>
      <p v-else>暂无会话</p>
    </section>

    <hr />

    <section>
      <h2>消息列表</h2>
      <p v-if="currentConversationId && loadingMessages">消息加载中...</p>
      <ul v-else-if="messages.length > 0">
        <li v-for="message in messages" :key="message.id">
          <strong>{{ message.role }}</strong>
          <div>{{ message.content }}</div>
          <small>{{ message.status }} / {{ message.created_at }}</small>
        </li>
      </ul>
      <p v-else-if="currentConversationId">当前会话暂无消息</p>
      <p v-else>请选择会话后查看消息</p>
    </section>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

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
const submittingConversation = ref(false)
const errorMessage = ref('')
const models = ref<UserModel[]>([])
const conversations = ref<ConversationSummary[]>([])
const messages = ref<MessageItem[]>([])
const createForm = reactive({
  title: '',
  modelId: ''
})

const currentConversationId = computed(() =>
  typeof route.params.conversationId === 'string' ? route.params.conversationId : ''
)

onMounted(async () => {
  await reloadAll()
})

watch(currentConversationId, async (conversationID) => {
  if (!conversationID) {
    messages.value = []
    return
  }
  await loadMessages(conversationID)
}, { immediate: true })

async function reloadAll() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [modelResponse, conversationResponse] = await Promise.all([
      apiRequest<UserModel[]>('/models', {
        accessToken: auth.accessToken
      }),
      apiRequest<ConversationSummary[]>('/conversations', {
        accessToken: auth.accessToken
      })
    ])

    models.value = modelResponse.data
    conversations.value = conversationResponse.data

    if (!createForm.modelId && models.value.length > 0) {
      createForm.modelId = models.value[0].id
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

async function createConversation() {
  submittingConversation.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<ConversationSummary>('/conversations', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        title: createForm.title,
        model_id: createForm.modelId || null
      }
    })

    createForm.title = ''
    await reloadAll()
    await router.push(`/chat/${data.id}`)
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submittingConversation.value = false
  }
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
