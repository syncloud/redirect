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
            <p class="product-from">from {{ money(device.price + shipping) }}</p>
            <ul class="product-points">
              <li>Assembled and tested, with the system already written to the disk</li>
              <li>Two drive bays, so a second disk can be added later</li>
              <li>Arrives on this account, domain name and certificate ready</li>
            </ul>
          </div>
        </div>

        <details class="spec" data-testid="device-spec">
          <summary>More details</summary>

          <table class="spec-table" data-testid="device-spec-table">
            <tbody>
              <tr v-for="spec in device.specs" :key="spec.name">
                <th scope="row">{{ spec.name }}</th>
                <td>{{ spec.value }}</td>
              </tr>
            </tbody>
          </table>

          <h4 class="spec-heading">In the box</h4>
          <ul>
            <li>Odroid HC4 board in its case</li>
            <li>Boot memory, an SD card with Syncloud already written</li>
            <li>The SSD you choose below, in the first bay</li>
            <li>Power cable and ethernet cable</li>
          </ul>
          <p>
            The second bay is empty. Put a disk in it whenever you like and turn it on from
            Settings; nothing has to be decided now.
          </p>
          <p>
            It suits one household running apps for family or friends, or a small business
            sharing things between colleagues and customers.
          </p>
        </details>

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

        <el-button
          v-if="loggedIn"
          type="primary"
          size="large"
          class="continue"
          data-testid="shop-continue"
          @click="step = 'address'"
        >
          Continue
        </el-button>
      </div>

      <div v-if="loggedIn === false && step === 'choose'" class="sc-card" data-testid="shop-signin">
        <h2 class="sc-h2">Sign in to order</h2>
        <p>
          A device is tied to the account that owns it, so ordering needs one. It takes a
          moment and the same account runs the device afterwards.
        </p>
        <router-link
          class="sc-btn"
          to="/login?next=/shop"
          data-testid="shop-signin-link"
        >
          Sign in or create an account
        </router-link>
      </div>

      <div v-if="loggedIn && step !== 'choose'" class="sc-card chosen" data-testid="shop-chosen">
        <div>
          <span class="chosen-name">{{ device.name }}, {{ optionName }}</span>
          <span class="chosen-total">{{ money(total) }}</span>
        </div>
        <button type="button" class="chosen-change" data-testid="shop-change" @click="step = 'choose'">
          Change
        </button>
      </div>

      <div v-if="loggedIn && step === 'pay'" class="sc-card chosen" data-testid="shop-address-chosen">
        <div>
          <span class="chosen-name">{{ name }}</span>
          <span class="chosen-total chosen-address">{{ addressLine }}</span>
        </div>
        <button type="button" class="chosen-change" data-testid="shop-address-change" @click="step = 'address'">
          Change
        </button>
      </div>

      <div v-if="loggedIn && step === 'address'" class="sc-card sc-form" data-testid="device-address">
        <h2 class="sc-h2">Ship it to</h2>

        <div class="field">
          <input
            id="device-name"
            v-model="name"
            type="text"
            placeholder=" "
            autocomplete="name"
            data-testid="device-name"
          >
          <label for="device-name">Full name</label>
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
          <label for="device-address-line">Address</label>
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
            <label for="device-city">City</label>
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
            <label for="device-postcode">Postcode</label>
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
          <label for="device-country">Country</label>
        </div>

        <p v-if="incomplete" class="sc-help" data-testid="device-incomplete">
          We need the whole address before you can pay.
        </p>

        <el-button
          type="primary"
          size="large"
          class="continue"
          :disabled="incomplete"
          data-testid="shop-address-continue"
          @click="step = 'pay'"
        >
          Continue
        </el-button>
      </div>

      <div v-if="loggedIn && step === 'pay'" class="sc-card" data-testid="device-pay">
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

          <div class="pay-paypal-wrap">
            <div
              id="device-paypal"
              v-show="!incomplete"
              class="pay-paypal"
              data-testid="device-pay-paypal"
            />

            <div
              v-if="busy === 'paypal'"
              class="pay-overlay"
              data-testid="device-pay-busy"
            >
              <span class="pay-spinner" aria-label="Opening PayPal"/>
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
      ordered: false,
      error: '',
      busy: '',
      step: 'choose',
      paypalLoaded: false,
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
        })
        .catch(this.onError)

      if (this.loggedIn) {
        axios.get('/api/user')
          .then(response => { this.email = response.data.data.email })
          .catch(() => {})
      }
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
