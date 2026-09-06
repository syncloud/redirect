<template>
  <div class="sc-page">
    <h1 class="sc-h1">All orders</h1>
    <p class="sc-lead">Every paid order. Only administrators can see this.</p>

    <div v-if="loading" class="sc-card" data-testid="admin-orders-loading">Loading orders.</div>

    <div v-else-if="orders.length === 0" class="sc-card" data-testid="admin-orders-empty">
      Nobody has ordered anything yet.
    </div>

    <template v-else>
      <OrderCard
        v-for="order in orders"
        :key="order.number"
        :order="order"
        account
      />
    </template>

    <p v-if="error" class="sc-warn" data-testid="admin-orders-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import OrderCard from '../components/OrderCard.vue'

export default {
  name: 'AdminOrders',
  components: { OrderCard },
  data () {
    return {
      orders: [],
      loading: true,
      error: ''
    }
  },
  mounted () {
    axios.get('/api/device/orders/all')
      .then(response => {
        this.orders = response.data.data
        this.loading = false
      })
      .catch(this.onError)
  },
  methods: {
    onError (error) {
      this.loading = false
      const status = error.response && error.response.status
      if (status === 401) {
        this.$router.push('/login?next=/admin/orders')
        return
      }
      if (status === 403) {
        this.$router.push('/orders')
        return
      }
      this.error = 'Something went wrong.'
    }
  }
}
</script>
