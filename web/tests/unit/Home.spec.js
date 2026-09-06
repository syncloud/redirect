import { mount } from '@vue/test-utils'
import Home from '../../src/views/Home.vue'

const RouterLinkStub = { props: ['to'], template: '<a :href="to"><slot /></a>' }

function mountHome (loggedIn) {
  return mount(Home, {
    props: { loggedIn, checkUserSession: jest.fn() },
    global: {
      components: { RouterLink: RouterLinkStub },
      stubs: { Devices: true }
    }
  })
}

test('a visitor with no account is told what this is and where to go', () => {
  const wrapper = mountHome(false)

  expect(wrapper.find('[data-testid="home-intro"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="home-login"]').attributes('href')).toBe('/login')
  expect(wrapper.find('[data-testid="home-shop"]').attributes('href')).toBe('/shop')
  expect(wrapper.find('[data-testid="home-build"]').attributes('href'))
    .toBe('https://syncloud.org/setup')
})

test('a signed in visitor gets their devices, not the introduction', () => {
  const wrapper = mountHome(true)

  expect(wrapper.find('[data-testid="home-intro"]').exists()).toBe(false)
  expect(wrapper.findComponent({ name: 'Devices' }).exists()).toBe(true)
})

test('nothing is claimed before the session answers', () => {
  const wrapper = mountHome(undefined)

  expect(wrapper.find('[data-testid="home-intro"]').exists()).toBe(false)
  expect(wrapper.findComponent({ name: 'Devices' }).exists()).toBe(false)
})
