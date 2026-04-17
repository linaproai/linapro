<script setup lang="ts">
import type {
  HostServicePermissionItem,
  HostServicePermissionResourceItem,
  HostServicePermissionTableItem,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

const hostServiceOrder: Record<string, number> = {
  data: 0,
  storage: 1,
  network: 2,
  runtime: 3,
};

const currentPlugin = ref<SystemPlugin | null>(null);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onOpenChange: handleOpenChange,
});

const requestedServices = computed(() => {
  return sortHostServices(currentPlugin.value?.requestedHostServices);
});

const authorizedServices = computed(() => {
  return sortHostServices(currentPlugin.value?.authorizedHostServices);
});

const hasHostServiceDetails = computed(() => {
  return requestedServices.value.length > 0 || authorizedServices.value.length > 0;
});

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
}

function sortHostServices(items?: HostServicePermissionItem[]) {
  return [...(items ?? [])].sort((left, right) => {
    const leftOrder = hostServiceOrder[left.service] ?? Number.MAX_SAFE_INTEGER;
    const rightOrder = hostServiceOrder[right.service] ?? Number.MAX_SAFE_INTEGER;
    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder;
    }
    return left.service.localeCompare(right.service);
  });
}

function formatPluginType(type: string) {
  if (type === 'source') {
    return '源码插件';
  }
  if (type === 'dynamic') {
    return '动态插件';
  }
  return type || '-';
}

function formatInstalledStatus(installed: number) {
  return installed === 1 ? '已接入' : '未安装';
}

function getInstalledStatusColor(installed: number) {
  return installed === 1 ? 'green' : 'default';
}

function formatEnabledStatus(enabled: number) {
  return enabled === 1 ? '启用' : '禁用';
}

function getEnabledStatusColor(enabled: number) {
  return enabled === 1 ? 'green' : 'default';
}

function formatAuthorizationRequirement(required: number) {
  return required === 1 ? '需要确认' : '无需确认';
}

function getAuthorizationRequirementColor(required: number) {
  return required === 1 ? 'gold' : 'default';
}

function formatAuthorizationStatus(status: string) {
  switch (status) {
    case 'confirmed': {
      return '已确认';
    }
    case 'not_required': {
      return '无需确认';
    }
    case 'pending': {
      return '待确认';
    }
    default: {
      return status || '-';
    }
  }
}

function getAuthorizationStatusColor(status: string) {
  switch (status) {
    case 'confirmed': {
      return 'green';
    }
    case 'pending': {
      return 'gold';
    }
    default: {
      return 'default';
    }
  }
}

function formatServiceLabel(service: string) {
  switch (service) {
    case 'data': {
      return '数据服务';
    }
    case 'network': {
      return '网络服务';
    }
    case 'runtime': {
      return '运行时服务';
    }
    case 'storage': {
      return '存储服务';
    }
    default: {
      return service;
    }
  }
}

function hasServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'storage') {
    return (service.paths ?? []).length > 0;
  }
  if (service.service === 'data') {
    return (service.tableItems ?? []).length > 0 || (service.tables ?? []).length > 0;
  }
  return (service.resources ?? []).length > 0;
}

function resolveDataTableItems(service: HostServicePermissionItem) {
  if ((service.tableItems ?? []).length > 0) {
    return (service.tableItems ?? []).map((item) => formatDataTableLabel(item));
  }
  return (service.tables ?? []).map((table) => table || '-');
}

function formatDataTableLabel(table: HostServicePermissionTableItem) {
  return table.comment ? `${table.name} (${table.comment})` : table.name;
}

function formatResourceLabel(resource: HostServicePermissionResourceItem) {
  const methods = resource.allowMethods ?? [];
  if (methods.length === 0) {
    return resource.ref;
  }
  return `${resource.ref} [${methods.join(', ')}]`;
}

function resolveServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'storage') {
    return [...(service.paths ?? [])];
  }
  if (service.service === 'data') {
    return resolveDataTableItems(service);
  }
  return (service.resources ?? []).map((item) => formatResourceLabel(item));
}

function resolveTargetLabel(service: HostServicePermissionItem) {
  switch (service.service) {
    case 'data': {
      return '数据表边界';
    }
    case 'network': {
      return '网络资源边界';
    }
    case 'storage': {
      return '存储目录边界';
    }
    default: {
      return '资源边界';
    }
  }
}
</script>

