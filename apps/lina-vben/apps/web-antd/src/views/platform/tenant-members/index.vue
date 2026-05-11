<script setup lang="ts">
import { Page } from '@vben/common-ui';

import { Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { platformTenantMemberList } from '#/api/platform/tenant';
import { $t } from '#/locales';

const statusOptions = [
  { label: $t('pages.multiTenant.member.status.active'), value: 1 },
  { label: $t('pages.multiTenant.member.status.suspended'), value: 0 },
];

const [Grid] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'InputNumber',
        fieldName: 'tenantId',
        label: $t('pages.multiTenant.fields.tenantId'),
      },
      {
        component: 'InputNumber',
        fieldName: 'userId',
        label: $t('pages.multiTenant.fields.userId'),
      },
      {
        component: 'Select',
        componentProps: {
          options: statusOptions,
        },
        fieldName: 'status',
        label: $t('pages.common.status'),
      },
    ],
    commonConfig: { componentProps: { allowClear: true }, labelWidth: 90 },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
  },
  gridOptions: {
    columns: [
      {
        field: 'tenantId',
        title: $t('pages.multiTenant.fields.tenantId'),
        width: 120,
      },
      {
        field: 'userId',
        title: $t('pages.multiTenant.fields.userId'),
        width: 120,
      },
      { field: 'username', minWidth: 160, title: $t('pages.fields.username') },
      { field: 'nickname', minWidth: 160, title: $t('pages.fields.nickname') },
      {
        field: 'status',
        slots: { default: 'status' },
        title: $t('pages.common.status'),
        width: 140,
      },
    ],
    emptyRender: {
      name: 'Empty',
      props: { description: $t('pages.multiTenant.empty.members') },
    },
    height: 'auto',
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) =>
          await platformTenantMemberList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          }),
      },
    },
    rowConfig: { keyField: 'id' },
    id: 'platform-tenant-member-index',
  },
});
</script>

<template>
  <Page :auto-content-height="true" data-testid="platform-tenant-members-page">
    <Grid :table-title="$t('pages.multiTenant.platformMembers.tableTitle')">
      <template #status="{ row }">
        <Tag :color="row.status === 1 ? 'green' : 'orange'">
          {{
            row.status === 1
              ? $t('pages.multiTenant.member.status.active')
              : $t('pages.multiTenant.member.status.suspended')
          }}
        </Tag>
      </template>
    </Grid>
  </Page>
</template>
