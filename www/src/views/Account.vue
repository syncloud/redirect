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
                  <div class="pull-right" id="subscription_active"
                       v-if="this.userLoaded && this.subscriptionId !== undefined">
                    <span class="label label-success" style="font-size: 16px;">Active</span>
                  </div>
                  <div class="pull-right" id="subscription_inactive"
                       v-if="this.userLoaded && this.subscriptionId === undefined">
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
                <div v-show="this.userLoaded && this.subscriptionId === undefined">
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
                    <el-row  style="padding: 20px 0 20px 0">
                      <el-col :span="24" style="text-align: center">
                        <el-radio-group v-if="this.userLoaded" v-model="this.subscriptionType">
                          <el-radio-button label="paypal_month">£5/Month</el-radio-button>
                          <el-radio-button label="paypal_year" >£60/Year</el-radio-button>
                          <el-radio-button label="crypto_year" id="crypto_year">ETH 0.05/Year</el-radio-button>
                        </el-radio-group>
                      </el-col>
                    </el-row>
                    <div v-show="this.subscriptionType.startsWith('paypal')" style="margin: auto; max-width: 200px" id="paypal-buttons">

                    </div>
                    <div v-show="this.subscriptionType.startsWith('crypto')" style="margin: auto; max-width: 400px" >
                      <el-row class="crypto-row" style="border-top: 1px solid var(--el-border-color); padding-top: 5px" >
                        <el-col :span="16" style="border-bottom: 1px solid var(--el-border-color); padding-bottom: 5px">
                          Amount (Ethereum)
                        </el-col>
                        <el-col :span="8" style="text-align: right; border-bottom: 4px solid #409EFF; padding-bottom: 5px">
                          0.05 ETH
                        </el-col>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col :span="24">Please send to address:</el-col>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col :span="24" style="text-align: center">
                          <code class="wallet">{{ wallet }}</code>
                          <el-button text :icon="CopyDocument" size="small" @click="copy" v-show="!copied"></el-button>
                          <el-icon color="green" style="padding: 0 10px 0 10px; vertical-align: middle; height: 24px" :size="34" v-show="copied">
                            <CircleCheck />
                          </el-icon>
                        </el-col>
                      </el-row>
                      <el-row class="crypto-row" style="padding-top: 2px">
                        <el-col :span="24">
                          or Scan the QR code
                        </el-col>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col :span="4"/>
                        <el-col :span="16">
                          <el-image src="/assets/crypto-wallet-qr.png" ></el-image>
                        </el-col>
                        <el-col :span="4"/>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col>
                          Enter transaction ID:
                        </el-col>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col>
                          <el-input v-model="cryptoTransactionId" id="crypto_transaction_id"></el-input>
                        </el-col>
                      </el-row>
                      <el-row class="crypto-row">
                        <el-col style="text-align:right">
                          <el-button
                            @click="cryptoSubscribe"
                            type="primary"
                            :disabled="cryptoTransactionId.length<10"
                            id="crypto_subscribe_btn"
                          >
                            Subscribe
                          </el-button>
                        </el-col>
                      </el-row>
                    </div>
                </div>

                <div v-if="this.userLoaded && this.subscriptionId !== undefined">
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

                  <button type="button" class="btn btn-danger pull-right" id="cancel" @click="cancelSubscription">
                    <span class="glyphicon glyphicon-remove" aria-hidden="true" style="padding-right: 5px"></span>Cancel
                  </button>

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

  <CustomDialog :visible="deleteConfirmationVisible" @cancel="deleteConfirmationVisible = false"
          id="delete_confirmation" @confirm="accountDeleteConfirm">
    <template v-slot:title>Delete Account</template>
    <template v-slot:text>
      <div>Once you delete your account, there's no going back. All devices you have will be deactivated and domains
        will
        be released. Proceed with caution!
      </div>
      <br>
      <div>Are you sure?</div>
    </template>
  </CustomDialog>

  <CustomDialog :visible="cancelConfirmationVisible" @cancel="cancelConfirmationVisible = false"
          id="cancel_confirmation" @confirm="cancelSubscriptionConfirm">
    <template v-slot:title>Cancel subscription</template>
    <template v-slot:text>
      <div>
        You are about to cancel your subscription
      </div>
      <br>
      <div>Are you sure?</div>
    </template>
  </CustomDialog>

