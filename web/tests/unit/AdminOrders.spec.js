import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import AdminOrders from '../../src/views/AdminOrders.vue'

jest.setTimeout(30000)

const ALL = [
  {
    reference: 'ref-1',
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

  const table = wrapper.find('[data-testid="orders-admin-table"]')
  expect(table.text()).toContain('me@example.com')
  expect(table.text()).toContain('other@example.com')
  expect(table.text()).toContain('2 Road')
})

test('every admin cell carries the label the mobile card view shows', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders/all').reply(200, { data: ALL })

  const wrapper = mountAdmin()
  await flushPromises()

  const headers = wrapper.findAll('[data-testid="orders-admin-table"] thead th').map(th => th.text())
  const labels = wrapper.findAll('[data-testid="admin-order-ref-1"] td').map(td => td.attributes('data-label'))

  expect(labels).toEqual(headers)
})

test('a non admin refused by the server is sent back to their own orders', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/device/orders/all').reply(403, { message: 'Forbidden' })
  const push = jest.fn()

  mountAdmin(push)
  await flushPromises()

  expect(push).toHaveBeenCalledWith('/orders')
})
