<template>
  <div class="sc-page">
    <div class="sc-card sc-prose" data-testid="privacy-content">
      <h1 class="sc-h1" data-testid="privacy-heading">{{ $t('privacy.title') }}</h1>
<div>

          {{ $t('privacy.intro') }}
          <h3>{{ $t('privacy.dataWeCollectTitle') }}</h3>
          {{ $t('privacy.dataWeCollectBody') }}
          <ul>
          <li>{{ $t('privacy.accountName') }}</li>
          <li>{{ $t('privacy.email') }}</li>
          <li>{{ $t('privacy.accountOperations') }}</li>
          </ul>

          {{ $t('privacy.deviceDataBody') }}
          <ul>
          <li>{{ $t('privacy.deviceDomain') }}</li>
          <li>{{ $t('privacy.deviceModel') }}</li>
          <li>{{ $t('privacy.deviceMac') }}</li>
          <li>{{ $t('privacy.deviceInternalIp') }}</li>
          <li>{{ $t('privacy.deviceExternalIp') }}</li>
          <li>{{ $t('privacy.deviceLiveFlag') }}</li>
          </ul>

          {{ $t('privacy.noTraffic') }}
          <h3>{{ $t('privacy.whenWeCollectTitle') }}</h3>
          {{ $t('privacy.whenWeCollectBody') }}
          <h3>{{ $t('privacy.whereWeStoreTitle') }}</h3>
          {{ $t('privacy.whereWeStoreBody') }}
          <h3>{{ $t('privacy.howWeUseTitle') }}</h3>
          {{ $t('privacy.howWeUseBody') }}
          <h3>{{ $t('privacy.doWeShareTitle') }}</h3>
          {{ $t('privacy.doWeShareBody') }}
          <h3>{{ $t('privacy.thirdPartyTitle') }}</h3>
          {{ $t('privacy.thirdPartyBody') }}

  </div>
    </div>
  </div>
</template>
<style>
@import '../style/form-center.css';
ul {
  margin: 10px 1px 1px -15px;
}
</style>
