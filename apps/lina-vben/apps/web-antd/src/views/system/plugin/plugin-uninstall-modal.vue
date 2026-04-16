<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Alert,
  Checkbox,
  Descriptions,
  DescriptionsItem,
  Tag,
  message,
} from 'ant-design-vue';

import { pluginUninstall } from '#/api/system/plugin';

const emit = defineEmits<{ reload: [] }>();

const currentPlugin = ref<SystemPlugin | null>(null);
const purgeStorageData = ref(true);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

const isSourcePlugin = computed(() => currentPlugin.value?.type === 'source');

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  purgeStorageData.value = currentPlugin.value?.type === 'source';
}

async function handleConfirm() {
  if (!currentPlugin.value) {
    return;
  }

  try {
    modalApi.lock(true);
    await pluginUninstall(
      currentPlugin.value.id,
      isSourcePlugin.value ? purgeStorageData.value : undefined,
    );
    message.success('插件已卸载');
    emit('reload');
    handleClosed();
  } finally {
    modalApi.lock(false);
  }
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
  purgeStorageData.value = true;
}
</script>

<template>
  <BasicModal title="卸载插件">
    <div
      v-if="currentPlugin"
      data-testid="plugin-uninstall-modal"
      class="flex flex-col gap-4"
    >
      <Alert
        v-if="isSourcePlugin"
        show-icon
        type="warning"
        message="源码插件卸载时可选择是否同时执行卸载 SQL 与插件自定义清理逻辑。勾选后会同步清除示例数据表数据和插件自有存储文件。"
      />
      <Alert
        v-else
        show-icon
        type="info"
        message="动态插件卸载默认仅移除治理挂载、菜单和运行时产物，不会删除插件业务数据。"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem label="插件标识">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem label="插件版本">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem label="插件类型">
          <Tag :color="isSourcePlugin ? 'blue' : 'green'">
            {{ isSourcePlugin ? '源码插件' : '动态插件' }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="当前状态">
          <Tag :color="currentPlugin.enabled === 1 ? 'green' : 'default'">
            {{ currentPlugin.enabled === 1 ? '启用' : '禁用' }}
          </Tag>
        </DescriptionsItem>
      </Descriptions>

      <Checkbox
        v-if="isSourcePlugin"
        v-model:checked="purgeStorageData"
        data-testid="plugin-uninstall-purge-checkbox"
      >
        卸载时同时清除插件存储数据（数据表数据与插件自有文件）
      </Checkbox>
    </div>
  </BasicModal>
</template>
