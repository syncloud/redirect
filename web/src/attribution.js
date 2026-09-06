const KEY = 'syncloud.gclid'
const MAX_AGE_MS = 90 * 24 * 60 * 60 * 1000

export function captureGclid (search) {
  const query = search === undefined ? window.location.search : search
  const gclid = new URLSearchParams(query).get('gclid')
  if (!gclid) {
    return
  }
  try {
    window.localStorage.setItem(KEY, JSON.stringify({ gclid, at: Date.now() }))
  } catch {

  }
}

export function storedGclid () {
  let entry
  try {
    entry = JSON.parse(window.localStorage.getItem(KEY))
  } catch {
    return null
  }
  if (!entry || !entry.gclid || !entry.at) {
    return null
  }
  if (Date.now() - entry.at > MAX_AGE_MS) {
    return null
  }
  return entry.gclid
}
