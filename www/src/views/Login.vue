<template>
  <div class="container">
  <form class="form-horizontal" @submit="login">
    <h2>Log in</h2>
    <br/>

    <fieldset>

      <div id="errors_placeholder">
        <div class="alert alert-danger" :class="{ visible:  isError, invisible:  !isError}">{{ error }}</div>
      </div>

      <div id="group-email" class="form-group" :class="{ 'has-error':  isEmailError}">
        <div class="col-3 col-md-3 col-sm-3 col-lg-3">
          <label class="control-label" for="email">Email</label>
        </div>
        <div class="col-9 col-md-9 col-sm-9 col-lg-9">
          <input id="email" type="text" placeholder="user@mail.com" class="form-control input-md" required="" v-model="email">
          <span id="help-email" class="help-block">{{ emailError }}</span>
        </div>
      </div>

      <div id="group-password" class="form-group" :class="{ 'has-error':  isPasswordError}">
        <div class="col-3 col-md-3 col-sm-3 col-lg-3">
          <label class="control-label" for="password">Password</label>
        </div>
        <div class="col-9 col-md-9 col-sm-9 col-lg-9">
          <input id="password" type="password" placeholder="" class="form-control input-md" required="" v-model="password">
          <span id="help-password" class="help-block">{{ passwordError }}</span>
        </div>
      </div>

      <div class="form-group">
        <div class="button-block col-12 col-md-12 col-sm-12 col-lg-12" style="padding-right:15px; padding-left:15px;">
          <router-link to="/forgot" id="forgot" class="pull-left" style="padding-top: 10px;">Forgot your password?</router-link>
          <button id="submit" class="btn btn-primary pull-right" >Log in</button>
        </div>
      </div>

    </fieldset>
  </form>
  </div>
</template>

<script>
import axios from 'axios'
import querystring from 'querystring'

function showError (component, error) {
  if ('parameters_messages' in error) {
    for (var i = 0; i < error.parameters_messages.length; i++) {
      var pm = error.parameters_messages[i]
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
    onLogin: Function,
    onLogout: Function
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
  methods: {
    login: function (event) {
      console.log('login click')
      this.isError = false
      axios.post('/api/login', querystring.stringify({ email: this.email, password: this.password }))
        .then(response => {
          this.onLogin()
          this.$router.push('/')
        })
        .catch(err => {
          if ('content-type' in err.response.headers &&
            err.response.headers['content-type'].indexOf('application/json') !== -1 &&
            'data' in err.response) {
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

.visible {
  display: block;
}
.invisible {
  display: none;
}
</style>
