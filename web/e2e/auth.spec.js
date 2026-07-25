const { test, expect } = require('./fixtures')
const { uniqueEmail, registerUser, activateLatestUser, loginUser } = require('./helpers/user')

test('user can register, activate, and log in', async ({ page }) => {
  const email = uniqueEmail('auth')
  const password = 'password123'

  await registerUser(page, email, password)
  await activateLatestUser(page)
  await loginUser(page, email, password)

  await expect(page.locator('#no_domains')).toContainText('You do not have any activated devices')
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
