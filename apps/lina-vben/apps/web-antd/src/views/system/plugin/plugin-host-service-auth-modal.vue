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
import { $t } from '#/locales';

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
    return authorizationRequired.value
      ? $t('pages.system.plugin.auth.installWithAuthTitle')
      : $t('pages.system.plugin.auth.installTitle');
  }
  return authorizationRequired.value
    ? $t('pages.system.plugin.auth.enableWithAuthTitle')
    : $t('pages.system.plugin.auth.enableTitle');
});

const currentConfirmText = computed(() => {
  if (currentMode.value === 'install') {
    return authorizationRequired.value
      ? $t('pages.system.plugin.auth.confirmInstallWithAuth')
      : $t('pages.system.plugin.auth.confirmInstall');
  }
  return authorizationRequired.value
    ? $t('pages.system.plugin.auth.confirmEnableWithAuth')
    : $t('pages.system.plugin.auth.confirmEnable');
});

const currentBannerMessage = computed(() => {
  if (currentMode.value === 'install') {
    return authorizationRequired.value
      ? $t('pages.system.plugin.auth.installBannerWithAuth')
      : $t('pages.system.plugin.auth.installBanner');
  }
  return $t('pages.system.plugin.auth.enableBanner');
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
          message.success($t('pages.system.plugin.messages.installedAndEnabled'));
        } catch {
          emit('reload');
          handleClosed();
          message.warning($t('pages.system.plugin.messages.installSucceededEnableFailed'));
          return;
        }
      } else {
        message.success($t('pages.system.plugin.messages.installed'));
      }
    } else {
      await pluginEnable(pluginID, payload);
      message.success($t('pages.system.plugin.messages.enabled'));
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
    return $t('pages.system.plugin.type.source');
  }
  if (type === 'dynamic') {
    return $t('pages.system.plugin.type.dynamic');
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
        {{ $t('pages.system.plugin.actions.installAndEnable') }}
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
        <DescriptionsItem :label="$t('pages.system.plugin.fields.name')">
          {{ currentPlugin.name || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.id')">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.type')">
          {{ formatPluginType(currentPlugin.type) }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.description')" :span="2">
          {{ currentPlugin.description || '-' }}
        </DescriptionsItem>
      </Descriptions>

      <template v-if="requestedServices.length > 0">
        <PluginSectionTitle test-id="plugin-host-service-section-title">
          {{
            authorizationRequired
              ? $t('pages.system.plugin.auth.hostServiceAuthTitle')
              : $t('pages.system.plugin.auth.hostServiceDeclareTitle')
          }}
        </PluginSectionTitle>

        <PluginHostServiceCards :cards="hostServiceCards" />
      </template>

      <template v-if="showDeclaredRoutes">
        <PluginSectionTitle test-id="plugin-route-section-title">
          {{ $t('pages.system.plugin.detail.routeListTitle') }}
        </PluginSectionTitle>

        <PluginRouteReviewList :routes="declaredRoutes" />
      </template>
    </div>
  </BasicModal>
</template>
