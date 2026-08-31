<template>
  <div class="sc-page">
    <h1 class="sc-h1">Shop</h1>
    <p class="sc-lead">
      A Syncloud device, assembled and ready to plug in. It arrives on this account, so
      there is nothing else to sign up for.
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
            <p class="product-from">from {{ money(device.price + shipping) }}</p>
            <ul class="product-points">
              <li>Assembled and tested, with the system already written to the disk</li>
              <li>Your files, mail, photos, passwords and notes, on hardware you own</li>
              <li>Arrives on this account, domain name and certificate ready</li>
            </ul>
          </div>
        </div>

        <h3 class="option-title">Storage</h3>
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
            <span class="option-extra">{{ each.extra ? `+ ${money(each.extra)}` : 'included' }}</span>
          </button>
        </div>

        <div class="sc-summary">
          <div class="sc-summary-row">
            <span>{{ device.name }}, {{ optionName }}</span>
            <span data-testid="device-price">{{ money(device.price + extra) }}</span>
          </div>
          <div class="sc-summary-row">
            <span>Delivery</span>
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
            :disabled="incomplete || paying"
            :loading="busy === 'stripe'"
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

          <p
            v-if="busy"
            class="pay-busy"
            data-testid="device-pay-busy"
          >
            {{ busy === 'paypal' ? 'Opening PayPal, this can take a few seconds.' : 'Opening the card checkout.' }}
          </p>
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
      busy: '',
      paypalLoaded: false,
      CreditCard: markRaw(CreditCard)
    }
  },
  computed: {
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
      this.busy = ''
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
      this.busy = 'stripe'
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
            onClick: () => {
              this.error = ''
              this.busy = 'paypal'
            },
            onCancel: () => {
              this.busy = ''
            },
            createOrder: () => {
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
            onApprove: () => this.complete(this.reference)
              .then(() => { this.busy = '' })
              .catch(this.onError),
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

.product-head {
  display: flex;
  align-items: flex-start;
  gap: 22px;
  flex-wrap: wrap;
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

.pay-busy {
  margin: 4px 0 0;
  color: var(--sc-muted);
  font-size: 0.9rem;
}
</style>
