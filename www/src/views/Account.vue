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
                <div id="request_premium" v-if="!this.subscriptionId">
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
                    <h4 style="text-align: center">Subscribe for £5/month</h4>
                    <div style="margin: auto; max-width: 200px" id="paypal-buttons"></div>
                  </div>
                </div>

                <div id="premium_active" v-if="this.subscriptionId">
                  <span class="label label-success" style="font-size: 16px;padding-top: 8px">Active</span>
                  <div style="padding-top: 10px">
                  You can now activate your device in a premium mode:<br>
                  </div>
                  <ol>
                    <li>Update system on the device from Settings - Updates</li>
                    <li>ReActivate from Settings - Activation and select a Premium mode</li>
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
import Confirmation from '@/components/Confirmation'
import { loadScript } from '@paypal/paypal-js'

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
      notificationEnabled: Boolean,
      premiumStatusId: Number,
      subscriptionId: String,
      domainGroups: Array,
      planId: String,
      clientId: String
    }
  },
  mounted () {
    this.reload()
  },
  methods: {
    reload: function () {
      axios.get('api/user')
        .then(response => {
          this.notificationEnabled = response.data.data.notification_enabled
          this.subscriptionId = response.data.data.subscription_id
          this.loadPlan(this.subscriptionId)
        })
        .catch(this.onError)
    },
    loadPlan: function (subscriptionId) {
      axios.get('api/plan')
        .then(response => {
          this.planId = response.data.data.plan_id
          this.clientId = response.data.data.client_id
          if (!subscriptionId) {
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
                axios.post('api/plan/subscribe', { subscription_id: data.subscriptionID })
                  .then(_ => {
                    this.reload()
                  })
                  .catch(this.onError)
              }
            })
            .render('#paypal-buttons')
        })
        .catch((err) => {
          console.error('failed to load the PayPal JS SDK script', err)
        })
    },
    notificationSave: function () {
      const action = this.notificationEnabled ? 'enable' : 'disable'
      axios.post('api/notification/' + action)
        .then(_ => {
          this.reload()
        })
        .catch(this.onError)
    },
    accountDelete: function () {
      this.$refs.delete_confirmation.show()
    },
    accountDeleteConfirm: function () {
      axios.delete('api/user')
        .then(_ => {
          this.onLogout()
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
