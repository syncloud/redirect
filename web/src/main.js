import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './jQuery'
import 'bootstrap'

async function start () {
  if (import.meta.env.VITE_STUB) {
    const { mock } = await import('./stub/api')
    mock()
  }

  createApp(App)
    .use(router)
    .mount('#app')
}

start()
