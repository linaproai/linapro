<script setup lang="ts">
import type {
  PluginListItem,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { defineAsyncComponent, h, ref } from 'vue';
import { useRouter } from 'vue-router';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';
import { useAccessStore } from '@vben/stores';

import { message, Modal, Space, Switch, Tag, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  pluginDisable,
  pluginDetail,
  pluginEnable,
  pluginList,
  pluginSync,
  pluginUpdateTenantProvisioningPolicy,
} from '#/api/system/plugin';
import { $t } from '#/locales';
import {
  hasPluginManagementPage,
  resolvePluginManagementPath,
} from '#/plugins/plugin-management-route';
import { notifyPluginRegistryChanged } from '#/plugins/slot-registry';
import { closePluginTabs } from '#/plugins/tabbar-cleanup';
import { formatTimestamp } from '#/utils/time';

const PluginDetailModal = defineAsyncComponent(
  () => import('./plugin-detail-modal.vue'),
);
const PluginDynamicUploadModal = defineAsyncComponent(
  () => import('./plugin-dynamic-upload-modal.vue'),
);
const PluginHostServiceAuthModal = defineAsyncComponent(
  () => import('./plugin-host-service-auth-modal.vue'),
);
const PluginUninstallModal = defineAsyncComponent(
  () => import('./plugin-uninstall-modal.vue'),
);
const PluginUpgradeModal = defineAsyncComponent(
  () => import('./plugin-upgrade-modal.vue'),
);
const LifecyclePreconditionDialog = defineAsyncComponent(
  () => import('#/views/platform/plugins/lifecycle-precondition-dialog.vue'),
);

const [DetailModal, detailModalApi] = useVbenModal({
  connectedComponent: PluginDetailModal,
});

const [DynamicUploadModal, dynamicUploadModalApi] = useVbenModal({
  connectedComponent: PluginDynamicUploadModal,
});

const [HostServiceAuthModal, hostServiceAuthModalApi] = useVbenModal({
  connectedComponent: PluginHostServiceAuthModal,
});

const [UninstallModal, uninstallModalApi] = useVbenModal({
  connectedComponent: PluginUninstallModal,
});

const [UpgradeModal, upgradeModalApi] = useVbenModal({
  connectedComponent: PluginUpgradeModal,
});

const [LifecyclePreconditionModal, lifecyclePreconditionModalApi] = useVbenModal({
  connectedComponent: LifecyclePreconditionDialog,
});

const typeColorMap: Record<string, string> = {
  dynamic: 'green',
  source: 'blue',
};

const pluginAccessCodes = {
  disable: 'plugin:disable',
  edit: 'plugin:edit',
  enable: 'plugin:enable',
  install: 'plugin:install',
  uninstall: 'plugin:uninstall',
} as const;

const { hasAccessByCodes } = useAccess();
const accessStore = useAccessStore();
const router = useRouter();
const statusChangingPluginIds = ref<Record<string, boolean>>({});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Input',
        fieldName: 'id',
        label: $t('pages.system.plugin.fields.id'),
      },
      {
        component: 'Input',
        fieldName: 'name',
        label: $t('pages.system.plugin.fields.name'),
      },
      {
        component: 'Select',
        fieldName: 'type',
        label: $t('pages.system.plugin.fields.type'),
        componentProps: {
          options: [
            {
              label: $t('pages.system.plugin.type.source'),
              value: 'source',
            },
            {
              label: $t('pages.system.plugin.type.dynamic'),
              value: 'dynamic',
            },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'installed',
        label: $t('pages.system.plugin.fields.installed'),
        componentProps: {
          options: [
            {
              label: $t('pages.system.plugin.installed.connected'),
              value: 1,
            },
            {
              label: $t('pages.system.plugin.installed.notInstalled'),
              value: 0,
            },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: $t('pages.common.status'),
        componentProps: {
          options: [
            { label: $t('pages.status.enabled'), value: 1 },
            { label: $t('pages.status.disabled'), value: 0 },
          ],
        },
      },
    ],
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    columns: [
      {
        align: 'left',
        field: 'id',
        headerAlign: 'center',
        // Slightly wider than name for long plugin IDs without dominating the row.
        minWidth: 220,
        title: $t('pages.system.plugin.fields.id'),
      },
      {
        align: 'left',
        className: 'plugin-name-column',
        field: 'name',
        headerAlign: 'center',
        // Compact width; builtin / auto-enable badges share the cell flex row.
        minWidth: 180,
        showOverflow: false,
        slots: { default: 'name' },
        title: $t('pages.system.plugin.fields.name'),
      },
      {
        align: 'left',
        className: 'plugin-description-column',
        field: 'description',
        headerAlign: 'center',
        // Wider than name so capability blurbs stay more readable than titles.
        minWidth: 220,
        showOverflow: false,
        slots: { default: 'description' },
        title: $t('pages.system.plugin.fields.description'),
      },
      {
        field: 'version',
        showOverflow: false,
        slots: { default: 'version' },
        title: $t('pages.system.plugin.fields.version'),
        width: 148,
      },
      {
        field: 'type',
        slots: { default: 'type', header: 'typeHeader' },
        title: $t('pages.system.plugin.fields.type'),
        width: 108,
      },
      {
        field: 'enabled',
        slots: { default: 'enabled' },
        title: $t('pages.common.status'),
        width: 96,
      },
      {
        field: 'runtimeState',
        slots: { default: 'runtimeState', header: 'runtimeStateHeader' },
        title: $t('pages.system.plugin.fields.runtimeState'),
        width: 112,
      },
      {
        field: 'hasMockData',
        slots: { default: 'hasMockData', header: 'hasMockDataHeader' },
        title: $t('pages.system.plugin.fields.hasMockData'),
        width: 104,
      },
      {
        field: 'supportsMultiTenant',
        slots: {
          default: 'supportsMultiTenant',
          header: 'supportsMultiTenantHeader',
        },
        title: $t('pages.system.plugin.fields.supportsMultiTenant'),
        width: 122,
      },
      {
        field: 'autoEnableForNewTenants',
        slots: {
          default: 'tenantProvisioning',
          header: 'tenantProvisioningHeader',
        },
        title: $t('pages.system.plugin.fields.tenantProvisioning'),
        width: 126,
      },
      {
        field: 'installedAt',
        formatter: ({
          cellValue,
        }: {
          cellValue?: null | number | string;
        }) => formatTimestamp(cellValue),
        title: $t('pages.system.plugin.fields.installedAt'),
        width: 180,
      },
      {
        field: 'updatedAt',
        formatter: ({
          cellValue,
        }: {
          cellValue?: null | number | string;
        }) => formatTimestamp(cellValue),
        title: $t('pages.common.updatedAt'),
        width: 180,
      },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: $t('pages.common.actions'),
        // At most three small ghost buttons (detail + manage + one lifecycle
        // action); keep width tight so the column does not waste horizontal
        // space while still fitting a single non-wrapping row.
        width: 200,
      },
    ],
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    showOverflow: 'ellipsis',
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await pluginList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    rowClassName: 'cursor-pointer',
    id: 'system-plugin-index',
  },
  gridEvents: {
    cellClick: ({
      $event,
      column,
      row,
    }: {
      $event?: Event;
      column?: { field?: string };
      row: PluginListItem;
    }) => {
      if (shouldIgnorePluginRowClick(column?.field, $event)) {
        return;
      }
      void handleDetail(row);
    },
  },
});

