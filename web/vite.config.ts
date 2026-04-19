import { fileURLToPath, URL } from 'node:url'

import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  build: {
    rollupOptions: {
      output: {
        manualChunks(id) {
          if (!id.includes('node_modules')) {
            return
          }

          if (id.includes('element-plus') || id.includes('@element-plus')) {
            return 'vendor-element-plus'
          }

          if (
            id.includes('/vue/') ||
            id.includes('/pinia/') ||
            id.includes('/vue-router/') ||
            id.includes('/@vue/')
          ) {
            return 'vendor-vue'
          }

          if (
            id.includes('/markdown-it/') ||
            id.includes('/highlight.js/') ||
            id.includes('/dompurify/') ||
            id.includes('/qrcode/')
          ) {
            return 'vendor-content'
          }

          if (id.includes('/axios/') || id.includes('/qs/')) {
            return 'vendor-network'
          }

          return 'vendor-misc'
        }
      }
    }
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://192.168.1.144:8080',
        changeOrigin: true,
        secure: false,
        rewrite: (path) => path.replace(/^\/api/, '/api')
      }
    }
  }
})
