const { test, expect } = require('./fixtures')

test('mobile navbar opens and navigates to register and login', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('#navbar')).toBeVisible()

  await page.locator('#navbar').click()
  await page.getByRole('link', { name: 'Register' }).click()
  await expect(page.getByRole('heading', { name: 'Register' })).toBeVisible()

  await page.locator('#navbar').click()
  await page.getByRole('link', { name: 'Log in' }).click()
  await expect(page.getByRole('heading', { name: 'Log in' })).toBeVisible()
})
