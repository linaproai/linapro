<script lang="ts" setup>
import { About } from '@vben/common-ui';

import { $t } from '#/locales';
import { PROJECT_INFO } from '#/views/about/config';

defineOptions({ name: 'About' });
</script>

<template>
  <About
    :description="$t(PROJECT_INFO.descriptionKey)"
    :name="PROJECT_INFO.name"
  />
</template>
