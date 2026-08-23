const { test, expect } = require('./fixtures')

test('mobile navbar opens and navigates to register and login', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('#navbar')).toBeVisible()

  await page.locator('#navbar').click()
  await page.getByTestId('nav-register').click()
  await expect(page.getByTestId('register-heading')).toBeVisible()

  await page.locator('#navbar').click()
  await page.getByTestId('nav-login').click()
  await expect(page.getByTestId('login-heading')).toBeVisible()
})
