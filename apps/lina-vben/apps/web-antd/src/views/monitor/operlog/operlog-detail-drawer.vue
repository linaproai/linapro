<script setup lang="ts">
import type { OperLog } from '#/api/monitor/operlog/model';

import { computed, shallowRef } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { DictTag } from '#/components/dict';
import JsonPreview from '#/components/json-preview/index.vue';
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
  return dictStore.dictOptionsMap.get('sys_oper_type') || [];
});

const operStatusDicts = computed(() => {
  return dictStore.dictOptionsMap.get('sys_oper_status') || [];
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
  <BasicDrawer :footer="false" class="w-[600px]" title="操作日志详情">
    <Descriptions
      v-if="currentLog"
      size="small"
      bordered
      :column="1"
    >
      <DescriptionsItem label="日志编号" :label-style="{ minWidth: '120px' }">
        {{ currentLog.id }}
      </DescriptionsItem>
      <DescriptionsItem label="操作结果">
        <DictTag
          :dicts="(operStatusDicts as any)"
          :value="currentLog.status"
        />
      </DescriptionsItem>
      <DescriptionsItem label="模块名称">
        {{ currentLog.title }}
      </DescriptionsItem>
      <DescriptionsItem label="操作名称">
        {{ currentLog.operSummary }}
      </DescriptionsItem>
      <DescriptionsItem label="操作类型">
        <DictTag
          :dicts="(operTypeDicts as any)"
          :value="currentLog.operType"
        />
      </DescriptionsItem>
      <DescriptionsItem label="操作人员">
        {{ currentLog.operName }}
      </DescriptionsItem>
      <DescriptionsItem label="请求地址">
        {{ currentLog.operUrl }}
      </DescriptionsItem>
      <DescriptionsItem label="IP地址">
        {{ currentLog.operIp }}
      </DescriptionsItem>
      <DescriptionsItem v-if="currentLog.operParam" label="请求参数">
        <div class="max-h-[300px] overflow-y-auto">
          <JsonPreview
            v-if="parseJson(currentLog.operParam)"
            class="break-normal"
            :data="parseJson(currentLog.operParam)"
          />
          <span v-else>{{ currentLog.operParam }}</span>
        </div>
      </DescriptionsItem>
      <DescriptionsItem v-if="currentLog.jsonResult" label="响应结果">
        <div class="max-h-[300px] overflow-y-auto">
          <JsonPreview
            v-if="parseJson(currentLog.jsonResult)"
            class="break-normal"
            :data="parseJson(currentLog.jsonResult)"
          />
          <span v-else>{{ currentLog.jsonResult }}</span>
        </div>
      </DescriptionsItem>
      <DescriptionsItem v-if="currentLog.errorMsg" label="异常信息">
        <span class="font-semibold text-red-600">
          {{ currentLog.errorMsg }}
        </span>
      </DescriptionsItem>
      <DescriptionsItem label="操作耗时">
        {{ currentLog.costTime }} ms
      </DescriptionsItem>
      <DescriptionsItem label="操作时间">
        {{ currentLog.operTime }}
      </DescriptionsItem>
    </Descriptions>
  </BasicDrawer>
</template>
