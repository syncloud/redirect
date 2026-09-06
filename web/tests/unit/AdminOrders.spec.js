import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import AdminOrders from '../../src/views/AdminOrders.vue'

jest.setTimeout(30000)

const ALL = [
  {
    number: 1,
    device: 'Syncloud H4',
    option: '1 TB SSD',
    total: '£322.00',
    status: 'ordered',
    ordered: '2026-09-05',
    email: 'me@example.com',
    name: 'Ada',
    address: '1 Road',
    city: 'London',
    postcode: 'E1',
    country: 'UK'
  },
  {
    number: 2,
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

function mountAdmin (push = jest.fn()) {
  return mount(AdminOrders, {
    global: {
      components: { RouterLink: RouterLinkStub },
      mocks: { $router: { push } }
    }
  })
}

test('an admin sees every order with who ordered it', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })

  const wrapper = mountAdmin()
  await flushPromises()

  const cards = wrapper.findAll('[data-testid="order"]')
  expect(cards).toHaveLength(2)
  expect(cards[0].text()).toContain('me@example.com')
  expect(cards[1].text()).toContain('other@example.com')
  expect(cards[1].text()).toContain('2 Road')
})

test('every order card opens that order', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })

  const wrapper = mountAdmin()
  await flushPromises()

  const links = wrapper.findAllComponents(RouterLinkStub).map(link => link.props().to)
  expect(links).toEqual(['/orders/1', '/orders/2'])
})

test('a non admin refused by the server is sent back to their own orders', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders/all').reply(403, { message: 'Forbidden' })
  const push = jest.fn()

  mountAdmin(push)
  await flushPromises()

  expect(push).toHaveBeenCalledWith('/orders')
})
