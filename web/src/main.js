import { createApp } from 'vue'
import { captureGclid } from './attribution'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { useThemeStore } from './stores/theme'

async function start () {
  if (import.meta.env.VITE_STUB) {
    const { mock } = await import('./stub/api')
    mock()
  }

  const pinia = createPinia()

  captureGclid()

  createApp(App)
    .use(pinia)
    .use(router)
    .mount('#app')

  useThemeStore(pinia).init()
}

start()
