<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, RadioGroup } from 'ant-design-vue';

import { $t } from '#/locales';

const emit = defineEmits<{
  confirm: [payload: { installMode: 'global' | 'tenant_scoped'; row: SystemPlugin }];
}>();

const currentPlugin = ref<SystemPlugin | null>(null);
const installMode = ref<'global' | 'tenant_scoped'>('tenant_scoped');

const [Modal, modalApi] = useVbenModal({
  onConfirm() {
    if (!currentPlugin.value) {
      return;
    }
    emit('confirm', {
      installMode: installMode.value,
      row: currentPlugin.value,
    });
    modalApi.close();
  },
  onOpenChange(open) {
    if (!open) {
      return;
    }
    const data = modalApi.getData<{ row: SystemPlugin }>();
    currentPlugin.value = data?.row ?? null;
    installMode.value =
      currentPlugin.value?.scopeNature === 'platform_only'
        ? 'global'
        : ((currentPlugin.value?.installMode as 'global' | 'tenant_scoped') ||
          'tenant_scoped');
  },
});

const isPlatformOnly = computed(
  () => currentPlugin.value?.scopeNature === 'platform_only',
);
</script>

<template>
  <Modal :title="$t('pages.multiTenant.plugin.installModeTitle')">
    <div class="space-y-4" data-testid="install-mode-selector">
      <Alert
        v-if="isPlatformOnly"
        show-icon
        type="info"
        :message="$t('pages.multiTenant.plugin.platformOnlyGlobalHint')"
      />
      <RadioGroup
        v-model:value="installMode"
        class="w-full"
        :disabled="isPlatformOnly"
        data-testid="install-mode-radio"
      >
        <div class="grid gap-3 md:grid-cols-2">
          <a-radio value="global" class="m-0 rounded border p-3">
            <div class="font-medium">
              {{ $t('pages.multiTenant.plugin.installModes.global') }}
            </div>
            <div class="text-xs text-[var(--ant-color-text-secondary)]">
              {{ $t('pages.multiTenant.plugin.installModeDescriptions.global') }}
            </div>
          </a-radio>
          <a-radio
            value="tenant_scoped"
            class="m-0 rounded border p-3"
            :disabled="isPlatformOnly"
          >
            <div class="font-medium">
              {{ $t('pages.multiTenant.plugin.installModes.tenant_scoped') }}
            </div>
            <div class="text-xs text-[var(--ant-color-text-secondary)]">
              {{
                $t(
                  'pages.multiTenant.plugin.installModeDescriptions.tenant_scoped',
                )
              }}
            </div>
          </a-radio>
        </div>
      </RadioGroup>
    </div>
  </Modal>
</template>
