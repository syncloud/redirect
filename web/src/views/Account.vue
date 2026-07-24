<template>

  <div class="container">
    <div id="has_domains">
      <h2 data-testid="account-title">Account</h2>
      <br/>
      <el-row :gutter="20">
        <el-col :xs="24" :md="12">
          <el-card class="account-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>Subscription</span>
                <el-tag
                  id="subscription_active"
                  type="success"
                  size="large"
                  v-if="userLoaded && subscriptionId !== undefined"
                >Active</el-tag>
                <span
                  id="subscription_inactive"
                  class="trial-note"
                  v-if="userLoaded && subscriptionId === undefined"
                >You have 30 days to subscribe</span>
              </div>
            </template>

            <div v-if="userLoaded && subscriptionId === undefined">
              Subscription is required after 30 days of a free trial period.<br>
              Additionally you can use your personal domain on active subscription (like example.com)<br><br>
            </div>
            <div>
              We provide the following features for your device:
            </div>
            <ul>
              <li>Automatic IP DNS updates</li>
              <li>Automatic mail DNS records</li>
              <li>Email support for your device</li>
            </ul>

            <div v-show="userLoaded && subscriptionId === undefined">
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
              <div class="pay-section-label">Billing</div>
              <el-radio-group v-if="userLoaded" v-model="period" size="large">
                <el-radio-button label="month">Monthly · £5</el-radio-button>
                <el-radio-button label="year">Annual · £60</el-radio-button>
              </el-radio-group>

              <div class="pay-section-label">Pay with</div>
              <div class="pay-methods">
                <el-button
                  type="primary"
                  size="large"
                  class="pay-button"
                  id="stripe_subscribe_btn"
                  data-testid="stripe-subscribe"
                  :icon="CreditCard"
                  @click="stripeCheckout"
                >Card</el-button>

                <div id="paypal-buttons" class="pay-paypal"></div>

                <div class="pay-crypto">
                  <el-button text id="crypto_year" data-testid="crypto-toggle" @click="cryptoOpen = !cryptoOpen">
                    Or pay with crypto (0.05 ETH / year)
                  </el-button>
                  <div v-show="cryptoOpen" class="crypto-details">
                    <el-row class="crypto-row" style="border-top: 1px solid var(--el-border-color); padding-top: 5px">
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
                        <el-image src="/assets/crypto-wallet-qr.png"></el-image>
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
              </div>
            </div>

            <div v-if="userLoaded && subscriptionId !== undefined">
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

              <div style="text-align: right">
                <el-button type="danger" id="cancel" :icon="Close" @click="cancelSubscription">Cancel</el-button>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card class="account-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>Email notifications</span>
              </div>
            </template>
            <div class="card-actions">
              <el-switch
                id="chk_email"
                data-testid="notification-toggle"
                v-model="notificationEnabled"
                active-text="Send me notifications"
              />
              <el-button type="primary" id="save" :icon="Check" @click="notificationSave">Save</el-button>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card class="account-card danger-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>Danger Zone</span>
              </div>
            </template>
            <h4>Delete this account</h4>
            <div class="card-actions">
              <span>Delete your account all domains and personal data.</span>
              <el-button type="danger" id="delete" :icon="Delete" @click="accountDelete">Delete</el-button>
            </div>
          </el-card>
        </el-col>

      </el-row>
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
import { CircleCheck, CopyDocument, Check, Close, Delete, CreditCard } from '@element-plus/icons-vue'
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
      period: 'month',
      cryptoOpen: false,
      cryptoTransactionId: '',
      wallet: '0x1c644443EA113Ef5aA17255a777EB909e2217566',
      copied: false,
      CopyDocument: markRaw(CopyDocument),
      Check: markRaw(Check),
      Close: markRaw(Close),
      Delete: markRaw(Delete),
      CreditCard: markRaw(CreditCard)
    }
  },
  mounted () {
    this.subscriptionId = undefined
    this.paypalLoaded = false
    this.userLoaded = false
    const sessionId = this.$route && this.$route.query ? this.$route.query.stripe_session_id : undefined
    if (sessionId) {
      this.confirmStripe(sessionId)
    } else {
      this.reload()
    }
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
    stripeCheckout: function () {
      const plan = this.period === 'year' ? 'annual' : 'monthly'
      axios.post('/api/plan/subscribe/stripe/checkout', { plan: plan })
        .then(response => {
          window.location.href = response.data.data.url
        })
        .catch(this.onError)
    },
    confirmStripe: function (sessionId) {
      axios.post('/api/plan/subscribe/stripe', { subscription_id: sessionId })
        .then(_ => {
          this.$router.replace({ query: {} })
          this.reload()
        })
        .catch(this.onError)
    },
    enablePayPal: function (clientId) {
      loadScript({
        clientId: clientId,
        vault: true,
        intent: 'subscription',
        disableFunding: 'card'
      })
        .then((paypal) => {
          paypal
            .Buttons({
              createSubscription: (data, actions) => {
                return actions.subscription.create({
                  plan_id: this.period === 'year' ? this.planAnnualId : this.planMonthlyId
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
.account-card {
  margin-bottom: 20px;
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.trial-note {
  color: var(--el-text-color-secondary);
  font-size: 14px;
}
.card-actions {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}
.danger-card {
  --el-card-border-color: var(--el-color-danger);
}
.pay-section-label {
  font-weight: 600;
  color: var(--el-text-color-secondary);
  margin: 20px 0 8px 0;
}
.pay-methods {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-width: 320px;
}
.pay-button {
  width: 100%;
}
.pay-paypal {
  min-height: 1px;
}
.pay-crypto {
  margin-top: 4px;
}
.crypto-details {
  max-width: 400px;
  padding-top: 8px;
}
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
