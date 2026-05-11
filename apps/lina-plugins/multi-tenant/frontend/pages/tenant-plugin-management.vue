<script lang="ts">
export const pluginPageMeta = {
  capabilities: ['tenant.management'],
  routePath: '/tenant/plugins',
  title: 'Tenant Plugins',
};
</script>

<script setup lang="ts">
import { watch } from 'vue';

import { Page } from '@vben/common-ui';

import { message, Space, Switch, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { useTenantStore } from '#/store';
import {
  tenantPluginDisable,
  tenantPluginEnable,
  tenantPluginList,
} from './tenant-plugin-client';

const tenantStore = useTenantStore();

const [Grid, gridApi] = useVbenVxeGrid({
  gridOptions: {
    columns: [
      { field: 'id', minWidth: 180, title: $t('pages.system.plugin.fields.id') },
      {
        field: 'name',
        minWidth: 220,
        title: $t('pages.system.plugin.fields.name'),
      },
      {
        field: 'installMode',
        slots: { default: 'installMode' },
        title: $t('pages.multiTenant.plugin.installMode'),
        width: 160,
      },
      {
        field: 'tenantEnabled',
        slots: { default: 'enabled' },
        title: $t('pages.common.status'),
        width: 140,
      },
      {
        field: 'description',
        minWidth: 260,
        title: $t('pages.fields.description'),
      },
    ],
    emptyRender: {
      name: 'Empty',
      props: { description: $t('pages.multiTenant.empty.plugins') },
    },
    height: 'auto',
    pagerConfig: {},
    proxyConfig: {
      ajax: { query: tenantPluginList },
    },
    rowConfig: { keyField: 'id' },
    id: 'tenant-plugin-index',
  },
});

async function togglePlugin(row: any, checked: boolean) {
  if (checked) {
    await tenantPluginEnable(row.id);
  } else {
    await tenantPluginDisable(row.id);
  }
  message.success($t('pages.common.updateSuccess'));
  await gridApi.query();
}

watch(
  () => tenantStore.currentTenant?.id,
  async (tenantId, previousTenantId) => {
    if (!tenantId || tenantId === previousTenantId) {
      return;
    }
    await gridApi.query();
  },
);
</script>

<template>
  <Page :auto-content-height="true" data-testid="tenant-plugins-page">
    <Grid :table-title="$t('pages.multiTenant.plugin.tableTitle')">
      <template #installMode="{ row }">
        <Tag color="blue">
          {{
            $t(
              `pages.multiTenant.plugin.installModes.${row.installMode || 'tenant_scoped'}`,
            )
          }}
        </Tag>
      </template>
      <template #enabled="{ row }">
        <Space>
          <Switch
            :checked="row.tenantEnabled === 1 || row.enabled === 1"
            :data-testid="`tenant-plugin-toggle-${row.id}`"
            size="small"
            @change="(checked) => togglePlugin(row, Boolean(checked))"
          />
        </Space>
      </template>
    </Grid>
  </Page>
</template>
