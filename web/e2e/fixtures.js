import base from '@playwright/test'
import fs from 'node:fs'
import path from 'node:path'

const test = base.test.extend({
  page: async ({ page }, use, testInfo) => {
    const noise = []
    page.on('console', message => {
      if (message.type() === 'error') {
        noise.push(`console: ${message.text()}`)
      }
    })
    page.on('pageerror', error => noise.push(`pageerror: ${error.message}`))
    page.on('requestfailed', request => {
      noise.push(`requestfailed: ${request.method()} ${request.url()} ${request.failure()?.errorText}`)
    })
    page.on('response', response => {
      if (response.status() >= 400) {
        noise.push(`response: ${response.status()} ${response.url()}`)
      }
    })

    await use(page)

    if (testInfo.status !== testInfo.expectedStatus && noise.length) {
      await testInfo.attach('browser-problems', { body: noise.join('\n'), contentType: 'text/plain' })
      console.log(`\n--- browser problems in "${testInfo.title}" ---\n${noise.join('\n')}\n`)
    }
  }
})

function shotDir (testInfo) {
  const root = process.env.PLAYWRIGHT_ARTIFACT_DIR ||
    path.join(__dirname, '..', 'test-results')
  return path.join(root, `screenshots-${testInfo.project.name}`)
}

async function shoot (page, testInfo, name) {
  const dir = shotDir(testInfo)
  fs.mkdirSync(dir, { recursive: true })
  await page.screenshot({ path: path.join(dir, `${name}.png`), fullPage: true })
}

test.afterEach(async ({ page }, testInfo) => {
  if (testInfo.status !== testInfo.expectedStatus) {
    await page.screenshot({
      path: testInfo.outputPath('failure-full-page.png'),
      fullPage: true
    })
  }
})

const expect = base.expect

export {
  test,
  expect,
  shoot
}
