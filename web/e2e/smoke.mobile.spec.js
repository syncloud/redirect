const { test, expect } = require('./fixtures')

test('mobile unauthenticated user lands on login flow', async ({ page }) => {
  await page.goto('/')
  await expect(page.locator('#email')).toBeVisible()
  await expect(page.locator('#password')).toBeVisible()
})
