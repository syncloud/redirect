const domain = process.env.PLAYWRIGHT_DOMAIN || 'syncloud.test'

function webBaseUrl () {
  return `https://www.${domain}`
}

function apiBaseUrl () {
  return `https://api.${domain}`
}

export {
  domain,
  webBaseUrl,
  apiBaseUrl
}
