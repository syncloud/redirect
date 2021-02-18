<template>

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
                  Services
                </div>
              </div>
              <div class="panel-body">
                <div class="pull-left">
                  <h4>Free services:</h4>
                  <ul>
                    <li>DNS (*.syncloud.it)</li>
                  </ul>
                  <h4>Premium services:</h4>
                  <div>Improve Syncloud device experience with cloud support.
                  </div>
                  <ul>
                    <li>Managed DNS (custom domain name IP updates)</li>
                  </ul>
                </div>

                <button id="request_premium" type="button" class="btn btn-default pull-right" @click="requestPremium"
                        v-if="isPremiumInActive">
                  <span class="glyphicon glyphicon-ok" style="padding-right: 5px"></span>Request Premium
                </button>

                <div id="premium_pending" v-if="isPremiumPending">
                  <span class="label label-warning pull-right" style="font-size: 16px;">Pending Premium</span>
                </div>

                <div id="premium_active" v-if="isPremiumActive">
                  <span class="label label-success pull-right" style="font-size: 16px;">Active Premium</span>
                </div>
              </div>
            </div>
          </div>
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
                  <input v-model="subscribed" type="checkbox" id="chk_email" :value="subscribed">
                  <label for="chk_email" style="font-weight: normal; padding-left: 5px">Send me Syncloud notifications,
                    including releases announcements</label>
                </div>
                <button type="button" class="btn btn-default pull-right" @click="notificationSave" id="save">
                  <span class="glyphicon glyphicon-ok" style="padding-right: 5px"></span>Save
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
                <button type="button" class="btn btn-default pull-right" id="delete" @click="accountDelete">
                  <span class="glyphicon glyphicon-remove" aria-hidden="true" style="padding-right: 5px"></span>Delete
                </button>
              </div>
            </div>
          </div>

        </div>
      </div>
    </div>
  </div>

  <Confirmation ref="premium_confirmation" id="premium_confirmation" @confirm="requestPremiumConfirm" @cancel="reload">
    <template v-slot:title>Premium account request</template>
    <template v-slot:text>
      By confirming the request you agree to the following:<br>
      <ul>
        <li>You own a domain (like example.com)</li>
        <li>You can change Nameservers for your domain</li>
        <li>You are ready to allow Syncloud to manage DNS records for that domain name by setting
          Syncloud Nameservers
        </li>
      </ul>
      <div>Service is currently provided for 1 year free with no guarantees and maybe canceled anytime depending on test
        results.
      </div>
      <br>
      <div>Are you sure?</div>
    </template>
  </Confirmation>

  <Confirmation ref="delete_confirmation" id="delete_confirmation" @confirm="accountDeleteConfirm" @cancel="reload">
    <template v-slot:title>Delete Account</template>
    <template v-slot:text>
      <div>Once you delete your account, there's no going back. All devices you have will be deactivated and domains
        will
        be released. Proceed with caution!
      </div>
      <br>
      <div>Are you sure?</div>
    </template>
  </Confirmation>

</template>

<script>
import axios from 'axios'
import Confirmation from '@/components/Confirmation'

const PREMIUM_STATUS_INACTIVE = 1
const PREMIUM_STATUS_PENDING = 2
const PREMIUM_STATUS_ACTIVE = 3

export default {
  name: 'Account',
  components: {
    Confirmation
  },
  props: {
    onLogin: Function,
    onLogout: Function
  },
  data () {
    return {
      subscribed: Boolean,
      premiumStatusId: Number,
      domainGroups: Array
    }
  },
  computed: {
    isPremiumActive: function () {
      return this.premiumStatusId === PREMIUM_STATUS_ACTIVE
    },
    isPremiumPending: function () {
      return this.premiumStatusId === PREMIUM_STATUS_PENDING
    },
    isPremiumInActive: function () {
      return this.premiumStatusId === PREMIUM_STATUS_INACTIVE
    }
  },
  mounted () {
    this.reload()
  },
  methods: {
    reload: function () {
      axios.get('api/user')
        .then(response => {
          this.subscribed = !response.data.data.unsubscribed
          this.premiumStatusId = response.data.data.premium_status_id
        })
        .catch(err => {
          if (err.response.status === 401) {
            this.$router.push('/login')
          } else {
            this.$router.push('/error')
          }
        })
    },
    notificationSave: function () {
      const action = this.subscribed ? 'subscribe' : 'unsubscribe'
      axios.post('api/notification/' + action)
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
      this.$refs.delete_confirmation.show()
    },
    accountDeleteConfirm: function () {
      axios.delete('api/user')
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
    },
    requestPremium: function () {
      this.$refs.premium_confirmation.show()
    },
    requestPremiumConfirm: function () {
      axios.post('api/premium/request')
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
    }
  }
}
</script>
<style>
</style>
