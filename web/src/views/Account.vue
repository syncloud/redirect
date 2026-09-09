<template>

  <div class="sc-page">
    <div id="has_domains">
      <h1 class="sc-h1" data-testid="account-title">{{ $t('account.title') }}</h1>
      <p class="sc-lead">
        {{ $t('account.signedInAs') }} <strong data-testid="account-email">{{ email }}</strong>
      </p>
      <el-row :gutter="20">
        <el-col :xs="24" :md="12">
          <el-card class="account-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>{{ $t('account.subscription') }}</span>
                <el-tag
                  id="subscription_active"
                  type="success"
                  size="large"
                  v-if="userLoaded && subscriptionId !== undefined"
                >{{ $t('account.active') }}</el-tag>
                <span
                  id="subscription_inactive"
                  class="trial-note"
                  v-if="userLoaded && subscriptionId === undefined"
                >{{ $t('account.trialNote') }}</span>
              </div>
            </template>

            <div v-if="userLoaded && subscriptionId !== undefined">
              {{ $t('account.subscriptionIncludes') }}
              <ul>
                <li>{{ $t('account.domain') }}</li>
                <li>{{ $t('account.autoDns') }}</li>
                <li>{{ $t('account.relay10gb') }}</li>
                <li>{{ $t('account.mailRelay') }}</li>
                <li>{{ $t('account.emailSupport') }}</li>
              </ul>

              <div class="billing-row">
                <span>
                  {{ $t('account.billing') }}
                  <strong data-testid="billing-current">{{ currentPriceText }}</strong>
                </span>
                <el-button
                  v-if="subscriptionPeriod === 'month'"
                  type="primary"
                  plain
                  size="small"
                  id="switch_annual"
                  data-testid="switch-annual"
                  :loading="busy === 'switch'"
                  @click="switchToAnnual"
                >{{ $t('account.switchToAnnual') }}</el-button>
              </div>
              <div v-if="switchError" class="switch-error" data-testid="switch-error">{{ switchError }}</div>
            </div>

            <div v-show="userLoaded && subscriptionId === undefined">
              <div class="pay-section-label">{{ $t('account.billingLabel') }}</div>
              <el-radio-group v-if="userLoaded" v-model="period" size="large">
                <el-radio-button label="month" data-testid="billing-month">{{ $t('account.monthly') }}</el-radio-button>
                <el-radio-button label="year" data-testid="billing-year">{{ $t('account.annual') }}</el-radio-button>
              </el-radio-group>

              <div class="pay-section-label">{{ $t('account.planLabel') }}</div>
              <div class="plan-grid">
                <div
                  class="plan-card"
                  :class="{ selected: tier === 'pro' }"
                  data-testid="plan-pro"
                  @click="tier = 'pro'"
                >
                  <div class="plan-head">
                    <span class="plan-name">{{ $t('account.proName') }}</span>
                    <span class="plan-price">{{ proPrice }}</span>
                  </div>
                  <ul class="plan-features">
                    <li>{{ $t('account.relay10gb') }}</li>
                    <li>{{ $t('account.mailRelay') }}</li>
                    <li>{{ $t('account.domain') }}</li>
                    <li>{{ $t('account.autoDns') }}</li>
                    <li>{{ $t('account.emailSupport') }}</li>
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
                    <span class="plan-name">{{ $t('account.maxName') }}</span>
                    <span class="plan-price">{{ maxPrice }}</span>
                  </div>
                  <ul class="plan-features">
                    <li>{{ $t('account.relay100gb') }}</li>
                    <li>{{ $t('account.mailRelay') }}</li>
                    <li>{{ $t('account.domain') }}</li>
                    <li>{{ $t('account.autoDns') }}</li>
                    <li>{{ $t('account.emailSupport') }}</li>
                  </ul>
                </div>
              </div>

              <div class="pay-section-label">{{ $t('account.payWith') }}</div>
              <div class="pay-methods">
                <el-button
                  type="primary"
                  size="large"
                  class="pay-button"
                  id="stripe_subscribe_btn"
                  data-testid="stripe-subscribe"
                  v-show="tier === 'pro' || stripeMaxEnabled"
                  :icon="CreditCard"
                  :disabled="paying"
                  :loading="busy === 'stripe'"
                  @click="stripeCheckout"
                >{{ $t('account.card') }}</el-button>

                <div class="pay-paypal-wrap">
                  <div id="paypal-buttons" class="pay-paypal" v-show="tier === 'pro' || paypalMaxEnabled"></div>

                  <div v-if="busy === 'paypal'" class="pay-overlay" data-testid="account-pay-busy">
                    <span class="pay-spinner" :aria-label="$t('account.openingPaypal')"/>
                  </div>
                </div>

                <div class="pay-crypto" v-show="tier === 'pro'">
                  <el-button text id="crypto_year" data-testid="crypto-toggle" @click="cryptoOpen = !cryptoOpen">
                    {{ $t('account.payCrypto') }}
                  </el-button>
                  <div v-show="cryptoOpen" class="crypto-details">
                    <el-row class="crypto-row" style="border-top: 1px solid var(--el-border-color); padding-top: 5px">
                      <el-col :span="16" style="border-bottom: 1px solid var(--el-border-color); padding-bottom: 5px">
                        {{ $t('account.cryptoAmount') }}
                      </el-col>
                      <el-col :span="8" style="text-align: right; border-bottom: 4px solid #409EFF; padding-bottom: 5px">
                        {{ $t('account.cryptoEthAmount') }}
                      </el-col>
                    </el-row>
                    <el-row class="crypto-row">
                      <el-col :span="24">{{ $t('account.cryptoSendTo') }}</el-col>
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
                        {{ $t('account.cryptoScanQr') }}
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
                        {{ $t('account.cryptoEnterTxId') }}
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
                          {{ $t('account.subscribe') }}
                        </el-button>
                      </el-col>
                    </el-row>
                  </div>
                </div>
              </div>
            </div>

            <div v-if="userLoaded && subscriptionId !== undefined">
              <div style="padding-top: 10px">
                {{ $t('account.usingYourDomain') }}
              </div>
              <ol>
                <li>
                  {{ $t('account.copyNameServers') }}
                  <router-link to="/">{{ $t('account.domainLink') }}</router-link>
                </li>
                <li>
                  {{ $t('account.setNameServers') }}
                </li>
              </ol>

              <div style="text-align: right">
                <el-button type="danger" id="cancel" :icon="Close" @click="cancelSubscription">{{ $t('account.cancel') }}</el-button>
              </div>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card class="account-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>{{ $t('account.emailNotifications') }}</span>
              </div>
            </template>
            <div class="card-actions">
              <el-switch
                id="chk_email"
                data-testid="notification-toggle"
                v-model="notificationEnabled"
                :active-text="$t('account.sendNotifications')"
              />
              <el-button type="primary" id="save" :icon="Check" @click="notificationSave">{{ $t('account.save') }}</el-button>
            </div>
          </el-card>
        </el-col>

        <el-col :xs="24" :md="12">
          <el-card class="account-card danger-card" shadow="never">
            <template #header>
              <div class="card-header">
                <span>{{ $t('account.dangerZone') }}</span>
              </div>
            </template>
            <h4>{{ $t('account.deleteThisAccount') }}</h4>
            <div class="card-actions">
              <span>{{ $t('account.deleteDescription') }}</span>
              <el-button type="danger" id="delete" data-testid="account-delete" :icon="Delete" @click="accountDelete">{{ $t('account.delete') }}</el-button>
            </div>
          </el-card>
        </el-col>

      </el-row>
    </div>
  </div>

  <CustomDialog :visible="deleteConfirmationVisible" @cancel="deleteConfirmationVisible = false"
          id="delete_confirmation" @confirm="accountDeleteConfirm">
    <template v-slot:title>{{ $t('account.deleteAccountTitle') }}</template>
    <template v-slot:text>
      <div>{{ $t('account.deleteConfirmBody') }}
      </div>
      <br>
      <div>{{ $t('account.areYouSure') }}</div>
    </template>
  </CustomDialog>

  <CustomDialog :visible="switchConfirmationVisible" @cancel="switchConfirmationVisible = false"
          id="switch_confirmation" @confirm="switchToAnnualConfirm">
    <template v-slot:title>{{ $t('account.switchTitle') }}</template>
    <template v-slot:text>
      <div>
        {{ $t('account.switchBody', { from: monthlyPriceText, to: annualPriceText }) }}
      </div>
      <br>
      <div>{{ $t('account.continueQuestion') }}</div>
    </template>
  </CustomDialog>

  <CustomDialog :visible="cancelConfirmationVisible" @cancel="cancelConfirmationVisible = false"
          id="cancel_confirmation" @confirm="cancelSubscriptionConfirm">
    <template v-slot:title>{{ $t('account.cancelTitle') }}</template>
    <template v-slot:text>
      <div>
        {{ $t('account.cancelBody') }}
      </div>
      <br>
      <div>{{ $t('account.areYouSure') }}</div>
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
      email: '',
      subscriptionId: String,
      subscriptionPeriod: '',
      subscriptionTier: '',
      tierPrices: {
        pro: { month: '£5 / month', year: '£60 / year' },
        max: { month: '£15 / month', year: '£180 / year' }
      },
      domainGroups: Array,
      planMonthlyId: String,
      planAnnualId: String,
      clientId: String,
      sdkUrl: '',
      paypalLoaded: Boolean,
      userLoaded: Boolean,
      deleteConfirmationVisible: false,
      cancelConfirmationVisible: false,
      switchConfirmationVisible: false,
      switchError: '',
      period: 'month',
      tier: 'pro',
      stripeMaxEnabled: false,
      paypalMaxEnabled: false,
      planMaxMonthlyId: String,
      planMaxAnnualId: String,
      busy: '',
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
      if (this.$route && this.$route.query && this.$route.query.paypal_switch) {
        this.$router.replace({ query: {} })
      }
      this.reload()
    }
  },
  computed: {
    paying () {
      return this.busy !== ''
    },
    maxEnabled: function () {
      return this.stripeMaxEnabled || this.paypalMaxEnabled
    },
    proPrice: function () {
      return this.period === 'year' ? '£60 / year' : '£5 / month'
    },
    maxPrice: function () {
      return this.period === 'year' ? '£180 / year' : '£15 / month'
    },
    currentTierKey: function () {
      return this.subscriptionTier === 'max' ? 'max' : 'pro'
    },
    currentPriceText: function () {
      return this.tierPrices[this.currentTierKey][this.subscriptionPeriod === 'year' ? 'year' : 'month']
    },
    monthlyPriceText: function () {
      return this.tierPrices[this.currentTierKey].month
    },
    annualPriceText: function () {
      return this.tierPrices[this.currentTierKey].year
    }
  },
  methods: {
    copy: function () {
      navigator.clipboard.writeText(this.wallet)
      this.copied = true
      setTimeout(() => { this.copied = false }, 2000)
    },
    reload: function () {
      this.busy = ''
      axios.get('/api/user')
        .then(response => {
          this.notificationEnabled = response.data.data.notification_enabled
          this.email = response.data.data.email
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
          this.sdkUrl = response.data.data.sdk_url
          this.stripeMaxEnabled = response.data.data.stripe_max_enabled
          this.paypalMaxEnabled = response.data.data.paypal_max_enabled
          this.subscriptionPeriod = response.data.data.current_period
          this.subscriptionTier = response.data.data.current_tier
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
      this.busy = 'stripe'
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
      const options = {
        clientId: clientId,
        vault: true,
        intent: 'subscription',
        disableFunding: 'card'
      }
      if (this.sdkUrl) {
        options.sdkBaseUrl = this.sdkUrl
      }
      loadScript(options)
        .then((paypal) => {
          paypal
            .Buttons({
              style: { layout: 'vertical', label: 'paypal', tagline: false, height: 44, borderRadius: 0 },
              onClick: () => {
                this.busy = 'paypal'
              },
              onCancel: () => {
                this.busy = ''
              },
              onError: () => {
                this.busy = ''
              },
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
    switchToAnnual: function () {
      this.switchConfirmationVisible = true
    },
    switchToAnnualConfirm: function () {
      this.switchConfirmationVisible = false
      this.switchError = ''
      this.busy = 'switch'
      axios.post('/api/plan/switch')
        .then(response => {
          const url = response.data.data.url
          if (url) {
            window.location.href = url
          } else {
            this.reload()
          }
        })
        .catch(err => {
          this.busy = ''
          if (err.response && err.response.status === 401) {
            this.$router.push('/login')
            return
          }
          this.switchError = (err.response && err.response.data && err.response.data.message) ||
            this.$t('account.switchFailed')
        })
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
      this.busy = ''
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
.billing-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px solid var(--el-border-color);
}
.billing-row #switch_annual {
  margin-left: auto;
}
.switch-error {
  margin-top: 8px;
  color: var(--el-color-danger);
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
.pay-paypal-wrap {
  position: relative;
}
.pay-overlay {
  position: absolute;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: var(--el-border-radius-base, 4px);
  background: color-mix(in srgb, var(--el-bg-color) 78%, transparent);
}
.pay-spinner {
  width: 22px;
  height: 22px;
  border: 2px solid rgba(0, 0, 0, 0.18);
  border-top-color: var(--el-color-primary);
  border-radius: 50%;
  animation: pay-spin 0.7s linear infinite;
}
@keyframes pay-spin {
  to {
    transform: rotate(360deg);
  }
}
@media (prefers-reduced-motion: reduce) {
  .pay-spinner {
    animation-duration: 2.4s;
  }
}
.pay-paypal {
  height: 44px;
  border-radius: var(--el-border-radius-base, 4px);
  overflow: hidden;
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
