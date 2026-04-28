<script setup lang="ts">
import type {
  ServerMonitorResult,
  ServerNodeInfo,
} from '#/api/monitor/server/model';

import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Progress, Table, Tooltip } from 'ant-design-vue';

import { getServerMonitor } from '#/api/monitor/server';
import { $t } from '#/locales';

defineOptions({ name: 'ServerMonitor' });

const nodes = ref<ServerNodeInfo[]>([]);
const dbInfo = ref<ServerMonitorResult['dbInfo'] | null>(null);
const loading = ref(false);
const expandedNodes = ref<Set<string>>(new Set());

const hasData = computed(() => nodes.value.length > 0);

onMounted(async () => {
  await loadData();
});

async function loadData() {
  loading.value = true;
  try {
    const resp = await getServerMonitor();
    nodes.value = resp.nodes ?? [];
    dbInfo.value = resp.dbInfo ?? null;
    // Auto-expand all nodes
    expandedNodes.value = new Set(
      nodes.value.map((n) => `${n.nodeName}|${n.nodeIp}`),
    );
  } finally {
    loading.value = false;
  }
}

function toggleNode(key: string) {
  const set = new Set(expandedNodes.value);
  if (set.has(key)) {
    set.delete(key);
  } else {
    set.add(key);
  }
  expandedNodes.value = set;
}

function isExpanded(key: string): boolean {
  return expandedNodes.value.has(key);
}

