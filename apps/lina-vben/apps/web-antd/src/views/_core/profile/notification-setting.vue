<script setup lang="ts">
import { computed } from 'vue';

import { ProfileNotificationSetting } from '@vben/common-ui';
import { $t } from '#/locales';

const formSchema = computed(() => {
  return [
    {
      value: true,
      fieldName: 'accountPassword',
      label: $t('pages.profile.notification.accountPassword'),
      description: $t('pages.profile.notification.accountPasswordDesc'),
    },
    {
      value: true,
      fieldName: 'systemMessage',
      label: $t('pages.profile.notification.systemMessage'),
      description: $t('pages.profile.notification.systemMessageDesc'),
    },
    {
      value: true,
      fieldName: 'todoTask',
      label: $t('pages.profile.notification.todoTask'),
      description: $t('pages.profile.notification.todoTaskDesc'),
    },
  ];
});
</script>
<template>
  <ProfileNotificationSetting :form-schema="formSchema" />
</template>
