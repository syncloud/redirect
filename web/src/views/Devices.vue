<template>
  <div class="sc-page">
    <div id="has_domains" :class="{ invisible: !hasDomains }">
      <h1 class="sc-h1">Devices</h1>
      <p class="sc-lead">Your activated Syncloud devices and how much of your plan you are using.</p>

      <div class="sc-card sc-usage" data-testid="relay-usage">
        <div class="sc-usage-row" data-testid="usage-traffic">
          <div class="sc-usage-head">
            <span class="sc-usage-name">Device access</span>
            <span class="sc-usage-value" data-testid="relay-usage-text">{{ relayText }}</span>
          </div>
          <div v-if="relayEnabled" class="sc-meter">
            <div
              class="sc-meter-fill"
              :class="relayBarClass"
              :style="{ width: Math.min(relayPercent, 100) + '%' }"
              data-testid="relay-usage-bar"
            ></div>
          </div>
          <p v-else class="sc-usage-off" data-testid="usage-traffic-off">
            Your device is reached directly, with no traffic limit. Turn the relay on if your
            connection has no public address.
          </p>
        </div>

        <div class="sc-usage-row" data-testid="usage-email">
          <div class="sc-usage-head">
            <span class="sc-usage-name">Email sending</span>
            <span class="sc-usage-value" data-testid="usage-email-text">{{ emailText }}</span>
          </div>
          <div v-if="emailEnabled" class="sc-meter">
            <div
              class="sc-meter-fill"
              :class="emailBarClass"
              :style="{ width: Math.min(emailPercent, 100) + '%' }"
              data-testid="usage-email-bar"
            ></div>
          </div>
          <p v-else class="sc-usage-off" data-testid="usage-email-off">
            Your device sends email directly, with no limit. Turn the relay on if providers
            reject mail from your address.
          </p>
        </div>

        <router-link v-if="nearLimit" to="/account" data-testid="relay-upgrade" class="sc-usage-upgrade">
          Approaching your limit &mdash; upgrade for more
        </router-link>
      </div>

      <div class="sc-device-grid">
        <div v-for="(domain, index) in allDomains" :key="index" class="sc-card sc-device">
          <div class="sc-device-head">
            <span id="name" class="sc-device-name" data-testid="domain-name">{{ domain.name }}</span>
            <span class="sc-dot" :class="domain.online ? 'online' : 'offline'"></span>
          </div>

          <div class="sc-device-title-row">
            <h3 id="title" class="sc-device-title" data-testid="device-title">{{ domain.device_title }}</h3>
            <button
              type="button"
              class="sc-btn-quiet"
              id="delete"
              data-testid="device-delete"
              @click="domainDeleteConfirm(domain.name)"
            >Deactivate</button>
          </div>

          <dl class="sc-kv">
            <dt>Domain address</dt>
            <dd>
              <a v-if="domain.has_domain_address" :href="domain.domain_address">{{ domain.domain_address }}</a>
              <span v-else class="sc-kv-empty">Not mapped</span>
            </dd>

            <dt>External address</dt>
            <dd>
              <a id="external_address" v-if="domain.has_external_address" :href="domain.external_address">{{ domain.external_address }}</a>
              <span v-else class="sc-kv-empty">Not provided</span>
            </dd>

            <dt>Internal address</dt>
            <dd>
              <a id="internal_address" v-if="domain.has_internal_address" :href="domain.internal_address">{{ domain.internal_address }}</a>
              <span v-else class="sc-kv-empty">Not provided</span>
            </dd>

            <dt>IPv6 address</dt>
            <dd>
              <a id="ipv6_address" v-if="domain.has_ipv6_address" :href="domain.ipv6_address">{{ domain.ipv6_address }}</a>
              <span id="ipv6_address_not_available" v-else class="sc-kv-empty">Not provided</span>
            </dd>

            <template v-if="domain.name_servers">
              <dt>Name servers</dt>
              <dd>
                <span v-if="domain.ns_check_state === 'matched'" class="sc-tag ok" data-testid="ns-status-matched">Matched</span>
                <span v-if="domain.ns_check_state === 'mismatched'" class="sc-tag bad" data-testid="ns-status-mismatched">Not set at registrar</span>
                <span v-if="domain.ns_check_state === 'checking'" class="sc-tag" data-testid="ns-status-checking">Checking...</span>
                <button
                  type="button"
                  class="sc-btn-quiet sc-btn-small"
                  data-testid="ns-revalidate"
                  :disabled="domain.ns_check_state === 'checking'"
                  @click="checkNameServers(domain)"
                >Revalidate</button>
                <div v-for="(name_server, i) in domain.name_servers" :key="i">
                  <code>{{ name_server }}</code>
                </div>
                <div
                  v-if="domain.ns_check_state === 'mismatched' && domain.ns_check_actual && domain.ns_check_actual.length > 0"
                  class="sc-kv-note"
                >
                  Currently set at registrar:
                  <div v-for="(ns, i) in domain.ns_check_actual" :key="i"><code>{{ ns }}</code></div>
                </div>
              </dd>
            </template>

            <dt>Updated</dt>
            <dd>{{ domain.nice_last_update }}</dd>
          </dl>
        </div>
      </div>
    </div>

    <div id="no_domains" data-testid="no-devices" :class="{ invisible: hasDomains }">
      <div class="sc-card sc-empty">
        <h1 class="sc-h1">No devices yet</h1>
        <p class="sc-lead" data-testid="no-devices-steps">
          Your account is ready. Next, install Syncloud on your own hardware and
          activate it with this account.
        </p>
        <ol data-testid="no-devices-list">
          <li>Use a Raspberry Pi, an old PC, or a ready-made device</li>
          <li>Write the Syncloud image to it and connect it to your router</li>
          <li>Open the device in your browser and activate it with this account</li>
        </ol>
        <a class="sc-btn sc-btn-inline" href="https://syncloud.org/setup" data-testid="no-devices-setup">
          How to set up your device
        </a>
      </div>
    </div>
  </div>

  <CustomDialog
    :visible="deleteConfirmationVisible"
    id="delete_confirmation"
    @cancel="deleteConfirmationVisible = false"
    @confirm="domainDelete"
  >
    <template v-slot:title>
      Deactivate {{ domainToDelete }}
    </template>
    <template v-slot:text>
      Device will be unlinked from the domain.<br>Domain will be released and might be taken by other user.<br>Proceed with caution!
    </template>
  </CustomDialog>
