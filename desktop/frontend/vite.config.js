import vue from '@vitejs/plugin-vue'
import { defineConfig } from 'vite'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'

const rootPath = new URL('.', import.meta.url).pathname

export default defineConfig({
  plugins: [
    vue(),
  ],
  resolve: {
    alias: {
      '@': rootPath + 'src',
      wailsjs: rootPath + 'wailsjs',
    },
  },
})
