<template>
  <div class="sc-page">
    <h1 class="sc-h1">{{ $t('shop.title') }}</h1>
    <p class="sc-lead">
      {{ $t('shop.lead') }}
    </p>

    <div v-if="ordered" class="sc-card" data-testid="device-ordered">
      <h2 class="sc-h2">{{ $t('shop.thankYou') }}</h2>
      <p>
        {{ $t('shop.orderPaid', { email: email }) }}
      </p>
      <p class="sc-muted" data-testid="device-reference">{{ $t('shop.orderNumber', { n: number }) }}</p>
    </div>

    <template v-else>
      <div v-show="step === 'choose'" class="sc-card" data-testid="device-choice">
        <div class="product-head">
          <img
            class="product-photo"
            src="/assets/syncloud-h4.jpg"
            :alt="device.name"
            width="180"
            height="180"
            data-testid="device-photo"
          >
          <div>
            <h2 class="sc-h2 product-name">{{ device.name }}</h2>
            <p class="product-from">{{ $t('shop.priceFrom', { price: money(device.price + shipping) }) }}</p>
            <ul class="product-points">
              <li>{{ $t('shop.point1') }}</li>
              <li>{{ $t('shop.point2') }}</li>
              <li>{{ $t('shop.point3') }}</li>
            </ul>
          </div>
        </div>

        <details class="spec" data-testid="device-spec">
          <summary>{{ $t('shop.moreDetails') }}</summary>

          <table class="spec-table" data-testid="device-spec-table">
            <tbody>
              <tr v-for="spec in device.specs" :key="spec.name">
                <th scope="row">{{ spec.name }}</th>
                <td>{{ spec.value }}</td>
              </tr>
            </tbody>
          </table>

          <h4 class="spec-heading">{{ $t('shop.inTheBox') }}</h4>
          <ul>
            <li>{{ $t('shop.boxItem1') }}</li>
            <li>{{ $t('shop.boxItem2') }}</li>
            <li>{{ $t('shop.boxItem3') }}</li>
            <li>{{ $t('shop.boxItem4') }}</li>
          </ul>
          <p>
            {{ $t('shop.secondBay') }}
          </p>
          <p>
            {{ $t('shop.suits') }}
          </p>
        </details>

        <h3 class="option-title">{{ $t('shop.storage') }}</h3>
        <div class="options" data-testid="device-options">
          <button
            v-for="each in device.options"
            :key="each.code"
            type="button"
            class="option"
            :class="{ 'option-on': option === each.code }"
            :data-testid="`device-option-${each.code}`"
            @click="option = each.code"
          >
            <span class="option-name">{{ each.name }}</span>
            <span class="option-extra">{{ each.extra ? $t('shop.optionExtra', { extra: money(each.extra) }) : $t('shop.optionIncluded') }}</span>
          </button>
        </div>

        <div class="sc-summary">
          <div class="sc-summary-row">
            <span>{{ device.name }}, {{ optionName }}</span>
            <span data-testid="device-price">{{ money(device.price + extra) }}</span>
          </div>
          <div class="sc-summary-row">
            <span>{{ $t('shop.delivery') }}</span>
            <span data-testid="device-shipping">{{ money(shipping) }}</span>
          </div>
          <div class="sc-summary-row sc-summary-total">
            <span>{{ $t('shop.total') }}</span>
            <span data-testid="device-total">{{ money(total) }}</span>
          </div>
        </div>

        <el-button
          v-if="loggedIn"
          type="primary"
          size="large"
          class="continue"
          data-testid="shop-continue"
          @click="step = 'address'"
        >
          {{ $t('shop.continue') }}
        </el-button>
      </div>

      <div v-if="loggedIn === false && step === 'choose'" class="sc-card" data-testid="shop-signin">
        <h2 class="sc-h2">{{ $t('shop.signInToOrder') }}</h2>
        <p>
          {{ $t('shop.signInBlurb') }}
        </p>
        <router-link
          class="sc-btn"
          to="/login?next=/shop"
          data-testid="shop-signin-link"
        >
          {{ $t('shop.signInLink') }}
        </router-link>
      </div>

      <div v-if="loggedIn && step !== 'choose'" class="sc-card chosen" data-testid="shop-chosen">
        <div>
          <span class="chosen-name">{{ device.name }}, {{ optionName }}</span>
          <span class="chosen-total">{{ money(total) }}</span>
        </div>
        <button type="button" class="chosen-change" data-testid="shop-change" @click="step = 'choose'">
          {{ $t('shop.change') }}
        </button>
      </div>

      <div v-if="loggedIn && step === 'pay'" class="sc-card chosen" data-testid="shop-address-chosen">
        <div>
          <span class="chosen-name">{{ name }}</span>
          <span class="chosen-total chosen-address">{{ addressLine }}</span>
        </div>
        <button type="button" class="chosen-change" data-testid="shop-address-change" @click="step = 'address'">
          {{ $t('shop.change') }}
        </button>
      </div>

      <div v-if="loggedIn && step === 'address'" class="sc-card sc-form" data-testid="device-address">
        <h2 class="sc-h2">{{ $t('shop.shipItTo') }}</h2>

        <div class="field">
          <input
            id="device-name"
            v-model="name"
            type="text"
            placeholder=" "
            autocomplete="name"
            data-testid="device-name"
          >
          <label for="device-name">{{ $t('shop.fullName') }}</label>
        </div>
        <div class="field">
          <input
            id="device-address-line"
            v-model="address"
            type="text"
            placeholder=" "
            autocomplete="street-address"
            data-testid="device-address-line"
          >
          <label for="device-address-line">{{ $t('shop.address') }}</label>
        </div>
        <div class="field-row">
          <div class="field">
            <input
              id="device-city"
              v-model="city"
              type="text"
              placeholder=" "
              autocomplete="address-level2"
              data-testid="device-city"
            >
            <label for="device-city">{{ $t('shop.city') }}</label>
          </div>
          <div class="field">
            <input
              id="device-postcode"
              v-model="postcode"
              type="text"
              placeholder=" "
              autocomplete="postal-code"
              data-testid="device-postcode"
            >
            <label for="device-postcode">{{ $t('shop.postcode') }}</label>
          </div>
        </div>
        <div class="field">
          <input
            id="device-country"
            v-model="country"
            type="text"
            placeholder=" "
            autocomplete="country-name"
            data-testid="device-country"
          >
          <label for="device-country">{{ $t('shop.country') }}</label>
        </div>

        <p v-if="incomplete" class="sc-help" data-testid="device-incomplete">
          {{ $t('shop.incomplete') }}
        </p>

        <el-button
          type="primary"
          size="large"
          class="continue"
          :disabled="incomplete"
          data-testid="shop-address-continue"
          @click="step = 'pay'"
        >
          {{ $t('shop.continue') }}
        </el-button>
      </div>

      <div v-if="loggedIn && step === 'pay'" class="sc-card" data-testid="device-pay">
        <h2 class="sc-h2">{{ $t('shop.pay', { amount: money(total) }) }}</h2>

        <div class="pay-methods">
          <el-button
            type="primary"
            size="large"
            class="pay-button"
            :icon="CreditCard"
            :disabled="incomplete || paying"
            :loading="busy === 'stripe'"
            data-testid="device-pay-stripe"
            @click="payWithStripe"
          >
            {{ $t('shop.card') }}
          </el-button>

          <div class="pay-paypal-wrap">
            <div
              id="device-paypal"
              v-show="!incomplete"
              class="pay-paypal"
              data-testid="device-pay-paypal"
            />

            <div
              v-if="paypalLoading"
              class="pay-overlay"
              data-testid="device-paypal-loading"
            >
              <span class="pay-spinner" :aria-label="$t('shop.loadingPaypal')"/>
            </div>

            <div
              v-if="busy === 'paypal'"
              class="pay-overlay"
              data-testid="device-pay-busy"
            >
              <span class="pay-spinner" :aria-label="$t('shop.openingPaypal')"/>
            </div>
          </div>
        </div>
      </div>
    </template>

    <p v-if="error" class="sc-warn" data-testid="device-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import { loadScript } from '@paypal/paypal-js'
