import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Shop from '../../src/views/Shop.vue'
import { loadScript } from '@paypal/paypal-js'
import { ElButton } from 'element-plus'
import { track } from '../../src/track'

jest.mock('@paypal/paypal-js')
jest.mock('../../src/track', () => ({
  ...jest.requireActual('../../src/track'),
  track: jest.fn()
}))

beforeEach(() => track.mockClear())

const CATALOG = {
  data: {
    devices: [{
      code: 'x1',
      name: 'Syncloud X1',
      price: 58200,
      specs: [
        { name: 'Base', value: 'GMKtec NucBox G10' },
        { name: 'CPU', value: 'AMD Ryzen 5 3500U' },
        { name: 'RAM', value: '16 GB DDR4, up to 64 GB' },
        { name: 'Disk', value: '1 TB NVMe SSD' },
        { name: 'Ethernet', value: '2.5 Gb' }
      ],
      options: [
        { code: '1t', name: '1 TB NVMe SSD', extra: 0 }
      ]
    }],
    shipping: 1500,
    currency: 'GBP',
    paypal_client_id: ''
  }
}

function catalog (mock) {
  mock.onGet('/api/device/catalog').reply(200, CATALOG)
  mock.onGet('/api/user').reply(200, { data: { email: 'a@b.c' } })
}

const RouterLinkStub = { props: ['to'], template: '<a :href="to"><slot /></a>' }

const UNKNOWN = 'unknown'

function mountShop (query = {}, replace = jest.fn(), loggedIn = true, push = jest.fn()) {
  return mount(Shop, {
    props: loggedIn === UNKNOWN ? { loggedIn: undefined } : { loggedIn },
    global: {
      components: { ElButton, RouterLink: RouterLinkStub },
      mocks: { $route: { query }, $router: { replace, push } }
    }
  })
}

async function proceed (wrapper) {
  await wrapper.find('[data-testid="shop-continue"]').trigger('click')
  await flushPromises()
}

async function fillAddress (wrapper) {
  await wrapper.find('[data-testid="device-name"]').setValue('A B')
  await wrapper.find('[data-testid="device-address-line"]').setValue('1 Road')
  await wrapper.find('[data-testid="device-city"]').setValue('Town')
  await wrapper.find('[data-testid="device-postcode"]').setValue('X1')
  await wrapper.find('[data-testid="device-country"]').setValue('Germany')
  await flushPromises()
}

async function fill (wrapper) {
  await fillAddress(wrapper)
  await wrapper.find('[data-testid="shop-address-continue"]').trigger('click')
  await flushPromises()
}

test('adds shipping to the one price there is', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()

  expect(wrapper.find('[data-testid="device-price"]').text()).toBe('£582.00')
  expect(wrapper.find('[data-testid="device-shipping"]').text()).toBe('£15.00')
  expect(wrapper.find('[data-testid="device-total"]').text()).toBe('£597.00')
})

test('offers nothing to choose, because the build is fixed', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()

  expect(wrapper.find('[data-testid="device-options"]').exists()).toBe(false)
})

test('opens the photo full size when it is pressed', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(false)

  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')
  expect(wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src'))
    .toBe(wrapper.find('[data-testid="device-photo"]').attributes('src'))

  await wrapper.find('[data-testid="device-lightbox"]').trigger('click')
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(false)
})

test('shows another picture of the device when a thumbnail is pressed', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  const first = wrapper.find('[data-testid="device-photo"]').attributes('src')

  await wrapper.find('[data-testid="device-thumb-2"]').trigger('click')
  const second = wrapper.find('[data-testid="device-photo"]').attributes('src')
  expect(second).not.toBe(first)

  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')
  expect(wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')).toBe(second)
})

test('steps through the pictures with the arrows, and wraps round', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  const shown = () => wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')
  const first = shown()

  await wrapper.find('[data-testid="device-photo-next"]').trigger('click')
  expect(shown()).not.toBe(first)
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(true)

  await wrapper.find('[data-testid="device-photo-previous"]').trigger('click')
  expect(shown()).toBe(first)

  await wrapper.find('[data-testid="device-photo-previous"]').trigger('click')
  expect(shown()).not.toBe(first)

  await wrapper.find('[data-testid="device-photo-next"]').trigger('click')
  expect(shown()).toBe(first)
})

