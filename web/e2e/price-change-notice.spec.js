import { test, expect } from './fixtures'
import { registerActivateAndLogin } from './helpers/user'

async function goToAccount (page) {
  const burger = page.getByTestId('menu-burger')
  if (await burger.isVisible()) {
    await burger.click()
  }
  await page.getByTestId('nav-account').click()
  await expect(page.getByTestId('account-title')).toBeVisible()
}

test('the price change notice shows before the change date and not after', async ({ page }) => {
  await page.clock.setFixedTime(new Date('2026-10-01T10:00:00Z'))
  await registerActivateAndLogin(page, 'price-change')

  await goToAccount(page)
  await expect(page.getByTestId('price-change-notice')).toBeVisible()
  await expect(page.getByTestId('price-change-notice')).toContainText('£70 / year')

  await page.clock.setFixedTime(new Date('2026-10-20T10:00:00Z'))
  await page.reload()
  await expect(page.getByTestId('plan-pro')).toBeVisible()
  await expect(page.getByTestId('price-change-notice')).toHaveCount(0)
})
