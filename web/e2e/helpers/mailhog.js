async function check (response, message) {
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

function bodyFromMessage (message) {
  return message.Content.Body
    .replace(/=\r\n/g, '')
    .replace(/=([0-9A-Fa-f]{2})/g, (_, hex) => String.fromCharCode(parseInt(hex, 16)))
}

async function waitForMessage (extract, timeoutMs = 30000, pollMs = 250) {
  const deadline = Date.now() + timeoutMs
  for (;;) {
    const messages = await fetchMessages()
    if (messages.length > 0) {
      const result = extract(bodyFromMessage(messages[0]))
      if (result) {
        return result
      }
    }
    if (Date.now() >= deadline) {
      throw new Error(`Timed out after ${timeoutMs}ms waiting for email message`)
    }
    await new Promise(resolve => setTimeout(resolve, pollMs))
  }
}

function extractActivateUrl (body) {
  const match = body.match(/activate your account: (https:\/\/.*)\r/)
  return match ? match[1] : null
}

function extractResetUrl (body) {
  const match = body.match(/reset your password: (https:\/\/.*)\r/)
  return match ? match[1] : null
}

async function waitForActivateUrl () {
  return await waitForMessage(extractActivateUrl)
}

async function waitForResetUrl () {
  return await waitForMessage(extractResetUrl)
}

module.exports = {
  clearEmails,
  waitForActivateUrl,
  waitForResetUrl
}
