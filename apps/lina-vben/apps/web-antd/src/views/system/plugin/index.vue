<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { useAccess } from '@vben/access';
import { Page } from '@vben/common-ui';
import { useVbenModal } from '@vben/common-ui';

import { message, Popconfirm, Space, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  pluginDisable,
  pluginEnable,
  pluginInstall,
  pluginList,
  pluginSync,
  pluginUninstall,
} from '#/api/system/plugin';
import { notifyPluginRegistryChanged } from '#/plugins/slot-registry';
import PluginHostServiceAuthModal from './plugin-host-service-auth-modal.vue';
import PluginDynamicUploadModal from './plugin-dynamic-upload-modal.vue';

const [DynamicUploadModal, dynamicUploadModalApi] = useVbenModal({
  connectedComponent: PluginDynamicUploadModal,
});

const [HostServiceAuthModal, hostServiceAuthModalApi] = useVbenModal({
  connectedComponent: PluginHostServiceAuthModal,
});

const typeColorMap: Record<string, string> = {
  dynamic: 'green',
  source: 'blue',
};

const pluginAccessCodes = {
  disable: 'plugin:disable',
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
        label: '插件标识',
      },
      {
        component: 'Input',
        fieldName: 'name',
        label: '插件名称',
      },
      {
        component: 'Select',
        fieldName: 'type',
        label: '插件类型',
        componentProps: {
          options: [
            { label: '源码插件', value: 'source' },
            { label: '动态插件', value: 'dynamic' },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'installed',
        label: '接入态',
        componentProps: {
          options: [
            { label: '已接入', value: 1 },
            { label: '未安装', value: 0 },
          ],
        },
      },
      {
        component: 'Select',
        fieldName: 'status',
        label: '状态',
        componentProps: {
          options: [
            { label: '启用', value: 1 },
            { label: '禁用', value: 0 },
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
      { field: 'id', minWidth: 160, title: '插件标识' },
      { field: 'name', minWidth: 160, title: '插件名称' },
      {
        field: 'type',
        slots: { default: 'type' },
        title: '插件类型',
        width: 120,
      },
      { field: 'version', title: '版本', width: 120 },
      {
        className: 'plugin-description-column',
        field: 'description',
        minWidth: 260,
        showOverflow: false,
        slots: { default: 'description' },
        title: '描述',
      },
      {
        field: 'enabled',
        slots: { default: 'enabled' },
        title: '状态',
        width: 130,
      },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: '操作',
        width: 180,
      },
      { field: 'installedAt', title: '安装时间', width: 180 },
      { field: 'updatedAt', title: '更新时间', width: 180 },
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
  return type === 'source' ? '源码插件' : '动态插件';
}

function getPluginTypeColor(type: string) {
  return typeColorMap[type === 'source' ? 'source' : 'dynamic'] || 'default';
}

function isSourcePlugin(row: SystemPlugin) {
  return row.type === 'source';
}

function canInstallPlugin() {
  return hasAccessByCodes([pluginAccessCodes.install]);
}

function canSyncPlugins() {
  return hasAccessByCodes([pluginAccessCodes.install]);
}

function canUninstallPlugin() {
  return hasAccessByCodes([pluginAccessCodes.uninstall]);
}

function canTogglePluginStatus(row: SystemPlugin) {
  return row.enabled === 1
    ? hasAccessByCodes([pluginAccessCodes.disable])
    : hasAccessByCodes([pluginAccessCodes.enable]);
}

async function handleStatusChange(row: SystemPlugin, checked: boolean) {
  if (row.installed !== 1) {
    message.warning('请先完成插件接入');
    return;
  }
  if (!canTogglePluginStatus(row)) {
    message.warning('当前账号缺少插件状态管理权限');
    return;
  }
  if (checked && row.authorizationRequired === 1) {
    hostServiceAuthModalApi.setData({ mode: 'enable', row });
    hostServiceAuthModalApi.open();
    return;
  }
  await (checked ? pluginEnable : pluginDisable)(row.id);
  row.enabled = checked ? 1 : 0;
  await notifyPluginRegistryChanged();
  message.success(checked ? '插件已启用' : '插件已禁用');
}

async function handleInstall(row: SystemPlugin) {
  if (!canInstallPlugin()) {
    message.warning('当前账号缺少插件安装权限');
    return;
  }
  if (row.authorizationRequired === 1) {
    hostServiceAuthModalApi.setData({ mode: 'install', row });
    hostServiceAuthModalApi.open();
    return;
  }
  await pluginInstall(row.id);
  row.installed = 1;
  row.enabled = 0;
  await notifyPluginRegistryChanged();
  message.success('动态插件已安装');
  await gridApi.query();
}

async function handleUninstall(row: SystemPlugin) {
  if (!canUninstallPlugin()) {
    message.warning('当前账号缺少插件卸载权限');
    return;
  }
  await pluginUninstall(row.id);
  row.installed = 0;
  row.enabled = 0;
  await notifyPluginRegistryChanged();
  message.success('动态插件已卸载');
  await gridApi.query();
}

async function handleSync() {
  if (!canSyncPlugins()) {
    message.warning('当前账号缺少插件安装权限');
    return;
  }
  const res = await pluginSync();
  await notifyPluginRegistryChanged();
  const total = typeof res?.total === 'number' ? res.total : 0;
  message.success(`已同步 ${total} 个源码插件`);
  await gridApi.query();
}

function handleOpenDynamicUpload() {
  if (!canInstallPlugin()) {
    message.warning('当前账号缺少插件安装权限');
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
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="插件列表">
      <template #toolbar-tools>
        <Space>
          <a-button
            data-testid="plugin-dynamic-upload-trigger"
            type="primary"
            v-access:code="pluginAccessCodes.install"
            @click="handleOpenDynamicUpload"
          >
            上传插件
          </a-button>
          <a-button
            v-access:code="pluginAccessCodes.install"
            type="primary"
            @click="handleSync"
          >
            同步插件
          </a-button>
        </Space>
      </template>

      <template #type="{ row }">
        <Tag :color="getPluginTypeColor(row.type)">
          {{ getPluginTypeLabel(row.type) }}
        </Tag>
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
        <Switch
          :checked="row.enabled === 1"
          :disabled="row.installed !== 1 || !canTogglePluginStatus(row)"
          checked-children="启用"
          un-checked-children="禁用"
          @change="(checked) => handleStatusChange(row, !!checked)"
        />
      </template>

      <template #action="{ row }">
        <Space v-if="isSourcePlugin(row)">
          <ghost-button
            v-if="canUninstallPlugin()"
            :data-testid="`plugin-source-uninstall-disabled-${row.id}`"
            danger
            disabled
            title="源码插件不支持页面动态卸载，如需移除请在源码中取消注册后重新构建宿主。"
            @click.stop=""
          >
            卸载
          </ghost-button>
        </Space>
        <Space v-else>
          <Popconfirm
            v-if="row.installed !== 1 && canInstallPlugin()"
            title="确认安装该插件？"
            @confirm="handleInstall(row)"
          >
            <ghost-button @click.stop="">安装</ghost-button>
          </Popconfirm>
          <Popconfirm
            v-else-if="canUninstallPlugin()"
            title="确认卸载该插件？"
            @confirm="handleUninstall(row)"
          >
            <ghost-button danger @click.stop="">卸载</ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
    <DynamicUploadModal @reload="handleDynamicUploadReload" />
    <HostServiceAuthModal @reload="handleHostServiceAuthReload" />
  </Page>
</template>
