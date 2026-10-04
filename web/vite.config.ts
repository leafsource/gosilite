import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

const apiTarget = process.env.GOSILITE_API_URL ?? 'http://127.0.0.1:18080'

export default defineConfig({
  plugins: [vue()],
  server: {
    port: Number(process.env.GOSILITE_WEB_PORT ?? 5173),
    proxy: {
      '/api': {
        target: apiTarget,
        changeOrigin: true,
      },
    },
  },
})
