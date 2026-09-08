import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import AutoImport from 'unplugin-auto-import/vite'
import Components from 'unplugin-vue-components/vite'
import { ElementPlusResolver } from 'unplugin-vue-components/resolvers'

function stubPaypalSdk () {
  const file = fileURLToPath(new URL('./src/stub/paypal-sdk.js', import.meta.url))
  return {
    name: 'stub-paypal-sdk',
    apply: 'serve',
    configureServer (server) {
      server.middlewares.use('/stub/paypal-sdk.js', (_req, res) => {
        res.setHeader('Content-Type', 'application/javascript')
        res.end(readFileSync(file, 'utf-8'))
      })
    }
  }
}

export default defineConfig(({ command, mode, ssrBuild }) => {
  return {
    plugins: [
      AutoImport({
        resolvers: [ElementPlusResolver()]
      }),
      Components({
        resolvers: [ElementPlusResolver()]
      }),
      vue(),
      stubPaypalSdk()
    ],
    server: {
      proxy: {
        '/api': {
          target: 'http://localhost:3002',
          changeOrigin: true,
          rewrite: (path) => path.replace(/^\/api/, '')
        }
      }
    }
  }
})
