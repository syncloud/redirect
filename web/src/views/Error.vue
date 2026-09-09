<template>
  <div class="sc-page">
    <div class="sc-card sc-prose">
      <h1 class="sc-h1">{{ $t('error.title') }}</h1>
      <p id="error" data-testid="error-message">
        {{ $t('error.message') }}
        <a href="mailto:support@syncloud.it?Subject=Site%20error">support@syncloud.it</a>.
      </p>
      <router-link class="sc-btn sc-btn-inline" to="/" data-testid="error-home">{{ $t('error.home') }}</router-link>
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
