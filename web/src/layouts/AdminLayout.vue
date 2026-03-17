<template>
  <div>
    <h1>Admin Console</h1>
    <nav>
      <RouterLink to="/admin/upstreams">Upstreams</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/models">Models</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/users">Users</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/redeem-codes">Redeem Codes</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/audit-logs">Audit Logs</RouterLink>
      <span> | </span>
      <RouterLink to="/chat">Back to App</RouterLink>
    </nav>
    <RouterView />
  </div>
</template>

<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
</script>
