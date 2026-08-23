<template>
  <button
    class="sc-icon-btn"
    type="button"
    aria-label="Toggle dark mode"
    data-testid="theme-toggle"
    @click="toggle"
  >{{ isDark ? '☀' : '☾' }}</button>
</template>

<script>
import { computed } from 'vue'
import { theme, toggleTheme } from '../theme'

export default {
  name: 'ThemeToggle',
  setup () {
    return {
      isDark: computed(() => theme.value === 'dark'),
      toggle: toggleTheme
    }
  }
}
</script>
