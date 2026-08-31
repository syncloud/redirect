const { test, expect, shoot } = require('./fixtures')
const { uniqueEmail, registerUser, activateLatestUser, loginUser } = require('./helpers/user')

async function signedIn (page, prefix) {
  const email = uniqueEmail(prefix)
  await registerUser(page, email, 'password123')
  await activateLatestUser(page)
  await loginUser(page, email, 'password123')
  return email
}

async function goToShop (page) {
  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await page.getByTestId('nav-shop').click()
  await expect(page).toHaveURL(/\/shop$/)
}

async function address (page) {
  await page.getByTestId('device-name').fill('Ada Lovelace')
  await page.getByTestId('device-address-line').fill('1 Analytical Street')
  await page.getByTestId('device-city').fill('London')
  await page.getByTestId('device-postcode').fill('E1 6AN')
  await page.getByTestId('device-country').fill('United Kingdom')
}

test('the shop is readable without an account', async ({ page }, testInfo) => {
  await page.goto('/shop')

  await expect(page.getByTestId('device-choice')).toBeVisible()
  await expect(page.getByTestId('device-photo')).toBeVisible()
  await expect(page.getByTestId('device-total')).toContainText('£')
  await shoot(page, testInfo, 'shop-signed-out')

  await expect(page.getByTestId('shop-signin')).toBeVisible()
  await expect(page.getByTestId('device-pay')).toHaveCount(0)

  await page.getByTestId('shop-signin-link').click()
  await expect(page).toHaveURL(/\/login$/)
})

test('the shop prices the device from the catalogue', async ({ page }, testInfo) => {
  await signedIn(page, 'buy-price')
  await goToShop(page)

  await expect(page.getByTestId('device-choice')).toBeVisible()
  await expect(page.getByTestId('device-photo')).toBeVisible()
  const cheapest = await page.getByTestId('device-total').textContent()
  await shoot(page, testInfo, 'device-choice')

  await page.getByTestId('device-option-2tx2').click()
  const dearer = await page.getByTestId('device-total').textContent()

  expect(cheapest).not.toBe(dearer)
  expect(cheapest).toMatch(/^£\d+\.\d\d$/)
})

test('paying cannot start until the address is complete', async ({ page }, testInfo) => {
  await signedIn(page, 'buy-address')
  await goToShop(page)

  await expect(page.getByTestId('device-pay-stripe')).toBeDisabled()
  await expect(page.getByTestId('device-incomplete')).toBeVisible()

  await address(page)
  await expect(page.getByTestId('device-pay-stripe')).toBeEnabled()
  await shoot(page, testInfo, 'device-address')
})

test('the payment faker is reachable from the browser', async ({ page }) => {
  const sdk = await page.request.get('https://payments.syncloud.test/paypal/sdk/js')
  expect(sdk.status(), 'the paypal sdk script must be served').toBe(200)
  expect(await sdk.text()).toContain('paypal-faker-button')
})

test('a card payment is taken and the order is confirmed', async ({ page }, testInfo) => {
  await signedIn(page, 'buy-card')
  await goToShop(page)
  await address(page)

  const total = await page.getByTestId('device-total').textContent()
  await page.getByTestId('device-pay-stripe').click()

  await expect(page.getByTestId('faker-pay')).toBeVisible()
  await expect(page.getByTestId('faker-amount')).toContainText('GBP')
  await shoot(page, testInfo, 'device-checkout')

  await page.getByTestId('faker-pay').click()

  await expect(page.getByTestId('device-ordered')).toBeVisible()
  await expect(page.getByTestId('device-reference')).toContainText('Reference')
  await shoot(page, testInfo, 'device-ordered')

  expect(total).toMatch(/^£\d+\.\d\d$/)
})

test('a paypal payment is taken and the order is confirmed', async ({ page }, testInfo) => {
  await signedIn(page, 'buy-paypal')
  await goToShop(page)
  await address(page)

  await expect(page.getByTestId('paypal-faker-button')).toBeVisible()
  await shoot(page, testInfo, 'device-paypal')
  await page.getByTestId('paypal-faker-button').click()

  await expect(page.getByTestId('device-ordered')).toBeVisible()
  await expect(page.getByTestId('device-reference')).toContainText('Reference')
})
