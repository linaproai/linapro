<script setup lang="ts">
import type { OperLog } from './operlog-client';

import { computed, shallowRef } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { DictTag } from '#/components/dict';
import JsonPreview from '#/components/json-preview/index.vue';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';

const dictStore = useDictStore();

const [BasicDrawer, drawerApi] = useVbenDrawer({
  onOpenChange: handleOpenChange,
  onClosed() {
    currentLog.value = null;
  },
});

const currentLog = shallowRef<null | OperLog>(null);

function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const { record } = drawerApi.getData() as { record: OperLog };
  currentLog.value = record;
}

const operTypeDicts = computed(() => {
  return dictStore.getDictOptions('sys_oper_type');
});

const operStatusDicts = computed(() => {
  return dictStore.getDictOptions('sys_oper_status');
});

function parseJson(str: string): any {
  if (!str) return null;
  try {
    const obj = JSON.parse(str);
    if (typeof obj === 'object') return obj;
    return null;
  } catch {
    return null;
  }
}
</script>

<template>
  <BasicDrawer
    :footer="false"
    class="w-[600px]"
    :title="$t('plugin.monitor-operlog.detail.title')"
  >
    <Descriptions v-if="currentLog" size="small" bordered :column="1">
      <DescriptionsItem
        :label="$t('plugin.monitor-operlog.fields.logId')"
        :label-style="{ minWidth: '120px' }"
      >
        {{ currentLog.id }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.operResult')">
        <DictTag :dicts="operStatusDicts as any" :value="currentLog.status" />
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.moduleName')">
        {{ currentLog.title }}
      </DescriptionsItem>
      <DescriptionsItem
        :label="$t('plugin.monitor-operlog.fields.operSummary')"
      >
        {{ currentLog.operSummary }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.operType')">
        <DictTag :dicts="operTypeDicts as any" :value="currentLog.operType" />
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.operator')">
        {{ currentLog.operName }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.requestUrl')">
        {{ currentLog.operUrl }}
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.ipAddress')">
        {{ currentLog.operIp }}
      </DescriptionsItem>
      <DescriptionsItem
        v-if="currentLog.operParam"
        :label="$t('plugin.monitor-operlog.fields.requestParams')"
      >
        <div class="max-h-[300px] overflow-y-auto">
          <JsonPreview
            v-if="parseJson(currentLog.operParam)"
            class="break-normal"
            :data="parseJson(currentLog.operParam)"
          />
          <span v-else>{{ currentLog.operParam }}</span>
        </div>
      </DescriptionsItem>
      <DescriptionsItem
        v-if="currentLog.jsonResult"
        :label="$t('plugin.monitor-operlog.fields.responseResult')"
      >
        <div class="max-h-[300px] overflow-y-auto">
          <JsonPreview
            v-if="parseJson(currentLog.jsonResult)"
            class="break-normal"
            :data="parseJson(currentLog.jsonResult)"
          />
          <span v-else>{{ currentLog.jsonResult }}</span>
        </div>
      </DescriptionsItem>
      <DescriptionsItem
        v-if="currentLog.errorMsg"
        :label="$t('plugin.monitor-operlog.fields.errorInfo')"
      >
        <span class="font-semibold text-red-600">
          {{ currentLog.errorMsg }}
        </span>
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.duration')">
        {{ currentLog.costTime }} ms
      </DescriptionsItem>
      <DescriptionsItem :label="$t('plugin.monitor-operlog.fields.operTime')">
        {{ currentLog.operTime }}
      </DescriptionsItem>
    </Descriptions>
  </BasicDrawer>
</template>