</template>

<script>
import axios from 'axios'
import moment from 'moment'
import CustomDialog from '../components/CustomDialog.vue'

function sameDay (date1, date2) {
  return (date1.getDate() === date2.getDate() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getFullYear() === date2.getFullYear())
}

function fullUrl (address, port) {
  let result = 'https://' + address
  if (port !== undefined && port !== 443) {
    result = result + ':' + port
  }
  return result
}

function niceTimestamp (ds, today) {
  if (ds === null) {
    return 'never'
  }
  const d = new Date(Date.parse(ds))
  if (sameDay(today, d)) {
    return 'Today ' + moment(d).format('H:mm')
  } else {
    return moment(d).format('MMM D, yyyy')
  }
}

function online (ds) {
  if (ds === null) {
    return false
  }

  const diff = new Date() - new Date(Date.parse(ds))
  const minutes = Math.floor((diff / 1000) / 60)

  return minutes < 10
}

function convert (domain) {
  domain.domain_address_port = domain.map_local_address ? 443 : domain.web_port
  domain.domain_address = fullUrl(domain.name, domain.domain_address_port)
  domain.has_domain_address = domain.name !== undefined
  domain.external_address = fullUrl(domain.ip, domain.domain_address_port)
  domain.has_external_address = domain.ip !== undefined
  domain.internal_address = 'https://' + domain.local_ip
  domain.has_internal_address = domain.local_ip !== undefined
  domain.ipv6_address = 'https://[' + domain.ipv6 + ']'
  domain.has_ipv6_address = !!domain.ipv6
  domain.online = online(domain.last_update)
  domain.nice_last_update = niceTimestamp(domain.last_update, new Date())
  domain.ns_check_state = 'unchecked'
  domain.ns_check_actual = []
  return domain
}

function gb (bytes) {
  return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
}

