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
    <h1 class="sc-h1">Your Syncloud account</h1>
    <p class="sc-lead">
      This is where a Syncloud device gets its name on the internet. The account holds the
      domain name for each device you own, keeps its certificate current, and carries the
      subscription.
    </p>

    <div class="sc-card">
      <h2 class="sc-h2">Already have a device</h2>
      <p>Sign in to see your devices, their addresses and how much of your plan you are using.</p>
      <router-link
        class="sc-btn"
        to="/login"
        data-testid="home-login"
      >
        Log in
      </router-link>
    </div>

    <div class="sc-card">
      <h2 class="sc-h2">Do not have one yet</h2>
      <p>
        Buy a device assembled and ready, or build your own from a disk image and activate it
        against a new account.
      </p>
      <router-link
        class="sc-btn"
        to="/shop"
        data-testid="home-shop"
      >
        See the shop
      </router-link>
      <a
        class="home-build"
        href="https://syncloud.org/setup"
        data-testid="home-build"
      >Build one yourself</a>
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
