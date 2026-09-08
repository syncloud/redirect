import { createServer, Model, Response } from 'miragejs'

let state = {
  loggedIn: true,
  unfinished: [
    {
      number: 7,
      device: 'Syncloud H4',
      option: '1 TB SSD',
      total: '£322.00',
      status: 'ordered',
      ordered: '2026-09-06'
    }
  ],
  orders: [
    {
      number: 1,
      device: 'Syncloud H4',
      option: '1 TB SSD',
      total: '£322.00',
      status: 'ordered',
      ordered: '2026-09-05',
      mine: true,
      email: 'test@example.com',
      name: 'Ada Lovelace',
      address: '1 Analytical Street',
      city: 'London',
      postcode: 'E1 6AN',
      country: 'United Kingdom'
    },
    {
      number: 2,
      device: 'Syncloud H4',
      option: '2 TB SSD x 2',
      total: '£687.00',
      status: 'sent',
      ordered: '2026-08-22',
      mine: true,
      email: 'test@example.com',
      name: 'Ada Lovelace',
      address: 'Flat 4, Somerset House, 12 Strand',
      city: 'London',
      postcode: 'WC2R 1LA',
      country: 'United Kingdom'
    },
    {
      number: 3,
      device: 'Syncloud H4',
      option: '120 GB SSD',
      total: '£244.00',
      status: 'sent',
      ordered: '2026-07-14',
      mine: true,
      email: 'test@example.com',
      name: 'Ada Lovelace',
      address: '1 Analytical Street',
      city: 'London',
      postcode: 'E1 6AN',
      country: 'United Kingdom'
    },
    {
      number: 4,
      device: 'Syncloud H4',
      option: '1 TB SSD x 2',
      total: '£422.00',
      status: 'ordered',
      ordered: '2026-09-06',
      mine: false,
      email: 'grace.hopper@navy.example.com',
      name: 'Grace Hopper',
      address: '2 Compiler Road',
      city: 'Manchester',
      postcode: 'M1 2AB',
      country: 'United Kingdom'
    },
    {
      number: 5,
      device: 'Syncloud H4',
      option: '120 GB SSD x 2',
      total: '£274.00',
      status: 'sent',
      ordered: '2026-09-02',
      mine: false,
      email: 'a-rather-long-address@some-long-domain.example.org',
      name: 'Karl-Friedrich von Habsburg-Lothringen',
      address: 'Bundesallee 187, Aufgang C, 3. Stock',
      city: 'Berlin-Wilmersdorf',
      postcode: '10717',
      country: 'Germany'
    },
    {
      number: 6,
      device: 'Syncloud H4',
      option: '2 TB SSD',
      total: '£444.00',
      status: 'ordered',
      ordered: '2026-08-28',
      mine: false,
      email: 'k.tanaka@example.jp',
      name: 'Kenji Tanaka',
      address: '3-2-1 Nishi-Shinjuku',
      city: 'Tokyo',
      postcode: '160-0023',
      country: 'Japan'
    }
  ],
  credentials: {
    user: 'test@example.com',
    password: 'test'
  },
  user: {
    data: {
      active: true,
      email: 'test@example.com',
      admin: true,
      notification_enabled: true,
      update_token: '0a',
      subscription_id: undefined,
      subscription_period: undefined,
      subscription_tier: undefined
    }
  },
  plan: {
    data: {
      plan_annual_id: 'P-3AV82824GF026134TMU772XQ', // paypal sandbox plan id (Annual)
      plan_monthly_id: 'P-88T8436193034834XMDZRP4A', // paypal sandbox plan id (Monthly)
      plan_max_annual_id: 'P-939294240R421883FNJSH54A', // paypal sandbox plan id (Max Annual)
      plan_max_monthly_id: 'P-1MN84195617128020NJSH3JI', // paypal sandbox plan id (Max Monthly)
      client_id: 'AbuA_mUz0LOkG36bf3fYl59N8xXSQU8M6Zufpq-z07fNLG4XEM01SXGGJRAEXZpN2ejsl45S4VrA9qLN', // paypal sandbox client id
      sdk_url: '/stub/paypal-sdk.js',
      stripe_max_enabled: true,
      paypal_max_enabled: true
    }
  },
  domains: {
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
        name: 'test.syncloud.test'
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
        name: 'test1.syncloud.test',
        name_servers: [
          'ns1.example.com',
          'ns2.example.com'
        ]
      }
    ]
  }
}

function paypalPlanInfo (subscriptionId) {
  const planId = String(subscriptionId).split('~')[1] || ''
  const p = state.plan.data
  if (planId === p.plan_annual_id) return { tier: 'pro', period: 'year' }
  if (planId === p.plan_max_monthly_id) return { tier: 'max', period: 'month' }
  if (planId === p.plan_max_annual_id) return { tier: 'max', period: 'year' }
  return { tier: 'pro', period: 'month' }
}

