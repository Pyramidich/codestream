import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  base: '/~s503255/codestream/',
  plugins: [react()],
  resolve: {
    alias: {
      'monaco-editor/esm/vs/editor/editor.api.js': 'monaco-editor',
    },
  },
  build: {
    rollupOptions: {
      output: {
        manualChunks(id: string) {
          if (id.includes('node_modules/react') || id.includes('node_modules/react-router')) {
            return 'react-vendor'
          }
          if (id.includes('node_modules/@monaco-editor/react') || id.includes('node_modules/monaco-editor')) {
            return 'editor'
          }
          if (id.includes('node_modules/yjs') || id.includes('node_modules/y-monaco') || id.includes('node_modules/y-protocols')) {
            return 'yjs'
          }
          if (id.includes('node_modules/axios')) {
            return 'vendor'
          }
        },
      },
    },
  },
})
