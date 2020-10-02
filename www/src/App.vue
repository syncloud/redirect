<template>
  <Menu v-bind:activeTab="currentPath" v-bind:onLogout="checkUserSession" v-bind:loggedIn="loggedIn" v-bind:email="email"/>
  <router-view v-bind:onLogin="checkUserSession" v-bind:onLogout="checkUserSession"/>
</template>
<script>
import axios from 'axios'
import Menu from '@/components/Menu'

global.jQuery = require('jquery')
var $ = global.jQuery
window.jQuery = window.$ = $

const publicRoutes = [
  '/register',
  '/activate',
  '/forgot',
  '/reset',
  '/error',
  '/login',
  ''
]

export default {
  data: function () {
    return {
      currentPath: '',
      loggedIn: undefined,
      email: ''
    }
  },
  name: 'app',
  components: {
    Menu
  },
  watch: {
    $route (to, from) {
      console.log('route change from ' + from.path + ' to ' + to.path)
      this.currentPath = to.path
    }
  },
  methods: {
    checkUserSession: function () {
      axios.get('/api/user/get')
        .then(response => {
          this.email = response.data.email
          this.loggedIn = true
          if (this.currentPath === '/login') {
            this.$router.push('/')
          }
        })
        .catch(_ => {
          this.email = ''
          this.loggedIn = false
          if (!publicRoutes.includes(this.currentPath)) {
            console.log('redirect to login from ' + this.currentPath)
            this.$router.push('/login')
          }
        })
    }
  },
  mounted () {
    this.checkUserSession()
  }
}
</script>
<style>
@import '~bootstrap/dist/css/bootstrap.css';
</style>
