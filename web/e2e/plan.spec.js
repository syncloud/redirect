import { readFileSync } from 'node:fs'
import { test, expect } from './fixtures'
import { registerActivateAndLogin } from './helpers/user'

async function openAccount (page) {
  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await page.getByTestId('nav-account').click()
  await expect(page.getByTestId('account-title')).toBeVisible()
}

async function captureBeacons (page) {
  await page.addInitScript(() => {
    window.beacons = []
    const send = navigator.sendBeacon && navigator.sendBeacon.bind(navigator)
    navigator.sendBeacon = (url, body) => {
      if (String(url).includes('/api/event')) {
        body.text().then(text => window.beacons.push(JSON.parse(text)))
      }
      return send ? send(url, body) : true
    }
  })
}

test('the plan page reports that the offer was shown and which plan was picked', async ({ page }) => {
  await captureBeacons(page)
  await registerActivateAndLogin(page, 'plan')

  await openAccount(page)

  await page.getByTestId('billing-year').click()

  await expect
    .poll(async () => (await page.evaluate(() => window.beacons)).map(each => each.event))
    .toEqual(expect.arrayContaining(['plan.view', 'plan.period']))

  const reported = await page.evaluate(() => window.beacons)
  for (const each of reported) {
    expect(each.gclid, 'no click id was set, so nothing may claim an ad').toBe(false)
    expect(each.language).toBe('en')
    expect(JSON.stringify(each), 'a beacon must not carry a session').not.toContain('session')

    const echoed = await page.request.post('/api/event', { data: each })
    expect(echoed.status(), `${each.event} must be a step the server counts`).toBe(204)
  }
})

test('a subscriber is not counted as having been shown the offer', async ({ page }) => {
  await captureBeacons(page)
  await registerActivateAndLogin(page, 'plan-subscribed')

  await openAccount(page)

  await page.locator('#crypto_year').click()
  await page.locator('#crypto_transaction_id').fill('12345678901')
  await page.locator('#crypto_subscribe_btn').click()
  await expect(page.getByTestId('billing-current')).toBeVisible()

  await expect
    .poll(async () => (await page.evaluate(() => window.beacons)).map(each => each.event))
    .toEqual(expect.arrayContaining(['plan.crypto', 'plan.crypto.sent', 'plan.subscribed']))

  await page.reload()
  await expect(page.getByTestId('billing-current')).toBeVisible()

  const afterReload = await page.evaluate(() => window.beacons)
  expect(afterReload.map(each => each.event),
    'a paying subscriber reloading the page has not been shown the offer').not.toContain('plan.view')
})

test('the event endpoint counts every step the pages can send and refuses anything else', async ({ page }) => {
  await page.goto('/shop')

  const source = readFileSync(new URL('../src/track.js', import.meta.url), 'utf8')
  const names = [...source.matchAll(/^export const [A-Z_0-9]+ = '([^']+)'$/gm)].map(each => each[1])
  expect(names.length, 'the tracker must declare the step names it can send').toBeGreaterThan(20)

  for (const event of names) {
    const known = await page.request.post('/api/event', { data: { event, gclid: false, language: 'en' } })
    expect(known.status(), `${event} must be a step the server counts`).toBe(204)
  }

  for (const event of ['plan', 'plan.', 'plan.anything', 'shop.anything', 'user.deleted', '']) {
    const refused = await page.request.post('/api/event', { data: { event } })
    expect(refused.status(), `${event} must not become a metric label`).toBe(404)
  }
})
