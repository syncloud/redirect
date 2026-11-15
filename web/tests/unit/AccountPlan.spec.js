import { mount, RouterLinkStub } from '@vue/test-utils'
import axios from 'axios'
import MockAdapter from 'axios-mock-adapter'
import flushPromises from 'flush-promises'
import Account from '../../src/views/Account.vue'
import { loadScript } from '@paypal/paypal-js'
import { ElButton, ElRadioGroup, ElRadioButton, ElRow, ElCol, ElImage, ElInput, ElIcon, ElSwitch, ElCard, ElTag } from 'element-plus'
import { track } from '../../src/track'

jest.mock('@paypal/paypal-js')
jest.mock('../../src/track', () => ({
  ...jest.requireActual('../../src/track'),
  track: jest.fn()
}))

const options = {
  attachTo: document.body,
  props: { checkUserSession: jest.fn() },
  global: {
    components: { RouterLink: RouterLinkStub },
    mocks: { $router: { push: jest.fn(), replace: jest.fn() } },
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

function server (subscriptionId) {
  const mock = new MockAdapter(axios)
  const user = {
    active: true,
    email: 'test@example.com',
    notification_enabled: true,
    update_token: '0a'
  }
  if (subscriptionId) {
    user.subscription_id = subscriptionId
  }
  mock.onGet('/api/user').reply(200, { data: user })
  mock.onGet('/api/plan').reply(200, {
    data: {
      client_id: '2',
      plan_monthly_id: 'P-M',
      plan_annual_id: 'P-A',
      plan_max_monthly_id: 'P-MM',
      plan_max_annual_id: 'P-MA',
      stripe_max_enabled: true,
      paypal_max_enabled: true
    }
  })
  return mock
}

function names () {
  return track.mock.calls.map(each => each[0])
}

let paypalButtons

beforeEach(() => {
  track.mockClear()
  paypalButtons = null
  loadScript.mockReset()
  loadScript.mockResolvedValue({
    Buttons: config => {
      paypalButtons = config
      return { render: jest.fn() }
    }
  })
})

test('a visitor without a subscription is counted as having seen the offer', async () => {
  server(null)
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(names()).toContain('plan.view')
  wrapper.unmount()
})

test('someone who already pays is not counted as seeing the offer again', async () => {
  server('I-SUB')
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(names()).not.toContain('plan.view')
  wrapper.unmount()
})

test('choosing a billing period and a tier is counted', async () => {
  server(null)
  const wrapper = mount(Account, options)
  await flushPromises()
  track.mockClear()

  await wrapper.find('[data-testid="billing-year"]').trigger('click')
  await wrapper.find('[data-testid="plan-max"]').trigger('click')
  await flushPromises()

  expect(names()).toEqual(['plan.period', 'plan.tier'])
  wrapper.unmount()
})

test('pressing the card button is counted before leaving for the provider', async () => {
  const mock = server(null)
  mock.onPost('/api/plan/subscribe/stripe/checkout').reply(200, { data: { url: 'https://stripe.test/pay' } })
  const original = window.location
  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: { href: '' } })

  const wrapper = mount(Account, options)
  await flushPromises()
  await wrapper.find('[data-testid="stripe-subscribe"]').trigger('click')
  await flushPromises()

  expect(names()).toContain('plan.stripe')
  expect(window.location.href).toBe('https://stripe.test/pay')

  Object.defineProperty(window, 'location', { configurable: true, writable: true, value: original })
  wrapper.unmount()
})

test('a PayPal button that never loads is reported and says so on the page', async () => {
  loadScript.mockRejectedValue(new Error('blocked'))
  const error = jest.spyOn(console, 'error').mockImplementation(() => {})
  server(null)

  const wrapper = mount(Account, options)
  await flushPromises()

  expect(names()).toContain('plan.paypal.failed')
  expect(wrapper.find('[data-testid="plan-paypal-failed"]').exists()).toBe(true)

  error.mockRestore()
  wrapper.unmount()
})

test('a PayPal button that loads says nothing about being broken', async () => {
  server(null)
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(names()).not.toContain('plan.paypal.failed')
  expect(wrapper.find('[data-testid="plan-paypal-failed"]').exists()).toBe(false)
  wrapper.unmount()
})

test('a completed PayPal subscription is counted at the press and at the finish', async () => {
  const mock = server(null)
  mock.onPost('/api/plan/subscribe/paypal').reply(200, { success: true })

  const wrapper = mount(Account, options)
  await flushPromises()
  track.mockClear()

  paypalButtons.onClick()
  await paypalButtons.onApprove({ subscriptionID: 'I-NEW' })
  await flushPromises()

  expect(names()).toEqual(['plan.paypal', 'plan.subscribed'])
  wrapper.unmount()
})

