import { test, expect } from './fixtures'
import { registerActivateAndLogin } from './helpers/user'

test('a paypal monthly subscriber switches to annual billing', async ({ page }) => {
  await registerActivateAndLogin(page, 'switch-annual')
  await page.goto('/account')

  await page.getByTestId('billing-month').click()
  await expect(page.getByTestId('paypal-faker-button')).toBeVisible()
  await page.getByTestId('paypal-faker-button').click()

  await expect(page.locator('#subscription_active')).toBeVisible()
  await expect(page.getByTestId('billing-current')).toHaveText('£5 / month')
  await expect(page.getByTestId('switch-annual')).toBeVisible()

  await page.getByTestId('switch-annual').click()
  await page.getByTestId('dialog-confirm').click()

  await expect(page.getByTestId('billing-current')).toHaveText('£60 / year')
  await expect(page.getByTestId('switch-annual')).toHaveCount(0)
})

test('a paypal max subscriber sees max pricing and switches to annual', async ({ page }) => {
  await registerActivateAndLogin(page, 'switch-max')
  await page.goto('/account')

  await page.getByTestId('billing-month').click()
  await page.getByTestId('plan-max').click()
  await expect(page.getByTestId('paypal-faker-button')).toBeVisible()
  await page.getByTestId('paypal-faker-button').click()

  await expect(page.locator('#subscription_active')).toBeVisible()
  await expect(page.getByTestId('billing-current')).toHaveText('£15 / month')

  await page.getByTestId('switch-annual').click()
  await page.getByTestId('dialog-confirm').click()

  await expect(page.getByTestId('billing-current')).toHaveText('£180 / year')
})
