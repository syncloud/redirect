import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '@/views/Account'

jest.setTimeout(30000)

test('Notifications', async () => {
  let subscribed

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user/get').reply(200,
    {
      user: {
        active: true,
        domains: [],
        email: 'test@example.com',
        unsubscribed: false,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/subscription').reply(function (config) {
    subscribed = JSON.parse(config.data).subscribed
    return [200, { success: true }]
  })

  const wrapper = mount(Account,
    {
      attachTo: document.body
    }
  )

  await flushPromises()

  await wrapper.find('#chk_email').trigger('toggle')
  await wrapper.find('#save').trigger('click')

  await flushPromises()

  expect(subscribed).toBe(true)
  wrapper.unmount()
})

test('Delete', async () => {
  let deleted

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user/get').reply(200,
    {
      user: {
        active: true,
        domains: [],
        email: 'test@example.com',
        unsubscribed: false,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/user_delete').reply(function (config) {
    deleted = true
    return [200, { success: true }]
  })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      props: {
        onLogout: jest.fn()
      }
    }
  )

  await flushPromises()

  await wrapper.find('#delete').trigger('click')
  await wrapper.find('#delete-confirm').trigger('click')

  await flushPromises()

  expect(deleted).toBe(true)
  wrapper.unmount()
})
