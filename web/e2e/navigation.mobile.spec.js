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
  await expect(page.getByTestId('menu-burger')).toHaveCount(0)

  await page.goto('/register')
  await expect(page.getByTestId('register-heading')).toBeVisible()
  await expect(page.getByTestId('menu-burger')).toHaveCount(0)
})

test('theme can be switched and is remembered', async ({ page }) => {
  await page.goto('/login')
  const toggle = page.getByTestId('theme-toggle')
  await expect(toggle).toBeVisible()

  const before = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
  await toggle.click()
  const after = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
  expect(after).not.toBe(before)

  await page.reload()
  const persisted = await page.evaluate(() => document.documentElement.getAttribute('data-theme'))
  expect(persisted).toBe(after)
})
