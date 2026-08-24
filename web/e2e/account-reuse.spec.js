const { test, expect } = require('./fixtures')
const { uniqueEmail, registerUser, activateLatestUser, loginUser } = require('./helpers/user')

test('deleting an account does not leave a session that opens a new account with the same email', async ({ page }) => {
  const email = uniqueEmail('reuse')
  const password = 'password123'

  await registerUser(page, email, password)
  await activateLatestUser(page)
  await loginUser(page, email, password)

  await page.goto('/account')
  await expect(page.getByTestId('account-title')).toBeVisible()
  await page.getByTestId('account-delete').click()
  await page.getByTestId('dialog-confirm').click()
  await expect(page.getByTestId('login-form')).toBeVisible()

  await registerUser(page, email, password)
  await activateLatestUser(page)

  await page.goto('/')
  await expect(page.getByTestId('login-form')).toBeVisible()
  await expect(page.getByTestId('no-devices')).toHaveCount(0)
})
