<script setup lang="ts">
/**
 * 通用导出确认弹窗组件
 * 用于在导出前二次确认用户的导出操作
 */
import { computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

const [BasicModal, modalApi] = useVbenModal();

const hasSelection = computed(() => {
  const data = modalApi.getData() as { selectedCount?: number };
  return (data?.selectedCount ?? 0) > 0;
});
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    :header="false"
    class="w-[400px]"
  >
    <div>
      <template v-if="hasSelection">
        {{ $t('pages.exportConfirm.selected') }}
      </template>
      <template v-else>
        {{ $t('pages.exportConfirm.all') }}
      </template>
    </div>
  </BasicModal>
</template>
