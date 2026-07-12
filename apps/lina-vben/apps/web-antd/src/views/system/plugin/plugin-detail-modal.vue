<script setup lang="ts">
import type {
  PluginRouteReviewItem,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import { $t } from '#/locales';
import { formatTimestamp } from '#/utils/time';
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

const isBuiltinPlugin = computed(() => {
  return currentPlugin.value?.distribution === 'builtin';
});

const pluginPrimaryStatus = computed(() => {
  return resolvePluginPrimaryStatus(currentPlugin.value);
});

const pluginScope = computed(() => {
  return resolvePluginScope(currentPlugin.value);
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

// Detail-only projection: collapse installed / enabled / runtimeState into one
// primary lifecycle status so operators are not asked to reconcile three labels.
type PluginPrimaryStatus =
  | 'not_installed'
  | 'disabled'
  | 'enabled'
  | 'pending_upgrade'
  | 'upgrade_running'
  | 'upgrade_failed'
  | 'abnormal';

function resolvePluginPrimaryStatus(
  plugin: SystemPlugin | null | undefined,
): PluginPrimaryStatus {
  if (!plugin || plugin.installed !== 1) {
    return 'not_installed';
  }
  switch (plugin.runtimeState) {
    case 'upgrade_running':
    case 'upgrade_failed':
    case 'abnormal':
    case 'pending_upgrade': {
      return plugin.runtimeState;
    }
    default: {
      return plugin.enabled === 1 ? 'enabled' : 'disabled';
    }
  }
}

function formatPluginPrimaryStatus(status: PluginPrimaryStatus) {
  switch (status) {
    case 'not_installed': {
      return $t('pages.system.plugin.pluginStatus.notInstalled');
    }
    case 'disabled': {
      return $t('pages.system.plugin.pluginStatus.disabled');
    }
    case 'enabled': {
      return $t('pages.system.plugin.pluginStatus.enabled');
    }
    case 'pending_upgrade':
    case 'upgrade_running':
    case 'upgrade_failed':
    case 'abnormal': {
      return $t(`pages.system.plugin.runtimeState.${status}`);
    }
    default: {
      return '-';
    }
  }
}

function getPluginPrimaryStatusColor(status: PluginPrimaryStatus) {
  switch (status) {
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
    case 'enabled': {
      return 'green';
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

function formatDistribution(distribution?: string) {
  if (distribution === 'builtin') {
    return $t('pages.system.plugin.builtinManaged');
  }
  if (distribution === 'managed') {
    return $t('pages.system.plugin.managedDistribution');
  }
  return distribution || '-';
}

function getDistributionColor(distribution?: string) {
  return distribution === 'builtin' ? 'purple' : 'default';
}

function formatBooleanValue(value: boolean) {
  return value ? $t('pages.common.yes') : $t('pages.common.no');
}

function getBooleanColor(value: boolean) {
  return value ? 'green' : 'default';
}

// Detail-only projection: installMode is the operational truth; scopeNature
// only distinguishes platform-only capability when the plugin runs globally.
type PluginScope =
  | 'platform_only'
  | 'global'
  | 'tenant_scoped';

function resolvePluginScope(
  plugin: SystemPlugin | null | undefined,
): PluginScope | '' {
  if (!plugin) {
    return '';
  }
  if (plugin.installMode === 'tenant_scoped') {
    return 'tenant_scoped';
  }
  if (plugin.scopeNature === 'platform_only') {
    return 'platform_only';
  }
  if (
    plugin.installMode === 'global' ||
    plugin.scopeNature === 'tenant_aware'
  ) {
    return 'global';
  }
  return '';
}

function formatPluginScope(scope: PluginScope | '') {
  switch (scope) {
    case 'platform_only': {
      return $t('pages.system.plugin.pluginScope.platformOnly');
    }
    case 'global': {
      return $t('pages.system.plugin.pluginScope.global');
    }
    case 'tenant_scoped': {
      return $t('pages.system.plugin.pluginScope.tenantScoped');
    }
    default: {
      return '-';
    }
  }
}

function getPluginScopeColor(scope: PluginScope | '') {
  return scope === 'tenant_scoped' ? 'green' : 'blue';
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
      <Descriptions
        bordered
        size="small"
        :column="2"
        class="plugin-detail-descriptions"
        data-testid="plugin-detail-descriptions"
      >
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
        <DescriptionsItem :label="$t('pages.system.plugin.fields.pluginStatus')">
          <Tag
            :color="getPluginPrimaryStatusColor(pluginPrimaryStatus)"
            data-testid="plugin-detail-plugin-status"
          >
            {{ formatPluginPrimaryStatus(pluginPrimaryStatus) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.effectiveVersion')">
          {{
            currentPlugin.effectiveVersion || currentPlugin.version || '-'
          }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.discoveredVersion')">
          {{
            currentPlugin.discoveredVersion || currentPlugin.version || '-'
          }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.distribution')">
          <Tag
            :color="getDistributionColor(currentPlugin.distribution)"
            data-testid="plugin-detail-distribution"
          >
            {{ formatDistribution(currentPlugin.distribution) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.startupManagement')">
          <Tag :color="getAutoEnableManagedColor(isAutoEnableManaged)">
            {{ formatAutoEnableManaged(isAutoEnableManaged) }}
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
        <DescriptionsItem :label="$t('pages.system.plugin.fields.pluginScope')">
          <Tag
            :color="getPluginScopeColor(pluginScope)"
            data-testid="plugin-detail-plugin-scope"
          >
            {{ formatPluginScope(pluginScope) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.installedAt')">
          {{ formatTimestamp(currentPlugin.installedAt) }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.updatedAt')">
          {{ formatTimestamp(currentPlugin.updatedAt) }}
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
        v-if="isBuiltinPlugin"
        data-testid="plugin-builtin-detail-alert"
        show-icon
        type="info"
        :message="$t('pages.system.plugin.messages.builtinDetailAlert')"
      />

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

<style scoped>
/* Keep the left-hand label column on one line; multi-word English labels
   (e.g. Authorization Status / Effective Version) otherwise wrap in the
   two-column bordered Descriptions table. */
:deep(.plugin-detail-descriptions .ant-descriptions-item-label) {
  min-width: 112px;
  white-space: nowrap;
}
</style>
