import { test, expect, shoot } from './fixtures'
import { uniqueEmail, registerUser, activateLatestUser, loginUser } from './helpers/user'
import { waitForEmailTo } from './helpers/mailhog'

async function signedIn (page, prefix) {
  const email = uniqueEmail(prefix)
  await registerUser(page, email, 'password123')
  await activateLatestUser(page)
  await loginUser(page, email, 'password123')
  return email
}

async function signOut (page) {
  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await page.getByTestId('nav-logout').click()
  await expect(page).toHaveURL(/\/login/)
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
  await page.getByTestId('shop-continue').click()
  await expect(page.getByTestId('shop-chosen')).toBeVisible()
  await page.getByTestId('device-name').fill('Ada Lovelace')
  await page.getByTestId('device-address-line').fill('1 Analytical Street')
  await page.getByTestId('device-city').fill('London')
  await page.getByTestId('device-postcode').fill('E1 6AN')
  await page.getByTestId('device-country').fill('United Kingdom')
  await page.getByTestId('shop-address-continue').click()
  await expect(page.getByTestId('shop-address-chosen')).toBeVisible()
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
  await expect(page).toHaveURL(/\/login\?next=\/shop$/)
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

  await page.getByTestId('shop-continue').click()
  await expect(page.getByTestId('device-pay')).toHaveCount(0)
  await expect(page.getByTestId('shop-address-continue')).toBeDisabled()
  await expect(page.getByTestId('device-incomplete')).toBeVisible()
  await page.getByTestId('shop-change').click()

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
  const buyer = await signedIn(page, 'buy-card')
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

  const reference = await page.getByTestId('device-reference').textContent()
  const number = reference.replace('Reference ', '').trim()

  const confirmation = await waitForEmailTo(buyer)
  expect(confirmation.subject, 'the buyer is told their reference').toContain(number)
  expect(confirmation.body).toContain('Syncloud H4')
  expect(confirmation.body).toContain('Ada Lovelace')
  expect(confirmation.body).toContain('1 Analytical Street')
  expect(confirmation.body).toContain(total.replace('£', ''))

  const support = await waitForEmailTo('support@syncloud.it')
  expect(support.subject).toContain(number)
  expect(support.body, 'support must see which account ordered').toContain(`Account: ${buyer}`)
  expect(support.body).toContain('Ada Lovelace')
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

test('a buyer sees the order they just paid for', async ({ page }, testInfo) => {
  await signedIn(page, 'buy-orders')
  await goToShop(page)
  await address(page)

  await page.getByTestId('device-pay-stripe').click()
  await expect(page.getByTestId('faker-pay')).toBeVisible()
  await page.getByTestId('faker-pay').click()
  await expect(page.getByTestId('device-ordered')).toBeVisible()

  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await page.getByTestId('nav-orders').click()
  await expect(page).toHaveURL(/\/orders$/)

  await expect(page.getByTestId('order')).toHaveCount(1)
  await expect(page.getByTestId('order-status')).toHaveText('Ordered')
  await shoot(page, testInfo, 'orders')
})

test('a buyer is refused the admin order endpoints', async ({ page }) => {
  await signedIn(page, 'buy-noadmin')

  const all = await page.request.get('/api/device/orders/all')
  expect(all.status(), 'listing every order must be admin only').toBe(403)

  const status = await page.request.post('/api/device/order/status', {
    data: { reference: 'anything', status: 'sent', comment: 'nope' }
  })
  expect(status.status(), 'changing a status must be admin only').toBe(403)
})

test('a buyer is offered no admin menu and cannot open the admin page', async ({ page }) => {
  await signedIn(page, 'buy-nomenu')

  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await expect(page.getByTestId('nav-orders')).toBeVisible()
  await expect(page.getByTestId('nav-admin-orders'),
    'a buyer must not be offered the admin screen').toHaveCount(0)

  await page.goto('/admin/orders')
  await expect(page, 'typing the admin url must not show it').toHaveURL(/\/orders$/)
  await expect(page.getByTestId('orders-admin-table')).toHaveCount(0)
})

test('a buyer cannot open an order that is not theirs', async ({ page }) => {
  const other = await signedIn(page, 'buy-other')
  await goToShop(page)
  await address(page)
  await page.getByTestId('device-pay-stripe').click()
  await expect(page.getByTestId('faker-pay')).toBeVisible()
  await page.getByTestId('faker-pay').click()
  await expect(page.getByTestId('device-ordered')).toBeVisible()
  const reference = (await page.getByTestId('device-reference').textContent())
    .replace('Reference ', '').trim()
  expect(other).toContain('buy-other')

  const owner = await page.request.get(`/api/device/order?reference=${reference}`)
  expect(owner.status(), 'the account that ordered can read it').toBe(200)

  await signOut(page)
  await signedIn(page, 'buy-nosy')
  const stolen = await page.request.get(`/api/device/order?reference=${reference}`)
  expect(stolen.status(), 'another account must not read this order').not.toBe(200)
})

test('the orders page needs an account', async ({ page }) => {
  await page.goto('/orders')
  await expect(page).toHaveURL(/\/login/)
})