import { CreditCard } from '@element-plus/icons-vue'
import { markRaw } from 'vue'

export default {
  name: 'DeviceView',
  props: {
    checkUserSession: Function,
    loggedIn: Boolean
  },
  data () {
    return {
      device: { name: '', price: 0, specs: [], options: [] },
      shipping: 0,
      currency: 'GBP',
      paypalClientId: '',
      paypalSdkUrl: '',
      option: '',
      name: '',
      address: '',
      city: '',
      postcode: '',
      country: '',
      email: '',
      reference: '',
      number: 0,
      resuming: 0,
      ordered: false,
      error: '',
      busy: '',
      step: 'choose',
      paypalLoaded: false,
      paypalLoading: false,
      CreditCard: markRaw(CreditCard)
    }
  },
  computed: {
    addressLine () {
      return [this.address, this.city, this.postcode, this.country]
        .filter(part => part !== '')
        .join(', ')
    },
    optionName () {
      const chosen = this.device.options.find(each => each.code === this.option)
      return chosen ? chosen.name : ''
    },
    extra () {
      const chosen = this.device.options.find(each => each.code === this.option)
      return chosen ? chosen.extra : 0
    },
    total () {
      return this.device.price + this.extra + this.shipping
    },
    paying () {
      return this.busy !== ''
    },
    incomplete () {
      return [this.name, this.address, this.city, this.postcode, this.country]
        .some(field => field.trim() === '')
    }
  },
  watch: {
    step (value) {
      if (value === 'pay') {
        this.$nextTick(() => this.loadPayPal())
      }
    }
  },
  mounted () {
    this.load()
    const returned = this.$route.query.reference
    if (returned) {
      this.complete(returned)
        .then(_ => this.$router.replace({ query: {} }))
        .catch(this.onError)
      return
    }
    const resume = this.$route.query.order
    if (resume) {
      this.resume(Number(resume))
    }
  },
  methods: {
    money (pence) {
      return `£${(pence / 100).toFixed(2)}`
    },
    onError (error) {
      this.busy = ''
      this.error = error.response && error.response.data
        ? error.response.data.message
        : this.$t('shop.somethingWentWrong')
    },
    load () {
      axios.get('/api/device/catalog')
        .then(response => {
          const catalog = response.data.data
          this.device = catalog.devices[0]
          this.shipping = catalog.shipping
          this.currency = catalog.currency
          this.paypalClientId = catalog.paypal_client_id
          this.paypalSdkUrl = catalog.paypal_sdk_url
          this.option = this.device.options[0].code
        })
        .catch(this.onError)

      if (this.loggedIn) {
        axios.get('/api/user')
          .then(response => { this.email = response.data.data.email })
          .catch(() => {})
      }
    },
    order (provider) {
      if (this.resuming) {
        return axios.post('/api/device/order/retry', {
          number: this.resuming,
          provider: provider
        })
      }
      return axios.post('/api/device/order', {
        device: this.device.code,
        option: this.option,
        provider: provider,
        name: this.name,
        address: this.address,
        city: this.city,
        postcode: this.postcode,
        country: this.country
      })
    },
    resume (number) {
      axios.get('/api/device/order', { params: { number: number } })
        .then(response => {
          const order = response.data.data.order
          this.resuming = number
          this.name = order.name
          this.address = order.address
          this.city = order.city
          this.postcode = order.postcode
          this.country = order.country
          this.step = 'address'
        })
        .catch(this.onError)
    },
    complete (reference) {
      return axios.post('/api/device/order/complete', { reference: reference })
        .then(response => {
          this.number = response.data.data.number
          this.ordered = true
        })
    },
    payWithStripe () {
      this.error = ''
      this.busy = 'stripe'
      this.order('stripe')
        .then(response => {
          window.location.href = response.data.data.url
        })
        .catch(this.onError)
    },
    loadPayPal () {
      if (this.paypalLoaded || this.paypalLoading || !this.paypalClientId) {
        return
      }
      this.paypalLoading = true
      const options = {
        clientId: this.paypalClientId,
        currency: this.currency,
        disableFunding: 'card,paylater'
      }
      if (this.paypalSdkUrl) {
        options.sdkBaseUrl = this.paypalSdkUrl
      }
      loadScript(options)
        .then(paypal => {
          return paypal.Buttons({
            style: { layout: 'vertical', label: 'paypal', tagline: false, height: 44, borderRadius: 0 },
            onClick: () => {
              this.error = ''
              this.busy = 'paypal'
            },
            onCancel: () => {
              this.busy = ''
            },
            createOrder: () => {
              this.error = ''
              this.busy = 'paypal'
              return this.order('paypal')
                .then(response => {
                  this.reference = response.data.data.reference
                  this.number = response.data.data.number
                  return response.data.data.provider_reference
                })
                .catch(error => {
                  this.onError(error)
                  throw error
                })
            },
            onApprove: () => this.complete(this.reference)
              .then(() => { this.busy = '' })
              .catch(this.onError),
            onError: error => this.onError(error)
          }).render('#device-paypal')
        })
        .then(() => {
          this.paypalLoaded = true
          this.paypalLoading = false
        })
        .catch(error => {
          this.paypalLoading = false
          console.error('failed to load the PayPal JS SDK script', error)
        })
    }
  }
}
</script>

