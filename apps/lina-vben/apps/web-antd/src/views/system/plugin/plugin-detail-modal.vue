<script setup lang="ts">
import type {
  PluginRouteReviewItem,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import PluginHostServiceCards from './plugin-host-service-cards.vue';
import { buildPluginDetailHostServiceCards } from './plugin-host-service-view';
import PluginRouteReviewList from './plugin-route-review-list.vue';
import PluginSectionTitle from './plugin-section-title.vue';

const currentPlugin = ref<SystemPlugin | null>(null);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onOpenChange: handleOpenChange,
});

const hostServiceCards = computed(() => {
  return buildPluginDetailHostServiceCards(
    currentPlugin.value?.requestedHostServices,
    currentPlugin.value?.authorizedHostServices,
  );
});

const hasHostServiceDetails = computed(() => {
  return hostServiceCards.value.length > 0;
});

const declaredRoutes = computed<PluginRouteReviewItem[]>(() => {
  return currentPlugin.value?.declaredRoutes ?? [];
});

const showDeclaredRoutes = computed(() => {
  return currentPlugin.value?.type === 'dynamic' && declaredRoutes.value.length > 0;
});

const showHostServiceSection = computed(() => {
  return currentPlugin.value?.type === 'dynamic';
});

const isAutoEnableManaged = computed(() => {
  return currentPlugin.value?.autoEnableManaged === 1;
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

function formatPluginType(type: string) {
  if (type === 'source') {
    return $t('pages.system.plugin.type.source');
  }
  if (type === 'dynamic') {
    return $t('pages.system.plugin.type.dynamic');
  }
  return type || '-';
}

function formatInstalledStatus(installed: number) {
  return installed === 1
    ? $t('pages.system.plugin.installed.connected')
    : $t('pages.system.plugin.installed.notInstalled');
}

function getInstalledStatusColor(installed: number) {
  return installed === 1 ? 'green' : 'default';
}

function formatEnabledStatus(enabled: number) {
  return enabled === 1 ? $t('pages.status.enabled') : $t('pages.status.disabled');
}

function getEnabledStatusColor(enabled: number) {
  return enabled === 1 ? 'green' : 'default';
}

function formatAuthorizationStatus(status: string) {
  switch (status) {
    case 'confirmed': {
      return $t('pages.system.plugin.authorization.confirmed');
    }
    case 'not_required': {
      return $t('pages.system.plugin.authorization.notRequired');
    }
    case 'pending': {
      return $t('pages.system.plugin.authorization.pending');
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

function formatAutoEnableManaged(managed: boolean) {
  return managed
    ? $t('pages.system.plugin.autoEnableManaged')
    : $t('pages.system.plugin.manualManaged');
}

function getAutoEnableManagedColor(managed: boolean) {
  return managed ? 'gold' : 'default';
}

function formatBooleanValue(value: boolean) {
  return value ? $t('pages.common.yes') : $t('pages.common.no');
}

function getBooleanColor(value: boolean) {
  return value ? 'green' : 'default';
}

function formatScopeNature(scopeNature?: string) {
  switch (scopeNature) {
    case 'platform_only': {
      return $t('pages.system.plugin.scopeNature.platformOnly');
    }
    case 'tenant_aware': {
      return $t('pages.system.plugin.scopeNature.tenantAware');
    }
    default: {
      return scopeNature || '-';
    }
  }
}

function getScopeNatureColor(scopeNature?: string) {
  return scopeNature === 'tenant_aware' ? 'green' : 'blue';
}

function formatInstallMode(installMode?: string) {
  switch (installMode) {
    case 'global': {
      return $t('pages.multiTenant.plugin.installModes.global');
    }
    case 'tenant_scoped': {
      return $t('pages.multiTenant.plugin.installModes.tenant_scoped');
    }
    default: {
      return installMode || '-';
    }
  }
}

function getInstallModeColor(installMode?: string) {
  return installMode === 'tenant_scoped' ? 'green' : 'blue';
}

function formatRuntimeState(state?: string) {
  const key = `pages.system.plugin.runtimeState.${state || 'normal'}`;
  const label = $t(key);
  return label === key ? state || '-' : label;
}

function getRuntimeStateColor(state?: string) {
  switch (state) {
    case 'pending_upgrade': {
      return 'gold';
    }
    case 'upgrade_failed':
    case 'abnormal': {
      return 'red';
    }
    case 'upgrade_running': {
      return 'blue';
    }
    default: {
      return 'green';
    }
  }
}

function formatAbnormalReason(reason?: string) {
  const key = `pages.system.plugin.abnormalReason.${reason || 'unknown'}`;
  const label = $t(key);
  return label === key ? reason || '-' : label;
}

function formatFailurePhase(phase?: string) {
  const key = `pages.system.plugin.failurePhase.${phase || 'unknown'}`;
  const label = $t(key);
  return label === key ? phase || '-' : label;
}
</script>

<template>
  <BasicModal
    :footer="false"
    :title="$t('pages.system.plugin.detail.title')"
    class="w-[860px] max-w-[calc(100vw-32px)]"
  >
    <div
      v-if="currentPlugin"
      data-testid="plugin-detail-modal"
      class="flex flex-col gap-4"
    >
      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem :label="$t('pages.system.plugin.fields.name')">
          {{ currentPlugin.name || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.id')">
          {{ currentPlugin.id || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.type')">
          <Tag color="blue">
            {{ formatPluginType(currentPlugin.type) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
          {{ currentPlugin.version || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.runtimeState')">
          <Tag
            :color="getRuntimeStateColor(currentPlugin.runtimeState)"
            data-testid="plugin-detail-runtime-state"
          >
            {{ formatRuntimeState(currentPlugin.runtimeState) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.effectiveVersion')">
          {{ currentPlugin.effectiveVersion || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.discoveredVersion')">
          {{ currentPlugin.discoveredVersion || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.installed')">
          <Tag :color="getInstalledStatusColor(currentPlugin.installed)">
            {{ formatInstalledStatus(currentPlugin.installed) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.status')">
          <Tag :color="getEnabledStatusColor(currentPlugin.enabled)">
            {{ formatEnabledStatus(currentPlugin.enabled) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.startupManagement')">
          <Tag :color="getAutoEnableManagedColor(isAutoEnableManaged)">
            {{ formatAutoEnableManaged(isAutoEnableManaged) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.authorizationStatus')">
          <Tag
            :color="
              getAuthorizationStatusColor(currentPlugin.authorizationStatus)
            "
          >
            {{ formatAuthorizationStatus(currentPlugin.authorizationStatus) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.hasMockData')">
          <Tag
            :color="getBooleanColor(currentPlugin.hasMockData === 1)"
            data-testid="plugin-detail-has-mock-data"
          >
            {{ formatBooleanValue(currentPlugin.hasMockData === 1) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.supportsMultiTenant')">
          <Tag
            :color="getBooleanColor(currentPlugin.supportsMultiTenant === true)"
            data-testid="plugin-detail-supports-multi-tenant"
          >
            {{ formatBooleanValue(currentPlugin.supportsMultiTenant === true) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.tenantProvisioning')">
          <Tag
            :color="
              getBooleanColor(currentPlugin.autoEnableForNewTenants === true)
            "
            data-testid="plugin-detail-tenant-provisioning"
          >
            {{
              formatBooleanValue(currentPlugin.autoEnableForNewTenants === true)
            }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.scopeNature')">
          <Tag
            :color="getScopeNatureColor(currentPlugin.scopeNature)"
            data-testid="plugin-detail-scope-nature"
          >
            {{ formatScopeNature(currentPlugin.scopeNature) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.installMode')">
          <Tag
            :color="getInstallModeColor(currentPlugin.installMode)"
            data-testid="plugin-detail-install-mode"
          >
            {{ formatInstallMode(currentPlugin.installMode) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.installedAt')">
          {{ currentPlugin.installedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.updatedAt')">
          {{ currentPlugin.updatedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.description')" :span="2">
          <div
            data-testid="plugin-detail-description-row"
            class="whitespace-pre-wrap break-words text-[13px] leading-6 text-[var(--ant-color-text-secondary)]"
          >
            {{ currentPlugin.description || '-' }}
          </div>
        </DescriptionsItem>
      </Descriptions>

      <Alert
        v-if="isAutoEnableManaged"
        data-testid="plugin-auto-enable-detail-alert"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.messages.autoEnableDetailAlert')"
      />

      <Alert
        v-if="currentPlugin.runtimeState === 'abnormal'"
        data-testid="plugin-detail-abnormal-alert"
        show-icon
        type="error"
        :message="$t('pages.system.plugin.messages.abnormalManualRepair')"
        :description="formatAbnormalReason(currentPlugin.abnormalReason)"
      />

      <Alert
        v-if="currentPlugin.runtimeState === 'upgrade_failed' && currentPlugin.lastUpgradeFailure"
        data-testid="plugin-detail-upgrade-failure-alert"
        show-icon
        type="error"
        :message="$t('pages.system.plugin.messages.upgradeFailed')"
        :description="
          [
            formatFailurePhase(currentPlugin.lastUpgradeFailure.phase),
            currentPlugin.lastUpgradeFailure.detail,
          ]
            .filter(Boolean)
            .join('：')
        "
      />

      <template v-if="showHostServiceSection">
        <Alert
          v-if="!hasHostServiceDetails"
          data-testid="plugin-detail-empty-host-services"
          show-icon
          type="info"
          :message="$t('pages.system.plugin.messages.emptyHostServices')"
        />

        <template v-else>
          <Alert
            show-icon
            type="info"
            :message="$t('pages.system.plugin.messages.hostServiceSnapshot')"
          />

          <PluginSectionTitle test-id="plugin-host-service-section-title">
            {{ $t('pages.system.plugin.detail.hostServicesTitle') }}
          </PluginSectionTitle>

          <PluginHostServiceCards :cards="hostServiceCards" />
        </template>

        <template v-if="showDeclaredRoutes">
          <PluginSectionTitle test-id="plugin-route-section-title">
            {{ $t('pages.system.plugin.detail.routeListTitle') }}
          </PluginSectionTitle>

          <PluginRouteReviewList :routes="declaredRoutes" />
        </template>
      </template>
    </div>
  </BasicModal>
</template>
