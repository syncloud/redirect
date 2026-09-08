package main

import (
	"fmt"
	"net/http"
)

const sdk = `
function fakerCreateSubscription (payload) {
  var planId = (payload && payload.plan_id) || ''
  return Promise.resolve('PAYPALSUB~' + planId + '~' + Date.now())
}

window.paypal = {
  Buttons: function (options) {
    return {
      render: function (selector) {
        var host = document.querySelector(selector)
        if (!host) { return Promise.resolve() }
        var button = document.createElement('button')
        button.type = 'button'
        button.textContent = 'PayPal'
        button.setAttribute('data-testid', 'paypal-faker-button')
        button.style.cssText =
          'width:100%%;height:44px;border-radius:8px;border:0;background:#ffc439;font-weight:700;cursor:pointer'
        button.addEventListener('click', function () {
          Promise.resolve()
            .then(function () {
              if (options.onClick) { return options.onClick({}, {}) }
            })
            .then(function () {
              if (options.createSubscription) {
                var actions = { subscription: { create: fakerCreateSubscription } }
                return options.createSubscription({}, actions).then(function (subscriptionID) {
                  return options.onApprove({ subscriptionID: subscriptionID })
                })
              }
              return options.createOrder().then(function (orderID) {
                return options.onApprove({ orderID: orderID })
              })
            })
            .catch(function (error) {
              if (options.onError) { options.onError(error) }
            })
        })
        host.appendChild(button)
        return Promise.resolve()
      }
    }
  }
}
`

func (p *PayPal) script(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	fmt.Fprint(w, sdk)
}
