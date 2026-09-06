import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Orders from '../../src/views/Orders.vue'

jest.setTimeout(30000)

const MINE = [{
  number: 1,
  device: 'Syncloud H4',
  option: '1 TB SSD',
  total: '£322.00',
  status: 'ordered',
  ordered: '2026-09-05'
}]

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
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/unfinished').reply(200, { data: [] })

  const wrapper = mountOrders()
  await flushPromises()

  expect(wrapper.findAll('[data-testid="order"]')).toHaveLength(1)
  expect(wrapper.find('[data-testid="order-total"]').text()).toBe('£322.00')
  expect(wrapper.find('[data-testid="order-status"]').text()).toBe('Ordered')
  expect(wrapper.find('[data-testid="order-account"]').exists()).toBe(false)
})

test('a buyer with nothing ordered is pointed at the shop', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders').reply(200, { data: [] })
  mock.onGet('/api/device/orders/unfinished').reply(200, { data: [] })

  const wrapper = mountOrders()
  await flushPromises()

  expect(wrapper.find('[data-testid="orders-empty"]').exists()).toBe(true)
  expect(wrapper.findAll('[data-testid="order"]')).toHaveLength(0)
})

test('a signed out visitor is sent to log in', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders').reply(401, { message: 'Unauthorized' })
  const push = jest.fn()

  mountOrders(push)
  await flushPromises()

  expect(push).toHaveBeenCalledWith('/login?next=/orders')
})

test('the buyer page never asks for every order', async () => {
  const mock = new MockAdapter(axios)
  let askedForAll = false
  mock.onGet('/api/device/orders').reply(200, { data: MINE })
  mock.onGet('/api/device/orders/unfinished').reply(200, { data: [] })
  mock.onGet('/api/device/orders/all').reply(() => {
    askedForAll = true
    return [200, { data: [] }]
  })

  mountOrders()
  await flushPromises()

  expect(askedForAll).toBe(false)
})

test('an unpaid order is offered back to the buyer to finish', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders').reply(200, { data: [] })
  mock.onGet('/api/device/orders/unfinished').reply(200, {
    data: [{ number: 9, device: 'Syncloud H4', option: '1 TB SSD', total: '£322.00' }]
  })

  const wrapper = mountOrders()
  await flushPromises()

  const finish = wrapper.findAllComponents(RouterLinkStub)
    .find(link => link.props().to === '/shop?order=9')
  expect(finish).toBeTruthy()
  expect(finish.text()).toContain('Not paid')
  expect(wrapper.find('[data-testid="orders-empty"]').exists()).toBe(false)
})
