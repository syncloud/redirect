import { createApp } from 'vue'
import { captureGclid } from './attribution'
import App from './App.vue'
import router from './router'
import { createPinia } from 'pinia'
import { useThemeStore } from './stores/theme'
import i18n, { detectLocale, setLocale } from './i18n'

async function start () {
  if (import.meta.env.VITE_STUB) {
    const { mock } = await import('./stub/api')
    mock()
  }

  const pinia = createPinia()

  captureGclid()
  await setLocale(detectLocale())

  createApp(App)
    .use(pinia)
    .use(router)
    .use(i18n)
    .mount('#app')

  useThemeStore(pinia).init()
}

start()
