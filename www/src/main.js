import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './jQuery'
import 'bootstrap'
import { mock } from './stub/api'

if (import.meta.env.DEV) {
  mock()
}

createApp(App)
  .use(router)
  .mount('#app')
