import { test, expect } from './fixtures'
import { acquireDomain } from './helpers/api'
import { domain } from './helpers/env'
import { registerActivateAndLogin } from './helpers/user'

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
