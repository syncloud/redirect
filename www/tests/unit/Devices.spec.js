import { shallowMount } from '@vue/test-utils'
import Devices from '@/views/Devices'

describe('Devices.vue', () => {
  it('timestamp format', () => {
    const wrapper = shallowMount(Devices)

    expect(wrapper.vm.timestamp('Sun, 02 Nov 2020 22:07:36 GMT', new Date(2020, 10, 1))).toMatch('Nov 2, 2020')

    expect(wrapper.vm.timestamp('Sun, 02 Nov 2020 22:07:36 GMT', new Date(2020, 10, 2))).toMatch('Today 22:07')

  })
})
