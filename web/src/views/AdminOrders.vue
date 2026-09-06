<template>
  <div class="sc-page">
    <h1 class="sc-h1">All orders</h1>
    <p class="sc-lead">Every paid order. Only administrators can see this.</p>

    <div v-if="loading" class="sc-card" data-testid="admin-orders-loading">Loading orders.</div>

    <div v-else-if="orders.length === 0" class="sc-card" data-testid="admin-orders-empty">
      Nobody has ordered anything yet.
    </div>

    <div v-else class="sc-card">
      <div class="admin-scroll">
        <table class="admin-table" data-testid="orders-admin-table">
          <thead>
            <tr>
              <th>Ordered</th>
              <th>Account</th>
              <th>Item</th>
              <th>Ship to</th>
              <th>Total</th>
              <th>Status</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="order in orders" :key="order.number" :data-testid="`admin-order-${order.number}`">
              <td data-label="Ordered">{{ order.ordered }}</td>
              <td data-label="Account">{{ order.email }}</td>
              <td data-label="Item">{{ order.device }}, {{ order.option }}</td>
              <td data-label="Ship to" class="cell-stacked">
                <span>{{ order.name }}</span>
                <span class="sc-muted">{{ order.address }}, {{ order.city }} {{ order.postcode }}, {{ order.country }}</span>
              </td>
              <td data-label="Total">{{ order.total }}</td>
              <td data-label="Status">
                <router-link
                  :to="`/orders/${order.number}`"
                  :data-testid="`admin-open-${order.number}`"
                >{{ statusLabel(order.status) }}</router-link>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <p v-if="error" class="sc-warn" data-testid="admin-orders-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import { STATUS_LABELS } from '../data/orderStatus'

export default {
  name: 'AdminOrders',
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
    statusLabel (status) {
      return STATUS_LABELS[status] || status
    },
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

<style scoped>
.admin-scroll {
  overflow-x: auto;
}

.admin-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 0.9rem;
}

.admin-table th,
.admin-table td {
  border-bottom: 1px solid var(--sc-border);
  padding: 8px 10px 8px 0;
  text-align: left;
  vertical-align: top;
}

.admin-table th {
  font-weight: 600;
  color: var(--sc-muted);
  white-space: nowrap;
}

.cell-stacked {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

@media (max-width: 640px) {
  .admin-table,
  .admin-table tbody,
  .admin-table tr,
  .admin-table td {
    display: block;
  }

  .admin-table thead {
    display: none;
  }

  .admin-table tr {
    border: 1px solid var(--sc-border);
    border-radius: 10px;
    padding: 10px 12px;
    margin-bottom: 10px;
  }

  .admin-table td {
    display: flex;
    justify-content: space-between;
    align-items: baseline;
    gap: 14px;
    border-bottom: 0;
    padding: 5px 0;
    text-align: right;
  }

  .admin-table td::before {
    content: attr(data-label);
    font-weight: 600;
    color: var(--sc-muted);
    text-align: left;
    white-space: nowrap;
  }

  .admin-table td.cell-stacked {
    display: block;
    text-align: left;
  }

  .admin-table td.cell-stacked::before {
    display: block;
    margin-bottom: 2px;
  }
}
</style>