test('opening the crypto details and sending a transaction are counted separately', async () => {
  const mock = server(null)
  mock.onPost('/api/plan/subscribe/crypto').reply(200, { success: true })

  const wrapper = mount(Account, options)
  await flushPromises()
  track.mockClear()

  await wrapper.find('[data-testid="crypto-toggle"]').trigger('click')
  await wrapper.find('#crypto_transaction_id').setValue('12345678901')
  await wrapper.find('#crypto_subscribe_btn').trigger('click')
  await flushPromises()

  expect(names()).toEqual(['plan.crypto', 'plan.crypto.sent', 'plan.subscribed'])
  wrapper.unmount()
})

test('closing the crypto details again is not counted as opening them', async () => {
  server(null)
  const wrapper = mount(Account, options)
  await flushPromises()
  track.mockClear()

  await wrapper.find('[data-testid="crypto-toggle"]').trigger('click')
  await wrapper.find('[data-testid="crypto-toggle"]').trigger('click')
  await flushPromises()

  expect(names()).toEqual(['plan.crypto'])
  wrapper.unmount()
})

test('a plan request that fails is counted as a plan error', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, {
    data: { active: true, email: 'test@example.com', notification_enabled: true, update_token: '0a' }
  })
  mock.onGet('/api/plan').reply(500)
  const log = jest.spyOn(console, 'log').mockImplementation(() => {})

  const wrapper = mount(Account, options)
  await flushPromises()

  expect(names()).toContain('plan.error')

  log.mockRestore()
  wrapper.unmount()
})

test('an error elsewhere on the page is not counted as a plan error', async () => {
  const mock = server(null)
  mock.onPost('/api/notification/disable').reply(500)
  const log = jest.spyOn(console, 'log').mockImplementation(() => {})

  const wrapper = mount(Account, options)
  await flushPromises()
  track.mockClear()

  await wrapper.find('#chk_email').trigger('click')
  await wrapper.find('#save').trigger('click')
  await flushPromises()

  expect(names()).not.toContain('plan.error')

  log.mockRestore()
  wrapper.unmount()
})

test('an account with no plan is offered the code that came with a device', async () => {
  server()
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(wrapper.find('[data-testid="redeem-code"]').isVisible()).toBe(false)

  await wrapper.find('[data-testid="redeem-toggle"]').trigger('click')
  expect(wrapper.find('[data-testid="redeem-code"]').isVisible()).toBe(true)
  expect(wrapper.find('[data-testid="redeem-submit"]').attributes('disabled')).toBeDefined()
})

test('a code is sent as typed and the plan is reloaded', async () => {
  const mock = server()
  let sent
  mock.onPost('/api/plan/redeem').reply(config => {
    sent = JSON.parse(config.data)
    return [200, { data: { plan: 'pro', ends_at: '2027-11-15T00:00:00Z' } }]
  })
  const wrapper = mount(Account, options)
  await flushPromises()

  await wrapper.find('[data-testid="redeem-toggle"]').trigger('click')
  await wrapper.find('[data-testid="redeem-code"]').setValue('aaaa-bbbb-cccc')
  await wrapper.find('[data-testid="redeem-submit"]').trigger('click')
  await flushPromises()

  expect(sent).toEqual({ code: 'aaaa-bbbb-cccc' })
  expect(wrapper.find('[data-testid="redeem-failed"]').exists()).toBe(false)
})

test('a refused code says why instead of throwing the buyer at the error page', async () => {
  const mock = server()
  mock.onPost('/api/plan/redeem').reply(400, { message: 'that code has already been used' })
  const push = jest.fn()
  const wrapper = mount(Account, { ...options, global: { ...options.global, mocks: { $router: { push, replace: jest.fn() } } } })
  await flushPromises()

  await wrapper.find('[data-testid="redeem-toggle"]').trigger('click')
  await wrapper.find('[data-testid="redeem-code"]').setValue('aaaa-bbbb-cccc')
  await wrapper.find('[data-testid="redeem-submit"]').trigger('click')
  await flushPromises()

  expect(wrapper.find('[data-testid="redeem-failed"]').text()).toBe('that code has already been used')
  expect(push).not.toHaveBeenCalledWith('/error')
})

test('a code field is not offered to an account that already has a plan', async () => {
  server('sub-1')
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(wrapper.find('[data-testid="redeem-toggle"]').isVisible()).toBe(false)
})

test('an included year is not quoted a price it will not be charged', async () => {
  const mock = new MockAdapter(axios)
  mock.onGet('/api/user').reply(200, {
    data: {
      active: true,
      email: 'test@example.com',
      notification_enabled: true,
      update_token: '0a',
      subscription_id: '7',
      subscription_type: 4
    }
  })
  mock.onGet('/api/plan').reply(200, { data: { client_id: '2' } })
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(wrapper.find('[data-testid="billing-included"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="billing-current"]').exists()).toBe(false)
  expect(wrapper.find('[data-testid="switch-annual"]').exists()).toBe(false)
})

test('a paying subscriber still sees what they are charged', async () => {
  server('sub-1')
  const wrapper = mount(Account, options)
  await flushPromises()

  expect(wrapper.find('[data-testid="billing-current"]').exists()).toBe(true)
  expect(wrapper.find('[data-testid="billing-included"]').exists()).toBe(false)
})