</template>
<script>
import axios from 'axios'
import CustomDialog from '../components/CustomDialog.vue'
import { loadScript } from '@paypal/paypal-js'
import { CircleCheck, CopyDocument } from '@element-plus/icons-vue'
import { markRaw } from 'vue'

export default {
  name: 'Account',
  components: {
    CircleCheck,
    CustomDialog
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
      planMonthlyId: String,
      planAnnualId: String,
      clientId: String,
      paypalLoaded: Boolean,
      userLoaded: Boolean,
      deleteConfirmationVisible: false,
      cancelConfirmationVisible: false,
      subscriptionType: 'paypal_month',
      cryptoTransactionId: '',
      wallet: '0x1c644443EA113Ef5aA17255a777EB909e2217566',
      copied: false,
      CopyDocument: markRaw(CopyDocument)
    }
  },
  mounted () {
    this.subscriptionId = undefined
    this.paypalLoaded = false
    this.userLoaded = false
    this.reload()
  },
  methods: {
    copy: function () {
      navigator.clipboard.writeText(this.wallet)
      this.copied = true
      setTimeout(() => { this.copied = false }, 2000)
    },
    reload: function () {
      axios.get('/api/user')
        .then(response => {
          this.notificationEnabled = response.data.data.notification_enabled
          this.subscriptionId = response.data.data.subscription_id
          this.userLoaded = true
          this.loadPlan(this.subscriptionId)
        })
        .catch(this.onError)
    },
    loadPlan: function (subscriptionId) {
      axios.get('/api/plan')
        .then(response => {
          this.planAnnualId = response.data.data.plan_annual_id
          this.planMonthlyId = response.data.data.plan_monthly_id
          this.clientId = response.data.data.client_id
          if (!subscriptionId && !this.paypalLoaded) {
            this.enablePayPal(this.clientId)
          }
        })
        .catch(this.onError)
    },
    subscribe: function () {

    },
    cryptoSubscribe: function () {
      axios.post('/api/plan/subscribe/crypto', { subscription_id: this.cryptoTransactionId })
        .then(_ => {
          this.reload()
        })
        .catch(this.onError)
    },
    enablePayPal: function (clientId) {
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
                  plan_id: this.subscriptionType === 'paypal_year' ? this.planAnnualId : this.planMonthlyId
                })
              },
              onApprove: (data, actions) => {
                axios.post('/api/plan/subscribe/paypal', { subscription_id: data.subscriptionID })
                  .then(_ => {
                    this.reload()
                  })
                  .catch(this.onError)
              }
            })
            .render('#paypal-buttons')
          this.paypalLoaded = true
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
    cancelSubscription: function () {
      this.cancelConfirmationVisible = true
    },
    cancelSubscriptionConfirm: function () {
      this.cancelConfirmationVisible = false
      axios.delete('/api/plan')
        .then(_ => {
          this.reload()
        })
        .catch(this.onError)
    },
    accountDelete: function () {
      this.deleteConfirmationVisible = true
    },
    accountDeleteConfirm: function () {
      this.deleteConfirmationVisible = false
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
.crypto-row {
  padding-bottom: 10px;
}
.wallet {
  border: 2px dashed var(--el-border-color);
  font-size: 90%;
}
@media (max-width: 1000px) {
  .wallet {
    border: 2px dashed var(--el-border-color);
    font-size: 10px;
  }
}
@media (max-width: 767px) {
  .wallet {
    border: 2px dashed var(--el-border-color);
    font-size: 90%;
  }
}
@media (max-width: 430px) {
  .wallet {
    border: 2px dashed var(--el-border-color);
    font-size: 10px;
  }
}
</style>
