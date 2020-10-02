import { shallowMount } from '@vue/test-utils'
import Menu from '@/components/Menu.vue'

describe('Menu.vue', () => {
  it('renders email when passed', () => {
    const email = 'test@example.com'
    const wrapper = shallowMount(Menu, {
      props: { email }
    })
    expect(wrapper.text()).toMatch(email)
  })
})
