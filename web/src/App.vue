<template>
  <CustomMenu v-if="!bare" v-bind:activeTab="currentPath" v-bind:checkUserSession="checkUserSession" v-bind:loggedIn="loggedIn"
        v-bind:email="email"/>
  <router-view v-bind:checkUserSession="checkUserSession" v-bind:loggedIn="loggedIn"/>
</template>
<script>
import axios from 'axios'
import CustomMenu from './components/CustomMenu.vue'

const bareRoutes = [
  '/login',
  '/register',
  '/forgot',
  '/reset',
  '/check-email',
  '/activate'
]

const publicRoutes = [
  '/shop',
  '/register',
  '/activate',
  '/forgot',
  '/reset',
  '/error',
  '/login',
  '/privacy',
  '/check-email',
  '/',
  ''
]

export default {
  name: 'app',
  components: {
    CustomMenu
  },
  data () {
    return {
      currentPath: '',
      loggedIn: undefined,
      email: ''
    }
  },
  computed: {
    bare () {
      return bareRoutes.includes(this.currentPath)
    }
  },
  mounted () {
  },
  beforeUnmount () {
  },
  watch: {
    $route (to, from) {
      // console.log('route change from ' + from.path + ' to ' + to.path)
      this.currentPath = to.path
      this.checkUserSession()
    }
  },
  methods: {
    checkUserSession: function () {
      axios.get('/api/user')
        .then(response => {
          this.email = response.data.email
          this.loggedIn = true
          if (this.currentPath === '/login') {
            this.$router.push(this.$route.query.next || '/')
          }
        })
        .catch(_ => {
          this.email = ''
          this.loggedIn = false
          if (!publicRoutes.includes(this.currentPath)) {
            // console.log('redirect to login from ' + this.currentPath)
            this.$router.push('/login')
          }
        })
    }
  }
}
</script>
<style>
@import 'element-plus/theme-chalk/dark/css-vars.css';
@import './style/design.css';
</style>
