import { storedGclid } from './attribution'
import i18n from './i18n'

export const SHOP_VIEW = 'shop.view'
export const SHOP_SIGNED_OUT = 'shop.signedout'
export const SHOP_OPTION = 'shop.option'
export const SHOP_SPEC = 'shop.spec'
export const SHOP_SIGN_IN = 'shop.signin'
export const SHOP_CONTINUE = 'shop.continue'
export const SHOP_ADDRESS = 'shop.address'
export const SHOP_PAY_VIEW = 'shop.pay.view'
export const SHOP_PAY_STRIPE = 'shop.pay.stripe'
export const SHOP_PAY_PAYPAL = 'shop.pay.paypal'
export const SHOP_PAYPAL_FAILED = 'shop.paypal.failed'
export const SHOP_ERROR = 'shop.error'
export const SHOP_ORDERED = 'shop.ordered'

export const PLAN_VIEW = 'plan.view'
export const PLAN_SUBSCRIBED = 'plan.subscribed'
export const PLAN_TIER = 'plan.tier'
export const PLAN_PERIOD = 'plan.period'
export const PLAN_STRIPE = 'plan.stripe'
export const PLAN_PAYPAL = 'plan.paypal'
export const PLAN_PAYPAL_FAILED = 'plan.paypal.failed'
export const PLAN_CRYPTO = 'plan.crypto'
export const PLAN_CRYPTO_SENT = 'plan.crypto.sent'
export const PLAN_SWITCH_ANNUAL = 'plan.switch.annual'
export const PLAN_CANCEL = 'plan.cancel'
export const PLAN_ERROR = 'plan.error'

function language () {
  try {
    return i18n.global.locale.value
  } catch {
    return ''
  }
}

export function track (event) {
  try {
    const body = JSON.stringify({
      event,
      gclid: storedGclid() != null,
      language: language()
    })
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/event', new Blob([body], { type: 'application/json' }))
      return
    }
    fetch('/api/event', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body,
      keepalive: true
    }).catch(() => {})
  } catch {

  }
}