function stripePlanInfo (plan) {
  return {
    tier: plan && plan.indexOf('max') !== -1 ? 'max' : 'pro',
    period: plan && plan.indexOf('annual') !== -1 ? 'year' : 'month'
  }
}

export function mock () {
  console.info(
    `stub mode: sign in with ${state.credentials.user} / ${state.credentials.password}`
  )
  createServer({
    models: {
      author: Model
    },
    routes () {
      this.post('/api/user/login', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        if (state.credentials.user === attrs.email && state.credentials.password === attrs.password) {
          state.loggedIn = true
          return new Response(200, {}, { message: 'OK' })
        } else {
          if (attrs.email.length < 2) {
            return new Response(400, {}, {
              message: 'There\'s an error in parameters',
              parameters_messages: [
                {
                  messages: [
                    'Not valid email'
                  ],
                  parameter: 'email'
                }
              ]
            })
          } else {
            return new Response(400, {}, { message: 'Authentication failed' })
          }
        }
      })
      this.post('/api/user/create', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        if (attrs.email.length < 2) {
          return new Response(400, {}, {
            message: 'There\'s an error in parameters',
            parameters_messages: [
              {
                messages: [
                  'Not valid email'
                ],
                parameter: 'email'
              }
            ]
          })
        } else {
          return new Response(200, {}, { success: true, message: 'OK' })
        }
      })
      this.get('/api/user', function (_schema, request) {
        if (state.loggedIn) {
          return new Response(200, {}, state.user)
        } else {
          return new Response(401, {}, {})
        }
      })
      this.get('/api/device/catalog', function (_schema, _request) {
        return new Response(200, {}, {
          data: {
            devices: [{
              code: 'h4',
              name: 'Syncloud H4',
              board: 'odroid-hc4',
              price: 22900,
              specs: [
                { name: 'CPU', value: 'Amlogic S905X3, Cortex-A55' },
                { name: 'RAM', value: '4 GB DDR4' },
                { name: 'Drive bays', value: '2 x SATA, 3.5 or 2.5 inch' },
                { name: 'Ethernet', value: '1 Gb' },
                { name: 'CPU cores', value: '4' },
                { name: 'CPU frequency', value: '1.8 GHz' },
                { name: 'Boot media', value: 'Micro SD card, included' },
                { name: 'USB', value: 'USB 2.0 x 1' },
                { name: 'Video', value: 'HDMI 2.0, 4K at 60 Hz' },
                { name: 'Size', value: '84 x 90.5 x 25 mm' }
              ],
              options: [
                { code: '120', name: '120 GB SSD', extra: 0 },
                { code: '120x2', name: '120 GB SSD x 2', extra: 3000 },
                { code: '1t', name: '1 TB SSD', extra: 8000 },
                { code: '1tx2', name: '1 TB SSD x 2', extra: 18000 },
                { code: '2t', name: '2 TB SSD', extra: 20000 },
                { code: '2tx2', name: '2 TB SSD x 2', extra: 43000 }
              ]
            }],
            shipping: 1500,
            currency: 'GBP',
            paypal_client_id: import.meta.env.VITE_PAYPAL_CLIENT_ID || 'test'
          }
        })
      })
      this.post('/api/device/order', function (_schema, _request) {
        return new Response(200, {}, {
          data: { reference: 'stub-reference', provider_reference: 'stub-provider-reference' }
        })
      }, { timing: 1500 })
      this.post('/api/device/order/complete', function (_schema, _request) {
        return new Response(200, {}, { data: { number: 1 } })
      })
      this.get('/api/device/order', function (_schema, request) {
        const order = state.orders.find(each => String(each.number) === request.queryParams.number)
        const history = [
          { status: 'ordered', at: `${order.ordered} 09:14` },
          {
            status: 'ordered',
            at: `${order.ordered} 11:02`,
            comment: 'Payment cleared. Queued for building.'
          },
          {
            status: 'ordered',
            at: `${order.ordered} 15:47`,
            comment: 'Waiting on the drives for this one, the supplier says Thursday. ' +
              'Nothing needed from you, we will email again when it goes out.'
          }
        ]
        if (order.status === 'sent') {
          history.push({
            status: 'ordered',
            at: `${order.ordered} 08:30`,
            comment: 'Built and tested overnight, system written to the disk.'
          })
          history.push({
            status: 'sent',
            at: `${order.ordered} 16:40`,
            comment: 'Posted first class. Tracking number to follow in a separate email.'
          })
        }
        return new Response(200, {}, { data: { order, history } })
      })
      this.get('/api/device/orders/unfinished', function (_schema, _request) {
        return new Response(200, {}, { data: state.unfinished })
      })
      this.post('/api/device/order/retry', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        return new Response(200, {}, {
          data: {
            number: attrs.number,
            reference: 'stub-reference',
            provider_reference: 'stub-provider-reference',
            url: '/shop?reference=stub-reference'
          }
        })
      })
      this.get('/api/device/orders', function (_schema, _request) {
        return new Response(200, {}, { data: state.orders.filter(order => order.mine) })
      })
      this.get('/api/device/orders/all', function (_schema, _request) {
        return new Response(200, {}, { data: state.orders })
      })
      this.post('/api/device/order/status', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        const order = state.orders.find(each => each.number === attrs.number)
        if (order) {
          order.status = attrs.status
        }
        return new Response(200, {}, { message: 'OK' })
      })
      this.get('/api/domains', function (_schema, request) {
        return new Response(200, {}, state.domains)
      })
      this.get('/api/relay/usage', function (_schema, request) {
        return new Response(200, {}, { data: { enabled: true, used_bytes: 7408818586, limit_bytes: 10737418240 } })
      })
      this.get('/api/mail/usage', function (_schema, request) {
        return new Response(200, {}, { data: { enabled: true, used_messages: 41, limit_messages: 50 } })
      })
      this.get('/api/domain/check_nameservers', function (_schema, request) {
        const name = request.queryParams.domain
        const domain = state.domains.data.find(d => d.name === name)
        if (!domain || !domain.name_servers) {
          return new Response(200, {}, { data: { matched: true } })
        }
        return new Response(200, {}, {
          data: {
            matched: false,
            expected: domain.name_servers,
            actual: ['ns1.registrar.example', 'ns2.registrar.example']
          }
        })
      })
      this.post('/api/logout', function (_schema, request) {
        state.loggedIn = false
        return new Response(200, {}, {})
      })
      this.delete('/api/domain', function (_schema, request) {
        state.domains.data = state.domains.data.filter(v => {
          return v.name !== request.queryParams.domain
        })
        return new Response(200, {}, {})
      })
      this.post('/api/notification/enable', function (_schema, request) {
        state.user.data.notification_enabled = true
        return new Response(200, {}, {})
      })
      this.post('/api/notification/disable', function (_schema, request) {
        state.user.data.notification_enabled = false
        return new Response(200, {}, {})
      })
      this.delete('/api/user', function (_schema, request) {
        return new Response(200, {}, {})
      })
      this.post('/api/user/reset_password', function (_schema, request) {
        state = {}
        return new Response(200, {}, {})
      })
      this.post('/api/user/activate', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        if (attrs.token === '1') {
          return new Response(400, {}, { message: 'No such token' })
        } else {
          return new Response(200, {}, { message: 'Activated' })
        }
      })
      this.get('/api/plan', function (_schema, request) {
        const data = Object.assign({}, state.plan.data)
        if (state.user.data.subscription_id !== undefined) {
          data.current_period = state.user.data.subscription_period
          data.current_tier = state.user.data.subscription_tier
        }
        return new Response(200, {}, { data: data })
      })
      this.post('/api/user/set_password', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        console.log('set_password')
        console.log(attrs.token)
        if (attrs.token === '1') {
          console.log('set_password failed')
          return new Response(400, {}, { message: 'No such token' })
        } else {
          return new Response(200, {}, { message: 'Activated' })
        }
      })
      this.post('/api/plan/subscribe/paypal', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        const info = paypalPlanInfo(attrs.subscription_id)
        state.user.data.subscription_id = attrs.subscription_id
        state.user.data.subscription_period = info.period
        state.user.data.subscription_tier = info.tier
        return new Response(200, {}, {})
      })
      this.post('/api/plan/subscribe/crypto', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        state.user.data.subscription_id = attrs.subscription_id
        state.user.data.subscription_period = 'year'
        state.user.data.subscription_tier = 'pro'
        return new Response(200, {}, {})
      })
      this.post('/api/plan/subscribe/stripe/checkout', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        return new Response(200, {}, { data: { url: '/account?stripe_session_id=stub-' + (attrs.plan || 'monthly') } })
      })
      this.post('/api/plan/subscribe/stripe', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        const info = stripePlanInfo(String(attrs.subscription_id).replace('stub-', ''))
        state.user.data.subscription_id = 'sub_stub'
        state.user.data.subscription_period = info.period
        state.user.data.subscription_tier = info.tier
        return new Response(200, {}, {})
      })
      this.post('/api/plan/switch', function (_schema, _request) {
        state.user.data.subscription_period = 'year'
        return new Response(200, {}, { data: {} })
      })
      this.delete('/api/plan', function (_schema, _request) {
        state.user.data.subscription_id = undefined
        return new Response(200, {}, {})
      })
    }
  })
}
