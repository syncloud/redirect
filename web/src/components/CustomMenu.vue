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

        <a class="navbar-brand" href="#"><img src="/assets/logo.png" style="display: inline" alt="syncloud"/>
          <span>SYNCLOUD</span>
        </a>
      </div>

      <div class="collapse navbar-collapse">
        <ul class="nav navbar-nav navbar-right">
          <li style="display:flex; align-items:center; padding:0 8px;">
            <ThemeToggle/>
          </li>
        </ul>
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
            <router-link to="/register" data-testid="nav-register">Register</router-link>
          </li>
          <li id="login" :class="{ invisible: loggedIn === undefined || loggedIn === true, active: activeTab === '/login' }">
            <router-link to="/login" data-testid="nav-login">Log in</router-link>
          </li>
        </ul>
      </div>

    </div>
  </nav>
</template>

<script>
import axios from 'axios'
import ThemeToggle from './ThemeToggle.vue'

export default {
  name: 'CustomMenu',
  components: { ThemeToggle },
  props: {
    activeTab: String,
    email: String,
    loggedIn: Boolean,
    checkUserSession: Function
  },
  methods: {
    logout: function (_) {
      axios.post('/api/logout')
        .then(_ => {
          this.checkUserSession()
        })
        .catch(err => {
          console.log(err)
        })
    }
  }
}

</script>
<style>
.navbar-default.navbar-static-top {
  background: var(--sc-surface);
  border: none;
  border-bottom: 1px solid var(--sc-border);
  box-shadow: 0 1px 2px rgba(22, 50, 92, 0.04);
  margin-bottom: 28px;
}

.navbar-brand {
  display: flex !important;
  align-items: center;
  height: 64px !important;
  line-height: 64px !important;
  padding: 0 !important;
  font-size: 19px !important;
  font-weight: 700;
  letter-spacing: 4px;
  color: var(--sc-primary) !important;
}

.navbar-brand img {
  height: 34px;
}

.navbar-brand span {
  padding-left: 12px !important;
  color: var(--sc-primary) !important;
}

.navbar-nav > li > a {
  margin: 14px 4px;
  height: 36px !important;
  line-height: 36px !important;
  padding: 0 14px !important;
  font-size: 15px !important;
  font-weight: 600;
  color: var(--sc-muted) !important;
  border-radius: 10px;
  transition: color 0.2s ease, background 0.2s ease;
}

.navbar-nav > li > a:hover,
.navbar-nav > li > a:focus {
  color: var(--sc-primary) !important;
  background: var(--sc-primary-soft) !important;
}

.navbar-nav > li.active > a,
.navbar-nav > li.active > a:hover,
.navbar-nav > li.active > a:focus {
  color: var(--sc-primary) !important;
  background: var(--sc-primary-soft) !important;
}

.navbar-nav.navbar-right > li {
  display: flex;
  align-items: center;
  height: 64px;
}

.navbar-nav > li > span {
  padding: 0 12px 0 4px !important;
  font-size: 14px;
  color: var(--sc-muted);
}

#logout.btn {
  border: 1px solid var(--sc-border);
  border-radius: 10px;
  background: var(--sc-surface-2);
  color: var(--sc-ink);
  font-weight: 600;
  padding: 7px 16px;
  transition: color 0.2s ease, border-color 0.2s ease;
}

#logout.btn:hover,
#logout.btn:focus {
  color: var(--sc-primary);
  border-color: var(--sc-primary);
}

.navbar-toggle {
  margin: 15px 0 !important;
  border: 1px solid var(--sc-border) !important;
  border-radius: 10px;
}

.navbar-toggle:hover,
.navbar-toggle:focus {
  background: var(--sc-primary-soft) !important;
}

.navbar-toggle .icon-bar {
  background: var(--sc-ink) !important;
}

@media (max-width: 767px) {
  .navbar-default.navbar-static-top {
    margin-bottom: 20px;
  }

  .navbar-static-top > .container {
    padding-left: 20px;
    padding-right: 20px;
  }

  .navbar > .container .navbar-brand {
    margin-left: 0 !important;
  }

  .navbar-collapse {
    border-top: 1px solid var(--sc-border);
    box-shadow: 0 12px 24px -10px rgba(22, 50, 92, 0.16);
    padding: 6px 0;
    background: var(--sc-surface);
  }

  .navbar-nav {
    margin: 0;
  }

  .navbar-nav > li,
  .navbar-nav.navbar-right > li {
    float: none;
    display: block;
    height: auto;
    border-top: 1px solid #f2f5fa;
  }

  .navbar-collapse .navbar-nav:first-child > li:first-child {
    border-top: none;
  }

  .navbar-nav > li > a {
    margin: 0 !important;
    height: auto !important;
    line-height: 1.4 !important;
    padding: 15px 4px !important;
    border-radius: 0;
  }

  .navbar-nav.navbar-right > li {
    padding: 14px 4px;
  }

  .navbar-nav > li > span {
    display: block;
    padding: 0 0 10px 0 !important;
  }

  #logout.btn {
    display: block;
    width: 100%;
    text-align: center;
  }
}

.invisible {
  display: none !important;
}
</style>
