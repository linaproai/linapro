<script setup lang="ts">
import type {
  HostServicePermissionItem,
  HostServicePermissionTableItem,
  PluginAuthorizationPayload,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, Tag, message } from 'ant-design-vue';

import { pluginEnable, pluginInstall } from '#/api/system/plugin';

type ReviewMode = 'enable' | 'install';

const hostServiceOrder: Record<string, number> = {
  data: 0,
  storage: 1,
  network: 2,
  runtime: 3,
};

const emit = defineEmits<{ reload: [] }>();

const currentPlugin = ref<SystemPlugin | null>(null);
const currentMode = ref<ReviewMode>('install');

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleSubmit,
  onOpenChange: handleOpenChange,
});

const requestedServices = computed<HostServicePermissionItem[]>(() => {
  const items = [...(currentPlugin.value?.requestedHostServices ?? [])];
  return items.sort((left, right) => {
    const leftOrder = hostServiceOrder[left.service] ?? Number.MAX_SAFE_INTEGER;
    const rightOrder =
      hostServiceOrder[right.service] ?? Number.MAX_SAFE_INTEGER;
    if (leftOrder !== rightOrder) {
      return leftOrder - rightOrder;
    }
    return left.service.localeCompare(right.service);
  });
});

const authorizationRequired = computed(() => {
  return currentPlugin.value?.authorizationRequired === 1;
});

const currentTitle = computed(() => {
  if (currentMode.value === 'install') {
    return authorizationRequired.value ? '安装插件并确认授权' : '安装插件';
  }
  return authorizationRequired.value ? '启用插件并确认授权' : '启用插件';
});

const currentConfirmText = computed(() => {
  if (currentMode.value === 'install') {
    return authorizationRequired.value ? '确认授权并安装' : '确认安装';
  }
  return authorizationRequired.value ? '确认授权并启用' : '确认启用';
});

const currentBannerMessage = computed(() => {
  if (currentMode.value === 'install') {
    return authorizationRequired.value
      ? '请先核对插件详情与宿主服务清单，确认后将默认授权该插件声明的全部服务。'
      : '请先核对插件详情，确认后开始安装插件。';
  }
  return '该插件当前 release 尚未形成最终授权快照；确认后将默认授权该 release 声明的全部服务并继续启用。';
});

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

function formatPluginType(type: string) {
  if (type === 'source') {
    return '源码插件';
  }
  if (type === 'dynamic') {
    return '动态插件';
  }
  return type || '-';
}

function hasServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'storage') {
    return (service.paths ?? []).length > 0;
  }
  if (service.service === 'data') {
    return (service.tables ?? []).length > 0;
  }
  return (service.resources ?? []).length > 0;
}

function resolveDataTableItems(service: HostServicePermissionItem) {
  if ((service.tableItems ?? []).length > 0) {
    return [...(service.tableItems ?? [])];
  }
  return (service.tables ?? []).map<HostServicePermissionTableItem>(
    (table) => ({
      name: table,
    }),
  );
}

function formatDataTableLabel(table: HostServicePermissionTableItem) {
  return table.comment ? `${table.name} (${table.comment})` : table.name;
}

function formatServiceResourceListLabel(service: HostServicePermissionItem) {
  switch (service.service) {
    case 'data': {
      return '数据表列表';
    }
    case 'network': {
      return 'URL 模式列表';
    }
    case 'storage': {
      return '存储目录前缀列表';
    }
    default: {
      return '资源列表';
    }
  }
}

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ mode: ReviewMode; row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  currentMode.value = data?.mode ?? 'install';
}

function buildAuthorizationPayload(): PluginAuthorizationPayload | undefined {
  if (!authorizationRequired.value) {
    return undefined;
  }
  return {
    authorization: {
      services: requestedServices.value
        .filter((service) => hasServiceTargets(service))
        .map((service) => ({
          methods: service.methods,
          paths:
            service.service === 'storage'
              ? [...(service.paths ?? [])]
              : undefined,
          resourceRefs:
            service.service === 'storage' || service.service === 'data'
              ? undefined
              : (service.resources ?? []).map((item) => item.ref),
          tables:
            service.service === 'data'
              ? [...(service.tables ?? [])]
              : undefined,
          service: service.service,
        })),
    },
  };
}

