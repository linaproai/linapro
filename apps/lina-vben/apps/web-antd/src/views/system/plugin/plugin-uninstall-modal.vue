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
import { $t } from '#/locales';

const emit = defineEmits<{ reload: [] }>();

const currentPlugin = ref<SystemPlugin | null>(null);
const purgeStorageData = ref(true);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

const isSourcePlugin = computed(() => currentPlugin.value?.type === 'source');
const isDynamicPlugin = computed(() => currentPlugin.value?.type === 'dynamic');
const isAutoEnableManaged = computed(
  () => currentPlugin.value?.autoEnableManaged === 1,
);
const supportsPurgeStorageData = computed(
  () => isSourcePlugin.value || isDynamicPlugin.value,
);

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  purgeStorageData.value = supportsPurgeStorageData.value;
}

async function handleConfirm() {
  if (!currentPlugin.value) {
    return;
  }

  try {
    modalApi.lock(true);
    await pluginUninstall(
      currentPlugin.value.id,
      supportsPurgeStorageData.value ? purgeStorageData.value : undefined,
    );
    message.success($t('pages.system.plugin.messages.uninstalled'));
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
  <BasicModal :title="$t('pages.system.plugin.uninstall.title')">
    <div
      v-if="currentPlugin"
      data-testid="plugin-uninstall-modal"
      class="flex flex-col gap-4"
    >
      <Alert
        v-if="isAutoEnableManaged"
        data-testid="plugin-auto-enable-uninstall-alert"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.messages.autoEnableUninstallAlert')"
      />
      <Alert
        v-if="isSourcePlugin"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.uninstall.sourceWarning')"
      />
      <Alert
        v-else-if="isDynamicPlugin"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.uninstall.dynamicWarning')"
      />
      <Alert
        v-else
        show-icon
        type="info"
        :message="$t('pages.system.plugin.uninstall.defaultWarning')"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem :label="$t('pages.system.plugin.fields.id')">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.type')">
          <Tag :color="isSourcePlugin ? 'blue' : 'green'">
            {{
              isSourcePlugin
                ? $t('pages.system.plugin.type.source')
                : $t('pages.system.plugin.type.dynamic')
            }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.status')">
          <Tag :color="currentPlugin.enabled === 1 ? 'green' : 'default'">
            {{
              currentPlugin.enabled === 1
                ? $t('pages.status.enabled')
                : $t('pages.status.disabled')
            }}
          </Tag>
        </DescriptionsItem>
      </Descriptions>

      <Checkbox
        v-if="supportsPurgeStorageData"
        v-model:checked="purgeStorageData"
        data-testid="plugin-uninstall-purge-checkbox"
      >
        {{ $t('pages.system.plugin.uninstall.purgeStorage') }}
      </Checkbox>
    </div>
  </BasicModal>
</template>
