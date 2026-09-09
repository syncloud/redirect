<template>
  <header class="sc-header">
    <div class="sc-header-inner">
      <router-link class="sc-logo" :to="loggedIn ? '/' : '/shop'" data-testid="menu-brand">
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
        >{{ $t('menu.devices') }}</router-link>
        <router-link
          id="shop"
          to="/shop"
          data-testid="nav-shop"
          :class="{ active: activeTab === '/shop' }"
          @click="open = false"
        >{{ $t('menu.shop') }}</router-link>
        <router-link
          v-if="loggedIn"
          id="orders"
          to="/orders"
          data-testid="nav-orders"
          :class="{ active: activeTab === '/orders' }"
          @click="open = false"
        >{{ $t('menu.orders') }}</router-link>
        <router-link
          v-if="loggedIn && admin"
          id="admin-orders"
          to="/admin/orders"
          data-testid="nav-admin-orders"
          :class="{ active: activeTab === '/admin/orders' }"
          @click="open = false"
        >{{ $t('menu.allOrders') }}</router-link>
        <router-link
          v-if="loggedIn"
          id="account"
          to="/account"
          data-testid="nav-account"
          :class="{ active: activeTab === '/account' }"
          @click="open = false"
        >{{ $t('menu.account') }}</router-link>
        <span v-if="loggedIn" class="sc-nav-email" data-testid="menu-email">{{ email }}</span>
        <router-link
          v-if="loggedIn === false"
          id="login"
          to="/login"
          data-testid="nav-login"
          :class="{ active: activeTab === '/login' }"
          @click="open = false"
        >{{ $t('menu.login') }}</router-link>
        <button
          v-if="loggedIn"
          id="logout"
          class="sc-nav-action"
          data-testid="nav-logout"
          @click="logout"
        >{{ $t('menu.logout') }}</button>
      </nav>

      <div class="sc-header-actions">
        <LanguageSwitcher/>
        <ThemeToggle/>
      </div>
    </div>
  </header>
</template>

<script>
import axios from 'axios'
import ThemeToggle from './ThemeToggle.vue'
import LanguageSwitcher from './LanguageSwitcher.vue'

export default {
  name: 'CustomMenu',
  components: { ThemeToggle, LanguageSwitcher },
  props: {
    activeTab: String,
    email: String,
    admin: Boolean,
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
          this.$router.push('/login')
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
}
</script>
