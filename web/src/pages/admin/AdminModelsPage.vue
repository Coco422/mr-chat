<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>模型管理</h1>
      <button class="primary-btn" @click="showForm = !showForm">
        {{ showForm ? '取消' : '+ 新建模型' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="showForm" class="form-card">
      <h2>创建模型</h2>
      <form @submit.prevent="createModel" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>Model Key</label>
            <input v-model.trim="form.modelKey" type="text" required />
          </div>
          <div class="form-group">
            <label>Display Name</label>
            <input v-model.trim="form.displayName" type="text" required />
          </div>
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Provider Type</label>
            <input v-model.trim="form.providerType" type="text" />
          </div>
          <div class="form-group">
            <label>Context Length</label>
            <input v-model.number="form.contextLength" type="number" min="1" />
          </div>
          <div class="form-group">
            <label>Max Output Tokens</label>
            <input v-model.number="form.maxOutputTokens" type="number" min="1" />
          </div>
        </div>

        <div class="form-group">
          <label>Visible User Group IDs</label>
          <input v-model.trim="form.visibleUserGroupIDsRaw" type="text" placeholder="uuid1,uuid2" />
        </div>

        <div class="form-row">
          <div class="form-group">
            <label>Channel</label>
            <el-select v-model="form.channelID" style="width: 100%">
              <el-option value="" label="默认路由" />
              <el-option v-for="channel in channels" :key="channel.id" :value="channel.id" :label="channel.name" />
            </el-select>
          </div>
          <div class="form-group">
            <label>Upstream</label>
            <el-select v-model="form.upstreamID" style="width: 100%">
              <el-option value="" label="请选择" />
              <el-option v-for="upstream in upstreams" :key="upstream.id" :value="upstream.id" :label="upstream.name" />
            </el-select>
          </div>
          <div class="form-group">
            <label>Status</label>
            <el-select v-model="form.status" style="width: 100%">
              <el-option value="active" label="active" />
              <el-option value="disabled" label="disabled" />
            </el-select>
          </div>
        </div>

        <button type="submit" :disabled="submitting" class="submit-btn">创建模型</button>
      </form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <h2>模型列表</h2>
        <button class="refresh-btn" @click="loadData" :disabled="loading">刷新</button>
      </div>

      <p v-if="loading" class="loading">加载中...</p>
      <table v-else-if="items.length > 0">
        <thead>
          <tr>
            <th>Display Name</th>
            <th>Model Key</th>
            <th>状态</th>
            <th>用户组</th>
            <th>路由绑定</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.display_name }}</td>
            <td class="model-key">{{ item.model_key }}</td>
            <td><span class="status-badge" :class="item.status">{{ item.status }}</span></td>
            <td class="user-groups">
              {{ item.visible_user_group_ids.length > 0 ? item.visible_user_group_ids.join(', ') : 'all users' }}
            </td>
            <td>
              <span v-for="binding in item.route_bindings" :key="binding.id" class="binding-tag">
                {{ binding.channel_id || 'default' }} → {{ binding.upstream_id }}#{{ binding.priority }}
              </span>
            </td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无模型</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError } from '@/lib/api'
import {
  createAdminModel,
  listAdminChannels,
  listAdminModels,
  listAdminUpstreams,
  listAdminUserGroups
} from '@/api/admin'
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
const showForm = ref(false)
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
      listAdminModels<ModelItem[]>(auth.accessToken),
      listAdminUpstreams<UpstreamItem[]>(auth.accessToken),
      listAdminChannels<ChannelItem[]>(auth.accessToken),
      listAdminUserGroups<UserGroupItem[]>(auth.accessToken)
    ])

    items.value = modelsResponse
    upstreams.value = upstreamsResponse
    channels.value = channelsResponse
    userGroups.value = userGroupsResponse

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
    await createAdminModel(auth.accessToken, {
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

<style scoped>
@import '@/styles/admin.css';

.model-key {
  font-family: monospace;
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.user-groups {
  font-size: 0.85rem;
  color: var(--text-secondary);
}

.binding-tag {
  display: inline-block;
  padding: 0.25rem 0.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 6px;
  font-size: 0.75rem;
  color: var(--text-secondary);
  margin-right: 0.5rem;
  font-family: monospace;
}
</style>
