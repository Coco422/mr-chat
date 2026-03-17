<template>
  <section>
    <h1>Upstreams</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="createUpstream">
      <div>
        <label>
          名称
          <input v-model.trim="form.name" type="text" required />
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
          Base URL
          <input v-model.trim="form.baseURL" type="text" required />
        </label>
      </div>
      <div>
        <label>
          Auth Type
          <input v-model.trim="form.authType" type="text" />
        </label>
      </div>
      <div>
        <label>
          API Key
          <input v-model.trim="form.apiKey" type="text" />
        </label>
      </div>
      <div>
        <label>
          状态
          <select v-model="form.status">
            <option value="active">active</option>
            <option value="disabled">disabled</option>
            <option value="maintenance">maintenance</option>
          </select>
        </label>
      </div>
      <div>
        <label>
          Timeout Seconds
          <input v-model.number="form.timeoutSeconds" type="number" min="1" />
        </label>
      </div>
      <div>
        <label>
          Cooldown Seconds
          <input v-model.number="form.cooldownSeconds" type="number" min="1" />
        </label>
      </div>
      <div>
        <label>
          Failure Threshold
          <input v-model.number="form.failureThreshold" type="number" min="1" />
        </label>
      </div>
      <button type="submit" :disabled="submitting">创建上游</button>
      <button type="button" @click="loadUpstreams" :disabled="loading">刷新</button>
    </form>

    <hr />

    <p v-if="loading">加载中...</p>
    <ul v-else-if="items.length > 0">
      <li v-for="item in items" :key="item.id">
        {{ item.name }} / {{ item.provider_type }} / {{ item.status }} / {{ item.base_url }}
      </li>
    </ul>
    <p v-else>暂无上游配置</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface UpstreamItem {
  id: string
  name: string
  provider_type: string
  base_url: string
  status: string
}

const auth = useAuthStore()
const loading = ref(false)
const submitting = ref(false)
const errorMessage = ref('')
const items = ref<UpstreamItem[]>([])
const form = reactive({
  name: '',
  providerType: 'openai_compatible',
  baseURL: '',
  authType: 'bearer',
  apiKey: '',
  status: 'active',
  timeoutSeconds: 60,
  cooldownSeconds: 60,
  failureThreshold: 3
})

onMounted(async () => {
  await loadUpstreams()
})

async function loadUpstreams() {
  loading.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<UpstreamItem[]>('/admin/upstreams', {
      accessToken: auth.accessToken
    })
    items.value = data
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function createUpstream() {
  submitting.value = true
  errorMessage.value = ''

  try {
    await apiRequest('/admin/upstreams', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        name: form.name,
        provider_type: form.providerType,
        base_url: form.baseURL,
        auth_type: form.authType,
        auth_config: form.apiKey ? { api_key: form.apiKey } : {},
        status: form.status,
        timeout_seconds: form.timeoutSeconds,
        cooldown_seconds: form.cooldownSeconds,
        failure_threshold: form.failureThreshold,
        metadata: {}
      }
    })

    form.name = ''
    form.baseURL = ''
    form.apiKey = ''
    await loadUpstreams()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submitting.value = false
  }
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
