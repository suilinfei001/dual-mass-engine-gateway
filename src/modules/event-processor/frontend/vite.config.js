import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: 3001,
    proxy: {
      '/api': {
        target: 'http://localhost:5002',
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
          // 将 node_modules 中的包打包到 vendor chunk
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