test('steps through the pictures with the keyboard', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  const shown = () => wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')
  const first = shown()

  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowRight' }))
  await wrapper.vm.$nextTick()
  expect(shown()).not.toBe(first)

  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowLeft' }))
  await wrapper.vm.$nextTick()
  expect(shown()).toBe(first)

  window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
  await wrapper.vm.$nextTick()
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(false)
})

test('swipes between the pictures on a touch screen', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  const lightbox = wrapper.find('[data-testid="device-lightbox"]')
  const shown = () => wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')
  const first = shown()

  await lightbox.trigger('touchstart', { changedTouches: [{ clientX: 300 }] })
  await lightbox.trigger('touchend', { changedTouches: [{ clientX: 100 }] })
  const second = shown()
  expect(second).not.toBe(first)

  await lightbox.trigger('touchstart', { changedTouches: [{ clientX: 100 }] })
  await lightbox.trigger('touchend', { changedTouches: [{ clientX: 300 }] })
  expect(shown()).toBe(first)
})

test('a tap that is not a swipe leaves the picture alone', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  const lightbox = wrapper.find('[data-testid="device-lightbox"]')
  const first = wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')

  await lightbox.trigger('touchstart', { changedTouches: [{ clientX: 200 }] })
  await lightbox.trigger('touchend', { changedTouches: [{ clientX: 210 }] })

  expect(wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')).toBe(first)
})

test('the picture itself does not close the full screen view', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  await wrapper.find('[data-testid="device-lightbox-photo"]').trigger('click')
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(true)

  await wrapper.find('[data-testid="device-lightbox-close"]').trigger('click')
  expect(wrapper.find('[data-testid="device-lightbox"]').exists()).toBe(false)
})

test('the picture chosen in the full screen view stays chosen on the card', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')
  await wrapper.find('[data-testid="device-photo-next"]').trigger('click')
  const chosen = wrapper.find('[data-testid="device-lightbox-photo"]').attributes('src')

  await wrapper.find('[data-testid="device-lightbox-close"]').trigger('click')

  expect(wrapper.find('[data-testid="device-photo"]').attributes('src')).toBe(chosen)
})

test('will not let you pay before the address is complete', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  expect(wrapper.find('[data-testid="device-pay"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="shop-address-continue"]').attributes('disabled')).toBeDefined()

  await fill(wrapper)
  expect(wrapper.find('[data-testid="device-pay-stripe"]').attributes('disabled')).toBeUndefined()
})

test('never tells the server what the order costs', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  let posted
  mock.onPost('/api/device/order').reply(config => {
    posted = JSON.parse(config.data)
    return [200, { data: { reference: 'OURREF', url: 'https://checkout.example/session' } }]
  })
  delete window.location
  window.location = { href: '' }

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await wrapper.find('[data-testid="device-pay-stripe"]').trigger('click')
  await flushPromises()

  expect(posted).not.toHaveProperty('total')
  expect(posted).not.toHaveProperty('price')
  expect(posted.device).toBe('x1')
  expect(window.location.href).toBe('https://checkout.example/session')
})

test('confirms the order when the buyer is sent back', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  let completed
  mock.onPost('/api/device/order/complete').reply(config => {
    completed = JSON.parse(config.data)
    return [200, { data: { number: 42 } }]
  })

  const wrapper = mountShop({ reference: 'OURREF' })
  await flushPromises()

  expect(completed).toEqual({ reference: 'OURREF' })
  expect(wrapper.find('[data-testid="device-ordered"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-reference"]').text()).toBe('Order 42')
})

test('shows the product to a visitor who is not signed in', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), false)
  await flushPromises()

  expect(wrapper.find('[data-testid="device-choice"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-total"]').text()).toBe('£597.00')

  expect(wrapper.find('[data-testid="shop-signin-link"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="shop-continue"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="device-pay"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="device-address"]').exists()).toBe(false)
})

