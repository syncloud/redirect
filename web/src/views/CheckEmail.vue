<template>
  <div class="sc-auth">
    <div class="sc-auth-wordmark">SYNCLOUD</div>
    <div class="sc-auth-card">
      <img class="sc-auth-logo" src="/logo.svg" alt="Syncloud">
      <h2 class="sc-auth-title" data-testid="check-email-complete">Check your email</h2>
      <p class="sc-auth-message">
        We have sent you a link to confirm your address.
        Open it to finish creating your account.
      </p>
      <p class="sc-auth-note">
        Nothing arrived? Check your spam folder, or
        <router-link to="/register" data-testid="check-email-register">try registering again</router-link>.
      </p>
    </div>
  </div>
</template>

<script>
export default {
  name: 'CheckEmail'
}
</script>
