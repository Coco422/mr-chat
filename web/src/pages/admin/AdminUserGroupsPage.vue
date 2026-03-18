<template>
  <section>
    <h1>User Groups</h1>
    <p v-if="errorMessage">{{ errorMessage }}</p>

    <form @submit.prevent="createUserGroup">
      <div>
        <label>
          名称
          <input v-model.trim="createForm.name" type="text" required />
        </label>
      </div>
      <div>
        <label>
          Description
          <input v-model.trim="createForm.description" type="text" />
        </label>
      </div>
      <div>
        <label>
          状态
          <select v-model="createForm.status">
            <option value="active">active</option>
            <option value="disabled">disabled</option>
          </select>
        </label>
      </div>
      <button type="submit" :disabled="submittingGroup">创建分组</button>
      <button type="button" @click="loadData" :disabled="loading">刷新</button>
    </form>

    <hr />

    <div>
      <label>
        当前分组
        <select v-model="selectedGroupID">
          <option value="">请选择</option>
          <option v-for="group in groups" :key="group.id" :value="group.id">
            {{ group.name }}
          </option>
        </select>
      </label>
      <button type="button" @click="loadPolicies" :disabled="!selectedGroupID || loadingPolicies">加载限额</button>
      <button type="button" @click="addPolicyRow" :disabled="!selectedGroupID">新增规则行</button>
      <button type="button" @click="savePolicies" :disabled="!selectedGroupID || submittingPolicies">保存全部规则</button>
    </div>

    <p v-if="loadingPolicies">限额加载中...</p>
    <div v-for="(policy, index) in policyRows" :key="`${policy.modelID}-${index}`">
      <hr />
      <label>
        Model
        <select v-model="policy.modelID">
          <option value="">默认模板</option>
          <option v-for="model in models" :key="model.id" :value="model.id">
            {{ model.display_name }}
          </option>
        </select>
      </label>
      <label>
        hour requests
        <input v-model.trim="policy.hourRequestLimit" type="number" />
      </label>
      <label>
        week requests
        <input v-model.trim="policy.weekRequestLimit" type="number" />
      </label>
      <label>
        lifetime requests
        <input v-model.trim="policy.lifetimeRequestLimit" type="number" />
      </label>
      <label>
        hour tokens
        <input v-model.trim="policy.hourTokenLimit" type="number" />
      </label>
      <label>
        week tokens
        <input v-model.trim="policy.weekTokenLimit" type="number" />
      </label>
      <label>
        lifetime tokens
        <input v-model.trim="policy.lifetimeTokenLimit" type="number" />
      </label>
      <label>
        status
        <select v-model="policy.status">
          <option value="active">active</option>
          <option value="disabled">disabled</option>
        </select>
      </label>
      <button type="button" @click="removePolicyRow(index)">删除</button>
    </div>

    <hr />

    <ul v-if="groups.length > 0">
      <li v-for="group in groups" :key="group.id">
        {{ group.name }} / {{ group.status }} / {{ group.description || '-' }}
      </li>
    </ul>
    <p v-else>暂无用户分组</p>
  </section>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'

import { ApiError, apiRequest } from '@/lib/api'
import { useAuthStore } from '@/stores/auth'

interface UserGroupItem {
  id: string
  name: string
  description: string | null
  status: string
}

interface ModelItem {
  id: string
  display_name: string
}

interface PolicyRow {
  modelID: string
  hourRequestLimit: string
  weekRequestLimit: string
  lifetimeRequestLimit: string
  hourTokenLimit: string
  weekTokenLimit: string
  lifetimeTokenLimit: string
  status: string
}

const auth = useAuthStore()
const loading = ref(false)
const loadingPolicies = ref(false)
const submittingGroup = ref(false)
const submittingPolicies = ref(false)
const errorMessage = ref('')
const groups = ref<UserGroupItem[]>([])
const models = ref<ModelItem[]>([])
const selectedGroupID = ref('')
const policyRows = ref<PolicyRow[]>([])
const createForm = reactive({
  name: '',
  description: '',
  status: 'active'
})

