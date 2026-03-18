<template>
  <section>
    <h1>Models</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="createModel">
      <div>
        <label>
          Model Key
          <input v-model.trim="form.modelKey" type="text" required />
        </label>
      </div>
      <div>
        <label>
          Display Name
          <input v-model.trim="form.displayName" type="text" required />
        </label>
      </div>
      <div>
        <label>
          Provider Type
          <input v-model.trim="form.providerType" type="text" />
        </label>
      </div>
      <div>
        <label>
          Context Length
          <input v-model.number="form.contextLength" type="number" min="1" />
        </label>
      </div>
      <div>
        <label>
          Max Output Tokens
          <input v-model.number="form.maxOutputTokens" type="number" min="1" />
        </label>
      </div>
      <div>
        <label>
          Visible User Group IDs
          <input v-model.trim="form.visibleUserGroupIDsRaw" type="text" placeholder="uuid1,uuid2" />
        </label>
      </div>
      <div>
        <label>
          Channel
          <select v-model="form.channelID">
            <option value="">默认路由</option>
            <option v-for="channel in channels" :key="channel.id" :value="channel.id">
              {{ channel.name }}
            </option>
          </select>
        </label>
      </div>
      <div>
        <label>
          Upstream
          <select v-model="form.upstreamID">
            <option value="">请选择</option>
            <option v-for="upstream in upstreams" :key="upstream.id" :value="upstream.id">
              {{ upstream.name }}
            </option>
          </select>
        </label>
      </div>
      <div>
        <label>
          状态
          <select v-model="form.status">
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </select>
        </label>
      </div>
      <button type="submit" :disabled="submitting">创建模型</button>
      <button type="button" @click="loadData" :disabled="loading">刷新</button>
    </form>

    <hr />

    <p v-if="loading">加载中...</p>
    <ul v-else-if="items.length > 0">
      <li v-for="item in items" :key="item.id">
        {{ item.display_name }} / {{ item.model_key }} / {{ item.status }}
        <div>
          visible_user_group_ids:
          {{ item.visible_user_group_ids.length > 0 ? item.visible_user_group_ids.join(', ') : 'all users' }}
        </div>
        <div>
          bindings:
          <span v-for="binding in item.route_bindings" :key="binding.id">
            {{ binding.channel_id || 'default' }} -> {{ binding.upstream_id }}#{{ binding.priority }}
          </span>
        </div>
      </li>
    </ul>
    <p v-else>暂无模型</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface UpstreamItem {
  id: string
  name: string
}

interface ChannelItem {
  id: string
  name: string
}

interface UserGroupItem {
  id: string
  name: string
}

interface ModelItem {
  id: string
  model_key: string
  display_name: string
  status: string
  visible_user_group_ids: string[]
  route_bindings: Array<{
    id: string
    channel_id: string | null
    upstream_id: string
    priority: number
  }>
}

const auth = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const upstreams = ref<UpstreamItem[]>([])
const channels = ref<ChannelItem[]>([])
const userGroups = ref<UserGroupItem[]>([])
const items = ref<ModelItem[]>([])
const form = reactive({
  modelKey: '',
  displayName: '',
  providerType: 'openai_compatible',
  contextLength: 32000,
  maxOutputTokens: 4096,
  visibleUserGroupIDsRaw: '',
  channelID: '',
  upstreamID: '',
  status: 'active'
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [modelsResponse, upstreamsResponse, channelsResponse, userGroupsResponse] = await Promise.all([
      apiRequest<ModelItem[]>('/admin/models', {
        accessToken: auth.accessToken
      }),
      apiRequest<UpstreamItem[]>('/admin/upstreams', {
        accessToken: auth.accessToken
      }),
      apiRequest<ChannelItem[]>('/admin/channels', {
        accessToken: auth.accessToken
      }),
      apiRequest<UserGroupItem[]>('/admin/user-groups', {
        accessToken: auth.accessToken
      })
    ])

    items.value = modelsResponse.data
    upstreams.value = upstreamsResponse.data
    channels.value = channelsResponse.data
    userGroups.value = userGroupsResponse.data

    if (!form.upstreamID && upstreams.value.length > 0) {
      form.upstreamID = upstreams.value[0].id
    }
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function createModel() {
  submitting.value = true
  errorMessage.value = ''

  try {
    await apiRequest('/admin/models', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        model_key: form.modelKey,
        display_name: form.displayName,
        provider_type: form.providerType,
        context_length: form.contextLength,
        max_output_tokens: form.maxOutputTokens,
        pricing: {},
        capabilities: {
          chat: true
        },
        visible_user_group_ids: parseCSV(form.visibleUserGroupIDsRaw),
        status: form.status,
        metadata: {},
        route_bindings: form.upstreamID
          ? [
              {
                channel_id: form.channelID || null,
                upstream_id: form.upstreamID,
                priority: 1,
                status: 'active'
              }
            ]
          : []
      }
    })

    form.modelKey = ''
    form.displayName = ''
    form.visibleUserGroupIDsRaw = ''
    await loadData()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function parseCSV(value: string) {
  return value
    .split(',')
    .map((item) => item.trim())
    .filter(Boolean)
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
