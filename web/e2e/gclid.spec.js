const { test, expect } = require('@playwright/test')
const { uniqueEmail } = require('./helpers/user')
const { clearEmails } = require('./helpers/mailhog')

async function submitRegistration (page, path) {
  await clearEmails()
  const created = page.waitForRequest(
    r => r.url().includes('/api/user/create') && r.method() === 'POST')
  await page.goto(path)
  await page.getByTestId('register-email').fill(uniqueEmail('gclid'))
  await page.getByTestId('register-password').fill('password123')
  await page.getByTestId('register-submit').click()
  return (await created).postDataJSON()
}

test('registration forwards the gclid from the landing url', async ({ page }) => {
  const body = await submitRegistration(page, '/register?gclid=e2e-test-gclid-123')
  expect(body.gclid).toBe('e2e-test-gclid-123')
  await expect(page.getByRole('heading', { name: 'Complete' })).toBeVisible()
})

test('registration omits gclid when the url has none', async ({ page }) => {
  const body = await submitRegistration(page, '/register')
  expect(body.gclid).toBeUndefined()
  await expect(page.getByRole('heading', { name: 'Complete' })).toBeVisible()
})

test('an unrelated query parameter does not become a gclid', async ({ page }) => {
  const body = await submitRegistration(page, '/register?utm_source=newsletter')
  expect(body.gclid).toBeUndefined()
  await expect(page.getByRole('heading', { name: 'Complete' })).toBeVisible()
})
