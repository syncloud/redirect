<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <div class="sc-auth-card">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="check-email-complete">{{ $t('checkEmail.heading') }}</h2>
      <p class="sc-auth-message">
        {{ $t('checkEmail.message') }}
      </p>
      <p class="sc-auth-note">
        {{ $t('checkEmail.noteBefore') }}
        <router-link to="/register" data-testid="check-email-register">{{ $t('checkEmail.noteLink') }}</router-link>.
      </p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CheckEmail'
}
</script>
