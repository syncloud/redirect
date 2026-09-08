(function () {
  function Buttons (config) {
    return {
      render: function (selector) {
        var el = typeof selector === 'string' ? document.querySelector(selector) : selector
        if (!el) return Promise.resolve()
        var button = document.createElement('button')
        button.type = 'button'
        button.textContent = 'PayPal'
        button.setAttribute('data-testid', 'stub-paypal-button')
        button.style.cssText = 'width:100%;height:100%;border:0;cursor:pointer;background:#ffc439;color:#003087;font-weight:700;font-size:15px'
        button.addEventListener('click', function () {
          if (config.onClick) config.onClick()
          var actions = {
            subscription: {
              create: function (payload) {
                var planId = (payload && payload.plan_id) || ''
                return Promise.resolve('PAYPALSUB~' + planId + '~' + Date.now())
              }
            }
          }
          Promise.resolve(
            config.createSubscription ? config.createSubscription({}, actions) : 'I-STUBPAYPAL'
          ).then(function (subscriptionID) {
            if (config.onApprove) config.onApprove({ subscriptionID: subscriptionID }, actions)
          })
        })
        el.appendChild(button)
        return Promise.resolve()
      }
    }
  }
  window.paypal = { Buttons: Buttons }
})()
