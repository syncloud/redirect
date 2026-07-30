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

            <div v-if="userLoaded && subscriptionId !== undefined">
              Your subscription includes:
              <ul>
                <li>Automatic IP DNS updates</li>
                <li>Automatic mail DNS records</li>
                <li>Email support for your device</li>
              </ul>
            </div>

            <div v-show="userLoaded && subscriptionId === undefined">
              <div class="pay-section-label">Billing</div>
              <el-radio-group v-if="userLoaded" v-model="period" size="large">
                <el-radio-button label="month" data-testid="billing-month">Monthly</el-radio-button>
                <el-radio-button label="year" data-testid="billing-year">Annual</el-radio-button>
              </el-radio-group>

              <div class="pay-section-label">Plan</div>
              <div class="plan-grid">
                <div
                  class="plan-card"
                  :class="{ selected: tier === 'pro' }"
                  data-testid="plan-pro"
                  @click="tier = 'pro'"
                >
                  <div class="plan-head">
                    <span class="plan-name">Pro</span>
                    <span class="plan-price">{{ proPrice }}</span>
                  </div>
                  <ul class="plan-features">
                    <li>10 GB relay traffic / month</li>
                    <li>Personal domain (example.com)</li>
                    <li>Automatic IP &amp; mail DNS</li>
                    <li>Email support</li>
                  </ul>
                </div>
                <div
                  v-if="maxEnabled"
                  class="plan-card"
                  :class="{ selected: tier === 'max' }"
                  data-testid="plan-max"
                  @click="tier = 'max'"
                >
                  <div class="plan-head">
                    <span class="plan-name">Max</span>
                    <span class="plan-price">{{ maxPrice }}</span>
                  </div>
                  <ul class="plan-features">
                    <li>100 GB relay traffic / month</li>
                    <li>Personal domain (example.com)</li>
                    <li>Automatic IP &amp; mail DNS</li>
                    <li>Email support</li>
                  </ul>
                </div>
              </div>

              <div class="pay-section-label">Pay with</div>
              <div class="pay-methods">
                <el-button
                  type="primary"
                  size="large"
                  class="pay-button"
                  id="stripe_subscribe_btn"
                  data-testid="stripe-subscribe"
                  v-show="tier === 'pro' || stripeMaxEnabled"
                  :icon="CreditCard"
                  @click="stripeCheckout"
                >Card</el-button>

                <div id="paypal-buttons" class="pay-paypal" v-show="tier === 'pro' || paypalMaxEnabled"></div>

                <div class="pay-crypto" v-show="tier === 'pro'">
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
                <span>Usage this month</span>
              </div>
            </template>

            <div class="usage-row" data-testid="usage-traffic">
              <div class="usage-head">
                <span class="usage-name">Device access</span>
                <span class="usage-value" data-testid="usage-traffic-text">{{ trafficText }}</span>
              </div>
              <el-progress
                v-if="trafficEnabled"
                :percentage="Math.min(trafficPercent, 100)"
                :status="usageStatus(trafficPercent)"
                :stroke-width="14"
                data-testid="usage-traffic-bar"
              />
              <div v-else class="usage-off" data-testid="usage-traffic-off">
                Your device is reached directly, with no traffic limit. Turn the relay on if your
                connection has no public address.
              </div>
            </div>

            <div class="usage-row" data-testid="usage-email">
              <div class="usage-head">
                <span class="usage-name">Email sending</span>
                <span class="usage-value" data-testid="usage-email-text">{{ emailText }}</span>
              </div>
              <el-progress
                v-if="emailEnabled"
                :percentage="Math.min(emailPercent, 100)"
                :status="usageStatus(emailPercent)"
                :stroke-width="14"
                data-testid="usage-email-bar"
              />
              <div v-else class="usage-off" data-testid="usage-email-off">
                Your device sends email directly, with no limit. Turn the relay on if providers
                reject mail from your address.
              </div>
            </div>

            <div class="usage-note" v-if="usageNearLimit" data-testid="usage-upgrade">
              You are approaching a limit. Upgrade for more.
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
      trafficUsed: 0,
      trafficLimit: 0,
      trafficEnabled: false,
      emailUsed: 0,
      emailLimit: 0,
      emailEnabled: false,
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
      tier: 'pro',
      stripeMaxEnabled: false,
      paypalMaxEnabled: false,
      planMaxMonthlyId: String,
      planMaxAnnualId: String,
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
    this.loadUsage()
  },
  computed: {
    maxEnabled: function () {
      return this.stripeMaxEnabled || this.paypalMaxEnabled
    },
    proPrice: function () {
      return this.period === 'year' ? '£60 / year' : '£5 / month'
    },
    maxPrice: function () {
      return this.period === 'year' ? '£180 / year' : '£15 / month'
    },
    trafficPercent: function () {
      return this.trafficLimit > 0 ? Math.round((this.trafficUsed / this.trafficLimit) * 100) : 0
    },
    emailPercent: function () {
      return this.emailLimit > 0 ? Math.round((this.emailUsed / this.emailLimit) * 100) : 0
    },
    trafficText: function () {
      if (!this.trafficEnabled) {
        return 'Relay off'
      }
      const gb = v => (v / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
      return gb(this.trafficUsed) + ' of ' + gb(this.trafficLimit)
    },
    emailText: function () {
      if (!this.emailEnabled) {
        return 'Relay off'
      }
      return this.emailUsed + ' of ' + this.emailLimit + ' emails'
    },
    usageNearLimit: function () {
      return (this.trafficEnabled && this.trafficPercent >= 80) ||
        (this.emailEnabled && this.emailPercent >= 80)
    }
  },
  methods: {
    usageStatus: function (percent) {
      if (percent >= 100) {
        return 'exception'
      }
      return percent >= 80 ? 'warning' : 'success'
    },
    loadUsage: function () {
      axios.get('/api/relay/usage')
        .then(response => {
          this.trafficUsed = response.data.data.used_bytes
          this.trafficLimit = response.data.data.limit_bytes
          this.trafficEnabled = response.data.data.enabled === true
        })
        .catch(_ => { this.trafficEnabled = false })
      axios.get('/api/mail/usage')
        .then(response => {
          this.emailUsed = response.data.data.used_messages
          this.emailLimit = response.data.data.limit_messages
          this.emailEnabled = response.data.data.enabled === true
        })
        .catch(_ => { this.emailEnabled = false })
    },
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
          this.planMaxAnnualId = response.data.data.plan_max_annual_id
          this.planMaxMonthlyId = response.data.data.plan_max_monthly_id
          this.clientId = response.data.data.client_id
          this.stripeMaxEnabled = response.data.data.stripe_max_enabled
          this.paypalMaxEnabled = response.data.data.paypal_max_enabled
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
      const annual = this.period === 'year'
      let plan
      if (this.tier === 'max') {
        plan = annual ? 'max_annual' : 'max_monthly'
      } else {
        plan = annual ? 'annual' : 'monthly'
      }
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
                let planId
                if (this.tier === 'max') {
                  planId = this.period === 'year' ? this.planMaxAnnualId : this.planMaxMonthlyId
                } else {
                  planId = this.period === 'year' ? this.planAnnualId : this.planMonthlyId
                }
                return actions.subscription.create({ plan_id: planId })
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
.plan-grid {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}
.plan-card {
  flex: 1 1 200px;
  border: 2px solid var(--el-border-color);
  border-radius: 12px;
  padding: 16px;
  cursor: pointer;
  transition: border-color 0.2s ease, box-shadow 0.2s ease;
}
.plan-card:hover {
  border-color: var(--el-color-primary-light-5);
}
.plan-card.selected {
  border-color: var(--el-color-primary);
  box-shadow: 0 0 0 1px var(--el-color-primary);
}
.plan-head {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  margin-bottom: 10px;
}
.plan-name {
  font-size: 18px;
  font-weight: 700;
}
.plan-price {
  font-size: 15px;
  font-weight: 600;
  color: var(--el-color-primary);
}
.plan-features {
  margin: 0;
  padding-left: 18px;
  color: var(--el-text-color-regular);
  font-size: 14px;
  line-height: 1.7;
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

.usage-row {
  margin-bottom: 18px;
}
.usage-row:last-of-type {
  margin-bottom: 0;
}
.usage-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 6px;
}
.usage-name {
  font-weight: 600;
}
.usage-value {
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}
.usage-off {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  line-height: 1.5;
}
.usage-note {
  margin-top: 14px;
  color: var(--el-color-warning);
}
</style>
