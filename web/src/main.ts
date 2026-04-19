import { createApp } from 'vue'
import { createPinia } from 'pinia'
import 'element-plus/es/components/icon/style/css'
import 'element-plus/es/components/message/style/css'

import App from './App.vue'
import { setupPerformanceMonitor } from './lib/performance'
import router from './router'
import './styles/theme.css'

// 初始化主题
const savedTheme = localStorage.getItem('theme') || 'dark'
document.documentElement.setAttribute('data-theme', savedTheme)

const app = createApp(App)

app.use(createPinia())
app.use(router)

setupPerformanceMonitor({
  // 默认开发环境开启，生产环境可以通过 VITE_ENABLE_PERF_MONITOR=true 手动开启。
  enabled: import.meta.env.DEV || import.meta.env.VITE_ENABLE_PERF_MONITOR === 'true',
  // 开发环境把指标打印到控制台，方便你先理解每个指标含义。
  debug: import.meta.env.DEV,
  router
})

app.mount('#app')
