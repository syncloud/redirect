import { createServer, Model, Response } from 'miragejs'

let state = {
  loggedIn: true,
  credentials: {
    user: '11',
    password: '2'
  },
  user: {
    data: {
      active: true,
      email: 'test@example.com',
      notification_enabled: true,
      update_token: '0a',
      subscription_id: undefined
    }
  },
  plan: {
    data: {
      plan_annual_id: 'P-3AV82824GF026134TMU772XQ', // paypal sandbox plan id (Annual)
      plan_monthly_id: 'P-88T8436193034834XMDZRP4A', // paypal sandbox plan id (Monthly)
      client_id: 'AbuA_mUz0LOkG36bf3fYl59N8xXSQU8M6Zufpq-z07fNLG4XEM01SXGGJRAEXZpN2ejsl45S4VrA9qLN' // paypal sandbox client id
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

export function mock () {
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
      this.get('/api/domains', function (_schema, request) {
        return new Response(200, {}, state.domains)
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
        return new Response(200, {}, state.plan)
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
      this.post('/api/plan/subscribe', function (_schema, request) {
        const attrs = JSON.parse(request.requestBody)
        state.user.data.subscription_id = attrs.subscription_id
        return new Response(200, {}, {})
      })
      this.delete('/api/plan', function (_schema, _request) {
        state.user.data.subscription_id = undefined
        return new Response(200, {}, {})
      })
    }
  })
}
