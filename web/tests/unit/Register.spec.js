import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Register from '../../src/views/Register.vue'

async function register (query) {
  let posted
  const mock = new MockAdapter(axios)
  mock.onPost('/api/user/create').reply(function (config) {
    posted = JSON.parse(config.data)
    return [200, { data: { success: true } }]
  })

  const wrapper = mount(Register, {
    attachTo: document.body,
    global: {
      components: { RouterLink: RouterLinkStub },
      mocks: {
        $route: { path: '/register', query },
        $router: { push: jest.fn() }
      }
    }
  })

  await flushPromises()
  await wrapper.find('#register_email').setValue('user@example.com')
  await wrapper.find('#register_password').setValue('password123')
  await wrapper.find('#btnregister').trigger('click')
  await flushPromises()

  mock.restore()
  wrapper.unmount()
  return posted
}

test('tells the user they will need their own hardware next', async () => {
  const wrapper = mount(Register, {
    global: {
      components: { RouterLink: RouterLinkStub },
      mocks: { $route: { path: '/register', query: {} }, $router: { push: jest.fn() } }
    }
  })
  const note = wrapper.find('[data-testid="register-next-steps"]')
  expect(note.exists()).toBe(true)
  expect(note.text()).toContain('Raspberry Pi')
  expect(wrapper.find('[data-testid="register-setup-link"]').attributes('href'))
    .toBe('https://syncloud.org/setup')
  wrapper.unmount()
})

test('forwards gclid from the query string', async () => {
  const posted = await register({ gclid: 'Cj0KCQjw-abc_123' })
  expect(posted.email).toBe('user@example.com')
  expect(posted.gclid).toBe('Cj0KCQjw-abc_123')
})

test('omits gclid when the query string has none', async () => {
  const posted = await register({})
  expect(posted.email).toBe('user@example.com')
  expect(posted.gclid).toBeUndefined()
})

test('does not treat other query parameters as gclid', async () => {
  const posted = await register({ utm_source: 'newsletter' })
  expect(posted.gclid).toBeUndefined()
})

test('omits an empty gclid', async () => {
  const posted = await register({ gclid: '' })
  expect(posted.gclid).toBeUndefined()
})
