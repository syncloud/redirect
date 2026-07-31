const { test, expect } = require('./fixtures')
const { waitForResetUrl } = require('./helpers/mailhog')
const { registerActivateAndLogin } = require('./helpers/user')

async function fillStable (locator, value) {
  await expect(async () => {
    await locator.fill(value)
    await expect(locator).toHaveValue(value)
  }).toPass()
}

test('user can reset password and log in with the new password', async ({ page }) => {
  const originalPassword = 'password123'
  const { email } = await registerActivateAndLogin(page, 'reset', originalPassword)

  const navbar = page.locator('#navbar')
  if (await navbar.isVisible()) {
    await navbar.click()
  }
  await page.locator('#logout').click()
  await expect(page.locator('#login')).toBeVisible()
  await expect(page.getByTestId('login-form')).toBeVisible()

  await page.getByTestId('login-forgot').click()
  await expect(page.getByTestId('forgot-form')).toBeVisible()
  await fillStable(page.getByTestId('forgot-email'), email)
  await page.getByTestId('forgot-send').click()
  await expect(page.getByTestId('check-email-complete')).toBeVisible()

  const resetUrl = await waitForResetUrl()
  const newPassword = 'password456'

  await page.goto(resetUrl)
  await expect(page.getByTestId('reset-form')).toBeVisible()
  await fillStable(page.getByTestId('reset-password'), newPassword)
  await page.getByTestId('reset-submit').click()
  await expect(page).toHaveURL(/\/login$/)

  await expect(page.getByTestId('login-form')).toBeVisible()
  await fillStable(page.getByTestId('login-email'), email)
  await fillStable(page.getByTestId('login-password'), newPassword)
  await page.getByTestId('login-submit').click()

  await expect(page.locator('#no_domains')).toBeVisible()
})
