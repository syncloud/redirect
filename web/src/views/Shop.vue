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
          <div class="product-gallery">
            <button
              type="button"
              class="product-open"
              data-testid="device-photo-open"
              @click="openPhoto"
            >
              <img
                class="product-photo"
                :src="photos[photo]"
                :alt="device.name"
                width="180"
                height="180"
                data-testid="device-photo"
              >
            </button>
            <div class="product-thumbs" data-testid="device-thumbs">
              <button
                v-for="(each, index) in photos"
                :key="each"
                type="button"
                class="product-thumb"
                :class="{ 'product-thumb-on': index === photo }"
                :data-testid="'device-thumb-' + index"
                @click="photo = index"
              >
                <img :src="each" :alt="device.name" width="40" height="40">
              </button>
            </div>
          </div>
          <div class="product-info">
            <h2 class="sc-h2 product-name">{{ device.name }}</h2>
            <ul class="product-points" data-testid="device-points">
              <li>{{ $t('shop.point1') }}</li>
              <li>{{ $t('shop.point2') }}</li>
              <li>{{ $t('shop.point3') }}</li>
            </ul>
          </div>
        </div>

        <details class="spec" data-testid="device-spec" @toggle="onSpecToggle">
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
          @click="toAddress"
        >
          {{ $t('shop.continue') }}
        </el-button>
        <el-button
          v-else-if="loggedIn === false"
          type="primary"
          size="large"
          class="continue"
          data-testid="shop-signin-link"
          @click="toSignIn"
        >
          {{ $t('shop.signInToOrder') }}
        </el-button>
        <el-button
          v-else
          type="primary"
          size="large"
          class="continue"
          disabled
          data-testid="shop-waiting"
        >
          {{ $t('shop.continue') }}
        </el-button>
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
          @click="toPay"
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

    <div
      v-if="zoomed"
      class="lightbox"
      data-testid="device-lightbox"
      @click="zoomed = false"
      @touchstart="onTouchStart"
      @touchend="onTouchEnd"
    >
      <button
        type="button"
        class="lightbox-step lightbox-back"
        :aria-label="$t('shop.previousPhoto')"
        data-testid="device-photo-previous"
        @click.stop="previousPhoto"
      >
        &lsaquo;
      </button>
      <img
        class="lightbox-photo"
        :src="photos[photo]"
        :alt="device.name"
        data-testid="device-lightbox-photo"
        @click.stop
      >
      <button
        type="button"
        class="lightbox-step lightbox-forward"
        :aria-label="$t('shop.nextPhoto')"
        data-testid="device-photo-next"
        @click.stop="nextPhoto"
      >
        &rsaquo;
      </button>
      <button
        type="button"
        class="lightbox-close"
        :aria-label="$t('shop.close')"
        data-testid="device-lightbox-close"
        @click.stop="zoomed = false"
      >
        &times;
      </button>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import { loadScript } from '@paypal/paypal-js'
import {
  track,
  SHOP_VIEW,
  SHOP_SIGNED_OUT,
  SHOP_SPEC,
  SHOP_PHOTO,
  SHOP_SIGN_IN,
  SHOP_CONTINUE,
  SHOP_ADDRESS,
  SHOP_PAY_VIEW,
  SHOP_PAY_STRIPE,
  SHOP_PAY_PAYPAL,
  SHOP_PAYPAL_FAILED,
  SHOP_ERROR,
  SHOP_ORDERED
} from '../track'
import { CreditCard } from '@element-plus/icons-vue'
import { markRaw } from 'vue'

