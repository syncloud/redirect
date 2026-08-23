<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <form class="sc-auth-card" data-testid="login-form" @submit="login">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="login-heading">Log in</h2>

      <div id="errors_placeholder">
        <div class="sc-alert" id="error" :class="{ invisible: !isError }">{{ error }}</div>
      </div>

      <div id="group-email" class="sc-field">
        <label for="email">Email</label>
        <input id="email" data-testid="login-email" type="text" placeholder="user@mail.com" required="" v-model="email">
        <span id="help-email" class="sc-help">{{ emailError }}</span>
      </div>

      <div id="group-password" class="sc-field">
        <label for="password">Password</label>
        <input id="password" data-testid="login-password" type="password" required="" v-model="password">
        <span id="help-password" class="sc-help">{{ passwordError }}</span>
      </div>

      <button id="submit" data-testid="login-submit" class="sc-btn">Log in</button>

      <div class="sc-auth-links">
        <router-link to="/forgot" id="forgot" data-testid="login-forgot">Forgot your password?</router-link>
        <router-link to="/register" id="register" data-testid="login-register">Create an account</router-link>
      </div>
    </form>
  </div>
</template>

<script>
import axios from 'axios'

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
  name: 'Login',
  props: {
    checkUserSession: Function
  },
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
  mounted () {
    // this.checkUserSession()
  },
  methods: {
    login: function (event) {
      this.isError = false
      axios.post('/api/user/login', { email: this.email, password: this.password })
        .then(_ => {
          this.checkUserSession()
          this.$router.push('/')
        })
        .catch(err => {
          showError(this, err.response.data)
        })
      event.preventDefault()
    }
  }
}
</script>
<style>
@import '../style/form-center.css';

.visible {
  display: block;
}
.invisible {
  display: none;
}
</style>
