import { mount, RouterLinkStub } from '@vue/test-utils'
import CustomMenu from '../../src/components/CustomMenu.vue'

test('CustomMenu.vue', async () => {
  const mockRouter = { push: jest.fn() }

  const email = 'test@example.com'
  const wrapper = mount(CustomMenu, {
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
