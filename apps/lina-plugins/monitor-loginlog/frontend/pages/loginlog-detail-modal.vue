<script setup lang="ts">
import type { LoginLog } from './loginlog-client';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { DictTag } from '#/components/dict';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';

const dictStore = useDictStore();

const [BasicModal, modalApi] = useVbenModal({
  onOpenChange: (isOpen) => {
    if (!isOpen) {
      return;
    }
    const record = modalApi.getData() as LoginLog;
    loginInfo.value = record;
  },
  onClosed() {
    loginInfo.value = undefined;
  },
});

const loginInfo = ref<LoginLog>();

const loginStatusDicts = computed(() => {
  return dictStore.getDictOptions('sys_login_status');
});
</script>

<template>
  <BasicModal
    :footer="false"
    :fullscreen-button="false"
    class="w-[550px]"
    :title="$t('plugin.monitor-loginlog.detail.title')"
  >
    <Descriptions v-if="loginInfo" size="small" :column="1" bordered>
      <DescriptionsItem
        :label="$t('plugin.monitor-loginlog.fields.userName')"
        :label-style="{ minWidth: '100px' }"
      >
        {{ loginInfo.userName }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.status')">
        <DictTag :dicts="loginStatusDicts as any" :value="loginInfo.status" />
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.ipAddress')">
        {{ loginInfo.ip }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.browser')">
        {{ loginInfo.browser }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.os')">
        {{ loginInfo.os }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.message')">
        <span
          class="font-semibold"
          :class="{ 'text-red-500': loginInfo.status !== 0 }"
        >
          {{ loginInfo.msg }}
        </span>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-loginlog.fields.loginTime')">
        {{ loginInfo.loginTime }}
      </DescriptionsItem>
    </Descriptions>
  </BasicModal>
</template>
