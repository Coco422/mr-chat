<template>
  <section>
    <h1>Channels</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="createChannel">
      <div>
        <label>
          名称
          <input v-model.trim="form.name" type="text" required />
        </label>
      </div>
      <div>
        <label>
          Description
          <input v-model.trim="form.description" type="text" />
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
      <div>
        <label>
          Billing Config JSON
          <textarea v-model="form.billingConfigText" rows="5" />
        </label>
      </div>
      <button type="submit" :disabled="submitting">创建渠道</button>
      <button type="button" @click="loadChannels" :disabled="loading">刷新</button>
    </form>

    <hr />

    <p v-if="loading">加载中...</p>
    <ul v-else-if="items.length > 0">
      <li v-for="item in items" :key="item.id">
        {{ item.name }} / {{ item.status }} / {{ item.description || '-' }}
      </li>
    </ul>
    <p v-else>暂无渠道</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
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
    const { data } = await apiRequest<ChannelItem[]>('/admin/channels', {
      accessToken: auth.accessToken
    })
    items.value = data
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
    await apiRequest('/admin/channels', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        name: form.name,
        description: form.description || null,
        status: form.status,
        billing_config: parseJSON(form.billingConfigText),
        metadata: {}
      }
    })

    form.name = ''
    form.description = ''
    form.billingConfigText = '{}'
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

