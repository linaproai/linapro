<script setup lang="ts">
import type {
  HostServicePermissionItem,
  PluginAuthorizationPayload,
  PluginManifestSnapshot,
  PluginUpgradeHostServiceChange,
  PluginUpgradePreview,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Alert,
  Descriptions,
  DescriptionsItem,
  Empty,
  message,
  Tag,
} from 'ant-design-vue';

import { pluginUpgrade, pluginUpgradePreview } from '#/api/system/plugin';
import { $t } from '#/locales';

import PluginDependencySummary from './plugin-dependency-summary.vue';
import PluginHostServiceCards from './plugin-host-service-cards.vue';
import PluginSectionTitle from './plugin-section-title.vue';
import {
  buildPluginAuthorizationHostServiceCards,
  formatServiceLabel,
  sortHostServices,
} from './plugin-host-service-view';

const emit = defineEmits<{ reload: [] }>();

const currentPlugin = ref<null | SystemPlugin>(null);
const preview = ref<null | PluginUpgradePreview>(null);
const previewLoading = ref(false);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

const currentTitle = computed(() => {
  if (!currentPlugin.value) {
    return $t('pages.system.plugin.upgrade.title');
  }
  return $t('pages.system.plugin.upgrade.titleWithName', {
    name: currentPlugin.value.name || currentPlugin.value.id,
  });
});

const targetRequestedServices = computed<HostServicePermissionItem[]>(() => {
  return sortHostServices(preview.value?.toManifest?.requestedHostServices);
});

const targetHostServiceCards = computed(() => {
  return buildPluginAuthorizationHostServiceCards(targetRequestedServices.value, {
    authorizationRequired:
      preview.value?.hostServicesDiff.authorizationRequired === true,
    buildScopeContainerTestId: (service) => {
      return currentPlugin.value
        ? `plugin-upgrade-host-service-${currentPlugin.value.id}-${service}`
        : undefined;
    },
    buildScopeItemTestIdPrefix: (service) => {
      return currentPlugin.value
        ? `plugin-upgrade-host-service-item-${currentPlugin.value.id}-${service}`
        : undefined;
    },
    targetSummaryBadgeColor: 'gold',
  });
});

const hostServiceChanges = computed(() => {
  const diff = preview.value?.hostServicesDiff;
  return [
    ...tagHostServiceChanges(diff?.added, 'added'),
    ...tagHostServiceChanges(diff?.changed, 'changed'),
    ...tagHostServiceChanges(diff?.removed, 'removed'),
  ];
});

const hasHostServiceReview = computed(() => {
  return targetHostServiceCards.value.length > 0 || hostServiceChanges.value.length > 0;
});

const hasDependencyContent = computed(() => {
  const check = preview.value?.dependencyCheck;
  return (
    previewLoading.value ||
    (check?.blockers ?? []).length > 0 ||
    (check?.autoInstallPlan ?? []).length > 0 ||
    (check?.manualInstallRequired ?? []).length > 0 ||
    (check?.softUnsatisfied ?? []).length > 0 ||
    (check?.cycle ?? []).length > 0 ||
    check?.framework?.status === 'unsatisfied'
  );
});

const dependencyBlocked = computed(() => {
  return (
    previewLoading.value ||
    (preview.value?.dependencyCheck?.blockers ?? []).length > 0
  );
});

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  preview.value = null;
  await refreshPreview();
}

async function refreshPreview() {
  if (!currentPlugin.value?.id) {
    return;
  }
  previewLoading.value = true;
  updateConfirmDisabled();
  try {
    preview.value = await pluginUpgradePreview(currentPlugin.value.id);
  } catch (error) {
    message.error(resolveRuntimeErrorMessage(error));
  } finally {
    previewLoading.value = false;
    updateConfirmDisabled();
  }
}

async function handleConfirm() {
  if (!currentPlugin.value || !preview.value) {
    return;
  }
  if (dependencyBlocked.value) {
    message.warning($t('pages.system.plugin.upgrade.resolveBeforeUpgrade'));
    return;
  }

  try {
    modalApi.lock(true);
    await pluginUpgrade(currentPlugin.value.id, buildAuthorizationPayload());
    message.success($t('pages.system.plugin.messages.upgraded'));
    emit('reload');
    handleClosed();
  } catch (error) {
    message.error(resolveRuntimeErrorMessage(error));
    await refreshPreview();
    emit('reload');
  } finally {
    modalApi.lock(false);
  }
}

function buildAuthorizationPayload(): PluginAuthorizationPayload | undefined {
  if (preview.value?.hostServicesDiff.authorizationRequired !== true) {
    return undefined;
  }
  return {
    authorization: {
      services: targetRequestedServices.value
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
          service: service.service,
          tables:
            service.service === 'data'
              ? [...(service.tables ?? [])]
              : undefined,
        })),
    },
  };
}

