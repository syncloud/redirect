import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Login from '@/views/Login'

jest.setTimeout(30000)

test('Login success', async () => {
  const showError = jest.fn()
  const mockRouter = { push: jest.fn() }

  let email
  let password

  const mock = new MockAdapter(axios)
  mock.onPost('/api/user/login').reply(function (config) {
    const request = JSON.parse(config.data)
    email = request.email
    password = request.password
    return [200, {
      data: {
        success: true
      }
    }]
  })

  const wrapper = mount(Login,
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
          Error: {
            template: '<span/>',
            methods: {
              showAxios: showError
            }
          }

        },
        mocks: {
          $route: { path: '/login' },
          $router: mockRouter
        }
      }
    }
  )

  await flushPromises()

  await wrapper.find('#email').setValue('username')
  await wrapper.find('#password').setValue('password')
  await wrapper.find('#submit').trigger('click')

  await flushPromises()

  expect(showError).toHaveBeenCalledTimes(0)
  expect(email).toBe('username')
  expect(password).toBe('password')
  expect(mockRouter.push).toHaveBeenCalledWith('/')

  wrapper.unmount()
})

test('Login failed', async () => {
  const showError = jest.fn()
  const mockRouter = { push: jest.fn() }

  let email
  let password

  const mock = new MockAdapter(axios)
  mock.onPost('/api/user/login').reply(function (config) {
    const request = JSON.parse(config.data)
    email = request.email
    password = request.password
    return [400, { message: 'login failed' }, { 'Content-Type': 'text/plain' }]
  })

  const wrapper = mount(Login,
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
          Error: {
            template: '<span/>',
            methods: {
              showAxios: showError
            }
          }

        },
        mocks: {
          $route: { path: '/login' },
          $router: mockRouter
        }
      }
    }
  )

  await flushPromises()

  await wrapper.find('#email').setValue('username')
  await wrapper.find('#password').setValue('password')
  await wrapper.find('#submit').trigger('click')

  await flushPromises()

  expect(showError).toHaveBeenCalledTimes(1)
  expect(email).toBe('username')
  expect(password).toBe('password')
  expect(mockRouter.push).toHaveBeenCalledTimes(0)

  wrapper.unmount()
})
