<script lang="ts" setup>
import { About } from '@vben/common-ui';

import { PROJECT_INFO } from '#/views/about/config';

defineOptions({ name: 'About' });
</script>

<template>
  <About :description="PROJECT_INFO.description" :name="PROJECT_INFO.name" />
</template>
