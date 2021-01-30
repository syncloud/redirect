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
      axios.post('api/user/activate', { token: token })
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
