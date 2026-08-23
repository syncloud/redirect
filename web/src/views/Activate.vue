<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <div class="sc-auth-card">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title">Activation</h2>
      <p class="sc-auth-message" id="activated" data-testid="activate-message">
        {{ message }}
      </p>
      <p class="sc-auth-note">
        <router-link to="/login" data-testid="activate-login">Continue to log in</router-link>
      </p>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Activate',
  data () {
    return {
      message: ''
    }
  },
  mounted () {
    const token = this.$route.query.token
    if (token === undefined) {
      this.message = 'Unknown token'
    } else {
      this.activate(token)
    }
  },
  methods: {
    activate: function (token) {
      axios.post('/api/user/activate', { token: token })
        .then(response => {
          if (response.data.data) {
            this.message = response.data.data
          }
        })
        .catch(err => {
          if (err.response.status === 400) {
            if (err.response.data.message) {
              this.message = err.response.data.message
              return
            }
          }
          this.$router.push('/error')
        })
    }
  }
}
</script>
<style>
</style>
