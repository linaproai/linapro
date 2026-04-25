<script setup lang="ts">
import { computed } from 'vue';

import { ProfileSecuritySetting } from '@vben/common-ui';
import { $t } from '#/locales';

const formSchema = computed(() => {
  return [
    {
      value: true,
      fieldName: 'accountPassword',
      label: $t('pages.profile.security.accountPassword'),
      description: $t('pages.profile.security.accountPasswordDesc'),
    },
    {
      value: true,
      fieldName: 'securityPhone',
      label: $t('pages.profile.security.securityPhone'),
      description: $t('pages.profile.security.securityPhoneDesc'),
    },
    {
      value: true,
      fieldName: 'securityQuestion',
      label: $t('pages.profile.security.securityQuestion'),
      description: $t('pages.profile.security.securityQuestionDesc'),
    },
    {
      value: true,
      fieldName: 'securityEmail',
      label: $t('pages.profile.security.securityEmail'),
      description: $t('pages.profile.security.securityEmailDesc'),
    },
    {
      value: false,
      fieldName: 'securityMfa',
      label: $t('pages.profile.security.securityMfa'),
      description: $t('pages.profile.security.securityMfaDesc'),
    },
  ];
});
</script>
<template>
  <ProfileSecuritySetting :form-schema="formSchema" />
</template>
