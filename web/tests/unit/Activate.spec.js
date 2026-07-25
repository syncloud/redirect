import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Activate from '../../src/views/Activate.vue'

test('Activate success', async () => {
  let token

  const mock = new MockAdapter(axios)
  mock.onPost('api/user/activate').reply(function (config) {
    token = JSON.parse(config.data).token
    return [200, { data: 'OK' }]
  })

  const wrapper = mount(Activate,
    {
      attachTo: document.body,
      global: {
        mocks: {
          $route: { query: { token: '123' } }
        }
      }

    }
  )

  await flushPromises()

  expect(token).toBe('123')
  expect(wrapper.get('#activated').text()).toBe('OK')
  wrapper.unmount()
})

test('Activate no token', async () => {
  const wrapper = mount(Activate,
    {
      attachTo: document.body,
      global: {
        mocks: {
          $route: { query: {} }
        }
      }

    }
  )

  await flushPromises()

  expect(wrapper.get('#activated').text()).toBe('Unknown token')
  wrapper.unmount()
})

test('Activate error', async () => {
  const mock = new MockAdapter(axios)
  mock.onPost('api/user/activate').reply(function (config) {
    return [400, { message: 'error' }]
  })

  const wrapper = mount(Activate,
    {
      attachTo: document.body,
      global: {
        mocks: {
          $route: { query: { token: '123' } }
        }
      }

    }
  )

  await flushPromises()

  expect(wrapper.get('#activated').text()).toBe('error')
  wrapper.unmount()
})

test('Activate unknown error', async () => {
  const mockRouter = { push: jest.fn() }
  const mock = new MockAdapter(axios)
  mock.onPost('api/user/activate').reply(function (config) {
    return [400, { }]
  })

  const wrapper = mount(Activate,
    {
      attachTo: document.body,
      global: {
        mocks: {
          $route: { query: { token: '123' } },
          $router: mockRouter
        }
      }

    }
  )

  await flushPromises()

  expect(mockRouter.push).toHaveBeenCalledWith('/error')
  wrapper.unmount()
})
