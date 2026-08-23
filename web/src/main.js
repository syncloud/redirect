import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './jQuery'
import { createPinia } from 'pinia'
import { useThemeStore } from './stores/theme'
import 'bootstrap'

async function start () {
  if (import.meta.env.VITE_STUB) {
    const { mock } = await import('./stub/api')
    mock()
  }

  const pinia = createPinia()

  createApp(App)
    .use(pinia)
    .use(router)
    .mount('#app')

  useThemeStore(pinia).init()
}

start()