export default {
  name: 'Devices',
  components: {
    CustomDialog
  },
  props: {
    checkUserSession: Function
  },
  data () {
    return {
      hasDomains: Boolean,
      domainGroups: [],
      domainToDelete: '',
      deleteConfirmationVisible: false,
      relayUsed: 0,
      relayLimit: 0,
      relayOn: false,
      emailUsed: 0,
      emailLimit: 0,
      emailOn: false
    }
  },
  mounted () {
    this.reload()
    this.loadRelayUsage()
  },
  computed: {
    allDomains () {
      return Array.isArray(this.domainGroups) ? this.domainGroups.flat() : []
    },
    relayEnabled () {
      return this.relayOn && this.relayLimit > 0
    },
    emailEnabled () {
      return this.emailOn && this.emailLimit > 0
    },
    emailPercent () {
      if (this.emailLimit <= 0) {
        return 0
      }
      return Math.round((this.emailUsed / this.emailLimit) * 100)
    },
    emailBarClass () {
      if (this.emailPercent >= 100) {
        return 'progress-bar-danger'
      }
      if (this.emailPercent >= 80) {
        return 'progress-bar-warning'
      }
      return 'progress-bar-success'
    },
    emailText () {
      if (!this.emailEnabled) {
        return 'Relay off'
      }
      return this.emailUsed + ' of ' + this.emailLimit + ' emails this month'
    },
    nearLimit () {
      return (this.relayEnabled && this.relayPercent >= 80) ||
        (this.emailEnabled && this.emailPercent >= 80)
    },
    relayPercent () {
      if (this.relayLimit <= 0) {
        return 0
      }
      return Math.round((this.relayUsed / this.relayLimit) * 100)
    },
    relayBarClass () {
      if (this.relayPercent >= 100) {
        return 'progress-bar-danger'
      }
      if (this.relayPercent >= 80) {
        return 'progress-bar-warning'
      }
      return 'progress-bar-success'
    },
    relayText () {
      if (!this.relayEnabled) {
        return 'Relay off'
      }
      return gb(this.relayUsed) + ' of ' + gb(this.relayLimit) + ' this month'
    }
  },
  methods: {
    loadRelayUsage: function () {
      axios.get('/api/relay/usage')
        .then(response => {
          this.relayUsed = response.data.data.used_bytes
          this.relayLimit = response.data.data.limit_bytes
          this.relayOn = response.data.data.enabled === true
        })
        .catch(_ => {
          this.relayOn = false
        })
      axios.get('/api/mail/usage')
        .then(response => {
          this.emailUsed = response.data.data.used_messages
          this.emailLimit = response.data.data.limit_messages
          this.emailOn = response.data.data.enabled === true
        })
        .catch(_ => {
          this.emailOn = false
        })
    },
    timestamp: function (ds, today) {
      return niceTimestamp(ds, today)
    },
    reload: function () {
      axios.get('/api/domains')
        .then(response => {
          const domains = response.data.data
          if (domains.length > 0) {
            this.hasDomains = true
            let group = []
            const groups = []
            domains.forEach(domain => {
              const converted = convert(domain)
              group.push(converted)
              if (converted.name_servers) {
                this.checkNameServers(converted)
              }
              if (group.length === 2) {
                groups.push(group)
                group = []
              }
            })
            if (group.length > 0) {
              groups.push(group)
            }
            this.domainGroups = groups
          } else {
            this.hasDomains = false
          }
        })
        .catch(err => {
          if (err.response.status === 401) {
            this.$router.push('/login')
          } else {
            this.$router.push('/error')
          }
        })
    },
    checkNameServers: function (domain) {
      domain.ns_check_state = 'checking'
      domain.ns_check_actual = []
      axios.get('/api/domain/check_nameservers', { params: { domain: domain.name } })
        .then(response => {
          const result = response.data.data
          domain.ns_check_state = result.matched ? 'matched' : 'mismatched'
          domain.ns_check_actual = result.actual || []
        })
        .catch(_ => {
          domain.ns_check_state = 'mismatched'
          domain.ns_check_actual = []
        })
    },
    domainDeleteConfirm: function (domainName) {
      this.domainToDelete = domainName
      this.deleteConfirmationVisible = true
    },
    domainDelete: function () {
      this.deleteConfirmationVisible = false
      axios.delete('/api/domain', { params: { domain: this.domainToDelete } })
        .then(_ => {
          this.reload()
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
}
</script>
<style>
.circle_offline {
  width: 20px;
  height: 20px;
  -webkit-border-radius: 10px;
  -moz-border-radius: 10px;
  border-radius: 10px;
  background: red;
}

.circle_online {
  width: 20px;
  height: 20px;
  -webkit-border-radius: 10px;
  -moz-border-radius: 10px;
  border-radius: 10px;
  background: green;
}

.invisible {
  display: none;
}

.usage-panel {
  margin-bottom: 20px;
}
.usage-row {
  margin-bottom: 14px;
}
.usage-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 5px;
}
.usage-name {
  font-weight: 600;
}
.usage-value {
  color: #666;
  font-variant-numeric: tabular-nums;
}
.usage-off {
  color: #666;
  font-size: 13px;
  line-height: 1.5;
}
.usage-upgrade {
  display: inline-block;
  margin-top: 4px;
}
</style>
