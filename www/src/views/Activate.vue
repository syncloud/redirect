<template>

  <div class="container">
    <h2>Activation</h2>
    <br/>
    <div class="row">
      <div id="activated" class="col-10 col-md-10 col-sm-10 col-lg-10" style="font-size: 18px">
        {{ message }}
      </div>
    </div>
  </div>

</template>

<script>
import axios from 'axios'
import querystring from 'querystring'

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
      console.log('activating with token: ' + token)
      axios.post('api/user/activate', querystring.stringify({ token: token }))
        .then(response => {
          console.log('activated')
          if ('message' in response.data) {
            this.message = response.data.message
          }
        })
        .catch(err => {
          console.log('activate error')
          if (err.response.status === 400) {
            if ('message' in err.response.data) {
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
