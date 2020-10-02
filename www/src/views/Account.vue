<template>

  <div class="modal fade" id="modalDeleteAccount" tabindex="-1" role="dialog" aria-labelledby="modalDeleteAccount">
    <div class="modal-dialog" role="document">
      <div class="modal-content">
        <div class="modal-header">
          <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
          <h4 class="modal-title">Delete Account</h4>
        </div>
        <div class="modal-body">
          Once you delete your account, there's no going back. All devices you have will be deactivated and domains will be released. Proceed with caution!
        </div>
        <div class="modal-footer">
          <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
          <button type="button" class="btn btn-danger" data-dismiss="modal" @click="accountDelete">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <div class="container">
    <div id="has_domains">
      <h2>Account</h2>
      <br/>

      <div>

        <div class="row">
          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  Notifications
                </div>
              </div>
              <div class="panel-body">
                <h4>Email</h4>
                <div class="pull-left">
                <input v-model="subscribed" type="checkbox" id="chk_email">
                Email me Syncloud notifications, including releases announcements.
                </div>
                <button type="button" class="btn btn-default pull-right" @click="accountSave">
                  <span class="glyphicon glyphicon-ok"></span>  Save
                </button>
              </div>
            </div>
          </div>

          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-danger">
              <div class="panel-heading">
                <div class="panel-title">
                  Danger Zone
                </div>
              </div>
              <div class="panel-body clearfix">
                <h4>Delete this account</h4>
                <div class="pull-left">
                  Delete your account all domains and personal data.
                </div>
                <button type="button" class="btn btn-default pull-right" data-toggle="modal" data-target="#modalDeleteAccount">
                  <span class="glyphicon glyphicon-remove" aria-hidden="true"></span> Delete
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import querystring from 'querystring'

export default {
  name: 'Account',
  props: {
    onLogin: Function,
    onLogout: Function
  },
  data () {
    return {
      subscribed: Boolean,
      domainGroups: Array
    }
  },
  mounted () {
    this.reload()
  },
  methods: {
    reload: function () {
      axios.get('api/user/get')
        .then(response => {
          this.subscribed = !response.data.unsubscribed
        })
        .catch(err => {
          if (err.response.status === 401) {
            this.$router.push('/login')
          } else {
            this.$router.push('/error')
          }
        })
    },
    accountSave: function () {
      axios.post('api/set_subscribed', querystring.stringify({ subscribed: this.subscribed }))
        .then(_ => {
          this.reload()
        })
        .catch(err => {
          if (err.response.status === 401) {
            this.$router.push('/login')
          } else {
            this.$router.push('/error')
          }
        })
    },
    accountDelete: function () {
      axios.post('api/user_delete')
        .then(_ => {
          this.onLogout()
        })
        .catch(err => {
          if (err.response.status === 401) {
            this.$router.push('/login')
          } else {
            this.$router.push('/error')
          }
        })
    }
  }
}
</script>
<style>
</style>
