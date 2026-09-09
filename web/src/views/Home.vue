<template>
  <Devices
    v-if="loggedIn"
    :check-user-session="checkUserSession"
  />

  <div
    v-else-if="loggedIn === false"
    class="sc-page"
    data-testid="home-intro"
  >
    <h1 class="sc-h1">{{ $t('home.accountTitle') }}</h1>
    <p class="sc-lead">
      {{ $t('home.accountLead') }}
    </p>

    <div class="sc-card">
      <h2 class="sc-h2">{{ $t('home.haveDeviceTitle') }}</h2>
      <p>{{ $t('home.haveDeviceText') }}</p>
      <router-link
        class="sc-btn"
        to="/login"
        data-testid="home-login"
      >
        {{ $t('home.login') }}
      </router-link>
    </div>

    <div class="sc-card">
      <h2 class="sc-h2">{{ $t('home.noDeviceTitle') }}</h2>
      <p>
        {{ $t('home.noDeviceText') }}
      </p>
      <router-link
        class="sc-btn"
        to="/shop"
        data-testid="home-shop"
      >
        {{ $t('home.shop') }}
      </router-link>
      <a
        class="home-build"
        href="https://syncloud.org/setup"
        data-testid="home-build"
      >{{ $t('home.build') }}</a>
    </div>
  </div>
</template>

<script>
import Devices from './Devices.vue'

export default {
  name: 'HomeView',
  components: { Devices },
  props: {
    checkUserSession: Function,
    loggedIn: Boolean
  }
}
</script>

<style scoped>
.sc-card + .sc-card {
  margin-top: 20px;
}

.home-build {
  display: inline-block;
  margin-left: 16px;
  color: var(--sc-primary);
}
</style>
