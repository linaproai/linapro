<script setup lang="ts">
import type {
  HostServicePermissionItem,
  PluginAuthorizationPayload,
  PluginRouteReviewItem,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, message } from 'ant-design-vue';

import { pluginEnable, pluginInstall } from '#/api/system/plugin';

import PluginHostServiceCards from './plugin-host-service-cards.vue';
import PluginRouteReviewList from './plugin-route-review-list.vue';
import PluginSectionTitle from './plugin-section-title.vue';
import {
  buildPluginAuthorizationHostServiceCards,
  sortHostServices,
} from './plugin-host-service-view';

type ReviewMode = 'enable' | 'install';
type SubmitAction = 'default' | 'install-and-enable';

const emit = defineEmits<{ reload: [] }>();

const currentPlugin = ref<null | SystemPlugin>(null);
const currentMode = ref<ReviewMode>('install');
const allowInstallAndEnable = ref(false);
const submittingAction = ref<null | SubmitAction>(null);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: () => handleSubmit('default'),
  onOpenChange: handleOpenChange,
});

const requestedServices = computed<HostServicePermissionItem[]>(() => {
  return sortHostServices(currentPlugin.value?.requestedHostServices);
});

const declaredRoutes = computed<PluginRouteReviewItem[]>(() => {
  return currentPlugin.value?.declaredRoutes ?? [];
});

const authorizationRequired = computed(() => {
  return currentPlugin.value?.authorizationRequired === 1;
});

const hostServiceCards = computed(() => {
  return buildPluginAuthorizationHostServiceCards(requestedServices.value, {
    authorizationRequired: authorizationRequired.value,
    buildScopeContainerTestId: (service) => {
      return currentPlugin.value
        ? `plugin-host-service-auth-list-${currentPlugin.value.id}-${service}`
        : undefined;
    },
    buildScopeItemTestIdPrefix: (service) => {
      return currentPlugin.value
        ? `plugin-host-service-auth-item-${currentPlugin.value.id}-${service}`
        : undefined;
    },
    targetSummaryBadgeColor: 'gold',
  });
});

const showInstallAndEnableAction = computed(() => {
  return currentMode.value === 'install' && allowInstallAndEnable.value;
});

const showDeclaredRoutes = computed(() => {
  return declaredRoutes.value.length > 0;
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
      ? '请先核对插件详情、宿主服务清单与注册路由列表，确认后将默认授权该插件声明的全部服务。'
      : '请先核对插件详情，确认后开始安装插件。';
  }
  return '该插件当前 release 尚未形成最终授权快照；确认后将默认授权该 release 声明的全部服务并继续启用。';
});

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{
    allowInstallAndEnable?: boolean;
    mode: ReviewMode;
    row: SystemPlugin;
  }>();
  currentPlugin.value = data?.row ?? null;
  currentMode.value = data?.mode ?? 'install';
  allowInstallAndEnable.value = data?.allowInstallAndEnable === true;
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

async function handleSubmit(action: SubmitAction) {
  if (!currentPlugin.value || submittingAction.value) {
    return;
  }
  submittingAction.value = action;
  try {
    modalApi.lock(true);
    const pluginID = currentPlugin.value.id;
    const payload = buildAuthorizationPayload();
    if (currentMode.value === 'install') {
      await pluginInstall(pluginID, payload);
      if (action === 'install-and-enable') {
        try {
          await pluginEnable(pluginID);
          message.success('插件已安装并启用');
        } catch {
          emit('reload');
          handleClosed();
          message.warning('插件已安装，但启用失败，请稍后重试。');
          return;
        }
      } else {
        message.success('插件已安装');
      }
    } else {
      await pluginEnable(pluginID, payload);
      message.success('插件已启用');
    }
    emit('reload');
    handleClosed();
  } finally {
    modalApi.lock(false);
    submittingAction.value = null;
  }
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
  currentMode.value = 'install';
  allowInstallAndEnable.value = false;
  submittingAction.value = null;
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
  return (
    (service.paths ?? []).length > 0 ||
    (service.tables ?? []).length > 0 ||
    (service.resources ?? []).length > 0
  );
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
    <template #append-footer>
      <a-button
        v-if="showInstallAndEnableAction"
        data-testid="plugin-install-enable-button"
        type="primary"
        :disabled="submittingAction !== null"
        :loading="submittingAction === 'install-and-enable'"
        @click="() => handleSubmit('install-and-enable')"
      >
        安装并启用
      </a-button>
    </template>
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
        <PluginSectionTitle test-id="plugin-host-service-section-title">
          {{ authorizationRequired ? '宿主服务授权范围' : '宿主服务声明概览' }}
        </PluginSectionTitle>

        <PluginHostServiceCards :cards="hostServiceCards" />
      </template>

      <template v-if="showDeclaredRoutes">
        <PluginSectionTitle test-id="plugin-route-section-title">
          注册路由列表
        </PluginSectionTitle>

        <PluginRouteReviewList :routes="declaredRoutes" />
      </template>
    </div>
  </BasicModal>
</template>
