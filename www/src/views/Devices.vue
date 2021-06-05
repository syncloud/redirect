<template>
  <div class="container">
    <div id="has_domains" v-bind:class="{ invisible:  !hasDomains}">
      <h2>Devices</h2>
      <br/>
      <div v-for="(domains, group_index) in domainGroups" :key="group_index">
        <div class="row">
          <div v-for="(domain, index) in domains" :key="index">
          <div class="modal fade" v-bind:id="'modalDeactivateDomain_' + index" tabindex="-1" role="dialog" v-bind:aria-labelledby="'modalDeactivateDomain_' + index">
            <div class="modal-dialog" role="document">
              <div class="modal-content">
                <div class="modal-header">
                  <button type="button" class="close" data-dismiss="modal" aria-label="Close"><span aria-hidden="true">&times;</span></button>
                  <h4 class="modal-title">Deactivate {{ domain.user_domain }}</h4>
                </div>
                <div class="modal-body">
                  Device will be unlinked from the domain. Domain will be released and might be taken by other user. Proceed with caution!
                </div>
                <div class="modal-footer">
                  <button type="button" class="btn btn-default" data-dismiss="modal">Cancel</button>
                  <button type="button" class="btn btn-danger" data-dismiss="modal" @click="domain_delete(domain.user_domain)">Deactivate</button>
                </div>
              </div>
            </div>
          </div>

          <div class="col-6 col-md-6 col-sm-6 col-lg-6">
            <div class="panel panel-default">
              <div class="panel-heading">
                <div class="panel-title">
                  <h3 style="margin-top: 5px; margin-bottom: 5px">
                    <span id="name">
                      {{ domain.domain }}
                    </span>
                    <span class="pull-right" :class="{ 'circle_online': domain.online, 'circle_offline': !domain.online }"></span>
                  </h3>
                </div>
              </div>
              <ul class="list-group">
                <li class="list-group-item clearfix">
                  <h3 id="title" class="pull-left" style="margin-top: 5px; margin-bottom: 5px">{{ domain.device_title }}</h3>

                  <button type="button" class="btn btn-default pull-right" data-toggle="modal" v-bind:data-target="'#modalDeactivateDomain_' + index">
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
                  <a v-if="domain.has_external_address" :href="domain.external_address">{{ domain.external_address }}</a>
                  <span v-if="!domain.has_external_address">Not provided</span>
                </li>
                <li class="list-group-item clearfix">
                  <span>Internal Address: </span>
                  <a v-if="domain.has_internal_address" :href="domain.internal_address">{{ domain.internal_address }}</a>
                  <span v-if="!domain.has_internal_address">Not provided</span>
                </li>
                <li class="list-group-item clearfix">
                  <span>IPv6 Address: </span>

                  <a v-if="domain.has_ipv6_address" :href="domain.ipv6_address">{{ domain.ipv6_address }}</a>
                  <span v-if="!domain.has_ipv6_address">Not provided</span>
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
              <a class="btn btn-primary btn-lg" href="http://syncloud.org" role="button">Learn more</a>
            </p>
          </div>
        </div>
        <div class="col-2 col-md-2 col-sm-2 col-lg-2"><span></span></div>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import moment from 'moment'
import querystring from 'querystring'

function sameDay (date1, date2) {
  return (date1.getDate() === date2.getDate() &&
    date1.getMonth() === date2.getMonth() &&
    date1.getFullYear() === date2.getFullYear())
}

function fullUrl (address, port) {
  let result = 'https://' + address
  if (port !== 443) {
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
  domain.domain_address = fullUrl(domain.domain, domain.domain_address_port)
  domain.has_domain_address = domain.domain !== null
  domain.external_address = fullUrl(domain.ip, domain.web_port)
  domain.has_external_address = domain.ip !== null
  domain.internal_address = 'https://' + domain.local_ip
  domain.has_internal_address = domain.local_ip !== null
  domain.ipv6_address = 'https://[' + domain.ipv6 + ']'
  domain.has_ipv6_address = domain.ipv6 !== null
  domain.online = online(domain.last_update)
  domain.nice_last_update = niceTimestamp(domain.last_update, new Date())
  return domain
}

export default {
  name: 'Devices',
  props: {
    onLogin: Function
  },
  data () {
    return {
      hasDomains: Boolean,
      domainGroups: Array
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
    domain_delete: function (userDomain) {
      axios.post('api/domain_delete', querystring.stringify({ user_domain: userDomain }))
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
