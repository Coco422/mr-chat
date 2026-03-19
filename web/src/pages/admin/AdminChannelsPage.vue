<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>渠道管理</h1>
      <button class="primary-btn" @click="showForm = !showForm">
        {{ showForm ? '取消' : '+ 新建渠道' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="showForm" class="form-card">
      <h2>创建渠道</h2>
      <form @submit.prevent="createChannel" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>名称</label>
            <input v-model.trim="form.name" type="text" required />
          </div>
          <div class="form-group">
            <label>状态</label>
            <el-select v-model="form.status" style="width: 100%">
              <el-option value="active" label="active" />
              <el-option value="disabled" label="disabled" />
            </el-select>
          </div>
        </div>

        <div class="form-group">
          <label>Description</label>
          <input v-model.trim="form.description" type="text" />
        </div>

        <div class="form-group">
          <label>Billing Config JSON</label>
          <textarea v-model="form.billingConfigText" rows="5" />
        </div>

        <button type="submit" :disabled="submitting" class="submit-btn">创建渠道</button>
      </form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <!-- <h2>渠道列表</h2> -->
        <button class="refresh-btn" @click="loadChannels" :disabled="loading">刷新</button>
      </div>

      <p v-if="loading" class="loading">加载中...</p>
      <table v-else-if="items.length > 0">
        <thead>
          <tr>
            <th>名称</th>
            <th>状态</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in items" :key="item.id">
            <td>{{ item.name }}</td>
            <td><span class="status-badge" :class="item.status">{{ item.status }}</span></td>
            <td>{{ item.description || '-' }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无渠道</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError } from '@/lib/api'
import { createAdminChannel, listAdminChannels } from '@/api/admin'
import { useAuthStore } from '@/stores/auth'

interface ChannelItem {
  id: string
  name: string
  description: string | null
  status: string
}

const auth = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const showForm = ref(false)
const items = ref<ChannelItem[]>([])
const form = reactive({
  name: '',
  description: '',
  status: 'active',
  billingConfigText: '{}'
})

onMounted(async () => {
  await loadChannels()
})

async function loadChannels() {
  loading.value = true
  errorMessage.value = ''

  try {
    items.value = await listAdminChannels<ChannelItem[]>(auth.accessToken)
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function createChannel() {
  submitting.value = true
  errorMessage.value = ''

  try {
    await createAdminChannel(auth.accessToken, {
      name: form.name,
      description: form.description || null,
      status: form.status,
      billing_config: parseJSON(form.billingConfigText),
      metadata: {}
    })

    form.name = ''
    form.description = ''
    form.billingConfigText = '{}'
    showForm.value = false
    await loadChannels()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function parseJSON(value: string) {
  const trimmed = value.trim()
  return trimmed ? JSON.parse(trimmed) : {}
}

function toErrorMessage(error: unknown) {
  if (error instanceof SyntaxError) {
    return 'billing_config 不是合法 JSON'
  }
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>

<style scoped>
@import '@/styles/admin.css';

textarea {
  padding: 0.75rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 8px;
  color: var(--text-primary);
  font-size: 0.9rem;
  font-family: monospace;
  resize: vertical;
}

textarea:focus {
  outline: none;
  border-color: var(--accent-primary);
}
</style>