onMounted(async () => {
  await loadData()
})

async function loadData() {
  loading.value = true
  errorMessage.value = ''

  try {
    const [groupResponse, modelResponse] = await Promise.all([
      apiRequest<UserGroupItem[]>('/admin/user-groups', {
        accessToken: auth.accessToken
      }),
      apiRequest<ModelItem[]>('/admin/models', {
        accessToken: auth.accessToken
      })
    ])

    groups.value = groupResponse.data
    models.value = modelResponse.data

    if (!selectedGroupID.value && groups.value.length > 0) {
      selectedGroupID.value = groups.value[0].id
    }
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loading.value = false
  }
}

async function createUserGroup() {
  submittingGroup.value = true
  errorMessage.value = ''

  try {
    await apiRequest('/admin/user-groups', {
      method: 'POST',
      accessToken: auth.accessToken,
      body: {
        name: createForm.name,
        description: createForm.description || null,
        status: createForm.status,
        permissions: {},
        metadata: {}
      }
    })

    createForm.name = ''
    createForm.description = ''
    await loadData()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submittingGroup.value = false
  }
}

async function loadPolicies() {
  if (!selectedGroupID.value) {
    return
  }

  loadingPolicies.value = true
  errorMessage.value = ''

  try {
    const { data } = await apiRequest<Array<Record<string, unknown>>>(`/admin/user-groups/${selectedGroupID.value}/limits`, {
      accessToken: auth.accessToken
    })
    policyRows.value = data.map((item) => ({
      modelID: (item.model_id as string | null) ?? '',
      hourRequestLimit: toText(item.hour_request_limit),
      weekRequestLimit: toText(item.week_request_limit),
      lifetimeRequestLimit: toText(item.lifetime_request_limit),
      hourTokenLimit: toText(item.hour_token_limit),
      weekTokenLimit: toText(item.week_token_limit),
      lifetimeTokenLimit: toText(item.lifetime_token_limit),
      status: String(item.status ?? 'active')
    }))

    if (policyRows.value.length === 0) {
      addPolicyRow()
    }
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    loadingPolicies.value = false
  }
}

async function savePolicies() {
  if (!selectedGroupID.value) {
    return
  }

  submittingPolicies.value = true
  errorMessage.value = ''

  try {
    await apiRequest(`/admin/user-groups/${selectedGroupID.value}/limits`, {
      method: 'PUT',
      accessToken: auth.accessToken,
      body: {
        policies: policyRows.value.map((row) => ({
          model_id: row.modelID || null,
          hour_request_limit: toNumberOrNull(row.hourRequestLimit),
          week_request_limit: toNumberOrNull(row.weekRequestLimit),
          lifetime_request_limit: toNumberOrNull(row.lifetimeRequestLimit),
          hour_token_limit: toNumberOrNull(row.hourTokenLimit),
          week_token_limit: toNumberOrNull(row.weekTokenLimit),
          lifetime_token_limit: toNumberOrNull(row.lifetimeTokenLimit),
          status: row.status
        }))
      }
    })

    await loadPolicies()
  } catch (error) {
    errorMessage.value = toErrorMessage(error)
  } finally {
    submittingPolicies.value = false
  }
}

function addPolicyRow() {
  policyRows.value.push({
    modelID: '',
    hourRequestLimit: '',
    weekRequestLimit: '',
    lifetimeRequestLimit: '',
    hourTokenLimit: '',
    weekTokenLimit: '',
    lifetimeTokenLimit: '',
    status: 'active'
  })
}

function removePolicyRow(index: number) {
  policyRows.value.splice(index, 1)
  if (policyRows.value.length === 0) {
    addPolicyRow()
  }
}

function toText(value: unknown) {
  return value == null ? '' : String(value)
}

function toNumberOrNull(value: string) {
  return value.trim() === '' ? null : Number(value)
}

function toErrorMessage(error: unknown) {
  if (error instanceof ApiError) {
    return `${error.code}: ${error.message}`
  }
  return '请求失败'
}
</script>