/** Columns / controls that must not open the detail modal on click. */
const pluginRowClickIgnoredFields = new Set([
  'action',
  'autoEnableForNewTenants',
  'enabled',
]);

function shouldIgnorePluginRowClick(
  columnField?: string,
  event?: Event,
): boolean {
  if (columnField && pluginRowClickIgnoredFields.has(columnField)) {
    return true;
  }
  const target = event?.target;
  if (!(target instanceof Element)) {
    return false;
  }
  return Boolean(
    target.closest(
      'button, a, input, textarea, select, .ant-switch, .ant-checkbox, .ant-radio, [role="switch"], [role="button"], [role="checkbox"]',
    ),
  );
}

function getPluginTypeLabel(type: string) {
  return type === 'source'
    ? $t('pages.system.plugin.type.source')
    : $t('pages.system.plugin.type.dynamic');
}

function getPluginTypeColor(type: string) {
  return typeColorMap[type === 'source' ? 'source' : 'dynamic'] || 'default';
}

function isAutoEnableManaged(row: PluginListItem) {
  return row.autoEnableManaged === 1;
}

function isBuiltinPlugin(row: PluginListItem) {
  return row.distribution === 'builtin';
}

function hasPluginMockData(row: PluginListItem) {
  return row.hasMockData === 1;
}

