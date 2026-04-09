import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    host: '0.0.0.0',
    port: 8083,
    proxy: {
      // 认证 API -> Auth Service (4007)
      '/api/login': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      '/api/logout': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      '/api/check-login': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
      // 事件相关 API -> Event Store (4002)
      '/api/events': {
        target: 'http://localhost:4002',
        changeOrigin: true
      },
      '/api/quality-checks': {
        target: 'http://localhost:4002',
        changeOrigin: true
      },
      // Webhook 相关 -> Webhook Gateway (4001)
      '/api/webhook': {
        target: 'http://localhost:4001',
        changeOrigin: true
      },
      '/api/mock': {
        target: 'http://localhost:4001',
        changeOrigin: true
      }
    }
  }
})
