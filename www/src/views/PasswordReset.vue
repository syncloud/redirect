<template>

  <form class="form-horizontal">
    <h2>Reset password</h2>
    <br/>
    <fieldset>

      <div id="errors_placeholder" v-if="error !== ''" >
        <div class="alert alert-danger">
          {{ error }}
        </div>
      </div>

      <div class="form-group" v-if="error === ''">
        <div class="col-4 col-md-4 col-sm-4 col-lg-4">
          <label class="control-label" for="password">New password</label>
        </div>
        <div class="col-8 col-md-8 col-sm-8 col-lg-8">
          <input id="password" type="password" placeholder="" class="form-control input-md" required="" v-model="password">
        </div>
      </div>

      <div class="form-group" v-if="error === ''">
        <div class="button-block col-12 col-md-12 col-sm-12 col-lg-12" style="padding-right:15px; padding-left:15px;">
          <button id="reset" class="btn btn-primary pull-right" @click="reset">Reset</button>
        </div>
      </div>

    </fieldset>
  </form>

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
        this.error = 'No token found'
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
