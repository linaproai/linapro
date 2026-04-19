<script setup lang="ts">
import type { JobLogRecord } from '#/api/system/job/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem, Spin, Tag } from 'ant-design-vue';

import { jobLogDetail } from '#/api/system/jobLog';
import JsonPreview from '#/components/json-preview/index.vue';

const currentLog = ref<JobLogRecord | null>(null);
const loading = ref(false);

const [Modal, modalApi] = useVbenModal({
  onClosed: () => {
    currentLog.value = null;
  },
  onOpenChange: async (open) => {
    if (!open) {
      return;
    }
    loading.value = true;
    try {
      const { id } = modalApi.getData<{ id: number }>();
      currentLog.value = await jobLogDetail(id);
    } finally {
      loading.value = false;
    }
  },
});

function parseJSON(raw?: string) {
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw);
  } catch {
    return null;
  }
}

const paramsSnapshot = computed(() => parseJSON(currentLog.value?.paramsSnapshot));
const resultSnapshot = computed(() => parseJSON(currentLog.value?.resultJson));
const shellResult = computed(() => {
  const result = resultSnapshot.value;
  if (!result || typeof result !== 'object') {
    return null;
  }
  if (
    Object.prototype.hasOwnProperty.call(result, 'stdout') ||
    Object.prototype.hasOwnProperty.call(result, 'stderr') ||
    Object.prototype.hasOwnProperty.call(result, 'exitCode')
  ) {
    return result as Record<string, any>;
  }
  return null;
});

function statusColor(status?: string) {
  const colorMap: Record<string, string> = {
    cancelled: 'default',
    failed: 'error',
    running: 'processing',
    skipped_max_concurrency: 'warning',
    skipped_not_primary: 'warning',
    skipped_singleton: 'warning',
    success: 'success',
    timeout: 'warning',
  };
  return colorMap[status || ''] || 'default';
}
</script>

<template>
  <Modal
    :footer="false"
    class="w-[920px]"
    data-testid="job-log-detail-modal"
    title="执行日志详情"
  >
    <Spin :spinning="loading">
      <Descriptions
        v-if="currentLog"
        :column="2"
        bordered
        size="small"
      >
        <DescriptionsItem label="日志编号">
          {{ currentLog.id }}
        </DescriptionsItem>
        <DescriptionsItem label="任务名称">
          {{ currentLog.jobName || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="执行状态">
          <Tag :color="statusColor(currentLog.status)">
            {{ currentLog.status }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="触发方式">
          {{ currentLog.trigger }}
        </DescriptionsItem>
        <DescriptionsItem label="执行节点">
          {{ currentLog.nodeId || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="执行耗时">
          {{ currentLog.durationMs }} ms
        </DescriptionsItem>
        <DescriptionsItem label="开始时间">
          {{ currentLog.startAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="结束时间">
          {{ currentLog.endAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem v-if="currentLog.errMsg" :span="2" label="异常信息">
          <span class="font-medium text-red-500">
            {{ currentLog.errMsg }}
          </span>
        </DescriptionsItem>
        <DescriptionsItem :span="2" label="任务快照">
          <div class="max-h-[260px] overflow-auto">
            <JsonPreview
              v-if="parseJSON(currentLog.jobSnapshot)"
              :data="parseJSON(currentLog.jobSnapshot)"
            />
            <span v-else>{{ currentLog.jobSnapshot || '-' }}</span>
          </div>
        </DescriptionsItem>
        <DescriptionsItem :span="2" label="参数快照">
          <div class="max-h-[260px] overflow-auto">
            <JsonPreview
              v-if="paramsSnapshot"
              :data="paramsSnapshot"
            />
            <span v-else>{{ currentLog.paramsSnapshot || '-' }}</span>
          </div>
        </DescriptionsItem>
        <DescriptionsItem :span="2" label="执行结果">
          <div class="space-y-3">
            <template v-if="shellResult">
              <div>
                <div class="mb-1 text-xs text-foreground/60">stdout</div>
                <pre class="max-h-[220px] overflow-auto rounded bg-accent px-3 py-2 text-xs">{{ shellResult.stdout || '' }}</pre>
              </div>
              <div>
                <div class="mb-1 text-xs text-foreground/60">stderr</div>
                <pre class="max-h-[220px] overflow-auto rounded bg-accent px-3 py-2 text-xs">{{ shellResult.stderr || '' }}</pre>
              </div>
              <div class="text-xs text-foreground/60">
                exitCode={{ shellResult.exitCode ?? '-' }}
              </div>
            </template>
            <template v-else>
              <div class="max-h-[260px] overflow-auto">
                <JsonPreview
                  v-if="resultSnapshot"
                  :data="resultSnapshot"
                />
                <span v-else>{{ currentLog.resultJson || '-' }}</span>
              </div>
            </template>
          </div>
        </DescriptionsItem>
      </Descriptions>
    </Spin>
  </Modal>
</template>
