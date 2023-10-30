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
                  Subscription
                  <div class="pull-right" id="premium_active" v-if="this.subscriptionId">
                    <span class="label label-success" style="font-size: 16px;">Active</span>
                  </div>
                  <div class="pull-right" id="premium_active" v-if="!this.subscriptionId">
                    You have 30 days to subscribe
                  </div>
                </div>
              </div>
              <div class="panel-body">
                <div>
                  Subscription is required after 30 days of a free trial period.<br>
                  Additionally you can use your personal domain on active subscription (like example.com)<br><br>
                  We provide the following features for your device:
                </div>
                <ul style="padding-top: 10px">
                  <li>Automatic IP DNS updates</li>
                  <li>Automatic mail DNS records</li>
                  <li>Email support for your device</li>
                </ul>
                <div id="request_premium" v-if="!this.subscriptionId">
                  <div>
                    For personal domain you need to:
                  </div>
                  <ul>
                    <li>Have you own a domain (like example.com)</li>
                    <li>Be able to change Nameservers for your domain</li>
                    <li>Allow Syncloud to manage DNS records for that domain name by setting
                      Syncloud Name Servers
                    </li>
                  </ul>
                  <div style="margin: auto">
                    <el-switch
                      style="margin: auto; max-width: 200px"
                      v-model="subscriptionAnnual"
                      active-text="Annual"
                      inactive-text="Monthly"
                    />
                    <h4 style="text-align: center" v-if="this.payPalLoaded && this.subscriptionAnnual">Subscribe for £60/year</h4>
                    <h4 style="text-align: center" v-if="this.payPalLoaded && !this.subscriptionAnnual">Subscribe for £5/month</h4>
                    <div style="margin: auto; max-width: 200px" id="paypal-buttons"></div>
                  </div>
                </div>

                <div id="premium_active" v-if="this.subscriptionId">
                  <div style="padding-top: 10px">
                  You can activate your device with a personal domain:<br>
                  </div>
                  <ol>
                    <li>Reactivate from Settings - Activation and select a Premium mode</li>
                    <li>
                      Copy Name Servers for your
                      <router-link to="/">domain</router-link>
                      (Under this domain Name Servers list)
                    </li>
                    <li>
                      Update Name Servers on your domain registrar page (GoDaddy for example)
                    </li>
                  </ol>

                </div>
              </div>
            </div>
          </div>

          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  Email notifications
                </div>
              </div>
              <div class="panel-body">
                <div class="pull-left">
                  <input v-model="notificationEnabled" type="checkbox" id="chk_email" :value="notificationEnabled">
                  <label for="chk_email" style="font-weight: normal; padding-left: 5px">Send me notifications</label>
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
                <button type="button" class="btn btn-danger pull-right" id="delete" @click="accountDelete">
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
import Confirmation from '../components/Confirmation.vue'
import { loadScript } from '@paypal/paypal-js'
import { ElSwitch } from 'element-plus'

export default {
  name: 'Account',
  components: {
    Confirmation,
    ElSwitch
  },
  props: {
    checkUserSession: Function
  },
  data () {
    return {
      notificationEnabled: Boolean,
      premiumStatusId: Number,
      subscriptionId: String,
      domainGroups: Array,
      planId: String,
      clientId: String,
      payPalLoaded: Boolean,
      subscriptionAnnual: false
    }
  },
  mounted () {
    this.subscriptionId = null
    this.payPalLoaded = false
    this.reload()
  },
  methods: {
    reload: function () {
      axios.get('/api/user')
        .then(response => {
          this.notificationEnabled = response.data.data.notification_enabled
          this.subscriptionId = response.data.data.subscription_id
          this.loadPlan(this.subscriptionId)
        })
        .catch(this.onError)
    },
    loadPlan: function (subscriptionId) {
      axios.get('/api/plan')
        .then(response => {
          this.planId = response.data.data.plan_id
          this.clientId = response.data.data.client_id
          if (!subscriptionId && !this.payPalLoaded) {
            this.enablePayPal(this.clientId, this.planId)
          }
        })
        .catch(this.onError)
    },
    enablePayPal: function (clientId, planId) {
      loadScript({
        'client-id': clientId,
        vault: true,
        intent: 'subscription'
      })
        .then((paypal) => {
          paypal
            .Buttons({
              createSubscription: (data, actions) => {
                return actions.subscription.create({
                  plan_id: planId
                })
              },
              onApprove: (data, actions) => {
                axios.post('/api/plan/subscribe', { subscription_id: data.subscriptionID })
                  .then(_ => {
                    this.reload()
                  })
                  .catch(this.onError)
              }
            })
            .render('#paypal-buttons')
          this.payPalLoaded = true
        })
        .catch((err) => {
          console.error('failed to load the PayPal JS SDK script', err)
        })
    },
    notificationSave: function () {
      const action = this.notificationEnabled ? 'enable' : 'disable'
      axios.post('/api/notification/' + action)
        .then(_ => {
          this.reload()
        })
        .catch(this.onError)
    },
    accountDelete: function () {
      this.$refs.delete_confirmation.show()
    },
    accountDeleteConfirm: function () {
      axios.delete('/api/user')
        .then(_ => {
          this.checkUserSession()
        })
        .catch(this.onError)
    },
    onError: function (err) {
      console.log(err)
      if (err.response.status === 401) {
        this.$router.push('/login')
      } else {
        this.$router.push('/error')
      }
    }
  }
}
</script>
<style>
</style>
