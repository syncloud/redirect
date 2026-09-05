<template>
  <div class="sc-page">
    <h1 class="sc-h1">Orders</h1>

    <div v-if="loading" class="sc-card" data-testid="orders-loading">Loading your orders.</div>

    <template v-else>
      <div v-if="mine.length === 0" class="sc-card" data-testid="orders-empty">
        <p>You have not ordered anything yet.</p>
        <router-link class="sc-btn" to="/shop" data-testid="orders-shop-link">Visit the shop</router-link>
      </div>

      <router-link
        v-for="order in mine"
        :key="order.reference"
        :to="`/orders/${order.reference}`"
        class="sc-card order"
        data-testid="order"
      >
        <div class="order-head">
          <span class="order-name">{{ order.device }}, {{ order.option }}</span>
          <span class="order-status" :class="`status-${order.status}`" data-testid="order-status">
            {{ statusLabel(order.status) }}
          </span>
        </div>
        <div class="order-line">
          <span>{{ order.ordered }}</span>
          <span data-testid="order-total">{{ order.total }}</span>
        </div>
        <p class="sc-muted order-reference" data-testid="order-reference">
          Reference {{ order.reference }}
        </p>
      </router-link>

    </template>

    <p v-if="error" class="sc-warn" data-testid="orders-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import { STATUS_LABELS } from '../data/orderStatus'

export default {
  name: 'Orders',
  data () {
    return {
      mine: [],
      loading: true,
      error: ''
    }
  },
  mounted () {
    this.load()
  },
  methods: {
    statusLabel (status) {
      return STATUS_LABELS[status] || status
    },
    load () {
      axios.get('/api/device/orders')
        .then(response => {
          this.mine = response.data.data
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
.order {
  display: block;
  margin-bottom: 12px;
  color: inherit;
  text-decoration: none;
}

.order:hover {
  border-color: var(--sc-primary);
}

.order-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.order-name {
  font-weight: 600;
}

.order-status {
  border-radius: 999px;
  padding: 2px 10px;
  font-size: 0.8rem;
  font-weight: 600;
  background: var(--sc-surface-2);
  color: var(--sc-ink-2);
}

.status-sent {
  background: var(--sc-primary);
  color: #fff;
}

.order-line {
  display: flex;
  justify-content: space-between;
  margin-top: 8px;
}

.order-reference {
  margin: 6px 0 0;
  font-size: 0.85rem;
}

</style>
