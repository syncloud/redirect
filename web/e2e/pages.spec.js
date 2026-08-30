const { test, expect, shoot } = require('./fixtures')
const { uniqueEmail, registerUser, activateLatestUser } = require('./helpers/user')

test('the register page offers a way to create an account', async ({ page }, testInfo) => {
  await page.goto('/register')

  await expect(page.getByTestId('register-form')).toBeVisible()
  await expect(page.getByTestId('register-heading')).toBeVisible()
  await shoot(page, testInfo, 'register')
})

test('registering leads to the check email page and back to registering', async ({ page }, testInfo) => {
  await registerUser(page, uniqueEmail('pages-check'), 'password123')

  await expect(page.getByTestId('check-email-complete')).toBeVisible()
  await shoot(page, testInfo, 'check-email')

  await page.getByTestId('check-email-register').click()
  await expect(page.getByTestId('register-form')).toBeVisible()
})

test('the activation page confirms and leads to logging in', async ({ page }, testInfo) => {
  await registerUser(page, uniqueEmail('pages-activate'), 'password123')
  await activateLatestUser(page)

  await expect(page.getByTestId('activate-message')).toBeVisible()
  await shoot(page, testInfo, 'activate')

  await page.getByTestId('activate-login').click()
  await expect(page).toHaveURL(/\/login$/)
})

test('the privacy policy is readable without an account', async ({ page }, testInfo) => {
  await page.goto('/privacy')

  await expect(page.getByTestId('privacy-heading')).toBeVisible()
  await expect(page.getByTestId('privacy-content')).toContainText('personal data')
  await shoot(page, testInfo, 'privacy')
})

test('the error page says what happened and offers a way out', async ({ page }, testInfo) => {
  await page.goto('/error')

  await expect(page.getByTestId('error-message')).toBeVisible()
  await shoot(page, testInfo, 'error')

  await page.getByTestId('error-home').click()
  await expect(page).not.toHaveURL(/\/error$/)
})
