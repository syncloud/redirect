import { test, expect } from './fixtures'
import { uniqueEmail, registerUser, activateLatestUser, loginUser } from './helpers/user'

test('user can register, activate, and log in', async ({ page }) => {
  const email = uniqueEmail('auth')
  const password = 'password123'

  await registerUser(page, email, password)
  await activateLatestUser(page)
  await loginUser(page, email, password)

  await expect(page.getByTestId('no-devices')).toBeVisible()
  await expect(page.getByTestId('no-devices-steps')).toContainText('install Syncloud on your own hardware')
  await expect(page.getByTestId('no-devices-setup')).toHaveAttribute('href', 'https://syncloud.org/setup')
})

test('invalid email shows login validation error', async ({ page }) => {
  await page.goto('/login')
  await page.locator('#email').fill('wrong_user')
  await page.locator('#password').fill('wrong_password')
  await page.locator('#submit').click()

  await expect(page.locator('#help-email')).toContainText('Not valid email')
})

test('wrong password shows authentication error', async ({ page }) => {
  await page.goto('/login')
  await page.locator('#email').fill('wrong_user@example.com')
  await page.locator('#password').fill('wrong_password')
  await page.locator('#submit').click()

  await expect(page.locator('#error')).toContainText('authentication failed')
})
