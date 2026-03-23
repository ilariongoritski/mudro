import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import path from 'node:path'
import { defineConfig } from 'vite'

const apiProxyTarget = process.env.MUDRO_API_PROXY_TARGET ?? 'http://127.0.0.1:8080'
const bffProxyTarget = process.env.MUDRO_BFF_PROXY_TARGET ?? 'http://127.0.0.1:8086'

export default defineConfig({
  plugins: [tailwindcss(), react()],
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          'react-vendor': ['react', 'react-dom', 'react-router-dom'],
          'state-vendor': ['@reduxjs/toolkit', 'react-redux'],
          'motion-vendor': ['framer-motion'],
        },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      '/api/movie-catalog': {
        target: bffProxyTarget,
        changeOrigin: true,
      },
      '/api': apiProxyTarget,
      '/healthz': apiProxyTarget,
      '/feed': apiProxyTarget,
      '/media': apiProxyTarget,
    },
  },
})
