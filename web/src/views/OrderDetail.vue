<template>
  <div class="sc-page">
    <h1 class="sc-h1">Order</h1>

    <div v-if="loading" class="sc-card" data-testid="order-loading">Loading the order.</div>

    <template v-else-if="order">
      <div class="sc-card" data-testid="order-summary">
        <div class="order-head">
          <span class="order-name">{{ order.device }}, {{ order.option }}</span>
          <span class="order-status" :class="`status-${order.status}`" data-testid="order-status">
            {{ statusLabel(order.status) }}
          </span>
        </div>
        <div class="order-line">
          <span>Ordered {{ order.ordered }}</span>
          <span data-testid="order-total">{{ order.total }}</span>
        </div>
        <p class="sc-muted order-reference" data-testid="order-reference">
          Order {{ order.number }}
        </p>
        <p v-if="order.name" class="order-ship" data-testid="order-ship">
          {{ order.name }}, {{ order.address }}, {{ order.city }} {{ order.postcode }}, {{ order.country }}
        </p>
      </div>

      <p v-if="history.length === 0" class="sc-muted" data-testid="order-history">
        Nothing has happened yet.
      </p>
      <ol v-else class="history" data-testid="order-history">
        <li
          v-for="(event, index) in history"
          :key="index"
          class="event"
          data-testid="order-event"
        >
          <div class="event-head">
            <span class="event-status" :class="`status-${event.status}`">
              {{ statusLabel(event.status) }}
            </span>
            <span class="sc-muted event-at">{{ event.at }}</span>
          </div>
          <p v-if="event.comment" class="event-comment" data-testid="order-event-comment">
            {{ event.comment }}
          </p>
        </li>
      </ol>

      <div v-if="admin" class="sc-form" data-testid="order-admin">
        <div class="field">
          <textarea
            id="order-comment"
            v-model="comment"
            rows="3"
            placeholder="Anything the buyer should know, included in the email"
            data-testid="order-comment"
          />
        </div>
        <div class="status-row">
          <select id="order-new-status" v-model="status" data-testid="order-new-status">
            <option v-for="each in statuses" :key="each" :value="each">{{ statusLabel(each) }}</option>
          </select>
          <el-button
            type="primary"
            size="large"
            :loading="saving"
            data-testid="order-save-status"
            @click="save"
          >
            Save
          </el-button>
        </div>
      </div>
    </template>

    <p v-if="error" class="sc-warn" data-testid="order-error">{{ error }}</p>
  </div>
</template>

<script>
import axios from 'axios'
import { STATUS_LABELS, STATUSES } from '../data/orderStatus'

export default {
  name: 'OrderDetail',
  props: {
    admin: Boolean
  },
  data () {
    return {
      order: null,
      history: [],
      status: '',
      comment: '',
      loading: true,
      saving: false,
      error: '',
      statuses: STATUSES
    }
  },
  mounted () {
    this.load()
  },
  methods: {
    statusLabel (status) {
      return STATUS_LABELS[status] || status
    },
    number () {
      return this.$route.params.number
    },
    load () {
      axios.get('/api/device/order', { params: { number: this.number() } })
        .then(response => {
          this.order = response.data.data.order
          this.history = response.data.data.history
          this.status = this.order.status
          this.loading = false
        })
        .catch(this.onError)
    },
    save () {
      this.error = ''
      this.saving = true
      axios.post('/api/device/order/status', {
        number: Number(this.number()),
        status: this.status,
        comment: this.comment
      })
        .then(() => {
          this.comment = ''
          this.saving = false
          this.load()
        })
        .catch(error => {
          this.saving = false
          this.onError(error)
        })
    },
    onError (error) {
      this.loading = false
      const status = error.response && error.response.status
      if (status === 401) {
        this.$router.push(`/login?next=/orders/${this.number()}`)
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

.order-reference,
.order-ship {
  margin: 6px 0 0;
  font-size: 0.9rem;
}

.sc-card {
  margin-bottom: 12px;
}

.history {
  margin: 0 0 12px;
  padding: 0;
  list-style: none;
}

.event {
  border: 1px solid var(--sc-border);
  border-radius: 10px;
  padding: 10px 12px;
  margin-bottom: 8px;
}

.event:last-child {
  margin-bottom: 0;
}

.event-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.event-status {
  border-radius: 999px;
  padding: 2px 10px;
  font-size: 0.78rem;
  font-weight: 600;
  white-space: nowrap;
  background: var(--sc-surface-2);
  color: var(--sc-ink-2);
}

.event-at {
  font-size: 0.82rem;
  white-space: nowrap;
}

.event-comment {
  margin: 8px 0 0;
  white-space: pre-wrap;
}

.field {
  margin-bottom: 12px;
}

.status-row {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-row #order-new-status {
  flex: 1 1 auto;
  min-width: 0;
  height: var(--el-component-size-large, 40px);
  padding: 0 14px;
}

.status-row .el-button {
  flex: 0 0 auto;
  margin: 0;
}

#order-comment,
#order-new-status {
  width: 100%;
  font-family: var(--sc-font);
  font-size: 1rem;
  padding: 10px 14px;
  border: 1px solid var(--sc-border);
  border-radius: var(--sc-control-radius);
  background: var(--sc-field-bg);
  color: var(--sc-ink);
}
</style>
