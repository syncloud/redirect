import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '../../src/views/Account.vue'
import { ElButton, ElRadioGroup, ElRadioButton, ElRow, ElCol, ElImage, ElInput, ElIcon, ElSwitch, ElCard, ElTag } from 'element-plus'

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
          'el-icon': ElIcon,
          'el-switch': ElSwitch,
          'el-card': ElCard,
          'el-tag': ElTag
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
          'el-icon': ElIcon,
          'el-switch': ElSwitch,
          'el-card': ElCard,
          'el-tag': ElTag
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
          'el-icon': ElIcon,
          'el-switch': ElSwitch,
          'el-card': ElCard,
          'el-tag': ElTag
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
          'el-icon': ElIcon,
          'el-switch': ElSwitch,
          'el-card': ElCard,
          'el-tag': ElTag
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

function mountSubscribed (planData) {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, {
    data: {
      active: true,
      email: 'test@example.com',
      notification_enabled: true,
      update_token: '0a',
      subscription_id: 'I-SUB'
    }
  })
  mock.onGet('/api/plan').reply(200, { data: Object.assign({ client_id: '2' }, planData) })
  return { mock }
}

const subscribedStubs = {
  attachTo: document.body,
  props: { checkUserSession: jest.fn() },
  global: {
    components: { RouterLink: RouterLinkStub },
    stubs: {
      CustomDialog: { template: '<button :id="id" />', props: { id: String } },
      'el-col': ElCol,
      'el-row': ElRow,
      'el-radio-button': ElRadioButton,
      'el-radio-group': ElRadioGroup,
      'el-button': ElButton,
      'el-image': ElImage,
      'el-input': ElInput,
      'el-icon': ElIcon,
      'el-switch': ElSwitch,
      'el-card': ElCard,
      'el-tag': ElTag
    }
  }
}

test('Max monthly subscriber sees max price and switch option', async () => {
  mountSubscribed({ current_period: 'month', current_tier: 'max' })
  const wrapper = mount(Account, subscribedStubs)
  await flushPromises()
  expect(wrapper.find('[data-testid="billing-current"]').text()).toBe('£15 / month')
  expect(wrapper.find('[data-testid="switch-annual"]').exists()).toBe(true)
  wrapper.unmount()
})

test('Max annual subscriber sees max annual price and no switch option', async () => {
  mountSubscribed({ current_period: 'year', current_tier: 'max' })
  const wrapper = mount(Account, subscribedStubs)
  await flushPromises()
  expect(wrapper.find('[data-testid="billing-current"]').text()).toBe('£180 / year')
  expect(wrapper.find('[data-testid="switch-annual"]').exists()).toBe(false)
  wrapper.unmount()
})

test('Switch to annual redirects to the approval url', async () => {
  const { mock } = mountSubscribed({ current_period: 'month', current_tier: 'pro' })
  let switched = false
  mock.onPost('/api/plan/switch').reply(function (_) {
    switched = true
    return [200, { data: { url: 'https://paypal.test/approve' } }]
  })
  const originalLocation = window.location
  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: { href: '' } })

  const wrapper = mount(Account, subscribedStubs)
  await flushPromises()
  await wrapper.find('#switch_annual').trigger('click')
  await wrapper.find('#switch_confirmation').trigger('confirm')
  await flushPromises()

  expect(switched).toBe(true)
  expect(window.location.href).toBe('https://paypal.test/approve')

  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: originalLocation })
  wrapper.unmount()
})

test('Switch to annual shows the provider message on failure', async () => {
  const { mock } = mountSubscribed({ current_period: 'month', current_tier: 'pro' })
  mock.onPost('/api/plan/switch').reply(422, { message: 'Payment for the subscription is in progress.' })

  const wrapper = mount(Account, subscribedStubs)
  await flushPromises()
  await wrapper.find('#switch_annual').trigger('click')
  await wrapper.find('#switch_confirmation').trigger('confirm')
  await flushPromises()

  expect(wrapper.find('[data-testid="switch-error"]').text()).toBe('Payment for the subscription is in progress.')
  wrapper.unmount()
})

test('Stripe Checkout', async () => {
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

  let checkoutPlan
  mock.onPost('/api/plan/subscribe/stripe/checkout').reply(function (config) {
    checkoutPlan = JSON.parse(config.data).plan
    return [200, { data: { url: 'https://checkout.stripe.test/session' } }]
  })

  mock.onGet('/api/plan').reply(200, { data: { plan_id: '1', client_id: '2' } })

  const originalLocation = window.location
  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: { href: '' } })

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
          'el-icon': ElIcon,
          'el-switch': ElSwitch,
          'el-card': ElCard,
          'el-tag': ElTag
        }
      }

    }
  )

  await flushPromises()

  await wrapper.find('#stripe_subscribe_btn').trigger('click')

  await flushPromises()

  expect(checkoutPlan).toBe('monthly')
  expect(window.location.href).toBe('https://checkout.stripe.test/session')

  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: originalLocation })
  wrapper.unmount()
})
