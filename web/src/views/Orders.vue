<template>
  <div class="sc-page">
    <h1 class="sc-h1">Orders</h1>

    <div v-if="loading" class="sc-card" data-testid="orders-loading">Loading your orders.</div>

    <template v-else>
      <div v-if="mine.length === 0" class="sc-card" data-testid="orders-empty">
        <p>You have not ordered anything yet.</p>
        <router-link class="sc-btn" to="/shop" data-testid="orders-shop-link">Visit the shop</router-link>
      </div>

      <div
        v-for="order in mine"
        :key="order.reference"
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
      </div>

      <div v-if="admin" class="sc-card" data-testid="orders-admin">
        <h2 class="sc-h2">All orders</h2>
        <p class="sc-muted">Every paid order. Only administrators can see this.</p>

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
              <tr v-for="order in all" :key="order.reference" :data-testid="`admin-order-${order.reference}`">
                <td data-label="Ordered">{{ order.ordered }}</td>
                <td data-label="Account">{{ order.email }}</td>
                <td data-label="Item">{{ order.device }}, {{ order.option }}</td>
                <td data-label="Ship to" class="cell-stacked">
                  <span>{{ order.name }}</span>
                  <span class="sc-muted">{{ order.address }}, {{ order.city }} {{ order.postcode }}, {{ order.country }}</span>
                </td>
                <td data-label="Total">{{ order.total }}</td>
                <td data-label="Status">
                  <select
                    class="admin-status"
                    :value="order.status"
                    :data-testid="`admin-status-${order.reference}`"
                    @change="setStatus(order, $event.target.value)"
                  >
                    <option v-for="status in statuses" :key="status" :value="status">
                      {{ statusLabel(status) }}
                    </option>
                  </select>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <p v-if="error" class="sc-warn" data-testid="orders-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'

const LABELS = {
  ordered: 'Ordered',
  sent: 'Sent'
}

export default {
  name: 'Orders',
  data () {
    return {
      mine: [],
      all: [],
      admin: false,
      loading: true,
      error: '',
      statuses: Object.keys(LABELS)
    }
  },
  mounted () {
    this.load()
  },
  methods: {
    statusLabel (status) {
      return LABELS[status] || status
    },
    load () {
      axios.get('/api/user')
        .then(response => {
          this.admin = response.data.data.admin === true
          return axios.get('/api/device/orders')
        })
        .then(response => {
          this.mine = response.data.data
          return this.admin ? axios.get('/api/device/orders/all') : null
        })
        .then(response => {
          if (response) {
            this.all = response.data.data
          }
          this.loading = false
        })
        .catch(this.onError)
    },
    setStatus (order, status) {
      this.error = ''
      axios.post('/api/device/order/status', { reference: order.reference, status })
        .then(() => {
          order.status = status
          const own = this.mine.find(each => each.reference === order.reference)
          if (own) {
            own.status = status
          }
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
  margin-bottom: 12px;
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

.admin-scroll {
  overflow-x: auto;
}

.admin-table {
  width: 100%;
  margin-top: 12px;
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

.admin-status {
  font-family: var(--sc-font);
  font-size: 0.9rem;
  padding: 4px 6px;
  border: 1px solid var(--sc-border);
  border-radius: var(--el-border-radius-base, 4px);
  background: var(--sc-field-bg);
  color: var(--sc-ink);
}
</style>
