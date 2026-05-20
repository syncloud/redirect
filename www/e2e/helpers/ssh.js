const { execFileSync } = require('node:child_process')

const deviceHost = process.env.PLAYWRIGHT_DEVICE_HOST ?? 'www.syncloud.test'
const sshUser = process.env.PLAYWRIGHT_SSH_USER ?? 'root'
const sshPassword = process.env.PLAYWRIGHT_SSH_PASSWORD ?? 'syncloud'

const baseArgs = [
  '-o', 'StrictHostKeyChecking=no',
  '-o', 'UserKnownHostsFile=/dev/null',
  '-o', 'LogLevel=ERROR'
]

function ssh (cmd, opts = {}) {
  const args = ['-p', sshPassword, 'ssh', ...baseArgs, `${sshUser}@${deviceHost}`, cmd]
  try {
    return execFileSync('sshpass', args, { encoding: 'utf8', timeout: 120_000 })
  } catch (e) {
    if (opts.throw === false) {
      return (e.stdout?.toString() ?? '') + (e.stderr?.toString() ?? '')
    }
    throw e
  }
}

function scpFrom (remote, local, opts = {}) {
  const args = ['-p', sshPassword, 'scp', ...baseArgs, '-r', `${sshUser}@${deviceHost}:${remote}`, local]
  try {
    execFileSync('sshpass', args, { encoding: 'utf8', timeout: 120_000 })
  } catch (e) {
    if (opts.throw !== false) throw e
  }
}

module.exports = { ssh, scpFrom, deviceHost, sshUser, sshPassword }
