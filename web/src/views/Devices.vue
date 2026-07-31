<template>
  <div class="container">
    <div id="has_domains" v-bind:class="{ invisible:  !hasDomains}">
      <h2>Devices</h2>
      <br/>
      <div class="usage-panel" data-testid="relay-usage">
        <div class="usage-row" data-testid="usage-traffic">
          <div class="usage-head">
            <span class="usage-name">Device access</span>
            <span class="usage-value" data-testid="relay-usage-text">{{ relayText }}</span>
          </div>
          <div v-if="relayEnabled" class="progress">
            <div class="progress-bar" :class="relayBarClass" role="progressbar"
                 :style="{ width: Math.min(relayPercent, 100) + '%' }"
                 data-testid="relay-usage-bar">
              {{ relayPercent }}%
            </div>
          </div>
          <div v-else class="usage-off" data-testid="usage-traffic-off">
            Your device is reached directly, with no traffic limit. Turn the relay on if your
            connection has no public address.
          </div>
        </div>

        <div class="usage-row" data-testid="usage-email">
          <div class="usage-head">
            <span class="usage-name">Email sending</span>
            <span class="usage-value" data-testid="usage-email-text">{{ emailText }}</span>
          </div>
          <div v-if="emailEnabled" class="progress">
            <div class="progress-bar" :class="emailBarClass" role="progressbar"
                 :style="{ width: Math.min(emailPercent, 100) + '%' }"
                 data-testid="usage-email-bar">
              {{ emailPercent }}%
            </div>
          </div>
          <div v-else class="usage-off" data-testid="usage-email-off">
            Your device sends email directly, with no limit. Turn the relay on if providers
            reject mail from your address.
          </div>
        </div>

        <router-link v-if="nearLimit" to="/account" data-testid="relay-upgrade" class="usage-upgrade">
          Approaching your limit — upgrade for more
        </router-link>
      </div>
      <div v-for="(domains, group_index) in domainGroups" :key="group_index">
        <div class="row">
          <div v-for="(domain, index) in domains" :key="index">
          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  <h3 style="margin-top: 5px; margin-bottom: 5px">
                    <span id="name" data-testid="domain-name">
                      {{ domain.name }}
                    </span>
                    <span class="pull-right" :class="{ 'circle_online': domain.online, 'circle_offline': !domain.online }"></span>
                  </h3>
                </div>
              </div>
              <ul class="list-group">
                <li class="list-group-item clearfix">
                  <h3 id="title" data-testid="device-title" class="pull-left" style="margin-top: 5px; margin-bottom: 5px">{{ domain.device_title }}</h3>

                  <button type="button" class="btn btn-default pull-right" id="delete" data-testid="device-delete" @click="domainDeleteConfirm(domain.name)">
                    <span class="glyphicon glyphicon-remove" aria-hidden="true"></span> Deactivate
                  </button>

                </li>
                <li class="list-group-item clearfix">
                  <span>Domain Address: </span>
                  <a v-if="domain.has_domain_address" :href="domain.domain_address">{{ domain.domain_address }}</a>
                  <span v-if="!domain.has_domain_address">Not mapped</span>
                </li>
                <li class="list-group-item clearfix">
                  <span>External Address: </span>
                  <a id="external_address" v-if="domain.has_external_address" :href="domain.external_address">{{ domain.external_address }}</a>
                  <span v-if="!domain.has_external_address">Not provided</span>
                </li>
                <li class="list-group-item clearfix">
                  <span>Internal Address: </span>
                  <a id="internal_address" v-if="domain.has_internal_address" :href="domain.internal_address">{{ domain.internal_address }}</a>
                  <span v-if="!domain.has_internal_address">Not provided</span>
                </li>
                <li class="list-group-item clearfix">
                  <span>IPv6 Address: </span>

                  <a id="ipv6_address" v-if="domain.has_ipv6_address" :href="domain.ipv6_address">{{ domain.ipv6_address }}</a>
                  <span id="ipv6_address_not_available" v-if="!domain.has_ipv6_address">Not provided</span>
                </li>
                <li class="list-group-item clearfix" v-if="domain.name_servers">
                  <span>Name Servers: </span>
                  <span
                    v-if="domain.ns_check_state === 'matched'"
                    class="label label-success"
                    data-testid="ns-status-matched"
                    style="margin-left: 5px"
                  >Matched</span>
                  <span
                    v-if="domain.ns_check_state === 'mismatched'"
                    class="label label-danger"
                    data-testid="ns-status-mismatched"
                    style="margin-left: 5px"
                  >Not set at registrar</span>
                  <span
                    v-if="domain.ns_check_state === 'checking'"
                    class="label label-default"
                    data-testid="ns-status-checking"
                    style="margin-left: 5px"
                  >Checking...</span>
                  <button
                    type="button"
                    class="btn btn-default btn-xs pull-right"
                    data-testid="ns-revalidate"
                    :disabled="domain.ns_check_state === 'checking'"
                    @click="checkNameServers(domain)"
                  >Revalidate</button>
                  <div v-for="(name_server, name_server_index) in domain.name_servers" :key="name_server_index">
                    <code>{{ name_server }}</code>
                  </div>
                  <div
                    v-if="domain.ns_check_state === 'mismatched' && domain.ns_check_actual && domain.ns_check_actual.length > 0"
                    style="padding-top: 5px; font-size: 90%"
                  >
                    Currently set at registrar:
                    <div v-for="(ns, i) in domain.ns_check_actual" :key="i">
                      <code>{{ ns }}</code>
                    </div>
                  </div>
                </li>
                <li class="list-group-item clearfix">
                  <span>Updated: {{ domain.nice_last_update }}</span>
                </li>
              </ul>
            </div>
          </div>
        </div>
        </div>
      </div>
    </div>
    <div id="no_domains" data-testid="no-devices" v-bind:class="{ invisible:  hasDomains}">
      <div class="row">
        <div class="col-2 col-md-2 col-sm-2 col-lg-2"><span></span></div>
        <div class="col-8 col-md-8 col-sm-8 col-lg-8">
          <div class="jumbotron" style="margin: 40px; padding: 30px">
            <h1>No Devices</h1>
            <p>You do not have any activated devices.<br/>Buy or build your first Syncloud device and activate it.</p>
            <br/>
            <p style="text-align:center;">
              <a class="btn btn-primary btn-lg" href="https://www.syncloud.org" role="button">Learn more</a>
            </p>
          </div>
        </div>
        <div class="col-2 col-md-2 col-sm-2 col-lg-2"><span></span></div>
      </div>
    </div>
  </div>

  <CustomDialog :visible="deleteConfirmationVisible" @cancel="deleteConfirmationVisible = false"
                id="delete_confirmation" @confirm="domainDelete">
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
      domainGroups: Array,
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
