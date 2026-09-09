<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" data-testid="reset-form">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="reset-heading">{{ $t('passwordReset.heading') }}</h2>

      <div id="errors_placeholder" v-if="error !== ''">
        <div class="sc-alert">{{ error }}</div>
      </div>

      <div class="sc-field" v-if="error === ''">
        <label for="password">{{ $t('passwordReset.password') }}</label>
        <input id="password" data-testid="reset-password" type="password" required="" v-model="password">
      </div>

      <button id="reset" data-testid="reset-submit" class="sc-btn" v-if="error === ''" @click="reset">{{ $t('passwordReset.submit') }}</button>

      <div class="sc-auth-links">
        <router-link to="/login" data-testid="reset-login">{{ $t('passwordReset.backToLogin') }}</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'PasswordReset',
  data () {
    return {
      password: '',
      error: '',
      token: undefined
    }
  },
  mounted () {
    this.getToken()
  },
  methods: {
    getToken: function () {
      const token = this.$route.query.token
      if (token === undefined) {
        this.error = this.$t('passwordReset.noToken')
      }
      this.token = token
    },
    reset: function (event) {
      if (this.token !== undefined) {
        axios.post('api/user/set_password', { token: this.token, password: this.password })
          .then(_ => {
            this.$router.push('/login')
          })
          .catch(err => {
            if (err.response.status === 400) {
              if ('message' in err.response.data) {
                this.error = err.response.data.message
                return
              }
            }
            this.$router.push('/error')
          })
      }
      event.preventDefault()
    }
  }
}
</script>
<style>
@import '../style/form-center.css';
</style>