async function handleSubmit() {
  if (!currentPlugin.value) {
    return;
  }
  try {
    modalApi.lock(true);
    const payload = buildAuthorizationPayload();
    if (currentMode.value === 'install') {
      await pluginInstall(currentPlugin.value.id, payload);
      message.success('插件已安装');
    } else {
      await pluginEnable(currentPlugin.value.id, payload);
      message.success('插件已启用');
    }
    emit('reload');
    handleClosed();
  } finally {
    modalApi.lock(false);
  }
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
  currentMode.value = 'install';
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    :confirm-text="currentConfirmText"
    :title="currentTitle"
    class="w-[860px] max-w-[calc(100vw-32px)]"
  >
    <div
      v-if="currentPlugin"
      data-testid="plugin-host-service-auth-modal"
      class="flex flex-col gap-4"
    >
      <Alert
        show-icon
        :type="authorizationRequired ? 'info' : 'success'"
        :message="currentBannerMessage"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem label="插件名称">
          {{ currentPlugin.name || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件标识">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem label="插件类型">
          {{ formatPluginType(currentPlugin.type) }}
        </DescriptionsItem>
        <DescriptionsItem label="插件版本">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem label="插件描述" :span="2">
          {{ currentPlugin.description || '-' }}
        </DescriptionsItem>
      </Descriptions>

      <template v-if="requestedServices.length > 0">
        <div class="text-[13px] font-medium text-[var(--ant-color-text)]">
          {{
            authorizationRequired
              ? '宿主服务授权范围'
              : '宿主服务声明概览'
          }}
        </div>

        <div
          v-for="service in requestedServices"
          :key="service.service"
          class="rounded-md border border-[var(--ant-color-border)] p-4"
        >
          <div
            :class="[
              'flex flex-wrap items-center gap-2',
              hasServiceTargets(service) ? 'mb-3' : '',
            ]"
          >
            <span class="text-[15px] font-medium">
              {{ formatServiceLabel(service.service) }}
            </span>
            <Tag color="blue">{{ service.service }}</Tag>
            <Tag v-for="method in service.methods" :key="method">
              {{ method }}
            </Tag>
          </div>

          <template
            v-if="
              service.service === 'storage' && (service.paths ?? []).length > 0
            "
          >
            <div
              class="flex flex-col gap-2"
            >
              <div
                class="text-[12px] font-medium text-[var(--ant-color-text-description)]"
              >
                {{ formatServiceResourceListLabel(service) }}
              </div>
              <ul
                :data-testid="`plugin-host-service-auth-list-${currentPlugin.id}-${service.service}`"
                class="m-0 list-disc space-y-2 pl-5"
              >
                <li
                  v-for="storagePath in service.paths"
                  :key="storagePath"
                  class="rounded-md bg-[var(--ant-color-fill-quaternary)] px-3 py-2 marker:text-[var(--ant-color-primary)]"
                >
                  <div
                    :data-testid="`plugin-host-service-auth-item-${currentPlugin.id}-${service.service}-${storagePath}`"
                    class="break-all text-[14px] text-[var(--ant-color-text)]"
                  >
                    {{ storagePath }}
                  </div>
                </li>
              </ul>
            </div>
          </template>

          <template
            v-else-if="
              service.service === 'data' && (service.tables ?? []).length > 0
            "
          >
            <div
              class="flex flex-col gap-2"
            >
              <div
                class="text-[12px] font-medium text-[var(--ant-color-text-description)]"
              >
                {{ formatServiceResourceListLabel(service) }}
              </div>
              <ul
                :data-testid="`plugin-host-service-auth-list-${currentPlugin.id}-${service.service}`"
                class="m-0 list-disc space-y-2 pl-5"
              >
                <li
                  v-for="table in resolveDataTableItems(service)"
                  :key="table.name"
                  class="rounded-md bg-[var(--ant-color-fill-quaternary)] px-3 py-2 marker:text-[var(--ant-color-primary)]"
                >
                  <div
                    :data-testid="`plugin-host-service-auth-item-${currentPlugin.id}-${service.service}-${table.name}`"
                    class="break-all text-[14px] text-[var(--ant-color-text)]"
                  >
                    {{ formatDataTableLabel(table) }}
                  </div>
                </li>
              </ul>
            </div>
          </template>

          <template v-else-if="(service.resources ?? []).length > 0">
            <div
              class="flex flex-col gap-2"
            >
              <div
                class="text-[12px] font-medium text-[var(--ant-color-text-description)]"
              >
                {{ formatServiceResourceListLabel(service) }}
              </div>
              <ul
                :data-testid="`plugin-host-service-auth-list-${currentPlugin.id}-${service.service}`"
                class="m-0 list-disc space-y-2 pl-5"
              >
                <li
                  v-for="resource in service.resources"
                  :key="resource.ref"
                  class="rounded-md bg-[var(--ant-color-fill-quaternary)] px-3 py-2 marker:text-[var(--ant-color-primary)]"
                >
                  <div
                    :data-testid="`plugin-host-service-auth-item-${currentPlugin.id}-${service.service}-${resource.ref}`"
                    class="break-all text-[14px] text-[var(--ant-color-text)]"
                  >
                    {{ resource.ref }}
                  </div>
                  <div
                    v-if="service.service !== 'network'"
                    class="mt-1 text-[12px] text-[var(--ant-color-text-description)]"
                  >
                    治理目标: {{ resource.ref }}
                  </div>
                </li>
              </ul>
            </div>
          </template>

        </div>

      </template>
    </div>
  </BasicModal>
</template>
