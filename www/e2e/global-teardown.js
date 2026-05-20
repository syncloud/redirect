const { execSync } = require('child_process')
const fs = require('fs')
const path = require('path')

module.exports = async function () {
  if (!process.env.CI) return

  const logsDir = path.join(__dirname, '..', 'test-results', 'logs')
  fs.mkdirSync(logsDir, { recursive: true })

  const ssh = "sshpass -p syncloud ssh -o StrictHostKeyChecking=no root@www.syncloud.test"

  const grab = (label, cmd) => {
    const target = path.join(logsDir, label)
    try {
      const out = execSync(`${ssh} ${cmd}`, { encoding: 'utf-8', maxBuffer: 50 * 1024 * 1024 })
      fs.writeFileSync(target, out)
    } catch (e) {
      const stderr = (e.stderr && e.stderr.toString()) || ''
      const stdout = (e.stdout && e.stdout.toString()) || ''
      fs.writeFileSync(target, `# command failed: ${cmd}\n${e.message}\n${stdout}${stderr}`)
    }
  }

  for (const c of ['redirect-api', 'redirect-www', 'mail', 'node-exporter']) {
    grab(`${c}.log`, `'docker logs ${c} 2>&1'`)
  }

  for (const f of [
    'redirect_rest-error.log',
    'redirect_ssl_rest-error.log',
    'redirect_ssl_web-error.log',
    'redirect_rest-access.log',
    'redirect_ssl_rest-access.log',
    'redirect_ssl_web-access.log'
  ]) {
    grab(`apache-${f}`, `'tail -n 500 /var/log/apache2/${f} 2>/dev/null || true'`)
  }
}
