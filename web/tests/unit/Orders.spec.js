import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Orders from '../../src/views/Orders.vue'

jest.setTimeout(30000)

const MINE = [{
  reference: 'ref-1',
  device: 'Syncloud H4',
  option: '1 TB SSD',
  total: '£322.00',
  status: 'ordered',
  ordered: '2026-09-05'
}]

const ALL = [
  { ...MINE[0], email: 'me@example.com', name: 'Ada', address: '1 Road', city: 'London', postcode: 'E1', country: 'UK' },
  {
    reference: 'ref-2',
    device: 'Syncloud H4',
    option: '120 GB SSD',
    total: '£244.00',
    status: 'sent',
    ordered: '2026-08-30',
    email: 'other@example.com',
    name: 'Grace',
    address: '2 Road',
    city: 'Manchester',
    postcode: 'M1',
    country: 'UK'
  }
]

function mountOrders (push = jest.fn()) {
  return mount(Orders, {
    global: {
      components: { RouterLink: RouterLinkStub },
      mocks: { $router: { push } }
    }
  })
}

test('a buyer sees their own orders and no admin table', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: false } })
  mock.onGet('/api/device/orders').reply(200, { data: MINE })

  const wrapper = mountOrders()
  await flushPromises()

  expect(wrapper.findAll('[data-testid="order"]')).toHaveLength(1)
  expect(wrapper.find('[data-testid="order-total"]').text()).toBe('£322.00')
  expect(wrapper.find('[data-testid="order-status"]').text()).toBe('Ordered')
  expect(wrapper.find('[data-testid="orders-admin"]').exists()).toBe(false)
})

test('a buyer with nothing ordered is pointed at the shop', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: false } })
  mock.onGet('/api/device/orders').reply(200, { data: [] })

  const wrapper = mountOrders()
  await flushPromises()

  expect(wrapper.find('[data-testid="orders-empty"]').exists()).toBe(true)
  expect(wrapper.findAll('[data-testid="order"]')).toHaveLength(0)
})

test('the admin table is never asked for unless the account is an admin', async () => {
  const mock = new MockAdapter(axios)
  let askedForAll = false
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: false } })
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/all').reply(() => {
    askedForAll = true
    return [200, { data: ALL }]
  })

  mountOrders()
  await flushPromises()

  expect(askedForAll).toBe(false)
})

test('an admin sees every order with who ordered it', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: true } })
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })

  const wrapper = mountOrders()
  await flushPromises()

  const table = wrapper.find('[data-testid="orders-admin-table"]')
  expect(table.exists()).toBe(true)
  expect(table.text()).toContain('me@example.com')
  expect(table.text()).toContain('other@example.com')
  expect(table.text()).toContain('2 Road')
})

test('an admin marking an order sent sends the change and shows it', async () => {
  const mock = new MockAdapter(axios)
  let posted = null
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: true } })
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })
  mock.onPost('/api/device/order/status').reply(config => {
    posted = JSON.parse(config.data)
    return [200, { message: 'OK' }]
  })

  const wrapper = mountOrders()
  await flushPromises()

  const select = wrapper.find('[data-testid="admin-status-ref-1"]')
  select.element.value = 'sent'
  await select.trigger('change')
  await flushPromises()

  expect(posted).toEqual({ reference: 'ref-1', status: 'sent' })
  expect(wrapper.find('[data-testid="order-status"]').text()).toBe('Sent')
})

test('a signed out visitor is sent to log in', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(401, { message: 'Unauthorized' })
  const push = jest.fn()

  mountOrders(push)
  await flushPromises()

  expect(push).toHaveBeenCalledWith('/login?next=/orders')
})

test('every admin cell carries the label the mobile card view shows', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, { data: { email: 'me@example.com', admin: true } })
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })

  const wrapper = mountOrders()
  await flushPromises()

  const headers = wrapper.findAll('[data-testid="orders-admin-table"] thead th')
    .map(th => th.text())
  const labels = wrapper.findAll('[data-testid="admin-order-ref-1"] td')
    .map(td => td.attributes('data-label'))

  expect(labels).toEqual(headers)
})
