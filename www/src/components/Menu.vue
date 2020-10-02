<template>
  <nav class="navbar navbar-default navbar-static-top" role="navigation">
    <div class="container">

      <div class="navbar-header">
        <button id="navbar" type="button" class="navbar-toggle" data-toggle="collapse" data-target=".navbar-collapse"
        style="margin: 20px">
          <span class="sr-only">Toggle navigation</span>
          <span class="icon-bar"></span>
          <span class="icon-bar"></span>
          <span class="icon-bar"></span>
        </button>

        <a class="navbar-brand" href="#"><img src="../assets/logo.png" style="display: inline" alt="syncloud"/>
          <span>SYNCLOUD</span>
        </a>
      </div>

      <div class="collapse navbar-collapse">
        <ul class="nav navbar-nav navbar-right" :class="{ invisible:  !loggedIn}">
          <li>
            <span style="padding-right: 5px">{{ email }}</span>
            <button id="logout" class="btn btn-default" @click="logout">
              <span class="glyphicon glyphicon-log-out"></span> Log out
            </button>
          </li>
        </ul>
        <ul class="nav navbar-nav">
          <li id="account" :class="{ invisible: loggedIn === undefined || loggedIn === false, active: activeTab === '/account'}" >
            <router-link to="/account" >Account</router-link>
          </li>
          <li id="devices" :class="{ invisible: loggedIn === undefined || loggedIn === false, active: activeTab === '/' }">
            <router-link to="/" >Devices</router-link>
          </li>
          <li id="register" :class="{ invisible: loggedIn === undefined || loggedIn === true, active: activeTab === '/register' }">
            <router-link to="/register" >Register</router-link>
          </li>
          <li id="login" :class="{ invisible: loggedIn === undefined || loggedIn === true, active: activeTab === '/login' }">
            <router-link to="/login" >Log in</router-link>
          </li>
        </ul>
      </div>

    </div>
  </nav>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Menu',
  props: {
    activeTab: String,
    email: String,
    loggedIn: Boolean,
    onLogout: Function
  },
  methods: {
    logout: function (_) {
      axios.post('/api/logout')
        .then(_ => {
          this.onLogout()
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
}

</script>
<style>
.navbar-brand {
  line-height: 80px !important;
  height: 80px !important;
  padding-top: 0 !important;
  font-size: 32px !important;
}

.navbar-inverse .navbar-brand {
  color: #fff !important;
}

.navbar-brand span {
  padding-left: 10px !important;
}

.navbar-nav > li > span {
  padding-left: 15px !important;
}

.navbar-nav li, .navbar-nav li a {
  padding-top: 0 !important;
  font-size: 18px !important;
}

.navbar-nav li, .navbar-nav li a {
  line-height: 80px !important;
  height: 80px !important;
}

@media (max-width: 767px) {
  .navbar-nav li, .navbar-nav li a {
    line-height: 30px !important;
    height: 30px !important;
  }
}
.invisible {
  display: none !important;
}
</style>
