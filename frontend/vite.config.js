import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import { env } from 'process'

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue()],
  root: './',
  base: './',
  publicDir: './public',
  build: {
    outDir: '../cmd/build/frontend',
    emptyOutDir: true,
    sourcemap: true,
    rollupOptions: {
      input: {
        main: resolve(__dirname, 'index.html')
      }
    }
  },
  server: {
    port: 34125,
    strictPort: true,
    hmr: {
      overlay: false
    }
  },
  define: {
    'process.env': env
  }
})