<script setup lang="ts">
import type {
  ServerMonitorResult,
  ServerNodeInfo,
} from '#/api/monitor/server/model';

import { computed, onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Progress, Table, Tooltip } from 'ant-design-vue';

import { getServerMonitor } from '#/api/monitor/server';

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
  if (days > 0) parts.push(`${days}天`);
  if (hours > 0) parts.push(`${hours}小时`);
  if (mins > 0) parts.push(`${mins}分钟`);
  return parts.join(' ') || '刚启动';
}

function getProgressColor(percent: number): string {
  if (percent >= 90) return '#ff4d4f';
  if (percent >= 70) return '#faad14';
  return '#52c41a';
}

const diskColumns = [
  { title: '盘符', dataIndex: 'path', key: 'path' },
  { title: '文件系统', dataIndex: 'fsType', key: 'fsType' },
  {
    title: '总容量',
    dataIndex: 'total',
    key: 'total',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '已用',
    dataIndex: 'used',
    key: 'used',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '可用',
    dataIndex: 'free',
    key: 'free',
    customRender: ({ text }: any) => formatBytes(text),
  },
  {
    title: '使用率',
    dataIndex: 'usagePercent',
    key: 'usagePercent',
    width: 200,
  },
];
</script>

<template>
  <Page>
    <template v-if="hasData">
      <!-- 数据库信息 -->
      <div v-if="dbInfo" class="card-box p-5">
        <h5 class="text-lg text-foreground">数据库信息</h5>
        <div class="mt-4">
          <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">数据库版本</dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.version }}
              </dd>
            </div>
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">最大连接数</dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.maxOpenConns }}
              </dd>
            </div>
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                当前打开连接
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.openConns }}
              </dd>
            </div>
            <div
              class="border-t border-border px-4 py-3 sm:col-span-1 sm:px-0"
            >
              <dt class="text-sm/6 font-medium text-foreground">
                使用中 / 空闲
              </dt>
              <dd class="mt-1 text-sm/6 text-foreground">
                {{ dbInfo.inUse }} / {{ dbInfo.idle }}
              </dd>
            </div>
          </dl>
        </div>
      </div>

      <!-- 服务器信息 -->
      <div class="card-box mt-6 p-5">
        <div class="flex items-center gap-2">
          <h5 class="text-lg text-foreground">服务器信息</h5>
          <Tooltip title="Lina 支持多节点高可用部署，每个节点具有独立的服务器指标信息">
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
              <!-- 服务信息 -->
              <div class="py-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  服务信息
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      Go 版本
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.version }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      GoFrame 版本
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.gfVersion }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      Goroutines
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.goroutines }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      GC 暂停
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{
                        (
                          (node.goInfo?.gcPauseNs ?? 0) / 1_000_000
                        ).toFixed(2)
                      }}
                      ms
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      服务启动时间
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.startTime }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      服务运行时长
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.goInfo?.serviceUptime }}
                    </dd>
                  </div>
                </dl>

                <!-- 服务 CPU + 服务内存 -->
                <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
                  <!-- 服务 CPU -->
                  <div class="rounded-lg border border-border p-4">
                    <h6 class="mb-3 text-sm font-medium text-foreground/70">
                      服务 CPU
                    </h6>
                    <div class="flex items-center gap-6">
                      <Progress
                        :percent="
                          Number(
                            (node.goInfo?.processCpu ?? 0).toFixed(1),
                          )
                        "
                        :stroke-color="
                          getProgressColor(node.goInfo?.processCpu ?? 0)
                        "
                        :width="80"
                        type="circle"
                      />
                      <dl class="flex-1">
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">已使用</dt>
                          <dd class="text-sm text-foreground">
                            {{
                              (
                                ((node.goInfo?.processCpu ?? 0) *
                                  (node.cpu?.cores ?? 0)) /
                                100
                              ).toFixed(2)
                            }}
                            核
                          </dd>
                        </div>
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">总核心数</dt>
                          <dd class="text-sm text-foreground">
                            {{ node.cpu?.cores }} 核
                          </dd>
                        </div>
                      </dl>
                    </div>
                  </div>

                  <!-- 服务内存 -->
                  <div class="rounded-lg border border-border p-4">
                    <h6 class="mb-3 text-sm font-medium text-foreground/70">
                      服务内存
                    </h6>
                    <div class="flex items-center gap-6">
                      <Progress
                        :percent="
                          Number(
                            (node.goInfo?.processMemory ?? 0).toFixed(1),
                          )
                        "
                        :stroke-color="
                          getProgressColor(
                            node.goInfo?.processMemory ?? 0,
                          )
                        "
                        :width="80"
                        type="circle"
                      />
                      <dl class="flex-1">
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">已使用</dt>
                          <dd class="text-sm text-foreground">
                            {{
                              formatBytes(
                                (node.memory?.total ?? 0) *
                                  (node.goInfo?.processMemory ?? 0) /
                                  100,
                              )
                            }}
                          </dd>
                        </div>
                        <div class="py-1">
                          <dt class="text-xs text-foreground/50">
                            总内存量
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

              <!-- 基本信息 -->
              <div class="py-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  基本信息
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      主机名
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.hostname }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      操作系统
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.os }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      系统架构
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.arch }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      节点IP
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.nodeIp }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      系统运行时长
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatUptime(node.server?.uptime ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      系统启动时间
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.server?.bootTime }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      数据更新时间
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ node.collectAt }}
                    </dd>
                  </div>
                </dl>
              </div>

              <!-- CPU + 内存 -->
              <div class="mt-3 grid grid-cols-1 gap-4 md:grid-cols-2">
                <!-- CPU -->
                <div class="rounded-lg border border-border p-4">
                  <h6 class="mb-3 text-sm font-medium text-foreground/70">
                  系统CPU
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
                        <dt class="text-xs text-foreground/50">核心数</dt>
                        <dd class="text-sm text-foreground">
                          {{ node.cpu?.cores }} 核
                        </dd>
                      </div>
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">型号</dt>
                        <dd class="max-w-[300px] truncate text-sm text-foreground">
                          {{ node.cpu?.modelName }}
                        </dd>
                      </div>
                    </dl>
                  </div>
                </div>

                <!-- 内存 -->
                <div class="rounded-lg border border-border p-4">
                  <h6 class="mb-3 text-sm font-medium text-foreground/70">
                    系统内存
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
                        <dt class="text-xs text-foreground/50">已用 / 总量</dt>
                        <dd class="text-sm text-foreground">
                          {{ formatBytes(node.memory?.used ?? 0) }} /
                          {{ formatBytes(node.memory?.total ?? 0) }}
                        </dd>
                      </div>
                      <div class="py-1">
                        <dt class="text-xs text-foreground/50">可用</dt>
                        <dd class="text-sm text-foreground">
                          {{ formatBytes(node.memory?.available ?? 0) }}
                        </dd>
                      </div>
                    </dl>
                  </div>
                </div>
              </div>

              <!-- 磁盘 -->
              <div class="mt-3">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  磁盘使用
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

              <!-- 网络 -->
              <div class="mt-3 pb-2">
                <h6 class="mb-2 text-sm font-medium text-foreground/70">
                  网络流量
                </h6>
                <dl class="grid grid-cols-2 md:grid-cols-4">
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      总发送
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatBytes(node.network?.bytesSent ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      总接收
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatBytes(node.network?.bytesRecv ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      发送速率
                    </dt>
                    <dd class="mt-1 text-sm/6 text-foreground">
                      {{ formatRate(node.network?.sendRate ?? 0) }}
                    </dd>
                  </div>
                  <div
                    class="border-t border-border px-4 py-2 sm:col-span-1 sm:px-0"
                  >
                    <dt class="text-sm/6 font-medium text-foreground">
                      接收速率
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
      暂无监控数据，请等待数据采集...
    </div>
  </Page>
</template>
