<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>用户分组管理</h1>
      <button class="primary-btn" @click="showCreateForm = !showCreateForm">
        {{ showCreateForm ? '取消' : '+ 新建分组' }}
      </button>
    </div>

    <p v-if="errorMessage" class="error">{{ errorMessage }}</p>

    <div v-if="showCreateForm" class="form-card">
      <h2>创建分组</h2>
      <form @submit.prevent="createUserGroup" class="admin-form">
        <div class="form-row">
          <div class="form-group">
            <label>名称</label>
            <input v-model.trim="createForm.name" type="text" required />
          </div>
          <div class="form-group">
            <label>状态</label>
            <el-select v-model="createForm.status" style="width: 100%">
              <el-option value="active" label="active" />
              <el-option value="disabled" label="disabled" />
            </el-select>
          </div>
        </div>
        <div class="form-group">
          <label>Description</label>
          <input v-model.trim="createForm.description" type="text" />
        </div>
        <button type="submit" :disabled="submittingGroup" class="submit-btn">创建分组</button>
      </form>
    </div>

    <div class="table-card">
      <div class="table-header">
        <h2>限额策略配置</h2>
        <button class="refresh-btn" @click="loadData" :disabled="loading">刷新</button>
      </div>

      <div class="form-group" style="margin-bottom: 1.5rem;">
        <label>当前分组</label>
        <el-select v-model="selectedGroupID" style="width: 100%" @change="loadPolicies">
          <el-option value="" label="请选择" />
          <el-option v-for="group in groups" :key="group.id" :value="group.id" :label="group.name" />
        </el-select>
      </div>

      <div v-if="selectedGroupID" style="margin-bottom: 1rem; display: flex; gap: 0.5rem;">
        <button class="refresh-btn" @click="loadPolicies" :disabled="loadingPolicies">加载限额</button>
        <button class="refresh-btn" @click="addPolicyRow">新增规则行</button>
        <button class="primary-btn" @click="savePolicies" :disabled="submittingPolicies">保存全部规则</button>
      </div>

      <p v-if="loadingPolicies" class="loading">限额加载中...</p>
      <div v-else-if="policyRows.length > 0" class="policy-list">
        <div v-for="(policy, index) in policyRows" :key="`${policy.modelID}-${index}`" class="policy-row">
          <div class="form-row">
            <div class="form-group">
              <label>Model</label>
              <el-select v-model="policy.modelID" style="width: 100%">
                <el-option value="" label="默认模板" />
                <el-option v-for="model in models" :key="model.id" :value="model.id" :label="model.display_name" />
              </el-select>
            </div>
            <div class="form-group">
              <label>状态</label>
              <el-select v-model="policy.status" style="width: 100%">
                <el-option value="active" label="active" />
                <el-option value="disabled" label="disabled" />
              </el-select>
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Hour Requests</label>
              <input v-model.trim="policy.hourRequestLimit" type="number" />
            </div>
            <div class="form-group">
              <label>Week Requests</label>
              <input v-model.trim="policy.weekRequestLimit" type="number" />
            </div>
            <div class="form-group">
              <label>Lifetime Requests</label>
              <input v-model.trim="policy.lifetimeRequestLimit" type="number" />
            </div>
          </div>
          <div class="form-row">
            <div class="form-group">
              <label>Hour Tokens</label>
              <input v-model.trim="policy.hourTokenLimit" type="number" />
            </div>
            <div class="form-group">
              <label>Week Tokens</label>
              <input v-model.trim="policy.weekTokenLimit" type="number" />
            </div>
            <div class="form-group">
              <label>Lifetime Tokens</label>
              <input v-model.trim="policy.lifetimeTokenLimit" type="number" />
            </div>
          </div>
          <button type="button" @click="removePolicyRow(index)" class="refresh-btn">删除</button>
        </div>
      </div>
    </div>

    <div class="table-card">
      <h2>分组列表</h2>
      <table v-if="groups.length > 0">
        <thead>
          <tr>
            <th>名称</th>
            <th>状态</th>
            <th>描述</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="group in groups" :key="group.id">
            <td>{{ group.name }}</td>
            <td><span class="status-badge" :class="group.status">{{ group.status }}</span></td>
            <td>{{ group.description || '-' }}</td>
          </tr>
        </tbody>
      </table>
      <p v-else class="empty">暂无用户分组</p>
    </div>
  </div>
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
const showCreateForm = ref(false)
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
    showCreateForm.value = false
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

<style scoped>
@import '@/styles/admin.css';

.policy-list {
  display: flex;
  flex-direction: column;
  gap: 1.5rem;
}

.policy-row {
  padding: 1.5rem;
  background: var(--bg-primary);
  border: 1px solid var(--input-border);
  border-radius: 8px;
}
</style>
