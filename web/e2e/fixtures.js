const base = require('@playwright/test')
const fs = require('node:fs')
const path = require('node:path')

const test = base.test.extend({})

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

module.exports = {
  test,
  expect: base.expect,
  shoot
}
