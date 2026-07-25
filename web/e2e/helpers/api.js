const { request, expect } = require('@playwright/test')
const { apiBaseUrl } = require('./env')

async function apiContext() {
  return await request.newContext({
    baseURL: apiBaseUrl(),
    ignoreHTTPSErrors: true
  })
}

async function acquireDomain(domainName, email, password) {
  const context = await apiContext()
  const response = await context.post('/domain/acquire_v2', {
    data: {
      domain: domainName,
      email,
      password,
      device_mac_address: '00:00:00:00:00:00',
      device_name: 'some-device',
      device_title: 'Some Device'
    }
  })
  expect(response.ok()).toBeTruthy()
  const payload = await response.json()
  expect(payload.success).toBeTruthy()
  await context.dispose()
  return payload.data.update_token
}

module.exports = {
  acquireDomain
}
