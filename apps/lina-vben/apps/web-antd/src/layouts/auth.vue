<script lang="ts" setup>
import { computed } from 'vue';

import { AuthPageLayout } from '@vben/layouts';
import { preferences } from '@vben/preferences';

import { $t } from '#/locales';
import {
  publicFrontendSettings,
  resolveWorkspaceAssetURL,
} from '#/runtime/public-frontend';

const appName = computed(() => preferences.app.name);
const logo = computed(() => preferences.logo.source);
const logoDark = computed(
  () => preferences.logo.sourceDark || preferences.logo.source,
);
const pageDescription = computed(
  () => publicFrontendSettings.auth.pageDesc || $t('authentication.pageDesc'),
);
const pageTitle = computed(
  () => publicFrontendSettings.auth.pageTitle || $t('authentication.pageTitle'),
);
const sloganImage = computed(() => {
  const source = publicFrontendSettings.auth.sloganImage;
  return source ? resolveWorkspaceAssetURL(source) : '';
});
</script>

<template>
  <AuthPageLayout
    :app-name="appName"
    :logo="logo"
    :logo-dark="logoDark"
    :page-description="pageDescription"
    :page-title="pageTitle"
    :slogan-image="sloganImage"
  >
    <!-- 自定义工具栏 -->
    <!-- <template #toolbar></template> -->
  </AuthPageLayout>
</template>