<template>
  <BasicModal
    :footer="false"
    title="插件详情"
    class="w-[860px] max-w-[calc(100vw-32px)]"
  >
    <div
      v-if="currentPlugin"
      data-testid="plugin-detail-modal"
      class="flex flex-col gap-4"
    >
      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem label="插件名称">
          {{ currentPlugin.name || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件标识">
          {{ currentPlugin.id || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件类型">
          <Tag color="blue">
            {{ formatPluginType(currentPlugin.type) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="插件版本">
          {{ currentPlugin.version || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="接入状态">
          <Tag :color="getInstalledStatusColor(currentPlugin.installed)">
            {{ formatInstalledStatus(currentPlugin.installed) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="当前状态">
          <Tag :color="getEnabledStatusColor(currentPlugin.enabled)">
            {{ formatEnabledStatus(currentPlugin.enabled) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="授权要求">
          <Tag
            :color="
              getAuthorizationRequirementColor(
                currentPlugin.authorizationRequired,
              )
            "
          >
            {{
              formatAuthorizationRequirement(currentPlugin.authorizationRequired)
            }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="授权状态">
          <Tag
            :color="getAuthorizationStatusColor(currentPlugin.authorizationStatus)"
          >
            {{ formatAuthorizationStatus(currentPlugin.authorizationStatus) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="安装时间">
          {{ currentPlugin.installedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="更新时间">
          {{ currentPlugin.updatedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件描述" :span="2">
          {{ currentPlugin.description || '-' }}
        </DescriptionsItem>
      </Descriptions>

      <Alert
        v-if="!hasHostServiceDetails"
        data-testid="plugin-detail-empty-host-services"
        show-icon
        type="info"
        message="当前插件未声明额外宿主服务申请或授权快照。"
      />

      <template v-if="requestedServices.length > 0">
        <div class="text-[13px] font-medium text-[var(--ant-color-text)]">
          宿主服务申请清单
        </div>

        <div
          v-for="service in requestedServices"
          :key="`requested-${service.service}`"
          class="rounded-md border border-[var(--ant-color-border)] p-4"
        >
          <div class="mb-3 flex flex-wrap items-center gap-2">
            <span class="text-[15px] font-medium">
              {{ formatServiceLabel(service.service) }}
            </span>
            <Tag color="blue">{{ service.service }}</Tag>
            <Tag v-for="method in service.methods" :key="method">
              {{ method }}
            </Tag>
          </div>

          <template v-if="hasServiceTargets(service)">
            <div class="mb-2 text-[13px] text-[var(--ant-color-text-secondary)]">
              {{ resolveTargetLabel(service) }}
            </div>
            <div class="flex flex-wrap gap-2">
              <Tag v-for="item in resolveServiceTargets(service)" :key="item">
                {{ item }}
              </Tag>
            </div>
          </template>
          <div
            v-else
            class="text-[13px] text-[var(--ant-color-text-secondary)]"
          >
            当前服务仅声明方法，无额外资源边界。
          </div>
        </div>
      </template>

      <template v-if="authorizedServices.length > 0">
        <div class="text-[13px] font-medium text-[var(--ant-color-text)]">
          宿主服务授权快照
        </div>

        <div
          v-for="service in authorizedServices"
          :key="`authorized-${service.service}`"
          class="rounded-md border border-[var(--ant-color-border)] p-4"
        >
          <div class="mb-3 flex flex-wrap items-center gap-2">
            <span class="text-[15px] font-medium">
              {{ formatServiceLabel(service.service) }}
            </span>
            <Tag color="green">{{ service.service }}</Tag>
            <Tag v-for="method in service.methods" :key="method">
              {{ method }}
            </Tag>
          </div>

          <template v-if="hasServiceTargets(service)">
            <div class="mb-2 text-[13px] text-[var(--ant-color-text-secondary)]">
              {{ resolveTargetLabel(service) }}
            </div>
            <div class="flex flex-wrap gap-2">
              <Tag v-for="item in resolveServiceTargets(service)" :key="item">
                {{ item }}
              </Tag>
            </div>
          </template>
          <div
            v-else
            class="text-[13px] text-[var(--ant-color-text-secondary)]"
          >
            当前服务仅声明方法，无额外资源边界。
          </div>
        </div>
      </template>
    </div>
  </BasicModal>
</template>
