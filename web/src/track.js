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

function language () {
  try {
    return i18n.global.locale.value
  } catch {
    return ''
  }
}

export function track (event) {
  const body = JSON.stringify({
    event,
    gclid: storedGclid() != null,
    language: language()
  })
  try {
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/event', new Blob([body], { type: 'application/json' }))
      return
    }
  } catch {

  }
  fetch('/api/event', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body,
    keepalive: true
  }).catch(() => {})
}