<style scoped>
.sc-summary {
  margin-top: 20px;
  border-top: 1px solid var(--sc-border);
  padding-top: 14px;
}

.sc-summary-row {
  display: flex;
  justify-content: space-between;
  padding: 6px 0;
  color: var(--sc-ink-2);
}

.sc-summary-total {
  border-top: 1px solid var(--sc-border);
  margin-top: 8px;
  padding-top: 12px;
  font-weight: 700;
  color: var(--sc-ink);
  font-size: 1.05rem;
}

.sc-card + .sc-card {
  margin-top: 20px;
}

.product-head {
  display: flex;
  align-items: flex-start;
  gap: 22px;
  flex-wrap: wrap;
}

.field {
  position: relative;
  margin-bottom: 12px;
}

.field input {
  padding-top: 20px !important;
  padding-bottom: 4px !important;
}

.field label {
  position: absolute;
  left: 15px;
  top: 14px;
  color: var(--sc-muted);
  font-size: 1rem;
  pointer-events: none;
  transition: top 0.12s ease, font-size 0.12s ease, color 0.12s ease;
}

.field input:focus + label,
.field input:not(:placeholder-shown) + label {
  top: 6px;
  font-size: 0.72rem;
  font-weight: 600;
  color: var(--sc-ink-2);
}

