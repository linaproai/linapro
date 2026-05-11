<script setup lang="ts">
import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Checkbox, List, ListItem } from 'ant-design-vue';

import { $t } from '#/locales';

const emit = defineEmits<{ force: [] }>();

const reasons = ref<string[]>([]);
const forceConfirmed = ref(false);

const [Modal, modalApi] = useVbenModal({
  onConfirm() {
    if (forceConfirmed.value) {
      emit('force');
      modalApi.close();
    }
  },
  onOpenChange(open) {
    if (!open) {
      return;
    }
    const data = modalApi.getData<{ reasons?: string[] }>();
    reasons.value = data?.reasons?.length
      ? data.reasons
      : [$t('pages.multiTenant.plugin.lifecycleGuard.defaultReason')];
    forceConfirmed.value = false;
  },
});
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
      <Checkbox v-model:checked="forceConfirmed" data-testid="lifecycle-guard-force">
        {{ $t('pages.multiTenant.plugin.lifecycleGuard.forceConfirm') }}
      </Checkbox>
    </div>
  </Modal>
</template>
