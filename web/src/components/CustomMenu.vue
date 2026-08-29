<template>
  <header class="sc-header">
    <div class="sc-header-inner">
      <router-link class="sc-logo" to="/" data-testid="menu-brand">
        <img class="sc-logo-img" src="/logo.svg" alt="">
        <span class="sc-logo-name">SYNCLOUD</span>
      </router-link>

      <button
        class="sc-burger"
        type="button"
        aria-label="menu"
        data-testid="menu-burger"
        @click="open = !open"
      >
        <span></span><span></span><span></span>
      </button>

      <nav class="sc-nav" :class="{ open }" data-testid="menu-nav">
        <router-link
          v-if="loggedIn"
          id="devices"
          to="/"
          data-testid="nav-devices"
          :class="{ active: activeTab === '/' }"
          @click="open = false"
        >Devices</router-link>
        <router-link
          v-if="loggedIn"
          id="buy"
          to="/device"
          data-testid="nav-buy"
          :class="{ active: activeTab === '/device' }"
          @click="open = false"
        >Buy</router-link>
        <router-link
          v-if="loggedIn"
          id="account"
          to="/account"
          data-testid="nav-account"
          :class="{ active: activeTab === '/account' }"
          @click="open = false"
        >Account</router-link>
        <span v-if="loggedIn" class="sc-nav-email" data-testid="menu-email">{{ email }}</span>
        <button
          v-if="loggedIn"
          id="logout"
          class="sc-nav-action"
          data-testid="nav-logout"
          @click="logout"
        >Log out</button>
      </nav>

      <div class="sc-header-actions">
        <ThemeToggle/>
      </div>
    </div>
  </header>
</template>

<script>
import axios from 'axios'
import ThemeToggle from './ThemeToggle.vue'

export default {
  name: 'CustomMenu',
  components: { ThemeToggle },
  props: {
    activeTab: String,
    email: String,
    loggedIn: Boolean,
    checkUserSession: Function
  },
  data () {
    return { open: false }
  },
  methods: {
    logout: function (_) {
      this.open = false
      axios.post('/api/logout')
        .then(_ => {
          this.checkUserSession()
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
}
</script>
