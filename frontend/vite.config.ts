import react from '@vitejs/plugin-react'
import path from 'node:path'
import { defineConfig } from 'vite'

const apiProxyTarget = process.env.MUDRO_API_PROXY_TARGET ?? 'http://127.0.0.1:18080'

export default defineConfig({
  plugins: [react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api': apiProxyTarget,
      '/healthz': apiProxyTarget,
      '/feed': apiProxyTarget,
      '/media': apiProxyTarget,
    },
  },
})
