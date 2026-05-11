<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { h } from 'vue';

import { useAccess } from '@vben/access';
import { Page, useVbenModal } from '@vben/common-ui';

import { message, Modal, Space, Switch, Tag, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  pluginDisable,
  pluginEnable,
  pluginList,
  pluginSync,
  pluginUpdateTenantProvisioningPolicy,
} from '#/api/system/plugin';
import { $t } from '#/locales';
import { notifyPluginRegistryChanged } from '#/plugins/slot-registry';

import PluginDetailModal from './plugin-detail-modal.vue';
import PluginDynamicUploadModal from './plugin-dynamic-upload-modal.vue';
import PluginHostServiceAuthModal from './plugin-host-service-auth-modal.vue';
import PluginUninstallModal from './plugin-uninstall-modal.vue';
import LifecycleGuardDialog from '#/views/platform/plugins/lifecycle-guard-dialog.vue';

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

const [LifecycleGuardModal, lifecycleGuardModalApi] = useVbenModal({
  connectedComponent: LifecycleGuardDialog,
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
        field: 'id',
        minWidth: 160,
        title: $t('pages.system.plugin.fields.id'),
      },
      {
        className: 'plugin-name-column',
        field: 'name',
        minWidth: 280,
        slots: { default: 'name' },
        title: $t('pages.system.plugin.fields.name'),
      },
      {
        field: 'type',
        slots: { default: 'type', header: 'typeHeader' },
        title: $t('pages.system.plugin.fields.type'),
        width: 120,
      },
      {
        field: 'version',
        title: $t('pages.system.plugin.fields.version'),
        width: 120,
      },
      {
        className: 'plugin-description-column',
        field: 'description',
        minWidth: 260,
        showOverflow: false,
        slots: { default: 'description' },
        title: $t('pages.fields.description'),
      },
      {
        field: 'enabled',
        slots: { default: 'enabled' },
        title: $t('pages.common.status'),
        width: 130,
      },
      {
        field: 'hasMockData',
        slots: { default: 'hasMockData', header: 'hasMockDataHeader' },
        title: $t('pages.system.plugin.fields.hasMockData'),
        width: 120,
      },
      {
        field: 'supportsMultiTenant',
        slots: {
          default: 'supportsMultiTenant',
          header: 'supportsMultiTenantHeader',
        },
        title: $t('pages.system.plugin.fields.supportsMultiTenant'),
        width: 140,
      },
      {
        field: 'autoEnableForNewTenants',
        slots: {
          default: 'tenantProvisioning',
          header: 'tenantProvisioningHeader',
        },
        title: $t('pages.system.plugin.fields.tenantProvisioning'),
        width: 160,
      },
      {
        field: 'installedAt',
        title: $t('pages.system.plugin.fields.installedAt'),
        width: 180,
      },
      {
        field: 'updatedAt',
        title: $t('pages.common.updatedAt'),
        width: 180,
      },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: $t('pages.common.actions'),
        width: 240,
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
    id: 'system-plugin-index',
  },
});

function getPluginTypeLabel(type: string) {
  return type === 'source'
    ? $t('pages.system.plugin.type.source')
    : $t('pages.system.plugin.type.dynamic');
}

function getPluginTypeColor(type: string) {
  return typeColorMap[type === 'source' ? 'source' : 'dynamic'] || 'default';
}

function isAutoEnableManaged(row: SystemPlugin) {
  return row.autoEnableManaged === 1;
}

function hasPluginMockData(row: SystemPlugin) {
  return row.hasMockData === 1;
}

function supportsPluginMultiTenant(row: SystemPlugin) {
  return row.supportsMultiTenant === true;
}

function isTenantProvisioningPolicySupported(row: SystemPlugin) {
  return (
    supportsPluginMultiTenant(row) &&
    row.scopeNature === 'tenant_aware' &&
    row.installMode === 'tenant_scoped'
  );
}