test('asks a signed out visitor to sign in rather than to pay', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const push = jest.fn()
  const wrapper = mountShop({}, jest.fn(), false, push)
  await flushPromises()

  await wrapper.find('[data-testid="shop-signin-link"]').trigger('click')
  expect(push).toHaveBeenCalledWith('/login?next=/shop')
})

test('keeps checkout out of the way until the device is chosen', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()

  expect(wrapper.find('[data-testid="device-address"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="device-pay"]').exists()).toBe(false)

  await proceed(wrapper)

  expect(wrapper.find('[data-testid="device-address"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="shop-chosen"]').text()).toContain('£597.00')

  await wrapper.find('[data-testid="shop-change"]').trigger('click')
  expect(wrapper.find('[data-testid="device-address"]').exists()).toBe(false)
})

test('every address field keeps a label that survives typing', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)

  for (const [id, text] of [
    ['device-name', 'Full name'],
    ['device-address-line', 'Address'],
    ['device-city', 'City'],
    ['device-postcode', 'Postcode'],
    ['device-country', 'Country']
  ]) {
    const input = wrapper.find(`[data-testid="${id}"]`)
    expect(input.exists()).toBe(true)
    const label = wrapper.find(`label[for="${id}"]`)
    expect(label.exists()).toBe(true)
    expect(label.text()).toBe(text)
  }

  await fillAddress(wrapper)
  expect(wrapper.find('label[for="device-city"]').text()).toBe('City')
})

test('says what is in the box without pushing the choice off the screen', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()

  const spec = wrapper.find('[data-testid="device-spec"]')
  expect(spec.exists()).toBe(true)
  expect(spec.element.open).toBe(false)
  expect(spec.text()).toContain('GMKtec NucBox G10')
  expect(spec.text()).toContain('second disk slot')
})

test('pressing PayPal says something is happening', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/catalog').reply(200, {
    data: { ...CATALOG.data, paypal_client_id: 'test-client-id' }
  })
  mock.onGet('/api/user').reply(200, { data: { email: 'a@b.c' } })
  mock.onPost('/api/device/order').reply(200,
    { data: { number: 7, reference: 'OURREF', provider_reference: 'PP1' } })

  let buttons
  loadScript.mockResolvedValue({
    Buttons: options => { buttons = options; return { render: () => {} } }
  })

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await flushPromises()

  expect(wrapper.find('[data-testid="device-pay-busy"]').exists()).toBe(false)

  await buttons.onClick({}, {})
  await flushPromises()
  expect(wrapper.find('[data-testid="device-pay-busy"]').exists()).toBe(true)

  wrapper.vm.busy = ''
  await flushPromises()
  buttons.createOrder()
  await flushPromises()
  expect(wrapper.find('[data-testid="device-pay-busy"]').exists()).toBe(true)

  await buttons.onCancel()
  await flushPromises()
  expect(wrapper.find('[data-testid="device-pay-busy"]').exists()).toBe(false)
})

test('keeps the full specification behind the expander', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()

  expect(wrapper.find('[data-testid="device-highlights"]').exists()).toBe(false)

  const table = wrapper.find('[data-testid="device-spec-table"]')
  expect(table.text()).toContain('AMD Ryzen 5 3500U')
  expect(table.text()).toContain('2.5 Gb')
  expect(wrapper.find('[data-testid="device-spec"]').element.open).toBeFalsy()
})

test('shows progress while the PayPal button itself is loading', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/catalog').reply(200, {
    data: { ...CATALOG.data, paypal_client_id: 'test-client-id' }
  })
  mock.onGet('/api/user').reply(200, { data: { email: 'a@b.c' } })

  let finishRender
  loadScript.mockResolvedValue({
    Buttons: () => ({ render: () => new Promise(resolve => { finishRender = resolve }) })
  })

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await flushPromises()

  expect(wrapper.find('[data-testid="device-paypal-loading"]').exists()).toBe(true)

  finishRender()
  await flushPromises()

  expect(wrapper.find('[data-testid="device-paypal-loading"]').exists()).toBe(false)
})