function hasServiceTargets(service: HostServicePermissionItem) {
  return (
    (service.paths ?? []).length > 0 ||
    (service.tables ?? []).length > 0 ||
    (service.cronItems ?? []).length > 0 ||
    (service.resources ?? []).length > 0
  );
}

function updateConfirmDisabled() {
  modalApi.setState({ confirmDisabled: previewLoading.value || !preview.value });
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
  preview.value = null;
  previewLoading.value = false;
  updateConfirmDisabled();
}

function tagHostServiceChanges(
  items: PluginUpgradeHostServiceChange[] | undefined,
  kind: 'added' | 'changed' | 'removed',
) {
  return (items ?? []).map((item) => ({ ...item, kind }));
}

function formatPluginType(type?: string) {
  if (type === 'source') {
    return $t('pages.system.plugin.type.source');
  }
  if (type === 'dynamic') {
    return $t('pages.system.plugin.type.dynamic');
  }
  return type || '-';
}

function formatSnapshotName(snapshot?: PluginManifestSnapshot) {
  if (!snapshot) {
    return '-';
  }
  return snapshot.name || snapshot.id || '-';
}

function formatRiskHint(key: string) {
  const localized = $t(key);
  return localized === key ? key : localized;
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

function getHostServiceChangeColor(kind: string) {
  switch (kind) {
    case 'added': {
      return 'green';
    }
    case 'removed': {
      return 'red';
    }
    default: {
      return 'blue';
    }
  }
}

function formatHostServiceChangeKind(kind: string) {
  const key = `pages.system.plugin.upgrade.hostServiceChange.${kind}`;
  const label = $t(key);
  return label === key ? kind : label;
}

function summarizeHostServiceChange(change: PluginUpgradeHostServiceChange) {
  const methods = [
    ...(change.fromMethods ?? []),
    ...(change.toMethods ?? []),
  ].filter(Boolean);
  const resourceCount =
    change.fromResourceCount !== change.toResourceCount
      ? `${change.fromResourceCount} -> ${change.toResourceCount}`
      : `${change.toResourceCount}`;
  return [
    methods.length > 0 ? methods.join(', ') : '',
    $t('pages.system.plugin.upgrade.resourceCount', { count: resourceCount }),
  ]
    .filter(Boolean)
    .join(' · ');
}

function extractRuntimeErrorEnvelope(error: unknown): null | {
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
} {
  if (!error || typeof error !== 'object') {
    return null;
  }
  const response = (error as { response?: { data?: unknown } }).response;
  return (response?.data ?? error) as {
    message?: string;
    messageKey?: string;
    messageParams?: Record<string, unknown>;
  };
}

function resolveRuntimeErrorMessage(error: unknown) {
  const envelope = extractRuntimeErrorEnvelope(error);
  if (envelope?.messageKey) {
    const localized = $t(envelope.messageKey, envelope.messageParams || {});
    if (localized && localized !== envelope.messageKey) {
      return localized;
    }
  }
  if (envelope?.message) {
    return envelope.message;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return $t('ui.fallback.http.internalServerError');
}
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :confirm-text="$t('pages.system.plugin.upgrade.confirm')"
    :fullscreen-button="false"
    :title="currentTitle"
    class="w-[920px] max-w-[calc(100vw-32px)]"
  >
    <div
      v-if="currentPlugin"
      class="flex flex-col gap-4"
      data-testid="plugin-upgrade-modal"
    >
      <Alert
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.upgrade.banner')"
      />

      <Alert
        v-if="preview?.runtimeState === 'upgrade_failed'"
        show-icon
        type="error"
        :message="$t('pages.system.plugin.upgrade.retryBanner')"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem :label="$t('pages.system.plugin.fields.name')">
          {{ currentPlugin.name || currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.id')">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.type')">
          {{ formatPluginType(currentPlugin.type) }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.runtimeState')">
          <Tag :color="getRuntimeStateColor(preview?.runtimeState || currentPlugin.runtimeState)">
            {{ formatRuntimeState(preview?.runtimeState || currentPlugin.runtimeState) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.effectiveVersion')">
          {{ preview?.effectiveVersion || currentPlugin.effectiveVersion || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.discoveredVersion')">
          {{ preview?.discoveredVersion || currentPlugin.discoveredVersion || '-' }}
        </DescriptionsItem>
      </Descriptions>

      <template v-if="preview">
        <PluginSectionTitle test-id="plugin-upgrade-version-section-title">
          {{ $t('pages.system.plugin.upgrade.versionSection') }}
        </PluginSectionTitle>

        <div class="grid gap-3 lg:grid-cols-2">
          <div
            class="rounded-md border border-[var(--ant-color-border)] p-3"
            data-testid="plugin-upgrade-from-manifest"
          >
            <div class="mb-2 text-sm font-semibold text-[var(--ant-color-text)]">
              {{ $t('pages.system.plugin.upgrade.fromManifest') }}
            </div>
            <Descriptions size="small" :column="1">
              <DescriptionsItem :label="$t('pages.system.plugin.fields.name')">
                {{ formatSnapshotName(preview.fromManifest) }}
              </DescriptionsItem>
              <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
                {{ preview.fromManifest?.version || '-' }}
              </DescriptionsItem>
              <DescriptionsItem :label="$t('pages.system.plugin.fields.description')">
                {{ preview.fromManifest?.description || '-' }}
              </DescriptionsItem>
            </Descriptions>
          </div>

          <div
            class="rounded-md border border-[var(--ant-color-primary)] bg-[var(--ant-color-primary-bg)] p-3"
            data-testid="plugin-upgrade-to-manifest"
          >
            <div class="mb-2 text-sm font-semibold text-[var(--ant-color-primary)]">
              {{ $t('pages.system.plugin.upgrade.toManifest') }}
            </div>
            <Descriptions size="small" :column="1">
              <DescriptionsItem :label="$t('pages.system.plugin.fields.name')">
                {{ formatSnapshotName(preview.toManifest) }}
              </DescriptionsItem>
              <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
                {{ preview.toManifest?.version || '-' }}
              </DescriptionsItem>
              <DescriptionsItem :label="$t('pages.system.plugin.fields.description')">
                {{ preview.toManifest?.description || '-' }}
              </DescriptionsItem>
            </Descriptions>
          </div>
        </div>

        <template v-if="hasDependencyContent">
          <PluginSectionTitle test-id="plugin-upgrade-dependency-section-title">
            {{ $t('pages.system.plugin.dependency.title') }}
          </PluginSectionTitle>
          <PluginDependencySummary
            :check="preview.dependencyCheck"
            :loading="previewLoading"
            mode="install"
          />
        </template>

        <PluginSectionTitle test-id="plugin-upgrade-sql-section-title">
          {{ $t('pages.system.plugin.upgrade.sqlSection') }}
        </PluginSectionTitle>
        <div class="grid gap-2 md:grid-cols-4" data-testid="plugin-upgrade-sql-summary">
          <Tag color="blue">
            {{ $t('pages.system.plugin.upgrade.installSqlCount', { count: preview.sqlSummary.installSqlCount }) }}
          </Tag>
          <Tag color="purple">
            {{ $t('pages.system.plugin.upgrade.runtimeSqlCount', { count: preview.sqlSummary.runtimeSqlAssetCount }) }}
          </Tag>
          <Tag color="default">
            {{ $t('pages.system.plugin.upgrade.uninstallSqlCount', { count: preview.sqlSummary.uninstallSqlCount }) }}
          </Tag>
          <Tag color="gold">
            {{ $t('pages.system.plugin.upgrade.mockSqlExcluded', { count: preview.sqlSummary.mockSqlCount }) }}
          </Tag>
        </div>

        <template v-if="hasHostServiceReview">
          <PluginSectionTitle test-id="plugin-upgrade-host-service-section-title">
            {{ $t('pages.system.plugin.upgrade.hostServiceSection') }}
          </PluginSectionTitle>
          <div
            v-if="hostServiceChanges.length > 0"
            class="flex flex-wrap gap-2"
            data-testid="plugin-upgrade-host-service-diff"
          >
            <Tag
              v-for="change in hostServiceChanges"
              :key="`${change.kind}-${change.service}`"
              :color="getHostServiceChangeColor(change.kind)"
            >
              {{ formatHostServiceChangeKind(change.kind) }}
              · {{ formatServiceLabel(change.service) }}
              · {{ summarizeHostServiceChange(change) }}
            </Tag>
          </div>
          <PluginHostServiceCards
            v-if="targetHostServiceCards.length > 0"
            :cards="targetHostServiceCards"
          />
        </template>

        <PluginSectionTitle test-id="plugin-upgrade-risk-section-title">
          {{ $t('pages.system.plugin.upgrade.riskSection') }}
        </PluginSectionTitle>
        <div v-if="preview.riskHints.length > 0" class="flex flex-col gap-2">
          <Alert
            v-for="hint in preview.riskHints"
            :key="hint"
            show-icon
            type="warning"
            :message="formatRiskHint(hint)"
          />
        </div>
        <Empty
          v-else
          :description="$t('pages.system.plugin.upgrade.noRiskHints')"
        />
      </template>
    </div>
  </BasicModal>
</template>
