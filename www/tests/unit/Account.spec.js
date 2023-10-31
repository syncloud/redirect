import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '../../src/views/Account.vue'

test('Notifications disable', async () => {
  let notificationsEnabled

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200,
    {
      data: {
        active: true,
        email: 'test@example.com',
        notification_enabled: true,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/notification/disable').reply(function (config) {
    notificationsEnabled = false
    return [200, { success: true }]
  })

  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      global: {
        components: {
          RouterLink: RouterLinkStub
        },
        stubs: {
          Confirmation: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          }
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#chk_email').trigger('click')
  await wrapper.find('#save').trigger('click')

  await flushPromises()

  expect(notificationsEnabled).toBe(false)
  wrapper.unmount()
})

test('Notifications subscribe', async () => {
  let subscribed

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200,
    {
      data: {
        active: true,
        email: 'test@example.com',
        notification_enabled: false,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/notification/enable').reply(function (config) {
    subscribed = true
    return [200, { success: true }]
  })

  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      global: {
        components: {
          RouterLink: RouterLinkStub
        },
        stubs: {
          Confirmation: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          }
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#chk_email').trigger('click')
  await wrapper.find('#save').trigger('click')

  await flushPromises()

  expect(subscribed).toBe(true)
  wrapper.unmount()
})

test('Delete', async () => {
  let deleted

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200,
    {
      data: {
        active: true,
        email: 'test@example.com',
        notification_enabled: true,
        update_token: '0a'
      }
    }
  )

  mock.onDelete('/api/user').reply(function (_) {
    deleted = true
    return [200, { success: true }]
  })

  mock.onPost('/api/logout').reply(function (_) {
    return [200, { success: true }]
  })
  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      props: {
        checkUserSession: jest.fn()
      },
      global: {
        components: {
          RouterLink: RouterLinkStub
        },
        stubs: {
          Confirmation: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          }
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#delete').trigger('click')
  await wrapper.find('#delete_confirmation').trigger('confirm')

  await flushPromises()

  expect(deleted).toBe(true)
  wrapper.unmount()
})

