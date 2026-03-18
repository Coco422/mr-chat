<template>
  <div class="admin-page">
    <div class="page-header">
      <h1>兑换码管理</h1>
    </div>

    <div class="placeholder-card">
      <h2>功能开发中</h2>
      <p class="muted">后续承接批量生成、批次列表和兑换记录。</p>
      <ul class="feature-list">
        <li>批量生成兑换码</li>
        <li>按批次查看列表</li>
        <li>兑换历史记录</li>
      </ul>
    </div>
  </div>
</template>

<style scoped>
@import '@/styles/admin.css';

.placeholder-card {
  background: var(--input-bg);
  border: 1px solid var(--input-border);
  border-radius: 12px;
  padding: 3rem;
  text-align: center;
}

.placeholder-card h2 {
  font-size: 1.5rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 1rem;
}

.muted {
  color: var(--text-secondary);
  margin-bottom: 2rem;
}

.feature-list {
  list-style: none;
  padding: 0;
  display: inline-block;
  text-align: left;
}

.feature-list li {
  padding: 0.5rem 0;
  color: var(--text-secondary);
}

.feature-list li::before {
  content: '•';
  color: var(--accent-primary);
  font-weight: bold;
  display: inline-block;
  width: 1em;
  margin-left: -1em;
}
</style>