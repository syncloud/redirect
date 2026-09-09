<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" data-testid="register-form" @submit="register">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="register-heading">{{ $t('register.heading') }}</h2>
      <p class="sc-auth-sub">{{ $t('register.subtitle') }}</p>

      <div id="errors_placeholder">
        <div class="sc-alert" :class="{ invisible: !isError }">{{ error }}</div>
      </div>

      <div id="group-email" class="sc-field sc-field-float">
        <input id="register_email" data-testid="register-email" name="email" type="text" placeholder=" " required="" v-model="email">
        <label for="register_email">{{ $t('register.email') }}</label>
        <span id="help-email" class="sc-help">{{ emailError }}</span>
      </div>

      <div id="group-password" class="sc-field sc-field-float">
        <input id="register_password" data-testid="register-password" name="password" type="password" placeholder=" " required="" v-model="password">
        <label for="register_password">{{ $t('register.password') }}</label>
        <span id="help-password" class="sc-help">{{ passwordError }}</span>
      </div>

      <button id="btnregister" data-testid="register-submit" name="btnregister" class="sc-btn">{{ $t('register.submit') }}</button>

      <p class="sc-auth-note" data-testid="register-next-steps">
        {{ $t('register.nextSteps') }}
        <a href="https://syncloud.org/setup" data-testid="register-setup-link">{{ $t('register.setupLink') }}</a>
      </p>

      <div class="sc-auth-links">
        <router-link to="/login" data-testid="register-login">{{ $t('register.login') }}</router-link>
        <router-link to="/privacy" data-testid="register-privacy">{{ $t('register.privacy') }}</router-link>
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
