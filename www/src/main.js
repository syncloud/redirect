import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import 'bootstrap'
import { mock } from './stub/api'

if (process.env.NODE_ENV === 'development') {
  mock()
}

createApp(App)
  .use(router)
  .mount('#app')
