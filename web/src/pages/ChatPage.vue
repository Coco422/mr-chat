<template>
  <section>
    <h1>Chat</h1>
    <p>这里暂时只保留联调入口，Chat API 后续继续补。</p>
    <p>conversationId: {{ route.params.conversationId ?? 'new' }}</p>
    <p>current user: {{ auth.user?.username ?? 'unknown' }}</p>
  </section>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'

import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const route = useRoute()
</script>
