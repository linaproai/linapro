<script lang="ts">
export const pluginPageMeta = {
  capabilities: ['tenant.management'],
  routePath: 'platform/tenants',
  title: 'Tenant Management',
};
</script>

<script setup lang="ts">
import type { SwitchProps } from 'ant-design-vue';

import type { PlatformTenant, TenantStatus } from './tenant-client';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  Modal,
  Popconfirm,
  Space,
  Switch,
  Tag,
  Tooltip,
  message,
} from 'ant-design-vue';
import { useRouter } from 'vue-router';

import { useVbenVxeGrid, vxeCheckboxChecked } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { useTenantStore } from '#/store';

import TenantModal from './components/tenant-modal.vue';
import {
  platformTenantChangeStatus,
  platformTenantDelete,
  platformTenantList,
} from './tenant-client';

const tenantStore = useTenantStore();
const router = useRouter();

const [TenantModalRef, tenantModalApi] = useVbenModal({
  connectedComponent: TenantModal,
});

const statusColorMap: Record<string, string> = {
  active: 'green',
  deleted: 'red',
  suspended: 'orange',
};

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Input',
        componentProps: {
          'data-testid': 'tenant-search-code',
        },
        fieldName: 'code',
        label: $t('pages.multiTenant.fields.code'),
      },
      {
        component: 'Input',
        componentProps: {
          'data-testid': 'tenant-search-name',
        },
        fieldName: 'name',
        label: $t('pages.multiTenant.fields.name'),
      },
      {
        component: 'Select',
        componentProps: {
          'data-testid': 'tenant-search-status',
          options: ['active', 'suspended'].map((value) => ({
            label: $t(`pages.multiTenant.status.${value}`),
            value,
          })),
        },
        fieldName: 'status',
        label: $t('pages.common.status'),
      },
    ],
    commonConfig: { componentProps: { allowClear: true }, labelWidth: 90 },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: [
      { type: 'checkbox', width: 56 },
      {
        field: 'code',
        minWidth: 160,
        title: $t('pages.multiTenant.fields.code'),
      },
      {
        field: 'name',
        minWidth: 220,
        title: $t('pages.multiTenant.fields.name'),
      },
      {
        field: 'status',
        slots: { default: 'status' },
        title: $t('pages.common.status'),
        width: 150,
      },
      { field: 'createdAt', title: $t('pages.common.createdAt'), width: 180 },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: $t('pages.common.actions'),
        width: 380,
      },
    ],
    emptyRender: {
      name: 'Empty',
      props: { description: $t('pages.multiTenant.empty.tenants') },
    },
    height: 'auto',
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) =>
          await platformTenantList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          }),
      },
    },
    rowConfig: { keyField: 'id' },
    id: 'platform-tenant-index',
  },
});

function openCreate() {
  tenantModalApi.setData({});
  tenantModalApi.open();
}

function openEdit(row: PlatformTenant) {
  tenantModalApi.setData(row);
  tenantModalApi.open();
}

async function changeStatus(row: PlatformTenant, status: TenantStatus) {
  await platformTenantChangeStatus(row.id, status);
  message.success($t('pages.multiTenant.messages.statusUpdated'));
  await gridApi.query();
  tenantStore.setTenantContext({
    tenants: tenantStore.tenants.map((tenant) =>
      tenant.id === row.id ? { ...tenant, name: row.name } : tenant,
    ),
  });
}

async function handleSwitchStatus(
  checked: SwitchProps['checked'],
  row: PlatformTenant,
) {
  const nextStatus = checked ? 'active' : 'suspended';
  await changeStatus(row, nextStatus);
}

async function handleDelete(row: PlatformTenant) {
  await platformTenantDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = (gridApi.grid?.getCheckboxRecords?.() ?? []) as PlatformTenant[];
  if (rows.length === 0) {
    return;
  }
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    content: $t('pages.multiTenant.messages.deleteSelectedConfirm', {
      count: rows.length,
    }),
    okType: 'danger',
    onOk: async () => {
      await Promise.all(rows.map((row) => platformTenantDelete(row.id)));
      message.success($t('pages.common.deleteSuccess'));
      await gridApi.query();
    },
  });
}

async function impersonate(row: PlatformTenant) {
  await tenantStore.switchTenant(row.id, router);
}
</script>

<template>
  <Page :auto-content-height="true" data-testid="platform-tenants-page">
    <Grid :table-title="$t('pages.multiTenant.tenant.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            danger
            type="primary"
            :disabled="!vxeCheckboxChecked(gridApi)"
            data-testid="tenant-batch-delete"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button
            type="primary"
            data-testid="tenant-create"
            @click="openCreate"
          >
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>
      <template #status="{ row }">
        <Space>
          <Switch
            :checked="row.status === 'active'"
            :data-testid="
              row.status === 'active'
                ? `tenant-suspend-${row.id}`
                : `tenant-resume-${row.id}`
            "
            :disabled="row.status === 'deleted'"
            :checked-children="$t('pages.multiTenant.status.active')"
            :un-checked-children="$t('pages.multiTenant.status.suspended')"
            @change="(checked) => handleSwitchStatus(checked, row)"
          />
          <Tag
            v-if="row.status === 'deleted'"
            :color="statusColorMap[row.status] || 'default'"
          >
            {{ $t(`pages.multiTenant.status.${row.status || 'unknown'}`) }}
          </Tag>
        </Space>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button
            :data-testid="`tenant-edit-${row.id}`"
            @click="openEdit(row)"
          >
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <Tooltip
            :title="$t('pages.multiTenant.tenant.tooltips.impersonate')"
          >
            <span>
              <ghost-button
                :data-testid="`tenant-impersonate-${row.id}`"
                :disabled="row.status !== 'active'"
                @click="impersonate(row)"
              >
                {{ $t('pages.multiTenant.tenant.actions.impersonate') }}
              </ghost-button>
            </span>
          </Tooltip>
          <Popconfirm
            placement="left"
            :title="
              $t('pages.multiTenant.messages.deleteTenantConfirm', {
                name: row.name,
              })
            "
            @confirm="handleDelete(row)"
          >
            <ghost-button
              danger
              :data-testid="`tenant-delete-${row.id}`"
              @click.stop=""
            >
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
    <TenantModalRef @success="gridApi.query()" />
  </Page>
</template>
