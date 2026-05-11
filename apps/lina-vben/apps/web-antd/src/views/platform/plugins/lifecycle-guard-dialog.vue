<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Input } from 'ant-design-vue';

import { $t } from '#/locales';

const emit = defineEmits<{ force: [payload: { pluginId: string }] }>();

const pluginId = ref('');
const reasons = ref<string[]>([]);
const confirmText = ref('');

const canForce = computed(() => confirmText.value.trim() === pluginId.value);
const reasonDisplayKeys: Record<string, string> = {
  'plugin.multi-tenant.uninstall_blocked.tenants_exist':
    'pages.multiTenant.plugin.lifecycleGuard.reasons.multiTenantUninstallBlocked',
};
const localizedReasons = computed(() =>
  reasons.value.map((reason) => {
    const displayKey = reasonDisplayKeys[reason] ?? reason;
    const localized = $t(displayKey);
    if (localized !== displayKey) {
      return localized;
    }
    const fallback = $t(reason);
    return fallback === reason ? reason : fallback;
  }),
);
const blockedReasonText = computed(() =>
  $t('pages.multiTenant.plugin.lifecycleGuard.blockedReason'),
);

const [Modal, modalApi] = useVbenModal({
  onConfirm() {
    if (canForce.value) {
      emit('force', { pluginId: pluginId.value });
    }
  },
  onOpenChange(open) {
    if (!open) {
      return;
    }
    const data = modalApi.getData<{ pluginId?: string; reasons?: string[] }>();
    pluginId.value = data?.pluginId?.trim() ?? '';
    reasons.value = data?.reasons?.length
      ? data.reasons
      : [$t('pages.multiTenant.plugin.lifecycleGuard.defaultReason')];
    confirmText.value = '';
    modalApi.setState({ confirmDisabled: true });
  },
});

watch(canForce, (allowed) => {
  modalApi.setState({ confirmDisabled: !allowed });
});
</script>

<template>
  <Modal :title="$t('pages.multiTenant.plugin.lifecycleGuard.title')">
    <div
      class="flex flex-col gap-[10px]"
      data-testid="lifecycle-guard-dialog"
    >
      <Alert
        data-testid="lifecycle-guard-reason-alert"
        show-icon
        type="warning"
      >
        <template #description>
          <div data-testid="lifecycle-guard-reason">
            <div>{{ blockedReasonText }}</div>
            <div>{{ localizedReasons.join('；') }}</div>
          </div>
        </template>
      </Alert>

      <Alert
        data-testid="lifecycle-guard-force-alert"
        show-icon
        type="error"
      >
        <template #description>
          <div class="space-y-3">
            <div>
              {{
                $t('pages.multiTenant.plugin.lifecycleGuard.forceConfirm', {
                  pluginId,
                })
              }}
            </div>
            <div>
              {{
                $t('pages.multiTenant.plugin.lifecycleGuard.forceInputHint', {
                  pluginId,
                })
              }}
            </div>
            <Input
              v-model:value="confirmText"
              :placeholder="pluginId"
              data-testid="lifecycle-guard-force-plugin-id"
            />
          </div>
        </template>
      </Alert>
    </div>
  </Modal>
</template>
