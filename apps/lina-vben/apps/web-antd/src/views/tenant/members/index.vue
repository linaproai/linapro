<script setup lang="ts">
import { watch } from 'vue';

import { Page } from '@vben/common-ui';

import { message, Popconfirm, Space, Tag } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { tenantMemberList, tenantMemberRemove } from '#/api/tenant';
import { $t } from '#/locales';
import { useTenantStore } from '#/store';

const tenantStore = useTenantStore();

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: 'Input',
        fieldName: 'keyword',
        label: $t('pages.multiTenant.fields.keyword'),
      },
      {
        component: 'Select',
        componentProps: {
          options: ['active', 'suspended', 'removed'].map((value) => ({
            label: $t(`pages.multiTenant.member.status.${value}`),
            value,
          })),
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
      { field: 'username', minWidth: 160, title: $t('pages.fields.username') },
      { field: 'realName', minWidth: 160, title: $t('pages.fields.nickname') },
      { field: 'email', minWidth: 220, title: $t('pages.fields.email') },
      {
        field: 'status',
        slots: { default: 'status' },
        title: $t('pages.common.status'),
        width: 140,
      },
      { field: 'joinedAt', title: $t('pages.multiTenant.fields.joinedAt'), width: 180 },
      {
        field: 'action',
        fixed: 'right',
        slots: { default: 'action' },
        title: $t('pages.common.actions'),
        width: 120,
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
          await tenantMemberList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          }),
      },
    },
    rowConfig: { keyField: 'id' },
    id: 'tenant-member-index',
  },
});

async function removeMember(id: number) {
  await tenantMemberRemove(id);
  message.success($t('pages.common.deleteSuccess'));
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
  <Page :auto-content-height="true" data-testid="tenant-members-page">
    <Grid :table-title="$t('pages.multiTenant.member.tableTitle')">
      <template #status="{ row }">
        <Tag :color="row.status === 'active' ? 'green' : 'default'">
          {{ $t(`pages.multiTenant.member.status.${row.status || 'active'}`) }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Space>
          <Popconfirm
            :title="$t('pages.multiTenant.messages.removeMemberConfirm')"
            @confirm="removeMember(row.id)"
          >
            <ghost-button danger :data-testid="`tenant-member-remove-${row.id}`">
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
  </Page>
</template>
