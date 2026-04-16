<script setup lang="ts">
/**
 * 通用导出确认弹窗组件
 * 用于在导出前二次确认用户的导出操作
 */
import { computed } from 'vue';

import { useVbenModal } from '@vben/common-ui';

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
        是否导出选中的记录？
      </template>
      <template v-else>
        是否导出全部数据？
      </template>
    </div>
  </BasicModal>
</template>
