import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import { mock } from './stub/api'
import './jQuery'
import 'bootstrap'
if (import.meta.env.DEV) {
  mock()
}

createApp(App)
  .use(router)
  .mount('#app')
