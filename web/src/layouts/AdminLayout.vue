<template>
  <div class="admin-layout">
    <aside class="admin-sidebar">
      <div class="sidebar-header">
        <h1>管理后台</h1>
      </div>

      <nav class="admin-nav">
        <RouterLink to="/admin/upstreams" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <circle cx="12" cy="12" r="10" stroke-width="2"/><line x1="2" y1="12" x2="22" y2="12" stroke-width="2"/>
            <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" stroke-width="2"/>
          </svg>
          上游管理
        </RouterLink>
        <RouterLink to="/admin/models" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <rect x="3" y="3" width="7" height="7" stroke-width="2"/><rect x="14" y="3" width="7" height="7" stroke-width="2"/>
            <rect x="14" y="14" width="7" height="7" stroke-width="2"/><rect x="3" y="14" width="7" height="7" stroke-width="2"/>
          </svg>
          模型管理
        </RouterLink>
        <RouterLink to="/admin/users" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" stroke-width="2"/>
            <circle cx="9" cy="7" r="4" stroke-width="2"/><path d="M23 21v-2a4 4 0 0 0-3-3.87" stroke-width="2"/>
            <path d="M16 3.13a4 4 0 0 1 0 7.75" stroke-width="2"/>
          </svg>
          用户管理
        </RouterLink>
        <RouterLink to="/admin/redeem-codes" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <rect x="1" y="4" width="22" height="16" rx="2" ry="2" stroke-width="2"/>
            <line x1="1" y1="10" x2="23" y2="10" stroke-width="2"/>
          </svg>
          兑换码
        </RouterLink>
        <RouterLink to="/admin/audit-logs" class="nav-item">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" stroke-width="2"/>
            <polyline points="14 2 14 8 20 8" stroke-width="2"/><line x1="16" y1="13" x2="8" y2="13" stroke-width="2"/>
            <line x1="16" y1="17" x2="8" y2="17" stroke-width="2"/><polyline points="10 9 9 9 8 9" stroke-width="2"/>
          </svg>
          审计日志
        </RouterLink>
      </nav>

      <div class="sidebar-footer">
        <RouterLink to="/chat" class="back-btn">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor">
            <line x1="19" y1="12" x2="5" y2="12" stroke-width="2"/>
            <polyline points="12 19 5 12 12 5" stroke-width="2"/>
          </svg>
          返回应用
        </RouterLink>
      </div>
    </aside>

    <main class="admin-content">
      <RouterView />
    </main>
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
