const { test, expect } = require('./fixtures')
const { acquireDomain } = require('./helpers/api')
const { domain } = require('./helpers/env')
const { registerActivateAndLogin } = require('./helpers/user')

test('user can view and deactivate a device domain', async ({ page }) => {
  const { email, password } = await registerActivateAndLogin(page, 'devices')
  const userDomain = `pw-${Date.now()}.${domain}`

  await acquireDomain(userDomain, email, password)
  await page.goto('/')

  await expect(page.getByTestId('device-title')).toHaveText('Some Device')
  await expect(page.getByTestId('domain-name')).toHaveText(userDomain)

  await page.getByTestId('device-delete').first().click()
  await page.getByTestId('dialog-confirm').click()

  await expect(page.getByTestId('no-devices')).toBeVisible()
})
