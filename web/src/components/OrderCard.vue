<template>
  <router-link
    :to="`/orders/${order.number}`"
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

    <template v-if="account">
      <p class="order-account" data-testid="order-account">{{ order.email }}</p>
      <p class="sc-muted order-ship" data-testid="order-ship">
        {{ order.name }}, {{ order.address }}, {{ order.city }} {{ order.postcode }}, {{ order.country }}
      </p>
    </template>

    <p class="sc-muted order-number" data-testid="order-reference">
      Order {{ order.number }}
    </p>
  </router-link>
</template>

<script>
import { STATUS_LABELS } from '../data/orderStatus'

export default {
  name: 'OrderCard',
  props: {
    order: {
      type: Object,
      required: true
    },
    account: Boolean
  },
  methods: {
    statusLabel (status) {
      return STATUS_LABELS[status] || status
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
  white-space: nowrap;
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
  gap: 12px;
  margin-top: 8px;
}

.order-account {
  margin: 8px 0 0;
  word-break: break-word;
}

.order-ship,
.order-number {
  margin: 4px 0 0;
  font-size: 0.85rem;
}
</style>
