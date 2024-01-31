import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '../../src/views/Account.vue'
import { ElButton, ElRadioGroup, ElRadioButton, ElRow, ElCol, ElImage, ElInput, ElIcon } from 'element-plus'

test('Notifications disable', async () => {
  let notificationsEnabled

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

  mock.onPost('/api/notification/disable').reply(function (_) {
    notificationsEnabled = false
    return [200, { success: true }]
  })

  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      global: {
        components: {
          RouterLink: RouterLinkStub
        },
        stubs: {
          CustomDialog: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          },
          'el-col': ElCol,
          'el-row': ElRow,
          'el-radio-button': ElRadioButton,
          'el-radio-group': ElRadioGroup,
          'el-button': ElButton,
          'el-image': ElImage,
          'el-input': ElInput,
          'el-icon': ElIcon
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#chk_email').trigger('click')
  await wrapper.find('#save').trigger('click')

  await flushPromises()

  expect(notificationsEnabled).toBe(false)
  wrapper.unmount()
})

test('Notifications subscribe', async () => {
  let subscribed

  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200,
    {
      data: {
        active: true,
        email: 'test@example.com',
        notification_enabled: false,
        update_token: '0a'
      }
    }
  )

  mock.onPost('/api/notification/enable').reply(function (_) {
    subscribed = true
    return [200, { success: true }]
  })

  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
    {
      attachTo: document.body,
      global: {
        components: {
          RouterLink: RouterLinkStub
        },
        stubs: {
          CustomDialog: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          },
          'el-col': ElCol,
          'el-row': ElRow,
          'el-radio-button': ElRadioButton,
          'el-radio-group': ElRadioGroup,
          'el-button': ElButton,
          'el-image': ElImage,
          'el-input': ElInput,
          'el-icon': ElIcon
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#chk_email').trigger('click')
  await wrapper.find('#save').trigger('click')

  await flushPromises()

  expect(subscribed).toBe(true)
  wrapper.unmount()
})

test('Delete', async () => {
  let deleted

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

  mock.onDelete('/api/user').reply(function (_) {
    deleted = true
    return [200, { success: true }]
  })

  mock.onPost('/api/logout').reply(function (_) {
    return [200, { success: true }]
  })
  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
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
          CustomDialog: {
            template: '<button :id="id" />',
            props: { id: String },
            methods: {
              show () {
              }
            }
          },
          'el-col': ElCol,
          'el-row': ElRow,
          'el-radio-button': ElRadioButton,
          'el-radio-group': ElRadioGroup,
          'el-button': ElButton,
          'el-image': ElImage,
          'el-input': ElInput,
          'el-icon': ElIcon
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#delete').trigger('click')
  await wrapper.find('#delete_confirmation').trigger('confirm')

  await flushPromises()

  expect(deleted).toBe(true)
  wrapper.unmount()
})

test('Crypto Subscribe', async () => {
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

  let subscriptionId
  mock.onPost('api/plan/subscribe/crypto').reply(function (config) {
    subscriptionId = JSON.parse(config.data).subscription_id
    return [200, { success: true }]
  })

  mock.onPost('/api/logout').reply(function (_) {
    return [200, { success: true }]
  })
  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const wrapper = mount(Account,
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
          CustomDialog: true,
          'el-col': ElCol,
          'el-row': ElRow,
          'el-radio-button': ElRadioButton,
          'el-radio-group': ElRadioGroup,
          'el-button': ElButton,
          'el-image': ElImage,
          'el-input': ElInput,
          'el-icon': ElIcon
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#crypto_year').trigger('click')
  await flushPromises()
  expect(wrapper.find('#crypto_subscribe_btn').attributes('disabled')).toBe('')
  await wrapper.find('#crypto_transaction_id').setValue('12345678901')
  await wrapper.find('#crypto_subscribe_btn').trigger('click')

  await flushPromises()

  expect(subscriptionId).toBe('12345678901')
  wrapper.unmount()
})
