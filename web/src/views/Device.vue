<template>
  <div class="sc-page">
    <h1 class="sc-h1">Buy a device</h1>
    <p class="sc-lead">
      A Syncloud device, assembled and ready, with the image already written. It arrives
      linked to this account, so there is nothing else to sign up for.
    </p>

    <div v-if="ordered" class="sc-card" data-testid="device-ordered">
      <h2 class="sc-h2">Thank you</h2>
      <p>
        Your order is paid and we have it. We will email you at {{ email }} when it ships.
      </p>
      <p class="sc-muted" data-testid="device-reference">Reference {{ reference }}</p>
    </div>

    <template v-else>
      <div class="sc-card sc-form" data-testid="device-choice">
        <h2 class="sc-h2">{{ device.name }}</h2>

        <div class="sc-field">
          <label for="device-option">Storage</label>
          <select
            id="device-option"
            v-model="option"
            data-testid="device-option"
          >
            <option v-for="each in device.options" :key="each.code" :value="each.code">
              {{ each.name }} — {{ money(device.price + each.extra) }}
            </option>
          </select>
        </div>

        <div class="sc-summary">
          <div class="sc-summary-row">
            <span>Device</span>
            <span data-testid="device-price">{{ money(device.price + extra) }}</span>
          </div>
          <div class="sc-summary-row">
            <span>Shipping</span>
            <span data-testid="device-shipping">{{ money(shipping) }}</span>
          </div>
          <div class="sc-summary-row sc-summary-total">
            <span>Total</span>
            <span data-testid="device-total">{{ money(total) }}</span>
          </div>
        </div>
      </div>

      <div class="sc-card sc-form" data-testid="device-address">
        <h2 class="sc-h2">Ship it to</h2>

        <div class="sc-field">
          <label for="device-name">Full name</label>
          <input id="device-name" v-model="name" type="text" data-testid="device-name">
        </div>
        <div class="sc-field">
          <label for="device-address-line">Address</label>
          <input id="device-address-line" v-model="address" type="text" data-testid="device-address-line">
        </div>
        <div class="sc-field">
          <label for="device-city">City</label>
          <input id="device-city" v-model="city" type="text" data-testid="device-city">
        </div>
        <div class="sc-field">
          <label for="device-postcode">Postcode</label>
          <input id="device-postcode" v-model="postcode" type="text" data-testid="device-postcode">
        </div>
        <div class="sc-field">
          <label for="device-country">Country</label>
          <input id="device-country" v-model="country" type="text" data-testid="device-country">
        </div>

        <p v-if="incomplete" class="sc-help" data-testid="device-incomplete">
          We need the whole address before you can pay.
        </p>
      </div>

      <div class="sc-card" data-testid="device-pay">
        <h2 class="sc-h2">Pay {{ money(total) }}</h2>

        <div class="pay-section-label">Pay with</div>
        <div class="pay-methods">
          <el-button
            type="primary"
            size="large"
            class="pay-button"
            :icon="CreditCard"
            :disabled="incomplete"
            data-testid="device-pay-stripe"
            @click="payWithStripe"
          >
            Card
          </el-button>

          <div
            id="device-paypal"
            v-show="!incomplete"
            class="pay-paypal"
            data-testid="device-pay-paypal"
          />
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
    checkUserSession: Function
  },
  data () {
    return {
      device: { name: '', price: 0, options: [] },
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
      ordered: false,
      error: '',
      paypalLoaded: false,
      CreditCard: markRaw(CreditCard)
    }
  },
  computed: {
    extra () {
      const chosen = this.device.options.find(each => each.code === this.option)
      return chosen ? chosen.extra : 0
    },
    total () {
      return this.device.price + this.extra + this.shipping
    },
    incomplete () {
      return [this.name, this.address, this.city, this.postcode, this.country]
        .some(field => field.trim() === '')
    }
  },
  mounted () {
    this.load()
    const returned = this.$route.query.reference
    if (returned) {
      this.complete(returned)
        .then(_ => this.$router.replace({ query: {} }))
        .catch(this.onError)
    }
  },
  methods: {
    money (pence) {
      return `£${(pence / 100).toFixed(2)}`
    },
    onError (error) {
      this.error = error.response && error.response.data
        ? error.response.data.message
        : 'something went wrong'
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
          this.loadPayPal()
        })
        .catch(this.onError)

      axios.get('/api/user')
        .then(response => { this.email = response.data.data.email })
        .catch(this.onError)
    },
    order (provider) {
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
    complete (reference) {
      return axios.post('/api/device/order/complete', { reference: reference })
        .then(_ => {
          this.reference = reference
          this.ordered = true
        })
    },
    payWithStripe () {
      this.error = ''
      this.order('stripe')
        .then(response => {
          window.location.href = response.data.data.url
        })
        .catch(this.onError)
    },
    loadPayPal () {
      if (this.paypalLoaded || !this.paypalClientId) {
        return
      }
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
          paypal.Buttons({
            style: { layout: 'vertical', label: 'paypal', tagline: false, height: 44 },
            createOrder: () => {
              this.error = ''
              return this.order('paypal')
                .then(response => {
                  this.reference = response.data.data.reference
                  return response.data.data.provider_reference
                })
                .catch(error => {
                  this.onError(error)
                  throw error
                })
            },
            onApprove: () => this.complete(this.reference).catch(this.onError),
            onError: error => this.onError(error)
          }).render('#device-paypal')
          this.paypalLoaded = true
        })
        .catch(error => console.error('failed to load the PayPal JS SDK script', error))
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

.pay-section-label {
  font-weight: 600;
  color: var(--sc-muted);
  margin: 20px 0 8px;
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
</style>