function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${(bytes / k ** i).toFixed(2)} ${sizes[i]}`;
}

function formatRate(bytesPerSec: number): string {
  return `${formatBytes(bytesPerSec)}/s`;
}

function formatUptime(seconds: number): string {
  const days = Math.floor(seconds / 86400);
  const hours = Math.floor((seconds % 86400) / 3600);
  const mins = Math.floor((seconds % 3600) / 60);
  const parts: string[] = [];
  if (days > 0)
    parts.push($t('pages.monitor.server.units.days', { count: days }));
  if (hours > 0)
    parts.push($t('pages.monitor.server.units.hours', { count: hours }));
  if (mins > 0)
    parts.push($t('pages.monitor.server.units.minutes', { count: mins }));
  return parts.join(' ') || $t('pages.monitor.server.units.justStarted');
}

function getProgressColor(percent: number): string {
  if (percent >= 90) return '#ff4d4f';
  if (percent >= 70) return '#faad14';
  return '#52c41a';
}

const diskColumns = computed(() => [
  {
    title: $t('pages.monitor.server.diskColumns.path'),
    dataIndex: 'path',
    key: 'path',
  },
  {
    title: $t('pages.monitor.server.diskColumns.fsType'),
    dataIndex: 'fsType',
    key: 'fsType',
  },
  {
    title: $t('pages.monitor.server.diskColumns.total'),
    dataIndex: 'total',
    key: 'total',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: $t('pages.monitor.server.diskColumns.used'),
    dataIndex: 'used',
    key: 'used',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: $t('pages.monitor.server.diskColumns.free'),
    dataIndex: 'free',
    key: 'free',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: $t('pages.monitor.server.diskColumns.usagePercent'),
    dataIndex: 'usagePercent',
    key: 'usagePercent',
    width: 200,
  },
]);
</script>

<template>
  <Page>
    <template v-if="hasData">
      <!-- Database information. -->
      <div v-if="dbInfo" class="card-box p-5">
        <h5 class="text-lg text-foreground">
          {{ $t('pages.monitor.server.sections.database') }}
        </h5>
        <div class="mt-4">
          <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
              <dt class="text-sm/6 font-medium text-foreground">
                {{ $t('pages.monitor.server.fields.dbVersion') }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.version }}
              </dd>
            </div>
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
              <dt class="text-sm/6 font-medium text-foreground">
                {{ $t('pages.monitor.server.fields.maxOpenConns') }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.maxOpenConns }}
              </dd>
            </div>
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
              <dt class="text-sm/6 font-medium text-foreground">
                {{ $t('pages.monitor.server.fields.openConns') }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.openConns }}
              </dd>
            </div>
            <div class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0">
              <dt class="text-sm/6 font-medium text-foreground">
                {{ $t('pages.monitor.server.fields.inUseIdle') }}
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.inUse }} / {{ dbInfo.idle }}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- Server information. -->
      <div class="card-box mt-6 p-5">
        <div class="flex items-center gap-2">
          <h5 class="text-lg text-foreground">
            {{ $t('pages.monitor.server.sections.server') }}
          </h5>
          <Tooltip :title="$t('pages.monitor.server.tooltips.multiNode')">
            <span
              class="icon-[ant-design--question-circle-outlined] cursor-help text-foreground/40"
            ></span>
          </Tooltip>
        </div>
        <div class="mt-4">
          <div
            v-for="(node, index) in nodes"
            :key="`${node.nodeName}|${node.nodeIp}`"
            :class="{ 'mt-3': index > 0 }"
          >
            <!-- Node header (tree-like) -->
            <div
              class="flex cursor-pointer items-center gap-2 rounded px-2 py-2 hover:bg-accent"
              @click="toggleNode(`${node.nodeName}|${node.nodeIp}`)"
            >
              <span
                :class="
                  isExpanded(`${node.nodeName}|${node.nodeIp}`)
                    ? 'icon-[ant-design--caret-down-outlined]'
                    : 'icon-[ant-design--caret-right-outlined]'
                "
                class="text-foreground/50"
              ></span>
              <span class="font-medium text-foreground">
                {{ node.nodeName }}
              </span>
              <span class="text-sm text-foreground/50">
                ({{ node.nodeIp }})
              </span>
            </div>

            <!-- Expanded content -->
            <div
              v-if="isExpanded(`${node.nodeName}|${node.nodeIp}`)"
              class="ml-6 border-l border-border pl-4"
            >
              <!-- Service information. -->
              <div class="py-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  {{ $t('pages.monitor.server.sections.service') }}
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.goVersion') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.version }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.goframeVersion') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.gfVersion }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.goroutines') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.goroutines }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.gcPause') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{
                        ((node.goInfo?.gcPauseNs ?? 0) / 1_000_000).toFixed(2)
                      }}
                      ms
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.serviceStartTime') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.startTime }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.serviceUptime') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.serviceUptime }}
                    </dd>
                  </div>
                </dl>

                <!-- Service CPU and memory. -->
                <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
                  <!-- Service CPU. -->
                  <div class="rounded-lg border border-border p-4">
                    <h6 class="mb-3 text-sm font-medium text-foreground/70">
                      {{ $t('pages.monitor.server.sections.serviceCpu') }}
                    </h6>
                    <div class="flex items-center gap-6">
                      <Progress
                        :percent="
                          Number((node.goInfo?.processCpu ?? 0).toFixed(1))
                        "
                        :stroke-color="
                          getProgressColor(node.goInfo?.processCpu ?? 0)
                        "
                        :width="80"
                        type="circle"
                      />
                      <dl class="flex-1">
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">
                            {{ $t('pages.monitor.server.fields.used') }}
                          </dt>
                          <dd class="text-sm text-foreground">
                            {{
                              (
                                ((node.goInfo?.processCpu ?? 0) *
                                  (node.cpu?.cores ?? 0)) /
                                100
                              ).toFixed(2)
                            }}
                            {{ $t('pages.monitor.server.units.core') }}
                          </dd>
                        </div>
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">
                            {{ $t('pages.monitor.server.fields.totalCores') }}
                          </dt>
                          <dd class="text-sm text-foreground">
                            {{ node.cpu?.cores }}
                            {{ $t('pages.monitor.server.units.core') }}
                          </dd>
                        </div>
                      </dl>
                    </div>
                  </div>

                  <!-- Service memory. -->
                  <div class="rounded-lg border border-border p-4">
                    <h6 class="mb-3 text-sm font-medium text-foreground/70">
                      {{ $t('pages.monitor.server.sections.serviceMemory') }}
                    </h6>
                    <div class="flex items-center gap-6">
                      <Progress
                        :percent="
                          Number((node.goInfo?.processMemory ?? 0).toFixed(1))
                        "
                        :stroke-color="
                          getProgressColor(node.goInfo?.processMemory ?? 0)
                        "
                        :width="80"
                        type="circle"
                      />
                      <dl class="flex-1">
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">
                            {{ $t('pages.monitor.server.fields.used') }}
                          </dt>
                          <dd class="text-sm text-foreground">
                            {{
                              formatBytes(
                                ((node.memory?.total ?? 0) *
                                  (node.goInfo?.processMemory ?? 0)) /
                                  100,
                              )
                            }}
                          </dd>
                        </div>
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">
                            {{ $t('pages.monitor.server.fields.totalMemory') }}
                          </dt>
                          <dd class="text-sm text-foreground">
                            {{ formatBytes(node.memory?.total ?? 0) }}
                          </dd>
                        </div>
                      </dl>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Basic information. -->
              <div class="py-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  {{ $t('pages.monitor.server.sections.basic') }}
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.hostname') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.hostname }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.os') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.os }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.arch') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.arch }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.nodeIp') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.nodeIp }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.systemUptime') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatUptime(node.server?.uptime ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.bootTime') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.bootTime }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.collectAt') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.collectAt }}
                    </dd>
                  </div>
                </dl>
              </div>

              <!-- CPU and memory. -->
              <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
                <!-- CPU -->
                <div class="rounded-lg border border-border p-4">
                  <h6 class="mb-3 text-sm font-medium text-foreground/70">
                    {{ $t('pages.monitor.server.sections.systemCpu') }}
                  </h6>
                  <div class="flex items-center gap-6">
                    <Progress
                      :percent="
                        Number((node.cpu?.usagePercent ?? 0).toFixed(1))
                      "
                      :stroke-color="
                        getProgressColor(node.cpu?.usagePercent ?? 0)
                      "
                      :width="80"
                      type="circle"
                    />
                    <dl class="flex-1">
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">
                          {{ $t('pages.monitor.server.fields.cores') }}
                        </dt>
                        <dd class="text-sm text-foreground">
                          {{ node.cpu?.cores }}
                          {{ $t('pages.monitor.server.units.core') }}
                        </dd>
                      </div>
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">
                          {{ $t('pages.monitor.server.fields.modelName') }}
                        </dt>
                        <dd
                          class="max-w-[300px] truncate text-sm text-foreground"
                        >
                          {{ node.cpu?.modelName }}
                        </dd>
                      </div>
                    </dl>
                  </div>
                </div>

                <!-- Memory. -->
                <div class="rounded-lg border border-border p-4">
                  <h6 class="mb-3 text-sm font-medium text-foreground/70">
                    {{ $t('pages.monitor.server.sections.systemMemory') }}
                  </h6>
                  <div class="flex items-center gap-6">
                    <Progress
                      :percent="
                        Number((node.memory?.usagePercent ?? 0).toFixed(1))
                      "
                      :stroke-color="
                        getProgressColor(node.memory?.usagePercent ?? 0)
                      "
                      :width="80"
                      type="circle"
                    />
                    <dl class="flex-1">
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">
                          {{ $t('pages.monitor.server.fields.usedTotal') }}
                        </dt>
                        <dd class="text-sm text-foreground">
                          {{ formatBytes(node.memory?.used ?? 0) }} /
                          {{ formatBytes(node.memory?.total ?? 0) }}
                        </dd>
                      </div>
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">
                          {{ $t('pages.monitor.server.fields.available') }}
                        </dt>
                        <dd class="text-sm text-foreground">
                          {{ formatBytes(node.memory?.available ?? 0) }}
                        </dd>
                      </div>
                    </dl>
                  </div>
                </div>
              </div>

              <!-- Disk. -->
              <div class="mt-3">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  {{ $t('pages.monitor.server.sections.disk') }}
                </h6>
                <Table
                  :columns="diskColumns"
                  :data-source="node.disks"
                  :pagination="false"
                  row-key="path"
                  size="small"
                >
                  <template #bodyCell="{ column, record }">
                    <template v-if="column.key === 'usagePercent'">
                      <Progress
                        :percent="Number(record.usagePercent.toFixed(1))"
                        :stroke-color="getProgressColor(record.usagePercent)"
                        size="small"
                        status="active"
                      />
                    </template>
                  </template>
                </Table>
              </div>

              <!-- Network. -->
              <div class="mt-3 pb-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  {{ $t('pages.monitor.server.sections.network') }}
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.totalSent') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatBytes(node.network?.bytesSent ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.totalReceived') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatBytes(node.network?.bytesRecv ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.sendRate') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatRate(node.network?.sendRate ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      {{ $t('pages.monitor.server.fields.receiveRate') }}
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatRate(node.network?.recvRate ?? 0) }}
                    </dd>
                  </div>
                </dl>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <!-- Empty State -->
    <div
      v-else-if="!loading"
      class="flex h-[300px] items-center justify-center text-foreground/40"
    >
      {{ $t('pages.monitor.server.empty') }}
    </div>
  </Page>
</template>
