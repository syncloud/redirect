import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Shop from '../../src/views/Shop.vue'
import { ElButton } from 'element-plus'

const CATALOG = {
  data: {
    devices: [{
      code: 'h4',
      name: 'Syncloud H4',
      price: 22900,
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

async function fill (wrapper) {
  await wrapper.find('[data-testid="device-name"]').setValue('A B')
  await wrapper.find('[data-testid="device-address-line"]').setValue('1 Road')
  await wrapper.find('[data-testid="device-city"]').setValue('Town')
  await wrapper.find('[data-testid="device-postcode"]').setValue('X1')
  await wrapper.find('[data-testid="device-country"]').setValue('Germany')
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
  expect(wrapper.find('[data-testid="device-pay-stripe"]').attributes('disabled')).toBeDefined()

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

  expect(wrapper.find('[data-testid="shop-signin-link"]').attributes('href')).toBe('/login')
})
