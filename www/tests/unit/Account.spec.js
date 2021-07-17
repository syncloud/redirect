import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '@/views/Account'

jest.setTimeout(30000)

test('Notifications unsubscribe', async () => {
  let unsubscribed

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200,
    {
      data: {
        active: true,
        email: 'test@example.com',
        unsubscribed: false,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/notification/unsubscribe').reply(function (config) {
    unsubscribed = true
    return [200, { success: true }]
  })

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

  expect(unsubscribed).toBe(true)
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
        unsubscribed: true,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/notification/subscribe').reply(function (config) {
    subscribed = true
    return [200, { success: true }]
  })

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
        unsubscribed: false,
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

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      props: {
        onLogout: jest.fn()
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

