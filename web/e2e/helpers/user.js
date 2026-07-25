const { expect } = require('@playwright/test')
const { clearEmails, waitForActivateUrl } = require('./mailhog')

function uniqueEmail(prefix = 'playwright') {
  const suffix = `${Date.now()}-${Math.floor(Math.random() * 100000)}`
  return `${prefix}-${suffix}@syncloud.test`
}

async function registerUser(page, email, password) {
  await clearEmails()
  await page.goto('/register')
  await page.locator('#register_email').fill(email)
  await page.locator('#register_password').fill(password)
  await page.locator('#btnregister').click()
  await expect(page.getByRole('heading', { name: 'Complete' })).toBeVisible()
}

async function activateLatestUser(page) {
  const activateUrl = await waitForActivateUrl()
  await page.goto(activateUrl)
  await expect(page.locator('#activated')).toContainText('User was activated')
}

async function loginUser(page, email, password) {
  await page.goto('/login')
  await page.locator('#email').fill(email)
  await page.locator('#password').fill(password)
  await page.locator('#submit').click()
  await expect(page.locator('#no_domains')).toBeVisible()
}

async function registerActivateAndLogin(page, prefix = 'playwright', password = 'password123') {
  const email = uniqueEmail(prefix)
  await registerUser(page, email, password)
  await activateLatestUser(page)
  await loginUser(page, email, password)
  return { email, password }
}

module.exports = {
  uniqueEmail,
  registerUser,
  activateLatestUser,
  loginUser,
  registerActivateAndLogin
}
