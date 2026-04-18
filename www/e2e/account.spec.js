const { test, expect } = require('./fixtures')
const { registerActivateAndLogin } = require('./helpers/user')

test('user can toggle notifications, subscribe with crypto, and cancel subscription', async ({ page }) => {
  await registerActivateAndLogin(page, 'account')
  await page.goto('/account')

  await expect(page.getByRole('heading', { name: 'Account' })).toBeVisible()

  const checkbox = page.locator('#chk_email')
  const initialValue = await checkbox.isChecked()
  await checkbox.click()
  await page.locator('#save').click()
  await expect(checkbox).toHaveJSProperty('checked', !initialValue)

  await page.locator('#crypto_year').click()
  await page.locator('#crypto_transaction_id').fill('12345678901')
  await page.locator('#crypto_subscribe_btn').click()
  await expect(page.locator('#subscription_active')).toBeVisible()

  await page.locator('#cancel').click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await expect(page.locator('#subscription_inactive')).toBeVisible()
})
