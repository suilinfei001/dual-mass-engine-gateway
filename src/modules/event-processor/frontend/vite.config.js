import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    host: '0.0.0.0',
    port: 8084,
    proxy: {
      // 认证 API -> Auth Service (4007)
      '/api/auth': {
        target: 'http://localhost:4007',
        changeOrigin: true
      },
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
      // 任务相关 API -> Task Scheduler (4003)
      '/api/tasks': {
        target: 'http://localhost:4003',
        changeOrigin: true
      },
      '/api/executions': {
        target: 'http://localhost:4004',
        changeOrigin: true
      },
      // AI 分析 API -> AI Analyzer (4005)
      '/api/analyze': {
        target: 'http://localhost:4005',
        changeOrigin: true
      },
      // 资源池 API -> Resource Manager (4006)
      '/api/resources': {
        target: 'http://localhost:4006',
        changeOrigin: true
      },
      '/api/resource-pool': {
        target: 'http://localhost:4006',
        changeOrigin: true
      },
      '/api/testbeds': {
        target: 'http://localhost:4006',
        changeOrigin: true
      },
      // 事件查询 -> Event Store (4002)
      '/api/events': {
        target: 'http://localhost:4002',
        changeOrigin: true
      }
    }
  },
  build: {
    outDir: 'dist',
    // 优化代码分割，减少 chunk 大小和数量
    rollupOptions: {
      output: {
        manualChunks(id) {
          // 将 node_modules 中的包 打包到 vendor chunk
          if (id.includes('node_modules')) {
            return 'vendor'
          }
        }
      }
    },
    // 优化 chunk 加载策略
    chunkSizeWarningLimit: 1000
  }
})
