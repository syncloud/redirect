import { test, expect } from './fixtures'
import { uniqueEmail } from './helpers/user'
import { clearEmails } from './helpers/mailhog'

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
  await expect(page.getByTestId('check-email-complete')).toBeVisible()
})

test('registration omits gclid when the url has none', async ({ page }) => {
  const body = await submitRegistration(page, '/register')
  expect(body.gclid).toBeUndefined()
  await expect(page.getByTestId('check-email-complete')).toBeVisible()
})

test('an unrelated query parameter does not become a gclid', async ({ page }) => {
  const body = await submitRegistration(page, '/register?utm_source=newsletter')
  expect(body.gclid).toBeUndefined()
  await expect(page.getByTestId('check-email-complete')).toBeVisible()
})

test('the click id survives arriving anywhere and navigating to register', async ({ page }) => {
  await clearEmails()
  await page.goto('/login?gclid=landed-on-login-123')
  await expect(page.getByTestId('login-heading')).toBeVisible()

  const created = page.waitForRequest(
    r => r.url().includes('/api/user/create') && r.method() === 'POST')
  await page.getByTestId('login-register').click()
  await expect(page.getByTestId('register-heading')).toBeVisible()

  await page.getByTestId('register-email').fill(uniqueEmail('gclid-kept'))
  await page.getByTestId('register-password').fill('password123')
  await page.getByTestId('register-submit').click()

  expect((await created).postDataJSON().gclid).toBe('landed-on-login-123')
  await expect(page.getByTestId('check-email-complete')).toBeVisible()
})

test('a click id in the url still wins over a stored one', async ({ page }) => {
  await page.goto('/login?gclid=older-stored-id')
  const body = await submitRegistration(page, '/register?gclid=fresher-url-id')
  expect(body.gclid).toBe('fresher-url-id')
})
