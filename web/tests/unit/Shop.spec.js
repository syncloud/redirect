import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Shop from '../../src/views/Shop.vue'
import { loadScript } from '@paypal/paypal-js'
import { ElButton } from 'element-plus'

jest.mock('@paypal/paypal-js')

const CATALOG = {
  data: {
    devices: [{
      code: 'h4',
      name: 'Syncloud H4',
      price: 22900,
      specs: [
        { name: 'CPU', value: 'Amlogic S905X3, Cortex-A55' },
        { name: 'RAM', value: '4 GB DDR4' },
        { name: 'Drive bays', value: '2 x SATA, 3.5 or 2.5 inch' },
        { name: 'Ethernet', value: '1 Gb' },
        { name: 'USB', value: 'USB 2.0 x 1' }
      ],
      options: [
        { code: '120', name: '120 GB SSD', extra: 0 },
        { code: '2tx2', name: '2 TB SSD x 2', extra: 43000 }
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

function mountShop (query = {}, replace = jest.fn(), loggedIn = true) {
  return mount(Shop, {
    props: { loggedIn },
    global: {
      components: { ElButton, RouterLink: RouterLinkStub },
      mocks: { $route: { query }, $router: { replace } }
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

test('adds shipping to whichever option is chosen', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop()
  await flushPromises()
  expect(wrapper.find('[data-testid="device-total"]').text()).toBe('£244.00')

  await wrapper.find('[data-testid="device-option-2tx2"]').trigger('click')
  expect(wrapper.find('[data-testid="device-total"]').text()).toBe('£674.00')
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
  expect(posted.device).toBe('h4')
  expect(window.location.href).toBe('https://checkout.example/session')
})

test('confirms the order when the buyer is sent back', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)
  let completed
  mock.onPost('/api/device/order/complete').reply(config => {
    completed = JSON.parse(config.data)
    return [200, { data: 'ordered' }]
  })

  const wrapper = mountShop({ reference: 'OURREF' })
  await flushPromises()

  expect(completed).toEqual({ reference: 'OURREF' })
  expect(wrapper.find('[data-testid="device-ordered"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-reference"]').text()).toContain('OURREF')
})

test('shows the product to a visitor who is not signed in', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), false)
  await flushPromises()

  expect(wrapper.find('[data-testid="device-choice"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-total"]').text()).toBe('£244.00')

  expect(wrapper.find('[data-testid="shop-signin"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="device-pay"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="device-address"]').exists()).toBe(false)
})

test('asks a signed out visitor to sign in rather than to pay', async () => {
  const mock = new MockAdapter(axios)
  catalog(mock)

  const wrapper = mountShop({}, jest.fn(), false)
  await flushPromises()

  expect(wrapper.find('[data-testid="shop-signin-link"]').attributes('href')).toBe('/login?next=/shop')
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
  expect(wrapper.find('[data-testid="shop-chosen"]').text()).toContain('£244.00')

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
  expect(spec.text()).toContain('Odroid HC4')
  expect(spec.text()).toContain('second bay')
})

test('pressing PayPal says something is happening', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/catalog').reply(200, {
    data: { ...CATALOG.data, paypal_client_id: 'test-client-id' }
  })
  mock.onGet('/api/user').reply(200, { data: { email: 'a@b.c' } })
  mock.onPost('/api/device/order').reply(200,
    { data: { reference: 'OURREF', provider_reference: 'PP1' } })

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
  expect(table.text()).toContain('Amlogic S905X3')
  expect(table.text()).toContain('USB 2.0 x 1')
  expect(wrapper.find('[data-testid="device-spec"]').element.open).toBeFalsy()
})
