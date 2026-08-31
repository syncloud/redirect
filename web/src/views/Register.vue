<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" data-testid="register-form" @submit="register">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="register-heading">Create your account</h2>
      <p class="sc-auth-sub">First month free, then £5 a month. Cancel anytime.</p>

      <div id="errors_placeholder">
        <div class="sc-alert" :class="{ invisible: !isError }">{{ error }}</div>
      </div>

      <div id="group-email" class="sc-field">
        <label for="register_email">Email</label>
        <input id="register_email" data-testid="register-email" name="email" type="text" placeholder="user@mail.com" required="" v-model="email">
        <span id="help-email" class="sc-help">{{ emailError }}</span>
      </div>

      <div id="group-password" class="sc-field">
        <label for="register_password">Password</label>
        <input id="register_password" data-testid="register-password" name="password" type="password" required="" v-model="password">
        <span id="help-password" class="sc-help">{{ passwordError }}</span>
      </div>

      <button id="btnregister" data-testid="register-submit" name="btnregister" class="sc-btn">Create account</button>

      <p class="sc-auth-note" data-testid="register-next-steps">
        Next you will install Syncloud on your own hardware &mdash; a Raspberry Pi, an old PC
        or a ready-made device &mdash; and activate it with this account.
        <a href="https://syncloud.org/setup" data-testid="register-setup-link">See how it works</a>
      </p>

      <div class="sc-auth-links">
        <router-link to="/login" data-testid="register-login">Already have an account?</router-link>
        <router-link to="/privacy" data-testid="register-privacy">Privacy policy</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import axios from 'axios'
import { storedGclid } from '../attribution'

function showError (component, error) {
  if ('parameters_messages' in error) {
    for (let i = 0; i < error.parameters_messages.length; i++) {
      const pm = error.parameters_messages[i]
      switch (pm.parameter) {
        case 'email':
          component.emailError = pm.messages.join('\n')
          component.isEmailError = true
          break
        case 'password':
          component.passwordError = pm.messages.join('\n')
          component.isPasswordError = true
          break
      }
    }
  } else {
    component.isError = true
    component.error = error.message
  }
}

export default {
  name: 'Register',
  data () {
    return {
      email: '',
      isEmailError: false,
      emailError: '',
      password: '',
      isPasswordError: false,
      passwordError: '',
      error: '',
      isError: false
    }
  },
  methods: {
    register: function (event) {
      this.isError = false
      const request = { email: this.email, password: this.password }
      const gclid = this.$route.query.gclid || storedGclid()
      if (gclid) {
        request.gclid = gclid
      }
      axios.post('/api/user/create', request)
        .then(_ => {
          this.$router.push('/check-email')
        })
        .catch(err => {
          if ('data' in err.response) {
            showError(this, err.response.data)
          } else {
            this.$router.push('/error')
          }
        })
      event.preventDefault()
    }
  }
}
</script>
<style>
@import '../style/form-center.css';
.invisible {
  display: none;
}
</style>
