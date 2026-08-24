<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" id="form-forgot" data-testid="forgot-form" @submit="reset">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="forgot-heading">Forgot your password?</h2>
      <p class="sc-auth-sub">Enter your email address and we will send you a link to reset it.</p>

      <div class="sc-field">
        <label for="email">Email</label>
        <input id="email" name="email" data-testid="forgot-email" type="text" placeholder="user@mail.com" required="" v-model="email">
      </div>

      <button id="send" data-testid="forgot-send" class="sc-btn">Send reset link</button>

      <div class="sc-auth-links">
        <router-link to="/login" data-testid="forgot-login">Back to log in</router-link>
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
