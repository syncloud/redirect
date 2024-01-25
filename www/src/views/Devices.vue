<template>
  <div class="container">
    <div id="has_domains" v-bind:class="{ invisible:  !hasDomains}">
      <h2>Devices</h2>
      <br/>
      <div v-for="(domains, group_index) in domainGroups" :key="group_index">
        <div class="row">
          <div v-for="(domain, index) in domains" :key="index">
          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  <h3 style="margin-top: 5px; margin-bottom: 5px">
                    <span id="name">
                      {{ domain.name }}
                    </span>
                    <span class="pull-right" :class="{ 'circle_online': domain.online, 'circle_offline': !domain.online }"></span>
                  </h3>
                </div>
              </div>
              <ul class="list-group">
                <li class="list-group-item clearfix">
                  <h3 id="title" class="pull-left" style="margin-top: 5px; margin-bottom: 5px">{{ domain.device_title }}</h3>

                  <button type="button" class="btn btn-default pull-right" id="delete" @click="domainDeleteConfirm(domain.name)">
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
                  <div v-for="(name_server, name_server_index) in domain.name_servers" :key="name_server_index">
                    <code>{{ name_server }}</code>
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
    <div id="no_domains" v-bind:class="{ invisible:  hasDomains}">
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

  <Dialog :visible="deleteConfirmationVisible" @cancel="deleteConfirmationVisible = false"
                id="delete_confirmation" @confirm="domainDelete">
    <template v-slot:title>
      Deactivate {{ domainToDelete }}
    </template>
    <template v-slot:text>
      Device will be unlinked from the domain.<br>Domain will be released and might be taken by other user.<br>Proceed with caution!
    </template>
  </Dialog>
</template>

<script>
import axios from 'axios'
import moment from 'moment'
import Dialog from '../components/Dialog.vue'

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
  return domain
}

export default {
  name: 'Devices',
  components: {
    Dialog
  },
  props: {
    checkUserSession: Function
  },
  data () {
    return {
      hasDomains: Boolean,
      domainGroups: Array,
      domainToDelete: '',
      deleteConfirmationVisible: false
    }
  },
  mounted () {
    this.reload()
  },
  methods: {
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
              group.push(convert(domain))
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
</style>
