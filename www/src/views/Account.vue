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
                  Free Services
                </div>
              </div>
              <div class="panel-body">
                  DNS (*.syncloud.it)
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
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  Premium Services
                </div>
              </div>
              <div class="panel-body">
                <div>Improve Syncloud device experience with your domain name (ex. example.com).
                </div>
                <ul style="padding-top: 10px">
                  <li>Automatic IP DNS updates</li>
                  <li>Automatic mail DNS records</li>
                  <li>There will be more in future ...</li>
                </ul>
                <div id="request_premium" v-if="isPremiumInActive">
                  <div>
                    What you need to have:
                  </div>
                  <ul>
                    <li>You own a domain (like example.com)</li>
                    <li>You can change Nameservers for your domain</li>
                    <li>You are ready to allow Syncloud to manage DNS records for that domain name by setting
                      Syncloud Name Servers
                    </li>
                  </ul>
                  <div style="margin: auto">
                    <h4 style="text-align: center" >Subscribe for £5/month</h4>
                    <div style="margin: auto; max-width: 200px" id="paypal-buttons"
                         v-if="isPremiumInActive"
                         :on-approve="onApprove" :create-order="createOrder"></div>
                  </div>
                </div>

                <div id="premium_active" v-if="isPremiumActive">
                  <span class="label label-success" style="font-size: 16px;">Active</span>
                  You can now activate your device in a premium mode<br>
                  <ul>
                    <li>Update system on the device from Settings - Updates</li>
                    <li>ReActivate from Settings - Activation and select a Premium mode</li>
                    <li>
                      Copy Name Servers for your <router-link to="/">domain</router-link> (Under this domain Name Servers list)
                    </li>
                    <li>
                      Update Name Servers on your domain registrar page (GoDaddy for example)
                    </li>
                  </ul>

                </div>
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
import { loadScript } from '@paypal/paypal-js'

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
    loadScript({
      'client-id': 'AbuA_mUz0LOkG36bf3fYl59N8xXSQU8M6Zufpq-z07fNLG4XEM01SXGGJRAEXZpN2ejsl45S4VrA9qLN',
      vault: true,
      intent: 'subscription'
    })
      .then((paypal) => {
        paypal
          .Buttons({
            createSubscription: (data, actions) => {
              return actions.subscription.create({
                plan_id: 'P-88T8436193034834XMDZRP4A'
              })
            },
            onApprove: (data, actions) => {
              console.log(data)
            }
          })
          .render('#paypal-buttons')
      })
      .catch((err) => {
        console.error('failed to load the PayPal JS SDK script', err)
      })
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
