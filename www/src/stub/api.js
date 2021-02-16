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
      unsubscribed: false,
      premium_status_id: 3,
      update_token: '0a'
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
        user_domain: 'test',
        web_local_port: 443,
        web_port: 443,
        web_protocol: 'https'
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
        user_domain: 'test1',
        web_local_port: 443,
        web_port: 10001,
        web_protocol: 'https'
      }
    ]
  }
}

const express = require('express')
const bodyparser = require('body-parser')
const mock = function (app, server, compiler) {
  app.use(express.urlencoded())
  app.use(bodyparser.json())
  app.post('/api/login', function (req, res) {
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
  app.post('/api/domain_delete', function (req, res) {
    state.domains = state.domains.filter(v => {
      return v.user_domain !== req.body.user_domain
    })
    res.json({})
  })
  app.post('/api/notification/subscribe', function (req, res) {
    state.user.unsubscribed = false
    res.json({})
  })
  app.post('/api/notification/unsubscribe', function (req, res) {
    state.user.unsubscribed = true
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
  app.post('/api/premium/request', function (req, res) {
    req.user.premium_status_id = 3
    res.json({})
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
}

exports.mock = mock
