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
      update_token: '0a'
      // subscription_id: 'S-1'
    }
  },
  plan: {
    data: {
      plan_id: 'P-88T8436193034834XMDZRP4A', // paypal sandbox plan id
      client_id: 'AbuA_mUz0LOkG36bf3fYl59N8xXSQU8M6Zufpq-z07fNLG4XEM01SXGGJRAEXZpN2ejsl45S4VrA9qLN', // paypal sandbox client id
      subscribed: false
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

const express = require('express')
const bodyparser = require('body-parser')
const mock = function (app, server, compiler) {
  app.use(express.urlencoded())
  app.use(bodyparser.json())
  app.post('/api/user/login', function (req, res) {
    if (state.credentials.user === req.body.email && state.credentials.password === req.body.password) {
      state.loggedIn = true
      res.json({ message: 'OK' })
    } else {
      if (req.body.email.length < 2) {
        res.status(400).json({
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
        res.status(400).json({ message: 'Authentication failed' })
      }
    }
  })
  app.post('/api/user/create', function (req, res) {
    if (req.body.email.length < 2) {
      res.status(400).json({
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
      res.json({ success: true, message: 'OK' })
    }
  })
  app.get('/api/user', function (req, res) {
    if (state.loggedIn) {
      res.json(state.user)
    } else {
      res.sendStatus(401)
    }
  })
  app.get('/api/domains', function (req, res) {
    res.json(state.domains)
  })
  app.post('/api/logout', function (req, res) {
    state.loggedIn = false
    res.json({})
  })
  app.delete('/api/domain', function (req, res) {
    state.domains.data = state.domains.data.filter(v => {
      return v.name !== req.query.domain
    })
    res.json({})
  })
  app.post('/api/notification/enable', function (req, res) {
    state.user.data.notification_enabled = true
    res.json({})
  })
  app.post('/api/notification/disable', function (req, res) {
    state.user.data.notification_enabled = false
    res.json({})
  })
  app.delete('/api/user', function (req, res) {
    res.json({})
  })
  app.post('/api/user/reset_password', function (req, res) {
    state = {}
    res.json({})
  })
  app.post('/api/user/activate', function (req, res) {
    if (req.body.token === '1') {
      res.status(400).json({ message: 'No such token' })
    } else {
      res.json({ message: 'Activated' })
    }
  })
  app.get('/api/plan', function (req, res) {
    res.json(state.plan)
  })
  app.post('/api/user/set_password', function (req, res) {
    console.log('set_password')
    console.log(req.body.token)
    if (req.body.token === '1') {
      console.log('set_password failed')
      res.status(400).json({ message: 'No such token' })
    } else {
      res.json({ message: 'Activated' })
    }
  })
  app.post('/api/plan/subscribe', function (req, res) {
    state.user.data.subscription_id = req.body.subscription_id
    res.json({ })
  })
}

exports.mock = mock
