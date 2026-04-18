const { test, expect } = require('./fixtures')
const { waitForResetUrl } = require('./helpers/mailhog')
const { registerActivateAndLogin } = require('./helpers/user')

test('user can reset password and log in with the new password', async ({ page }) => {
  const originalPassword = 'password123'
  const { email } = await registerActivateAndLogin(page, 'reset', originalPassword)

  await page.locator('#logout').click()
  await expect(page.locator('#login')).toBeVisible()

  await page.locator('#forgot').click()
  await page.locator('#email').fill(email)
  await page.locator('#send').click()
  await expect(page.getByRole('heading', { name: 'Complete' })).toBeVisible()

  const resetUrl = await waitForResetUrl()
  const newPassword = 'password456'

  await page.goto(resetUrl)
  await page.locator('#password').fill(newPassword)
  await page.locator('#reset').click()
  await expect(page).toHaveURL(/\/login$/)

  await page.locator('#email').fill(email)
  await page.locator('#password').fill(newPassword)
  await page.locator('#submit').click()

  await expect(page.locator('#no_domains')).toBeVisible()
})
