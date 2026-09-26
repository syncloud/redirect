export const PRICE_CHANGE_DATE = new Date('2026-10-15T00:00:00Z')

const PRO_MONTHLY_AMOUNT = '£5'
const PRO_MONTHLY_AMOUNT_FROM_PRICE_CHANGE_DATE = '£7'

export const TIER_PRICES = {
  pro: { month: PRO_MONTHLY_AMOUNT + ' / month', year: '£60 / year' },
  max: { month: '£15 / month', year: '£180 / year' }
}

export const TIER_PRICES_FROM_PRICE_CHANGE_DATE = {
  pro: { month: PRO_MONTHLY_AMOUNT_FROM_PRICE_CHANGE_DATE + ' / month', year: '£70 / year' }
}

export function priceChangePending (now = Date.now()) {
  return now < PRICE_CHANGE_DATE.getTime()
}

export function proMonthlyAmount (now = Date.now()) {
  return priceChangePending(now) ? PRO_MONTHLY_AMOUNT : PRO_MONTHLY_AMOUNT_FROM_PRICE_CHANGE_DATE
}

export function priceChangeDateText (locale) {
  const tag = locale === 'en' ? 'en-GB' : locale
  return new Intl.DateTimeFormat(tag, {
    day: 'numeric',
    month: 'long',
    year: 'numeric',
    timeZone: 'UTC'
  }).format(PRICE_CHANGE_DATE)
}
