<template>
  <div class="container">

  <form class="form-horizontal" id="form-forgot" @submit="reset">
    <h2>Forgot your password?</h2>
    <br/>
    <fieldset>

      <p>
        Enter your email address to reset your password. We will send you a letter with a link.
      </p>

      <div class="form-group">
        <div class="col-12 col-md-12 col-sm-12 col-lg-12">
          <label class="control-label" for="email">Email</label>
        </div>
        <div class="col-12 col-md-12 col-sm-12 col-lg-12">
          <input id="email" name="email" type="text" placeholder="user@mail.com" class="form-control input-md" required="" v-model="email">
        </div>
      </div>

      <div class="form-group">
        <div class="button-block col-12 col-md-12 col-sm-12 col-lg-12" style="padding-right:15px; padding-left:15px;">
          <button id="send" class="btn btn-primary pull-right">Submit</button>
        </div>
      </div>

    </fieldset>
  </form>
  </div>
</template>

<script>
import axios from 'axios'
import querystring from 'querystring'

export default {
  name: 'PasswordForgot',
  data () {
    return {
      email: ''
    }
  },
  methods: {
    reset: function (event) {
      axios.post('api/user/reset_password', querystring.stringify({ email: this.email }))
        .then(_ => {
          this.$router.push('/login')
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
