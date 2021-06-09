import { mount } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Devices from '@/views/Devices'

jest.setTimeout(30000)

test('Show devices', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/domains').reply(200,
    {
      data: [
        {
          device_mac_address: '111',
          device_name: 'syncloud',
          device_title: 'Syncloud',
          dkim_key: 'dkim',
          ip: '111.111.111.111',
          ipv6: null,
          last_update: 'Mon, 19 Oct 2020 19:31:49 GMT',
          local_ip: '192.168.1.1',
          map_local_address: false,
          platform_version: '2',
          web_local_port: 443,
          web_port: 443,
          web_protocol: 'https',
          name: 'test.example.com'
        },
        {
          device_mac_address: '00:11:22:33:44:ff',
          device_name: 'odroid-xu3and4',
          device_title: 'ODROID-XU',
          dkim_key: null,
          ip: '111.222.333.444',
          ipv6: '[::1]',
          last_update: 'Mon, 19 Oct 2020 18:51:18 GMT',
          local_ip: '192.168.1.2',
          map_local_address: false,
          platform_version: '2',
          web_local_port: 443,
          web_port: 10001,
          web_protocol: 'https',
          name: 'test1.example.com'
        }
      ]
    }
  )

  const wrapper = mount(Devices,
    {
      attachTo: document.body,
      props: {
        onLogin: jest.fn()
      }
    }
  )

  await flushPromises()

  const deviceTitles = await wrapper.findAll('#title')
  expect(deviceTitles[0].text()).toBe('Syncloud')
  expect(deviceTitles[1].text()).toBe('ODROID-XU')
  const deviceNames = await wrapper.findAll('#name')
  expect(deviceNames[0].text()).toBe('test.example.com')
  expect(deviceNames[1].text()).toBe('test1.example.com')
  wrapper.unmount()
})

test('Delete', async () => {
  const mockRouter = { push: jest.fn() }

  const mock = new MockAdapter(axios)
  mock.onGet('/api/domains').reply(200,
    {
      data: [
        {
          device_mac_address: '111',
          device_name: 'syncloud',
          device_title: 'Syncloud',
          dkim_key: 'dkim',
          ip: '111.111.111.111',
          ipv6: null,
          last_update: 'Mon, 19 Oct 2020 19:31:49 GMT',
          local_ip: '192.168.1.1',
          map_local_address: false,
          platform_version: '2',
          web_local_port: 443,
          web_port: 443,
          web_protocol: 'https',
          name: 'test.example.com'
        }
      ]
    }
  )

  let deletedDomain
  mock.onDelete('/api/domain').reply(function (config) {
    deletedDomain = config.params.domain
    return [200, { success: true }]
  })

  const wrapper = mount(Devices,
    {
      attachTo: document.body,
      props: {
        onLogin: jest.fn()
      },
      global: {
        mocks: {
          $route: { path: '' },
          $router: mockRouter
        },
        stubs: {
          Confirmation: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          }
        }
      }
    }
  )

  await flushPromises()

  await wrapper.find('#delete').trigger('click')
  await wrapper.find('#delete_confirmation').trigger('confirm')
  await flushPromises()

  expect(deletedDomain).toBe('test.example.com')
  wrapper.unmount()
})
