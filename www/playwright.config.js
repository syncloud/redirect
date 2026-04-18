const { defineConfig, devices } = require('@playwright/test')

const domain = process.env.PLAYWRIGHT_DOMAIN || 'syncloud.test'

module.exports = defineConfig({
  testDir: './e2e',
  outputDir: 'test-results',
  timeout: 60 * 1000,
  expect: {
    timeout: 10 * 1000
  },
  fullyParallel: false,
  retries: process.env.CI ? 1 : 0,
  reporter: process.env.CI
    ? [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]]
    : [['list'], ['html', { open: 'never', outputFolder: 'playwright-report' }]],
  use: {
    baseURL: `https://www.${domain}`,
    ignoreHTTPSErrors: true,
    screenshot: 'off',
    trace: 'retain-on-failure',
    video: 'on'
  },
  projects: [
    {
      name: 'desktop',
      testIgnore: [/\.mobile\.spec\.js$/],
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1440, height: 960 }
      }
    },
    {
      name: 'mobile',
      testMatch: [/.*\.mobile\.spec\.js$/],
      use: {
        ...devices['Pixel 5']
      }
    }
  ]
})
