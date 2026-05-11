<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Input, List, ListItem } from 'ant-design-vue';

import { $t } from '#/locales';

const emit = defineEmits<{ force: [payload: { pluginId: string }] }>();

const pluginId = ref('');
const reasons = ref<string[]>([]);
const confirmText = ref('');

const canForce = computed(() => confirmText.value.trim() === pluginId.value);

const [Modal, modalApi] = useVbenModal({
  onConfirm() {
    if (canForce.value) {
      emit('force', { pluginId: pluginId.value });
      modalApi.close();
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

function handleConfirmTextChange() {
  modalApi.setState({ confirmDisabled: !canForce.value });
}
</script>

<template>
  <Modal :title="$t('pages.multiTenant.plugin.lifecycleGuard.title')">
    <div class="space-y-4" data-testid="lifecycle-guard-dialog">
      <Alert
        show-icon
        type="warning"
        :message="$t('pages.multiTenant.plugin.lifecycleGuard.summary')"
      />
      <List bordered size="small">
        <ListItem v-for="reason in reasons" :key="reason">
          {{ $t(reason) === reason ? reason : $t(reason) }}
        </ListItem>
      </List>
      <Alert
        show-icon
        type="error"
        :message="
          $t('pages.multiTenant.plugin.lifecycleGuard.forceConfirm', {
            pluginId,
          })
        "
      />
      <div class="space-y-1">
        <div class="text-xs text-muted-foreground">
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
          @change="handleConfirmTextChange"
          @input="handleConfirmTextChange"
        />
      </div>
    </div>
  </Modal>
</template>