.field input:focus + label {
  color: var(--sc-primary);
}

.field-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
}

.continue {
  width: 100%;
  margin-top: 20px;
}

.chosen {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.chosen-name {
  font-weight: 600;
}

.chosen-total {
  margin-left: 10px;
  color: var(--sc-muted);
}

.chosen-address {
  display: block;
  font-weight: 400;
  color: var(--sc-muted);
  font-size: 0.9rem;
}

.chosen-change {
  border: 0;
  background: none;
  color: var(--sc-primary);
  font-family: var(--sc-font);
  font-size: 0.95rem;
  cursor: pointer;
  padding: 0;
}

.product-photo {
  width: 180px;
  height: 180px;
  object-fit: contain;
  border-radius: var(--sc-control-radius);
  background: var(--sc-field-bg);
  flex: none;
}

.product-name {
  margin: 0;
}

.product-from {
  margin: 2px 0 0;
  color: var(--sc-muted);
  font-size: 0.95rem;
}

.product-points {
  margin: 12px 0 0;
  padding-left: 18px;
  color: var(--sc-ink-2);
  line-height: 1.7;
}

.spec-table {
  width: 100%;
  margin-top: 12px;
  border-collapse: collapse;
  font-size: 0.92rem;
}

.spec-table th,
.spec-table td {
  border-bottom: 1px solid var(--sc-border);
  padding: 8px 0;
  text-align: left;
  vertical-align: top;
}

.spec-table th {
  width: 45%;
  font-weight: 500;
  color: var(--sc-muted);
}

.spec-heading {
  margin: 18px 0 0;
  font-size: 0.95rem;
  color: var(--sc-ink);
}

.spec {
  margin-top: 20px;
  border-top: 1px solid var(--sc-border);
  padding-top: 14px;
  color: var(--sc-ink-2);
}

.spec summary {
  cursor: pointer;
  font-weight: 600;
  color: var(--sc-ink);
}

.spec ul {
  margin: 12px 0 0;
  padding-left: 18px;
  line-height: 1.7;
}

.spec p {
  margin: 12px 0 0;
}

.option-title {
  margin: 24px 0 10px;
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--sc-ink-2);
}

.options {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 10px;
}

.option {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 4px;
  padding: 14px;
  border: 1px solid var(--sc-border);
  border-radius: var(--sc-control-radius);
  background: var(--sc-surface);
  color: var(--sc-ink);
  font-family: var(--sc-font);
  cursor: pointer;
  text-align: left;
  transition: border-color 0.15s ease, box-shadow 0.15s ease;
}

.option:hover {
  border-color: var(--sc-primary);
}

.option-on {
  border-color: var(--sc-primary);
  box-shadow: 0 0 0 2px var(--sc-primary-soft);
}

.option-name {
  font-weight: 600;
}

.option-extra {
  color: var(--sc-muted);
  font-size: 0.85rem;
}

@media (max-width: 560px) {
  .product-head {
    gap: 14px;
  }

  .product-photo {
    width: 96px;
    height: 96px;
  }

  .product-points {
    display: none;
  }

  .options {
    grid-template-columns: repeat(2, 1fr);
  }
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
  height: 44px;
  border-radius: var(--el-border-radius-base, 4px);
  overflow: hidden;
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
  background: color-mix(in srgb, var(--sc-surface) 78%, transparent);
}

.pay-spinner {
  width: 22px;
  height: 22px;
  border: 2px solid rgba(0, 0, 0, 0.18);
  border-top-color: var(--sc-accent, #0070ba);
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
</style>