const names = () => track.mock.calls.map(call => call[0])

test('counts the shop view as soon as the page opens', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  mountShop()
  await flushPromises()

  expect(names()).toContain('shop.view')
})

test('counts a signed out visitor once, not on every render', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  mountShop({}, jest.fn(), false)
  await flushPromises()
  await flushPromises()

  expect(names().filter(name => name === 'shop.signedout')).toHaveLength(1)
})

test('does not call a visitor signed out while the session is still unknown', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), UNKNOWN)
  await flushPromises()

  expect(names()).not.toContain('shop.signedout')
  expect(wrapper.find('[data-testid="shop-waiting"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="shop-waiting"]').attributes('disabled')).toBeDefined()
  expect(wrapper.find('[data-testid="shop-signin-link"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="shop-continue"]').exists()).toBe(false)
})

test('counts the visitor signed out once the session comes back empty', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), UNKNOWN)
  await flushPromises()
  await wrapper.setProps({ loggedIn: false })
  await flushPromises()

  expect(names()).toContain('shop.signedout')
  expect(wrapper.find('[data-testid="shop-signin-link"]').exists()).toBe(true)
})

test('records every step a buyer takes on the way to paying by card', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  mock.onPost('/api/device/order').reply(200,
    { data: { reference: 'OURREF', url: 'https://checkout.example/session' } })
  delete window.location
  window.location = { href: '' }

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await wrapper.find('[data-testid="device-pay-stripe"]').trigger('click')
  await flushPromises()

  expect(names()).toEqual([
    'shop.view',
    'shop.continue',
    'shop.address',
    'shop.pay.view',
    'shop.pay.stripe'
  ])
})

test('counts the press on sign in to order', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), false)
  await flushPromises()
  await wrapper.find('[data-testid="shop-signin-link"]').trigger('click')

  expect(names()).toContain('shop.signin')
})

test('counts the specification being opened, not merely rendered', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  expect(names()).not.toContain('shop.spec')

  const spec = wrapper.find('[data-testid="device-spec"]')
  spec.element.open = true
  await spec.trigger('toggle')

  expect(names()).toContain('shop.spec')
})

test('counts the photo being opened', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  expect(names()).not.toContain('shop.photo')

  await wrapper.find('[data-testid="device-photo-open"]').trigger('click')

  expect(names()).toContain('shop.photo')
})

test('counts a completed order', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  mock.onPost('/api/device/order/complete').reply(200, { data: { number: 42 } })

  mountShop({ reference: 'OURREF' })
  await flushPromises()

  expect(names()).toContain('shop.ordered')
})

test('counts a failure and says so instead of only logging it', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/catalog').reply(200, {
    data: { ...CATALOG.data, paypal_client_id: 'test-client-id' }
  })
  mock.onGet('/api/user').reply(200, { data: { email: 'a@b.c' } })
  loadScript.mockRejectedValue(new Error('blocked'))
  const logged = jest.spyOn(console, 'error').mockImplementation(() => {})

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await flushPromises()

  expect(names()).toContain('shop.paypal.failed')
  expect(wrapper.find('[data-testid="device-error"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-paypal-loading"]').exists()).toBe(false)
  logged.mockRestore()
})

test('counts a server error on the way to checkout', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  mock.onPost('/api/device/order').reply(500, { message: 'no' })

  const wrapper = mountShop()
  await flushPromises()
  await proceed(wrapper)
  await fill(wrapper)
  await wrapper.find('[data-testid="device-pay-stripe"]').trigger('click')
  await flushPromises()

  expect(names()).toContain('shop.error')
})

test('treats a missing session prop as unknown rather than signed out', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mount(Shop, {
    global: {
      components: { ElButton, RouterLink: RouterLinkStub },
      mocks: { $route: { query: {} }, $router: { replace: jest.fn(), push: jest.fn() } }
    }
  })
  await flushPromises()

  expect(wrapper.find('[data-testid="shop-waiting"]').exists()).toBe(true)
  expect(names()).not.toContain('shop.signedout')
})
