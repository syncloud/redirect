import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './jQuery'
import { initTheme } from './theme'
import 'bootstrap'

async function start () {
  initTheme()
  if (import.meta.env.VITE_STUB) {
    const { mock } = await import('./stub/api')
    mock()
  }

  createApp(App)
    .use(router)
    .mount('#app')
}

start()
