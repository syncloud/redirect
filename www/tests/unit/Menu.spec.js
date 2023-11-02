import { mount, RouterLinkStub } from '@vue/test-utils'
import Menu from '../../src/components/Menu.vue'

test('Menu.vue', async () => {
  const mockRouter = { push: jest.fn() }

  const email = 'test@example.com'
  const wrapper = mount(Menu, {
    attachTo: document.body,
    props: { email },
    global: {
      components: {
        RouterLink: RouterLinkStub
      },
      mocks: {
        $route: { path: '' },
        $router: mockRouter
      }
    }
  })

  expect(wrapper.text()).toMatch(email)
  wrapper.unmount()
})