function supportsPluginMultiTenant(row: PluginListItem) {
  return row.supportsMultiTenant === true;
}

function formatPluginVersion(row: PluginListItem) {
  const effective = getPluginEffectiveVersion(row);
  const discovered = getPluginDiscoveredVersion(row);
  return effective === discovered ? effective : `${effective} -> ${discovered}`;
}

function getPluginEffectiveVersion(row: PluginListItem) {
  return row.effectiveVersion || row.version || '-';
}

function getPluginDiscoveredVersion(row: PluginListItem) {
  return row.discoveredVersion || row.version || '-';
}

function hasPluginVersionDiff(row: PluginListItem) {
  return getPluginEffectiveVersion(row) !== getPluginDiscoveredVersion(row);
}

function isRuntimeUpgradeAvailable(row: PluginListItem) {
  return (
    !isBuiltinPlugin(row) &&
    row.installed === 1 &&
    row.upgradeAvailable === true &&
    (row.runtimeState === 'pending_upgrade' || row.runtimeState === 'upgrade_failed')
  );
}

function isRuntimeAbnormal(row: PluginListItem) {
  return row.runtimeState === 'abnormal';
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

function buildRuntimeStateTooltip(row: PluginListItem) {
  if (row.runtimeState === 'pending_upgrade') {
    return $t('pages.system.plugin.runtimeStateHint.pendingUpgrade', {
      discoveredVersion: row.discoveredVersion || '-',
      effectiveVersion: row.effectiveVersion || '-',
    });
  }
  if (row.runtimeState === 'upgrade_failed') {
    return row.lastUpgradeFailure?.detail
      ? $t('pages.system.plugin.runtimeStateHint.upgradeFailedWithDetail', {
          detail: row.lastUpgradeFailure.detail,
        })
      : $t('pages.system.plugin.runtimeStateHint.upgradeFailed');
  }
  if (row.runtimeState === 'abnormal') {
    return $t('pages.system.plugin.runtimeStateHint.abnormal', {
      reason: formatAbnormalReason(row.abnormalReason),
    });
  }
  return $t('pages.system.plugin.runtimeStateHint.normal');
}

function formatAbnormalReason(reason?: string) {
  const key = `pages.system.plugin.abnormalReason.${reason || 'unknown'}`;
  const label = $t(key);
  return label === key ? reason || '-' : label;
}

function isTenantProvisioningPolicySupported(row: PluginListItem) {
  return (
    !isBuiltinPlugin(row) &&
    supportsPluginMultiTenant(row) &&
    row.scopeNature === 'tenant_aware' &&
    row.installMode === 'tenant_scoped'
  );
}

function buildBuiltinPluginTooltip(row: PluginListItem) {
  return $t('pages.system.plugin.messages.builtinTooltip', {
    pluginId: row.id,
  });
}

function buildAutoEnableManagedTooltip(row: PluginListItem) {
  return $t('pages.system.plugin.messages.autoEnableTooltip', {
    pluginId: row.id,
  });
}

function buildAutoEnableManagedRuntimeHint(actionLabel: string) {
  return $t('pages.system.plugin.messages.autoEnableRuntimeHint', {
    actionLabel,
  });
}

function getColumnHelpAriaLabel(label: string) {
  return $t('pages.system.plugin.columnHelp.ariaLabel', { label });
}

async function confirmAutoEnableManagedAction(actionLabel: string) {
  return await new Promise<boolean>((resolve) => {
    Modal.confirm({
      cancelText: $t('pages.common.cancel'),
      content: h('div', { class: 'whitespace-pre-line leading-6' }, [
        buildAutoEnableManagedRuntimeHint(actionLabel),
      ]),
      okText: $t('pages.system.plugin.actions.continueAction', { actionLabel }),
      title: $t('pages.system.plugin.messages.autoEnableConfirmTitle', {
        actionLabel,
      }),
      onCancel: () => resolve(false),
      onOk: () => resolve(true),
    });
  });
}

function canInstallPlugin() {
  return hasAccessByCodes([pluginAccessCodes.install]);
}

function canInstallAndEnablePlugin() {
  return [pluginAccessCodes.install, pluginAccessCodes.enable].every((code) =>
    hasAccessByCodes([code]),
  );
}

function canSyncPlugins() {
  return hasAccessByCodes([pluginAccessCodes.install]);
}

function canEditPluginPolicy() {
  return hasAccessByCodes([pluginAccessCodes.edit]);
}

function canUninstallPlugin() {
  return hasAccessByCodes([pluginAccessCodes.uninstall]);
}

function canTogglePluginStatus(row: PluginListItem) {
  if (isBuiltinPlugin(row)) {
    return false;
  }
  return row.enabled === 1
    ? hasAccessByCodes([pluginAccessCodes.disable])
    : hasAccessByCodes([pluginAccessCodes.enable]);
}

function isPluginStatusChanging(row: PluginListItem) {
  return statusChangingPluginIds.value[row.id] === true;
}

function setPluginStatusChanging(pluginId: string, changing: boolean) {
  const next = { ...statusChangingPluginIds.value };
  if (changing) {
    next[pluginId] = true;
  } else {
    delete next[pluginId];
  }
  statusChangingPluginIds.value = next;
}

async function loadPluginDetail(row: PluginListItem): Promise<SystemPlugin> {
  return await pluginDetail(row.id);
}

async function handleDetail(row: PluginListItem) {
  const detail = await loadPluginDetail(row);
  detailModalApi.setData({ row: detail });
  detailModalApi.open();
}

function canOpenPluginManagement(row: PluginListItem) {
  return row.installed === 1 && hasPluginManagementPage(row.id);
}

function buildPluginManagementDisabledTooltip(row: PluginListItem) {
  if (canOpenPluginManagement(row)) {
    return undefined;
  }
  if (row.installed !== 1) {
    return $t('pages.system.plugin.messages.installFirst');
  }
  return $t('pages.system.plugin.messages.noManagementPage');
}

async function handleOpenManagement(row: PluginListItem) {
  if (!canOpenPluginManagement(row)) {
    return;
  }
  const targetPath = resolvePluginManagementPath(
    row.id,
    router,
    accessStore.accessMenus,
  );
  if (!targetPath) {
    message.warning(
      $t('pages.system.plugin.messages.managementRouteUnavailable'),
    );
    return;
  }
  await router.push(targetPath);
}

async function handleStatusChange(row: PluginListItem, checked: boolean) {
  if (isBuiltinPlugin(row)) {
    return;
  }
  if (isPluginStatusChanging(row)) {
    return;
  }
  if (row.installed !== 1) {
    message.warning($t('pages.system.plugin.messages.installFirst'));
    return;
  }
  if (!canTogglePluginStatus(row)) {
    message.warning($t('pages.system.plugin.messages.noStatusPermission'));
    return;
  }
  if (
    checked &&
    row.authorizationRequired === 1 &&
    row.authorizationStatus !== 'confirmed'
  ) {
    const detail = await loadPluginDetail(row);
    hostServiceAuthModalApi.setData({ mode: 'enable', row: detail });
    hostServiceAuthModalApi.open();
    return;
  }
  if (!checked && isAutoEnableManaged(row)) {
    const confirmed = await confirmAutoEnableManagedAction(
      $t('pages.status.disabled'),
    );
    if (!confirmed) {
      return;
    }
  }
  const previousEnabled = row.enabled;
  const nextEnabled = checked ? 1 : 0;

  setPluginStatusChanging(row.id, true);
  row.enabled = nextEnabled;
  try {
    await (checked ? pluginEnable : pluginDisable)(row.id);
    if (!checked) {
      await closePluginTabs(row.id);
    }
    await notifyPluginRegistryChanged();
    message.success(
      checked
        ? $t('pages.system.plugin.messages.enabled')
        : $t('pages.system.plugin.messages.disabled'),
    );
  } catch {
    row.enabled = previousEnabled;
    await gridApi.query();
  } finally {
    setPluginStatusChanging(row.id, false);
  }
}

async function handleTenantProvisioningPolicyChange(
  row: PluginListItem,
  checked: boolean,
) {
  if (isBuiltinPlugin(row)) {
    return;
  }
  if (!canEditPluginPolicy()) {
    message.warning($t('pages.system.plugin.messages.noPolicyPermission'));
    return;
  }
  if (!isTenantProvisioningPolicySupported(row)) {
    message.warning(
      $t('pages.system.plugin.messages.tenantProvisioningUnsupported'),
    );
    return;
  }
  await pluginUpdateTenantProvisioningPolicy(row.id, checked);
  row.autoEnableForNewTenants = checked;
  message.success($t('pages.system.plugin.messages.tenantProvisioningUpdated'));
}

async function handleInstall(row: PluginListItem) {
  if (isBuiltinPlugin(row)) {
    return;
  }
  if (!canInstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noInstallPermission'));
    return;
  }
  const detail = await loadPluginDetail(row);
  hostServiceAuthModalApi.setData({
    allowInstallAndEnable: canInstallAndEnablePlugin(),
    mode: 'install',
    row: detail,
  });
  hostServiceAuthModalApi.open();
}

async function handleOpenUninstall(row: PluginListItem) {
  if (isBuiltinPlugin(row)) {
    return;
  }
  if (!canUninstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noUninstallPermission'));
    return;
  }
  const detail = await loadPluginDetail(row);
  uninstallModalApi.setData({ row: detail });
  uninstallModalApi.open();
}

async function handleOpenUpgrade(row: PluginListItem) {
  if (isBuiltinPlugin(row)) {
    return;
  }
  if (!canInstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noUpgradePermission'));
    return;
  }
  const detail = await loadPluginDetail(row);
  upgradeModalApi.setData({ row: detail });
  upgradeModalApi.open();
}

async function handleSync() {
  if (!canSyncPlugins()) {
    message.warning($t('pages.system.plugin.messages.noInstallPermission'));
    return;
  }
  const res = await pluginSync();
  await notifyPluginRegistryChanged();
  const total = typeof res?.total === 'number' ? res.total : 0;
  message.success($t('pages.system.plugin.messages.syncSuccess', { total }));
  await gridApi.query();
}

function handleOpenDynamicUpload() {
  if (!canInstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noInstallPermission'));
    return;
  }
  dynamicUploadModalApi.open();
}

async function handleDynamicUploadReload() {
  await notifyPluginRegistryChanged();
  await gridApi.query();
}

async function handleHostServiceAuthReload() {
  await notifyPluginRegistryChanged();
  await gridApi.query();
}

async function handleUninstallReload(payload?: { pluginId?: string }) {
  if (payload?.pluginId) {
    await closePluginTabs(payload.pluginId);
  }
  await notifyPluginRegistryChanged();
  await gridApi.query();
}

async function handleUpgradeReload() {
  await notifyPluginRegistryChanged();
  await gridApi.query();
}

function handleLifecyclePrecondition(payload: {
  force: () => Promise<void>;
  pluginId: string;
  reasons: string[];
}) {
  lifecyclePreconditionModalApi.setData(payload);
  lifecyclePreconditionModalApi.open();
}

async function handleLifecyclePreconditionForce(payload: { pluginId: string }) {
  const data = lifecyclePreconditionModalApi.getData<{
    force?: () => Promise<void>;
    pluginId?: string;
  }>();
  if (!data.force || data.pluginId !== payload.pluginId) {
    return;
  }
  lifecyclePreconditionModalApi.lock(true);
  try {
    await data.force();
    lifecyclePreconditionModalApi.close();
  } catch {
    // The force callback already surfaces the backend error locally.
  } finally {
    lifecyclePreconditionModalApi.lock(false);
  }
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('pages.system.plugin.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            data-testid="plugin-dynamic-upload-trigger"
            type="primary"
            v-access:code="pluginAccessCodes.install"
            @click="handleOpenDynamicUpload"
          >
            {{ $t('pages.system.plugin.actions.uploadPlugin') }}
          </a-button>
          <a-button
            v-access:code="pluginAccessCodes.install"
            type="primary"
            @click="handleSync"
          >
            {{ $t('pages.system.plugin.actions.syncPlugins') }}
          </a-button>
        </Space>
      </template>

      <template #typeHeader>
        <span class="inline-flex items-center gap-1">
          <span>{{ $t('pages.system.plugin.fields.type') }}</span>
          <Tooltip
            :title="$t('pages.system.plugin.columnHelp.type')"
            placement="top"
          >
            <span
              :aria-label="
                getColumnHelpAriaLabel($t('pages.system.plugin.fields.type'))
              "
              class="icon-[ant-design--question-circle-outlined] inline-flex size-4 cursor-help items-center justify-center text-[14px] leading-none text-[var(--ant-color-text-secondary)] transition-colors hover:text-[var(--ant-color-primary)]"
              data-testid="plugin-type-column-help-icon"
              role="img"
              tabindex="0"
            ></span>
          </Tooltip>
        </span>
      </template>

      <template #runtimeStateHeader>
        <span class="inline-flex items-center gap-1">
          <span>{{ $t('pages.system.plugin.fields.runtimeState') }}</span>
          <Tooltip
            :title="$t('pages.system.plugin.columnHelp.runtimeState')"
            placement="top"
          >
            <span
              :aria-label="
                getColumnHelpAriaLabel(
                  $t('pages.system.plugin.fields.runtimeState'),
                )
              "
              class="icon-[ant-design--question-circle-outlined] inline-flex size-4 cursor-help items-center justify-center text-[14px] leading-none text-[var(--ant-color-text-secondary)] transition-colors hover:text-[var(--ant-color-primary)]"
              data-testid="plugin-runtime-state-column-help-icon"
              role="img"
              tabindex="0"
            ></span>
          </Tooltip>
        </span>
      </template>

      <template #hasMockDataHeader>
        <span class="inline-flex items-center gap-1">
          <span>{{ $t('pages.system.plugin.fields.hasMockData') }}</span>
          <Tooltip
            :title="$t('pages.system.plugin.columnHelp.mockData')"
            placement="top"
          >
            <span
              :aria-label="
                getColumnHelpAriaLabel(
                  $t('pages.system.plugin.fields.hasMockData'),
                )
              "
              class="icon-[ant-design--question-circle-outlined] inline-flex size-4 cursor-help items-center justify-center text-[14px] leading-none text-[var(--ant-color-text-secondary)] transition-colors hover:text-[var(--ant-color-primary)]"
              data-testid="plugin-mock-data-column-help-icon"
              role="img"
              tabindex="0"
            ></span>
          </Tooltip>
        </span>
      </template>

      <template #supportsMultiTenantHeader>
        <span class="inline-flex items-center gap-1">
          <span>{{ $t('pages.system.plugin.fields.supportsMultiTenant') }}</span>
          <Tooltip
            :title="$t('pages.system.plugin.columnHelp.supportsMultiTenant')"
            placement="top"
          >
            <span
              :aria-label="
                getColumnHelpAriaLabel(
                  $t('pages.system.plugin.fields.supportsMultiTenant'),
                )
              "
              class="icon-[ant-design--question-circle-outlined] inline-flex size-4 cursor-help items-center justify-center text-[14px] leading-none text-[var(--ant-color-text-secondary)] transition-colors hover:text-[var(--ant-color-primary)]"
              data-testid="plugin-supports-multi-tenant-column-help-icon"
              role="img"
              tabindex="0"
            ></span>
          </Tooltip>
        </span>
      </template>

      <template #tenantProvisioningHeader>
        <span class="inline-flex items-center gap-1">
          <span>{{ $t('pages.system.plugin.fields.tenantProvisioning') }}</span>
          <Tooltip
            :title="$t('pages.system.plugin.columnHelp.tenantProvisioning')"
            placement="top"
          >
            <span
              :aria-label="
                getColumnHelpAriaLabel(
                  $t('pages.system.plugin.fields.tenantProvisioning'),
                )
              "
              class="icon-[ant-design--question-circle-outlined] inline-flex size-4 cursor-help items-center justify-center text-[14px] leading-none text-[var(--ant-color-text-secondary)] transition-colors hover:text-[var(--ant-color-primary)]"
              data-testid="plugin-tenant-provisioning-column-help-icon"
              role="img"
              tabindex="0"
            ></span>
          </Tooltip>
        </span>
      </template>

      <template #type="{ row }">
        <Tag :color="getPluginTypeColor(row.type)">
          {{ getPluginTypeLabel(row.type) }}
        </Tag>
      </template>

      <template #version="{ row }">
        <span
          :data-testid="`plugin-version-${row.id}`"
          :title="formatPluginVersion(row)"
          class="inline-flex max-w-full items-center gap-1 whitespace-nowrap font-mono text-xs tabular-nums"
        >
          <span class="shrink-0">{{ getPluginEffectiveVersion(row) }}</span>
          <span
            v-if="hasPluginVersionDiff(row)"
            class="inline-flex shrink-0 items-center gap-1"
          >
            <span class="text-[var(--ant-color-text-quaternary)]">-&gt;</span>
            <span class="text-[var(--ant-color-primary)]">
              {{ getPluginDiscoveredVersion(row) }}
            </span>
          </span>
        </span>
      </template>

      <template #name="{ row }">
        <!--
          Constrain the name cell to the column width so long titles truncate
          and badge tooltips stay hoverable. min-w-max + shrink-0 name previously
          spilled into the description column and intercepted pointer events.
        -->
        <div
          class="flex w-full min-w-0 max-w-full items-center gap-1.5 overflow-hidden whitespace-nowrap"
          :data-testid="`plugin-name-cell-${row.id}`"
        >
          <span class="min-w-0 truncate" :title="row.name">{{ row.name }}</span>
          <Tooltip
            v-if="isBuiltinPlugin(row)"
            :title="buildBuiltinPluginTooltip(row)"
          >
            <Tag
              class="m-0 shrink-0 whitespace-nowrap leading-5"
              :data-testid="`plugin-builtin-tag-${row.id}`"
              color="purple"
            >
              {{ $t('pages.system.plugin.builtinBadge') }}
            </Tag>
          </Tooltip>
          <Tooltip
            v-if="isAutoEnableManaged(row)"
            :title="buildAutoEnableManagedTooltip(row)"
          >
            <Tag
              class="m-0 shrink-0 whitespace-nowrap leading-5"
              :data-testid="`plugin-auto-enable-tag-${row.id}`"
              color="gold"
            >
              {{ $t('pages.system.plugin.autoEnableBadge') }}
            </Tag>
          </Tooltip>
        </div>
      </template>

      <template #description="{ row, isHidden }">
        <div
          v-if="!isHidden"
          :data-testid="`plugin-description-${row.id}`"
          class="max-w-full truncate"
          :title="row.description || '-'"
        >
          {{ row.description || '-' }}
        </div>
        <span v-else aria-hidden="true" class="sr-only"></span>
      </template>

      <template #runtimeState="{ row }">
        <Tooltip :title="buildRuntimeStateTooltip(row)">
          <Tag
            :color="getRuntimeStateColor(row.runtimeState)"
            :data-testid="`plugin-runtime-state-${row.id}`"
          >
            {{ formatRuntimeState(row.runtimeState) }}
          </Tag>
        </Tooltip>
      </template>

      <template #enabled="{ row }">
        <Tooltip
          v-if="!isBuiltinPlugin(row)"
          :title="
            isAutoEnableManaged(row)
              ? buildAutoEnableManagedRuntimeHint($t('pages.status.disabled'))
              : undefined
          "
        >
          <Switch
            :checked="row.enabled === 1"
            :disabled="
              row.installed !== 1 ||
              !canTogglePluginStatus(row) ||
              isPluginStatusChanging(row)
            "
            :loading="isPluginStatusChanging(row)"
            :checked-children="$t('pages.status.enabled')"
            :un-checked-children="$t('pages.status.disabled')"
            @change="(checked) => handleStatusChange(row, !!checked)"
          />
        </Tooltip>
      </template>

      <template #hasMockData="{ row }">
        <Tag
          :color="hasPluginMockData(row) ? 'green' : 'default'"
          :data-testid="`plugin-mock-data-value-${row.id}`"
        >
          {{
            hasPluginMockData(row)
              ? $t('pages.common.yes')
              : $t('pages.common.no')
          }}
        </Tag>
      </template>

      <template #supportsMultiTenant="{ row }">
        <Tag
          :color="supportsPluginMultiTenant(row) ? 'green' : 'default'"
          :data-testid="`plugin-supports-multi-tenant-${row.id}`"
        >
          {{
            supportsPluginMultiTenant(row)
              ? $t('pages.common.yes')
              : $t('pages.common.no')
          }}
        </Tag>
      </template>

      <template #tenantProvisioning="{ row }">
        <Tooltip
          v-if="!isBuiltinPlugin(row)"
          :title="
            isTenantProvisioningPolicySupported(row)
              ? $t('pages.system.plugin.messages.tenantProvisioningEffective')
              : $t('pages.system.plugin.messages.tenantProvisioningUnsupported')
          "
        >
          <Switch
            :checked="row.autoEnableForNewTenants === true"
            :disabled="
              !isTenantProvisioningPolicySupported(row) || !canEditPluginPolicy()
            "
            size="small"
            :data-testid="`plugin-tenant-provisioning-${row.id}`"
            @change="
              (checked) =>
                handleTenantProvisioningPolicyChange(row, Boolean(checked))
            "
          />
        </Tooltip>
      </template>

      <template #action="{ row }">
        <Space :size="4" :wrap="false">
          <ghost-button
            :data-testid="`plugin-detail-button-${row.id}`"
            @click.stop="handleDetail(row)"
          >
            {{ $t('pages.common.detail') }}
          </ghost-button>
          <Tooltip :title="buildPluginManagementDisabledTooltip(row)">
            <span
              class="inline-flex"
              :data-testid="`plugin-manage-wrapper-${row.id}`"
              @click.stop
            >
              <ghost-button
                :data-testid="`plugin-manage-button-${row.id}`"
                :disabled="!canOpenPluginManagement(row)"
                @click.stop="handleOpenManagement(row)"
              >
                {{ $t('pages.system.plugin.actions.manage') }}
              </ghost-button>
            </span>
          </Tooltip>
          <ghost-button
            v-if="isRuntimeUpgradeAvailable(row) && canInstallPlugin()"
            :data-testid="`plugin-upgrade-button-${row.id}`"
            @click.stop="handleOpenUpgrade(row)"
          >
            {{
              row.runtimeState === 'upgrade_failed'
                ? $t('pages.system.plugin.actions.retryUpgrade')
                : $t('pages.system.plugin.actions.upgrade')
            }}
          </ghost-button>
          <Tooltip
            v-else-if="!isBuiltinPlugin(row) && isRuntimeAbnormal(row)"
            :title="$t('pages.system.plugin.messages.abnormalManualRepair')"
          >
            <span
              class="inline-flex"
              :data-testid="`plugin-abnormal-repair-wrapper-${row.id}`"
              @click.stop
            >
              <ghost-button
                danger
                disabled
                :data-testid="`plugin-abnormal-repair-${row.id}`"
              >
                {{ $t('pages.system.plugin.actions.manualRepair') }}
              </ghost-button>
            </span>
          </Tooltip>
          <ghost-button
            v-else-if="
              !isBuiltinPlugin(row) && row.installed !== 1 && canInstallPlugin()
            "
            @click.stop="handleInstall(row)"
          >
            {{ $t('pages.system.plugin.actions.install') }}
          </ghost-button>
          <Tooltip
            v-else-if="
              !isBuiltinPlugin(row) &&
              canUninstallPlugin() &&
              isAutoEnableManaged(row)
            "
            :title="
              buildAutoEnableManagedRuntimeHint(
                $t('pages.system.plugin.actions.uninstall'),
              )
            "
          >
            <ghost-button danger @click.stop="handleOpenUninstall(row)">
              {{ $t('pages.system.plugin.actions.uninstall') }}
            </ghost-button>
          </Tooltip>
          <ghost-button
            v-else-if="!isBuiltinPlugin(row) && canUninstallPlugin()"
            danger
            @click.stop="handleOpenUninstall(row)"
          >
            {{ $t('pages.system.plugin.actions.uninstall') }}
          </ghost-button>
        </Space>
      </template>
    </Grid>
    <DetailModal />
    <DynamicUploadModal @reload="handleDynamicUploadReload" />
    <HostServiceAuthModal @reload="handleHostServiceAuthReload" />
    <UpgradeModal @reload="handleUpgradeReload" />
    <UninstallModal
      @lifecycle-precondition="handleLifecyclePrecondition"
      @reload="handleUninstallReload"
    />
    <LifecyclePreconditionModal @force="handleLifecyclePreconditionForce" />
  </Page>
</template>
