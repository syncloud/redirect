<template>
  <button
    class="sc-icon-btn"
    type="button"
    aria-label="Toggle dark mode"
    data-testid="theme-toggle"
    @click="theme.toggle()"
  >{{ theme.isDark ? '☀' : '☾' }}</button>
</template>

<script>
import { useThemeStore } from '../stores/theme'

export default {
  name: 'ThemeToggle',
  setup () {
    return { theme: useThemeStore() }
  }
}
</script>
