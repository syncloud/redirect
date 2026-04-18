const { test, expect } = require('./fixtures')
const { acquireDomain } = require('./helpers/api')
const { domain } = require('./helpers/env')
const { registerActivateAndLogin } = require('./helpers/user')

test('user can view and deactivate a device domain', async ({ page }) => {
  const { email, password } = await registerActivateAndLogin(page, 'devices')
  const userDomain = `pw-${Date.now()}.${domain}`

  await acquireDomain(userDomain, email, password)
  await page.goto('/')

  await expect(page.getByText('Some Device')).toBeVisible()
  await expect(page.getByText(userDomain)).toBeVisible()

  await page.locator('#delete').first().click()
  await page.getByRole('button', { name: 'Confirm' }).click()

  await expect(page.getByText('Some Device')).toHaveCount(0)
})
