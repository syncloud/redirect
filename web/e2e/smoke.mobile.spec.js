import { test, expect, shoot } from './fixtures'

test('mobile visitor with no account is given somewhere to go', async ({ page }, testInfo) => {
  await page.goto('/')

  await expect(page.getByTestId('home-intro')).toBeVisible()
  await shoot(page, testInfo, 'home-signed-out')

  await page.getByTestId('home-login').click()
  await expect(page.getByTestId('login-email')).toBeVisible()
  await expect(page.getByTestId('login-password')).toBeVisible()
})

test('mobile visitor can reach the shop without an account', async ({ page }, testInfo) => {
  await page.goto('/')

  await page.getByTestId('home-shop').click()
  await expect(page).toHaveURL(/\/shop$/)
  await expect(page.getByTestId('device-choice')).toBeVisible()
  await expect(page.getByTestId('device-total')).toContainText('£')
  await shoot(page, testInfo, 'shop-mobile-signed-out')
})
