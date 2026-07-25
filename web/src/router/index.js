import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  {
    path: '/privacy',
    name: 'Privacy',
    component: () => import('../views/PrivacyPolicy.vue')
  },
  {
    path: '/error',
    name: 'Error',
    component: () => import('../views/Error.vue')
  },
  {
    path: '/reset',
    name: 'Reset',
    component: () => import('../views/PasswordReset.vue')
  },
  {
    path: '/activate',
    name: 'Activate',
    component: () => import('../views/Activate.vue')
  },
  {
    path: '/forgot',
    name: 'Forgot',
    component: () => import('../views/PasswordForgot.vue')
  },
  {
    path: '/account',
    name: 'Account',
    component: () => import('../views/Account.vue')
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue')
  },
  {
    path: '/check-email',
    name: 'CheckEmail',
    component: () => import('../views/CheckEmail.vue')
  },
  {
    path: '/',
    name: 'Devices',
    component: () => import('../views/Devices.vue')
  }
]

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes
})

export default router
