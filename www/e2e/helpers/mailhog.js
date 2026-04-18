async function check(response, message) {
  if (!response.ok) {
    throw new Error(`${message}: ${response.status} ${await response.text()}`)
  }
}

async function clearEmails () {
  const response = await fetch('http://mail:8025/api/v1/messages', {
    method: 'DELETE'
  })
  await check(response, 'Failed to clear MailHog messages')
}

async function fetchMessages () {
  const response = await fetch('http://mail:8025/api/v1/messages')
  await check(response, 'Failed to fetch MailHog messages')
  return await response.json()
}

function bodyFromMessage(message) {
  return message.Content.Body.replace(/=\r\n/g, '')
}

async function waitForMessage(extract, attempts = 10) {
  for (let attempt = 0; attempt < attempts; attempt += 1) {
    const messages = await fetchMessages()
    if (messages.length > 0) {
      const result = extract(bodyFromMessage(messages[0]))
      if (result) {
        return result
      }
    }
    await new Promise(resolve => setTimeout(resolve, 1000))
  }
  throw new Error('Timed out waiting for email message')
}

function extractActivateUrl(body) {
  const match = body.match(/activate your account: (https:\/\/.*)\r/)
  return match ? match[1] : null
}

function extractResetUrl(body) {
  const match = body.match(/reset your password: (https:\/\/.*)\r/)
  return match ? match[1] : null
}

async function waitForActivateUrl() {
  return await waitForMessage(extractActivateUrl)
}

async function waitForResetUrl() {
  return await waitForMessage(extractResetUrl)
}

module.exports = {
  clearEmails,
  waitForActivateUrl,
  waitForResetUrl
}
