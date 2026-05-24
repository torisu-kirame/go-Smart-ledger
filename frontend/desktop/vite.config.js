import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { fileURLToPath, URL } from 'node:url'

export default defineConfig({
  plugins: [vue()],
  optimizeDeps: {
    exclude: ['sql.js'],
  },
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url)),
    },
  },
  server: {
    port: 25173,
    proxy: {
      '/api': {
        target: 'http://localhost:28080',
        changeOrigin: true,
      },
      '/dashboard': { target: 'http://localhost:24441', changeOrigin: true },
      '/status': { target: 'http://localhost:24441', changeOrigin: true },
      '/blocks': { target: 'http://localhost:24441', changeOrigin: true },
      '/tx': { target: 'http://localhost:24441', changeOrigin: true },
      '/state': { target: 'http://localhost:24441', changeOrigin: true },
      '/peers': { target: 'http://localhost:24441', changeOrigin: true },
      '/consensus': { target: 'http://localhost:24441', changeOrigin: true },
      '/search': { target: 'http://localhost:24441', changeOrigin: true },
      '/contracts': { target: 'http://localhost:24441', changeOrigin: true },
      '/proposals': { target: 'http://localhost:24441', changeOrigin: true },
      '/identity': { target: 'http://localhost:24441', changeOrigin: true },
    },
  },
})
