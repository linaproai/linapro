<script setup lang="ts">
import type {
  HostServicePermissionItem,
  HostServicePermissionTableItem,
  PluginAuthorizationPayload,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Alert,
  Checkbox,
  Descriptions,
  DescriptionsItem,
  Divider,
  Tag,
  message,
} from 'ant-design-vue';

import { pluginEnable, pluginInstall } from '#/api/system/plugin';

type ReviewMode = 'enable' | 'install';

const emit = defineEmits<{ reload: [] }>();

const selectedTargets = ref<Record<string, string[]>>({});
const currentPlugin = ref<SystemPlugin | null>(null);
const currentMode = ref<ReviewMode>('install');

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleSubmit,
  onOpenChange: handleOpenChange,
});

const requestedServices = computed<HostServicePermissionItem[]>(() => {
  return currentPlugin.value?.requestedHostServices ?? [];
});

const authorizationRequired = computed(() => {
  return currentPlugin.value?.authorizationRequired === 1;
});

const currentTitle = computed(() => {
  return currentMode.value === 'install'
    ? '安装插件并确认权限'
    : '启用插件并确认权限';
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

function resolveServiceTargets(service: HostServicePermissionItem) {
  if (service.service === 'storage') {
    return [...(service.paths ?? [])];
  }
  if (service.service === 'data') {
    return [...(service.tables ?? [])];
  }
  return (service.resources ?? []).map((item) => item.ref);
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

function resolveDefaultSelections(service: HostServicePermissionItem) {
  const authorized = currentPlugin.value?.authorizedHostServices?.find(
    (item) => item.service === service.service,
  );
  const requestedTargets = resolveServiceTargets(service);
  const authorizedTargets =
    service.service === 'storage'
      ? [...(authorized?.paths ?? [])]
      : service.service === 'data'
        ? [...(authorized?.tables ?? [])]
        : (authorized?.resources ?? []).map((item) => item.ref);
  if (currentPlugin.value?.authorizationStatus === 'confirmed') {
    return authorizedTargets;
  }
  return requestedTargets;
}

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ mode: ReviewMode; row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  currentMode.value = data?.mode ?? 'install';
  selectedTargets.value = {};

  for (const service of requestedServices.value) {
    const hasTargets =
      service.service === 'storage'
        ? (service.paths ?? []).length > 0
        : service.service === 'data'
          ? (service.tables ?? []).length > 0
          : (service.resources ?? []).length > 0;
    if (!hasTargets) {
      continue;
    }
    selectedTargets.value[service.service] = resolveDefaultSelections(service);
  }
}

function buildAuthorizationPayload(): PluginAuthorizationPayload | undefined {
  if (!authorizationRequired.value) {
    return undefined;
  }
  return {
    authorization: {
      services: requestedServices.value
        .filter((service) =>
          service.service === 'storage'
            ? (service.paths ?? []).length > 0
            : service.service === 'data'
              ? (service.tables ?? []).length > 0
              : (service.resources ?? []).length > 0,
        )
        .map((service) => ({
          methods: service.methods,
          paths:
            service.service === 'storage'
              ? [...(selectedTargets.value[service.service] ?? [])]
              : undefined,
          resourceRefs:
            service.service === 'storage' || service.service === 'data'
              ? undefined
              : [...(selectedTargets.value[service.service] ?? [])],
          tables:
            service.service === 'data'
              ? [...(selectedTargets.value[service.service] ?? [])]
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
      message.success('动态插件已安装');
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
  selectedTargets.value = {};
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :fullscreen-button="false"
    :title="currentTitle"
  >
    <div
      v-if="currentPlugin"
      data-testid="plugin-host-service-auth-modal"
      class="flex flex-col gap-4"
    >
      <Alert
        v-if="authorizationRequired"
        show-icon
        type="info"
        message="以下 hostServices 声明属于插件权限申请，安装/启用时需由宿主确认最终授权结果。"
      />
      <Alert
        v-else
        show-icon
        type="success"
        message="当前插件未声明需要额外确认的资源申请，本次仅展示宿主服务概览。"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem label="插件标识">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem label="插件版本">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem label="当前授权状态" :span="2">
          <Tag
            :color="
              currentPlugin.authorizationStatus === 'confirmed'
                ? 'green'
                : currentPlugin.authorizationStatus === 'pending'
                  ? 'gold'
                  : 'blue'
            "
          >
            {{
              currentPlugin.authorizationStatus === 'confirmed'
                ? '已确认'
                : currentPlugin.authorizationStatus === 'pending'
                  ? '待确认'
                  : '无需确认'
            }}
          </Tag>
        </DescriptionsItem>
      </Descriptions>

      <div
        v-for="service in requestedServices"
        :key="service.service"
        class="rounded-md border border-dashed border-[var(--ant-color-border)] p-4"
      >
        <div class="mb-3 flex flex-wrap items-center gap-2">
          <span class="text-[15px] font-medium">
            {{ formatServiceLabel(service.service) }}
          </span>
          <Tag color="blue">{{ service.service }}</Tag>
          <Tag v-for="method in service.methods" :key="method">
            {{ method }}
          </Tag>
          <Tag
            v-if="
              service.service === 'storage'
                ? (service.paths ?? []).length === 0
                : service.service === 'data'
                  ? (service.tables ?? []).length === 0
                  : (service.resources ?? []).length === 0
            "
            color="success"
          >
            无需额外确认
          </Tag>
        </div>

        <template
          v-if="
            service.service === 'storage' && (service.paths ?? []).length > 0
          "
        >
          <div
            class="mb-3 text-[13px] text-[var(--ant-color-text-description)]"
          >
            请选择允许该插件访问的逻辑路径或路径前缀。
          </div>
          <Checkbox.Group
            v-model:value="selectedTargets[service.service]"
            class="flex w-full flex-col gap-3"
          >
            <div
              v-for="storagePath in service.paths"
              :key="storagePath"
              class="rounded-md bg-[var(--ant-color-fill-quaternary)] p-3"
            >
              <Checkbox
                :value="storagePath"
                :data-testid="`plugin-host-service-auth-checkbox-${currentPlugin.id}-${service.service}-${storagePath}`"
              >
                {{ storagePath }}
              </Checkbox>
              <div
                class="mt-2 text-[12px] text-[var(--ant-color-text-description)]"
              >
                允许方法: {{ service.methods.join(', ') || '-' }}
              </div>
            </div>
          </Checkbox.Group>
        </template>

        <template
          v-else-if="
            service.service === 'data' && (service.tables ?? []).length > 0
          "
        >
          <div
            class="mb-3 text-[13px] text-[var(--ant-color-text-description)]"
          >
            请选择允许该插件访问的数据表。
          </div>
          <Checkbox.Group
            v-model:value="selectedTargets[service.service]"
            class="flex w-full flex-col gap-3"
          >
            <div
              v-for="table in resolveDataTableItems(service)"
              :key="table.name"
              class="rounded-md bg-[var(--ant-color-fill-quaternary)] p-3"
            >
              <Checkbox
                :value="table.name"
                :data-testid="`plugin-host-service-auth-checkbox-${currentPlugin.id}-${service.service}-${table.name}`"
              >
                {{ table.name }}
              </Checkbox>
              <div
                class="mt-2 text-[12px] text-[var(--ant-color-text-description)]"
              >
                <div v-if="table.comment">表说明: {{ table.comment }}</div>
                允许方法: {{ service.methods.join(', ') || '-' }}
              </div>
            </div>
          </Checkbox.Group>
        </template>

        <template v-else-if="(service.resources ?? []).length > 0">
          <div
            class="mb-3 text-[13px] text-[var(--ant-color-text-description)]"
          >
            {{
              service.service === 'network'
                ? '请选择允许该插件访问的 URL 模式。'
                : '请选择允许该插件访问的逻辑资源。'
            }}
          </div>
          <Checkbox.Group
            v-model:value="selectedTargets[service.service]"
            class="flex w-full flex-col gap-3"
          >
            <div
              v-for="resource in service.resources"
              :key="resource.ref"
              class="rounded-md bg-[var(--ant-color-fill-quaternary)] p-3"
            >
              <Checkbox
                :value="resource.ref"
                :data-testid="`plugin-host-service-auth-checkbox-${currentPlugin.id}-${service.service}-${resource.ref}`"
              >
                {{ resource.ref }}
              </Checkbox>
              <div
                class="mt-2 grid gap-2 text-[12px] text-[var(--ant-color-text-description)]"
              >
                <template v-if="service.service === 'network'">
                  <div>
                    该 URL 模式一旦授权，插件即可直接访问命中的 HTTP 地址。
                  </div>
                </template>
                <template v-else>
                  <div>治理目标: {{ resource.ref }}</div>
                </template>
              </div>
            </div>
          </Checkbox.Group>
        </template>
      </div>

      <Divider class="my-0" />
      <div class="text-[12px] text-[var(--ant-color-text-description)]">
        未勾选的资源申请将不会进入当前 release
        的最终授权快照，运行时调用会被宿主拒绝。
      </div>
    </div>
  </BasicModal>
</template>
