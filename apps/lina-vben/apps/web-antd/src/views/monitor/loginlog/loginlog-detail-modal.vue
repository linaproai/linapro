<script setup lang="ts">
import type { LoginLog } from '#/api/monitor/loginlog/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { DictTag } from '#/components/dict';
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

const operStatusDicts = computed(() => {
  return dictStore.dictOptionsMap.get('sys_oper_status') || [];
});
</script>

<template>
  <BasicModal
    :footer="false"
    :fullscreen-button="false"
    class="w-[550px]"
    title="登录日志详情"
  >
    <Descriptions v-if="loginInfo" size="small" :column="1" bordered>
      <DescriptionsItem label="用户账号" :label-style="{ minWidth: '100px' }">
        {{ loginInfo.userName }}
      </DescriptionsItem>
      <DescriptionsItem label="登录状态">
        <DictTag
          :dicts="(operStatusDicts as any)"
          :value="loginInfo.status"
        />
      </DescriptionsItem>
      <DescriptionsItem label="IP地址">
        {{ loginInfo.ip }}
      </DescriptionsItem>
      <DescriptionsItem label="浏览器">
        {{ loginInfo.browser }}
      </DescriptionsItem>
      <DescriptionsItem label="操作系统">
        {{ loginInfo.os }}
      </DescriptionsItem>
      <DescriptionsItem label="提示信息">
        <span
          class="font-semibold"
          :class="{ 'text-red-500': loginInfo.status !== 0 }"
        >
          {{ loginInfo.msg }}
        </span>
      </DescriptionsItem>
      <DescriptionsItem label="登录时间">
        {{ loginInfo.loginTime }}
      </DescriptionsItem>
    </Descriptions>
  </BasicModal>
</template>
