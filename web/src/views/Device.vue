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
      <div class="sc-card" data-testid="device-choice">
        <h2 class="sc-h2">{{ device.name }}</h2>
        <label class="sc-label" for="device-option">Storage</label>
        <select
          id="device-option"
          v-model="option"
          class="sc-input"
          data-testid="device-option"
        >
          <option v-for="each in device.options" :key="each.code" :value="each.code">
            {{ each.name }} — {{ money(device.price + each.extra) }}
          </option>
        </select>

        <dl class="sc-total">
          <dt>Device</dt><dd data-testid="device-price">{{ money(device.price + extra) }}</dd>
          <dt>Shipping</dt><dd data-testid="device-shipping">{{ money(shipping) }}</dd>
          <dt><strong>Total</strong></dt>
          <dd><strong data-testid="device-total">{{ money(total) }}</strong></dd>
        </dl>
      </div>

      <div class="sc-card" data-testid="device-address">
        <h2 class="sc-h2">Ship it to</h2>
        <input v-model="name" class="sc-input" placeholder="Full name" data-testid="device-name">
        <input v-model="address" class="sc-input" placeholder="Address" data-testid="device-address-line">
        <input v-model="city" class="sc-input" placeholder="City" data-testid="device-city">
        <input v-model="postcode" class="sc-input" placeholder="Postcode" data-testid="device-postcode">
        <input v-model="country" class="sc-input" placeholder="Country" data-testid="device-country">
        <p v-if="incomplete" class="sc-muted" data-testid="device-incomplete">
          We need the whole address before you can pay.
        </p>
      </div>

      <div class="sc-card" data-testid="device-pay">
        <h2 class="sc-h2">Pay</h2>
        <button
          class="sc-btn sc-btn-primary"
          :disabled="incomplete"
          data-testid="device-pay-stripe"
          @click="payWithStripe"
        >
          Pay {{ money(total) }} by card
        </button>
        <div id="device-paypal" v-show="!incomplete" data-testid="device-pay-paypal"></div>
      </div>
    </template>

    <p v-if="error" class="sc-warn" data-testid="device-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import { loadScript } from '@paypal/paypal-js'

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
      paypalLoaded: false
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
      loadScript({ clientId: this.paypalClientId, currency: this.currency })
        .then(paypal => {
          paypal.Buttons({
            createOrder: () => this.order('paypal').then(response => {
              this.reference = response.data.data.reference
              return response.data.data.provider_reference
            }),
            onApprove: () => this.complete(this.reference).catch(this.onError)
          }).render('#device-paypal')
          this.paypalLoaded = true
        })
        .catch(error => console.error('failed to load the PayPal JS SDK script', error))
    }
  }
}
</script>
