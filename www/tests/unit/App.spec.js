import { mount } from '@vue/test-utils'
import App from '@/App'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import { h } from 'vue'

test('index to login (not logged in)', async () => {
  const mockRouter = { push: jest.fn() }

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(400,
    { }
  )

  mount(App, {
    global: {
      components: {
        RouterView: { render () { return h('div') } }
      },
      stubs: {
        Menu: true
      },
      mocks: {
        $route: { path: '/' },
        $router: mockRouter
      }
    }
  })

  await flushPromises()
  expect(mockRouter.push).toHaveBeenCalledWith('/login')
})

test('index stay (logged in)', async () => {
  const mockRouter = { push: jest.fn() }

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

  mount(App, {
    global: {
      components: {
        RouterView: { render () { return h('div') } }
      },
      stubs: {
        Menu: true
      },
      mocks: {
        $route: { path: '/' },
        $router: mockRouter
      }
    }
  })

  await flushPromises()
  expect(mockRouter.push).toHaveBeenCalledTimes(0)
})

test('login to index (logged in)', async () => {
  const mockRouter = { push: jest.fn() }

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

  mount(App, {
    global: {
      components: {
        RouterView: { render () { return h('div') } }
      },
      stubs: {
        Menu: true
      },
      mocks: {
        $route: { path: '/login' },
        $router: mockRouter
      }
    }
  })

  await flushPromises()
  expect(mockRouter.push).toHaveBeenCalledWith('/')
})