export default {
  name: 'DeviceView',
  props: {
    checkUserSession: Function,
    loggedIn: { type: Boolean, default: undefined }
  },
  data () {
    return {
      device: { name: '', price: 0, specs: [], options: [] },
      photos: [
        '/assets/syncloud-x1.jpg',
        '/assets/syncloud-x1-ports.jpg',
        '/assets/syncloud-x1-inside.jpg',
        '/assets/syncloud-x1-cooling.jpg'
      ],
      photo: 0,
      zoomed: false,
      swipeFrom: 0,
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
    },
    loggedIn (value) {
      if (value === false) {
        track(SHOP_SIGNED_OUT)
      }
    },
    incomplete (value, was) {
      if (was && !value) {
        track(SHOP_ADDRESS)
      }
    }
  },
  mounted () {
    window.addEventListener('keydown', this.onKey)
    track(SHOP_VIEW)
    if (this.loggedIn === false) {
      track(SHOP_SIGNED_OUT)
    }
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
  unmounted () {
    window.removeEventListener('keydown', this.onKey)
  },
  methods: {
    money (pence) {
      return `£${(pence / 100).toFixed(2)}`
    },
    onError (error) {
      track(SHOP_ERROR)
      this.busy = ''
      this.error = error.response && error.response.data
        ? error.response.data.message
        : this.$t('shop.somethingWentWrong')
    },
    onSpecToggle (toggle) {
      if (toggle.target.open) {
        track(SHOP_SPEC)
      }
    },
    openPhoto () {
      track(SHOP_PHOTO)
      this.zoomed = true
    },
    onKey (press) {
      if (!this.zoomed) {
        return
      }
      if (press.key === 'Escape') {
        this.zoomed = false
      }
      if (press.key === 'ArrowRight') {
        this.nextPhoto()
      }
      if (press.key === 'ArrowLeft') {
        this.previousPhoto()
      }
    },
    nextPhoto () {
      this.photo = (this.photo + 1) % this.photos.length
    },
    previousPhoto () {
      this.photo = (this.photo + this.photos.length - 1) % this.photos.length
    },
    onTouchStart (touch) {
      this.swipeFrom = touch.changedTouches[0].clientX
    },
    onTouchEnd (touch) {
      const moved = touch.changedTouches[0].clientX - this.swipeFrom
      if (Math.abs(moved) < 40) {
        return
      }
      if (moved < 0) {
        this.nextPhoto()
      } else {
        this.previousPhoto()
      }
    },
    toSignIn () {
      track(SHOP_SIGN_IN)
      this.$router.push('/login?next=/shop')
    },
    toAddress () {
      track(SHOP_CONTINUE)
      this.step = 'address'
    },
    toPay () {
      track(SHOP_PAY_VIEW)
      this.step = 'pay'
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
          track(SHOP_ORDERED)
        })
    },
    payWithStripe () {
      track(SHOP_PAY_STRIPE)
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
              track(SHOP_PAY_PAYPAL)
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
          track(SHOP_PAYPAL_FAILED)
          this.paypalLoading = false
          this.error = this.$t('shop.somethingWentWrong')
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

.product-gallery {
  flex: none;
}

.product-open {
  display: block;
  padding: 0;
  border: 0;
  background: none;
  line-height: 0;
  cursor: zoom-in;
}

.product-thumbs {
  display: flex;
  gap: 8px;
  margin-top: 8px;
}

.product-thumb {
  padding: 0;
  border: 1px solid var(--sc-border);
  border-radius: var(--sc-control-radius);
  background: var(--sc-field-bg);
  line-height: 0;
  overflow: hidden;
  cursor: pointer;
}

.product-thumb-on {
  border-color: var(--sc-primary);
}

.product-thumb img {
  width: 40px;
  height: 40px;
  object-fit: contain;
}

.lightbox {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(0, 0, 0, 0.8);
  cursor: zoom-out;
}

.lightbox-photo {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
  background: #fff;
  border-radius: var(--sc-control-radius);
}

.lightbox-step {
  position: absolute;
  top: 50%;
  width: 48px;
  height: 64px;
  transform: translateY(-50%);
  border: 0;
  background: rgba(255, 255, 255, 0.16);
  color: #fff;
  font-size: 2.4rem;
  line-height: 1;
  cursor: pointer;
}

.lightbox-back {
  left: 0;
  border-radius: 0 var(--sc-control-radius) var(--sc-control-radius) 0;
}

.lightbox-forward {
  right: 0;
  border-radius: var(--sc-control-radius) 0 0 var(--sc-control-radius);
}

.lightbox-close {
  position: absolute;
  top: 10px;
  right: 14px;
  padding: 4px 10px;
  border: 0;
  background: none;
  color: #fff;
  font-size: 2rem;
  line-height: 1;
  cursor: pointer;
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

@media (max-width: 560px) {
  .product-head {
    gap: 14px;
  }

  .product-photo {
    width: 96px;
    height: 96px;
  }

  .product-gallery {
    display: contents;
  }

  .product-open {
    order: 1;
  }

  .product-info {
    display: contents;
  }

  .product-name {
    order: 2;
    flex: 1;
    min-width: 0;
  }

  .product-thumbs {
    order: 3;
    flex-basis: 100%;
    margin-top: 0;
  }

  .product-thumb {
    flex: 1;
  }

  .product-thumb img {
    width: 100%;
    height: 60px;
  }

  .product-points {
    order: 4;
    flex-basis: 100%;
    margin-top: 0;
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
