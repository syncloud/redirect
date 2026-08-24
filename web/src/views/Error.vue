<template>
  <div class="sc-page">
    <div class="sc-card sc-prose">
      <h1 class="sc-h1">Something went wrong</h1>
      <p id="error" data-testid="error-message">
        Sorry, something went wrong and we are fixing it.
        If this is urgent, contact us at
        <a href="mailto:support@syncloud.it?Subject=Site%20error">support@syncloud.it</a>.
      </p>
      <router-link class="sc-btn sc-btn-inline" to="/" data-testid="error-home">Back to devices</router-link>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Error'
}
</script>
<style>
</style>
