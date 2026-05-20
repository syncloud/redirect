const fs = require('node:fs')
const path = require('node:path')
const { ssh, scpFrom } = require('./helpers/ssh')

const REMOTE_TMP = '/tmp/syncloud/ui'

module.exports = async function () {
  if (!process.env.CI) return

  const out = path.join(__dirname, '..', 'test-results', 'logs')
  fs.mkdirSync(out, { recursive: true })

  ssh(`rm -rf ${REMOTE_TMP} && mkdir -p ${REMOTE_TMP}`, { throw: false })

  for (const c of ['redirect-api', 'redirect-www', 'mail', 'node-exporter']) {
    ssh(`docker logs ${c} > ${REMOTE_TMP}/${c}.log 2>&1`, { throw: false })
  }
  for (const f of [
    'redirect_rest-error.log',
    'redirect_ssl_rest-error.log',
    'redirect_ssl_web-error.log',
    'redirect_rest-access.log',
    'redirect_ssl_rest-access.log',
    'redirect_ssl_web-access.log'
  ]) {
    ssh(`tail -n 500 /var/log/apache2/${f} > ${REMOTE_TMP}/apache-${f} 2>/dev/null || true`, { throw: false })
  }

  scpFrom(`${REMOTE_TMP}/*`, out, { throw: false })
}
