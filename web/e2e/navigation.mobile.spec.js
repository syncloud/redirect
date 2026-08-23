const { test, expect } = require('./fixtures')

test('mobile user moves between login and register without a menu', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByTestId('login-heading')).toBeVisible()

  await page.getByTestId('login-register').click()
  await expect(page.getByTestId('register-heading')).toBeVisible()

  await page.getByTestId('register-login').click()
  await expect(page.getByTestId('login-heading')).toBeVisible()
})

test('auth pages show no site navigation', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByTestId('login-heading')).toBeVisible()
  await expect(page.locator('#navbar')).toHaveCount(0)

  await page.goto('/register')
  await expect(page.getByTestId('register-heading')).toBeVisible()
  await expect(page.locator('#navbar')).toHaveCount(0)
})
