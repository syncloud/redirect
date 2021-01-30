import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import PasswordForgot from '@/views/PasswordForgot'

jest.setTimeout(30000)

test('Request reset', async () => {
  const mockRouter = { push: jest.fn() }
  let email
  const mock = new MockAdapter(axios)
  mock.onPost('/api/user/reset_password').reply(function (config) {
    email = JSON.parse(config.data).email
    return [200, { success: true }]
  })

  const wrapper = mount(PasswordForgot,
    {
      attachTo: document.body,
      global: {
        mocks: {
          $route: { path: '' },
          $router: mockRouter
        }
      }
    }
  )

  await flushPromises()

  await wrapper.find('#email').setValue('test@example.com')
  await wrapper.find('#send').trigger('click')

  await flushPromises()

  expect(email).toBe('test@example.com')
  wrapper.unmount()
})