function buildAutoEnableManagedTooltip(row: SystemPlugin) {
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

function canTogglePluginStatus(row: SystemPlugin) {
  return row.enabled === 1
    ? hasAccessByCodes([pluginAccessCodes.disable])
    : hasAccessByCodes([pluginAccessCodes.enable]);
}

function handleDetail(row: SystemPlugin) {
  detailModalApi.setData({ row });
  detailModalApi.open();
}

async function handleStatusChange(row: SystemPlugin, checked: boolean) {
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
    hostServiceAuthModalApi.setData({ mode: 'enable', row });
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
  await (checked ? pluginEnable : pluginDisable)(row.id);
  row.enabled = checked ? 1 : 0;
  await notifyPluginRegistryChanged();
  message.success(
    checked
      ? $t('pages.system.plugin.messages.enabled')
      : $t('pages.system.plugin.messages.disabled'),
  );
}

async function handleTenantProvisioningPolicyChange(
  row: SystemPlugin,
  checked: boolean,
) {
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

async function handleInstall(row: SystemPlugin) {
  if (!canInstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noInstallPermission'));
    return;
  }
  hostServiceAuthModalApi.setData({
    allowInstallAndEnable: canInstallAndEnablePlugin(),
    mode: 'install',
    row,
  });
  hostServiceAuthModalApi.open();
}

function handleOpenUninstall(row: SystemPlugin) {
  if (!canUninstallPlugin()) {
    message.warning($t('pages.system.plugin.messages.noUninstallPermission'));
    return;
  }
  uninstallModalApi.setData({ row });
  uninstallModalApi.open();
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

async function handleUninstallReload() {
  await notifyPluginRegistryChanged();
  await gridApi.query();
}

function handleLifecycleGuard(payload: {
  force: () => Promise<void>;
  pluginId: string;
  reasons: string[];
}) {
  lifecycleGuardModalApi.setData(payload);
  lifecycleGuardModalApi.open();
}

async function handleLifecycleGuardForce(payload: { pluginId: string }) {
  const data = lifecycleGuardModalApi.getData<{
    force?: () => Promise<void>;
    pluginId?: string;
  }>();
  if (!data.force || data.pluginId !== payload.pluginId) {
    return;
  }
  lifecycleGuardModalApi.lock(true);
  try {
    await data.force();
    lifecycleGuardModalApi.close();
  } catch {
    // The force callback already surfaces the backend error locally.
  } finally {
    lifecycleGuardModalApi.lock(false);
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

      <template #name="{ row }">
        <div
          class="inline-flex min-w-max max-w-full items-center gap-1.5 whitespace-nowrap"
          :data-testid="`plugin-name-cell-${row.id}`"
        >
          <span class="shrink-0 whitespace-nowrap">{{ row.name }}</span>
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

      <template #enabled="{ row }">
        <Tooltip
          :title="
            isAutoEnableManaged(row)
              ? buildAutoEnableManagedRuntimeHint($t('pages.status.disabled'))
              : undefined
          "
        >
          <Switch
            :checked="row.enabled === 1"
            :disabled="row.installed !== 1 || !canTogglePluginStatus(row)"
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
        <Space>
          <ghost-button
            :data-testid="`plugin-detail-button-${row.id}`"
            @click.stop="handleDetail(row)"
          >
            {{ $t('pages.common.detail') }}
          </ghost-button>
          <ghost-button
            v-if="row.installed !== 1 && canInstallPlugin()"
            @click.stop="handleInstall(row)"
          >
            {{ $t('pages.system.plugin.actions.install') }}
          </ghost-button>
          <Tooltip
            v-else-if="canUninstallPlugin() && isAutoEnableManaged(row)"
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
            v-else-if="canUninstallPlugin()"
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
    <UninstallModal
      @lifecycle-guard="handleLifecycleGuard"
      @reload="handleUninstallReload"
    />
    <LifecycleGuardModal @force="handleLifecycleGuardForce" />
  </Page>
</template>
