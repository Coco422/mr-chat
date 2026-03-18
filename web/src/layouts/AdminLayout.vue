<template>
  <div>
    <h1>Admin Console</h1>
    <nav>
      <RouterLink to="/admin/upstreams">Upstreams</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/channels">Channels</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/models">Models</RouterLink>
      <span> | </span>
      <RouterLink to="/admin/user-groups">User Groups</RouterLink>
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

<style scoped>
.admin-layout {
  display: flex;
  height: 100vh;
  background: var(--bg-primary);
}

.admin-sidebar {
  width: 260px;
  background: var(--bg-secondary);
  border-right: 1px solid var(--glass-border);
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  padding: 1.5rem 1rem;
  border-bottom: 1px solid var(--glass-border);
}

.sidebar-header h1 {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.admin-nav {
  flex: 1;
  padding: 1rem 0.5rem;
}

.nav-item {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1rem;
  margin-bottom: 0.25rem;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.nav-item:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.nav-item.router-link-active {
  background: var(--input-bg);
  color: var(--accent-primary);
}

.sidebar-footer {
  padding: 1rem;
  border-top: 1px solid var(--glass-border);
}

.back-btn {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.75rem 1rem;
  border-radius: 8px;
  text-decoration: none;
  color: var(--text-secondary);
  font-size: 0.9rem;
  transition: all 0.2s ease;
}

.back-btn:hover {
  background: var(--input-bg);
  color: var(--text-primary);
}

.admin-content {
  flex: 1;
  overflow-y: auto;
  background: var(--bg-primary);
}
</style>
