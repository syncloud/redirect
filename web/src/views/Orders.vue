<template>
  <div class="sc-page">
    <h1 class="sc-h1">Orders</h1>

    <div v-if="loading" class="sc-card" data-testid="orders-loading">Loading your orders.</div>

    <template v-else>
      <div v-if="mine.length === 0 && unfinished.length === 0" class="sc-card" data-testid="orders-empty">
        <p>You have not ordered anything yet.</p>
        <router-link class="sc-btn" to="/shop" data-testid="orders-shop-link">Visit the shop</router-link>
      </div>

      <template v-if="unfinished.length > 0">
        <p class="unfinished-note" data-testid="orders-unfinished">
          You started {{ unfinished.length === 1 ? 'an order' : 'these orders' }} but never
          finished paying. Nothing has been charged. Tap to pick it up where you left off.
        </p>
        <OrderCard
          v-for="order in unfinished"
          :key="`unfinished-${order.number}`"
          :order="order"
          :to="`/shop?order=${order.number}`"
          badge="Not paid"
          :data-testid="`order-finish-${order.number}`"
        />
      </template>

      <OrderCard v-for="order in mine" :key="order.number" :order="order"/>

    </template>

    <p v-if="error" class="sc-warn" data-testid="orders-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import OrderCard from '../components/OrderCard.vue'

export default {
  name: 'Orders',
  components: { OrderCard },
  data () {
    return {
      mine: [],
      unfinished: [],
      loading: true,
      error: ''
    }
  },
  mounted () {
    this.load()
  },
  methods: {
    load () {
      axios.get('/api/device/orders')
        .then(response => {
          this.mine = response.data.data
          return axios.get('/api/device/orders/unfinished')
        })
        .then(response => {
          this.unfinished = response.data.data
          this.loading = false
        })
        .catch(this.onError)
    },
    onError (error) {
      this.loading = false
      if (error.response && error.response.status === 401) {
        this.$router.push('/login?next=/orders')
        return
      }
      this.error = error.response && error.response.data && error.response.data.message
        ? error.response.data.message
        : 'Something went wrong.'
    }
  }
}
</script>

<style scoped>
.unfinished-note {
  margin: 0 0 12px;
  color: var(--sc-ink-2);
}
</style>
