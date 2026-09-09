<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" id="form-forgot" data-testid="forgot-form" @submit="reset">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="forgot-heading">{{ $t('passwordForgot.heading') }}</h2>
      <p class="sc-auth-sub">{{ $t('passwordForgot.subtitle') }}</p>

      <div class="sc-field">
        <label for="email">{{ $t('passwordForgot.email') }}</label>
        <input id="email" name="email" data-testid="forgot-email" type="text" :placeholder="$t('passwordForgot.emailPlaceholder')" required="" v-model="email">
      </div>

      <button id="send" data-testid="forgot-send" class="sc-btn">{{ $t('passwordForgot.submit') }}</button>

      <div class="sc-auth-links">
        <router-link to="/login" data-testid="forgot-login">{{ $t('passwordForgot.backToLogin') }}</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'PasswordForgot',
  data () {
    return {
      email: ''
    }
  },
  methods: {
    reset: function (event) {
      axios.post('api/user/reset_password', { email: this.email })
        .then(_ => {
          this.$router.push('/check-email')
        })
        .catch(_ => {
          this.$router.push('/error')
        })
      event.preventDefault()
    }
  }
}
</script>
<style>
@import '../style/form-center.css';
</style>
