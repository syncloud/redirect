import { track, SHOP_VIEW, SHOP_PAY_STRIPE } from '../../src/track'

const KEY = 'syncloud.gclid'

function sent (beacon) {
  return JSON.parse(beacon.mock.calls[0][1].parts[0])
}

describe('the shop beacon', () => {
  let beacon

  beforeEach(() => {
    window.localStorage.clear()
    beacon = jest.fn()
    navigator.sendBeacon = beacon
    global.Blob = class {
      constructor (parts, options) {
        this.parts = parts
        this.type = options.type
      }
    }
  })

  it('posts the step to the event endpoint', () => {
    track(SHOP_VIEW)
    expect(beacon).toHaveBeenCalledTimes(1)
    expect(beacon.mock.calls[0][0]).toBe('/api/event')
    expect(sent(beacon).event).toBe(SHOP_VIEW)
  })

  it('says the visitor came from an ad only when a click id was kept', () => {
    track(SHOP_PAY_STRIPE)
    expect(sent(beacon).gclid).toBe(false)

    beacon.mockClear()
    window.localStorage.setItem(KEY, JSON.stringify({ gclid: 'abc', at: Date.now() }))
    track(SHOP_PAY_STRIPE)
    expect(sent(beacon).gclid).toBe(true)
  })

  it('forgets a click id older than the attribution window', () => {
    const ancient = Date.now() - 91 * 24 * 60 * 60 * 1000
    window.localStorage.setItem(KEY, JSON.stringify({ gclid: 'abc', at: ancient }))
    track(SHOP_VIEW)
    expect(sent(beacon).gclid).toBe(false)
  })

  it('names the language the page was shown in', () => {
    track(SHOP_VIEW)
    expect(sent(beacon).language).toBe('en')
  })

  it('never sends the click id itself, only whether there is one', () => {
    window.localStorage.setItem(KEY, JSON.stringify({ gclid: 'secret-click-id', at: Date.now() }))
    track(SHOP_VIEW)
    expect(JSON.stringify(sent(beacon))).not.toContain('secret-click-id')
  })

  it('falls back to fetch where sendBeacon is missing', () => {
    navigator.sendBeacon = undefined
    global.fetch = jest.fn(() => Promise.resolve())
    track(SHOP_VIEW)
    expect(global.fetch).toHaveBeenCalledWith('/api/event', expect.objectContaining({ method: 'POST' }))
  })
})

describe('a beacon that cannot be sent', () => {
  let beacon
  let fetching

  beforeEach(() => {
    beacon = navigator.sendBeacon
    fetching = global.fetch
  })

  afterEach(() => {
    navigator.sendBeacon = beacon
    global.fetch = fetching
  })

  it('never throws into the payment it was measuring', () => {
    navigator.sendBeacon = () => {
      throw new Error('blocked by the browser')
    }

    expect(() => track(SHOP_PAY_STRIPE)).not.toThrow()
  })

  it('never throws when the browser offers no transport at all', () => {
    delete navigator.sendBeacon
    delete global.fetch

    expect(() => track(SHOP_PAY_STRIPE)).not.toThrow()
  })
})
