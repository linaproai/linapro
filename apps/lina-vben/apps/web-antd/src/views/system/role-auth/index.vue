<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { RoleUser } from '#/api/system/role';

import { onMounted, ref } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { Page } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { Modal, Popconfirm, Space, Tag, message } from 'ant-design-vue';

import { useVbenVxeGrid, vxeCheckboxChecked } from '#/adapter/vxe-table';
import { roleInfo, roleUnassignUser, roleUnassignUsers, roleUsers } from '#/api/system/role';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';

const route = useRoute();
const router = useRouter();
const roleId = Number(route.params.id);

const roleName = ref('');
const roleKey = ref('');

// 加载角色信息
onMounted(async () => {
  const role = await roleInfo(roleId);
  roleName.value = role.name;
  roleKey.value = role.key;
});

const formOptions: VbenFormProps = {
  commonConfig: {
    labelWidth: 80,
    componentProps: {
      allowClear: true,
    },
  },
  schema: [
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.username'),
      },
      fieldName: 'username',
      label: $t('pages.system.user.labels.userAccount'),
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: $t('pages.system.user.placeholders.phone'),
      },
      fieldName: 'phone',
      label: $t('pages.fields.phone'),
    },
    {
      component: 'Select',
      componentProps: {
        placeholder: $t('pages.system.role.placeholders.status'),
        options: [],
      },
      fieldName: 'status',
      label: $t('pages.common.status'),
    },
  ],
  wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
};

const gridOptions: VxeGridProps = {
  checkboxConfig: {
    highlight: true,
    reserve: true,
  },
  columns: [
    { type: 'checkbox', width: 60 },
    {
      title: $t('pages.system.user.labels.userAccount'),
      field: 'username',
      minWidth: 100,
    },
    {
      title: $t('pages.system.user.labels.userNickname'),
      field: 'nickname',
      minWidth: 100,
    },
    {
      title: $t('pages.fields.email'),
      field: 'email',
      minWidth: 120,
    },
    {
      title: $t('pages.fields.phone'),
      field: 'phone',
      minWidth: 110,
    },
    {
      title: $t('pages.common.status'),
      field: 'status',
      width: 80,
      slots: { default: 'status' },
    },
    {
      title: $t('pages.common.createdAt'),
      field: 'createdAt',
      width: 170,
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: $t('pages.common.actions'),
      resizable: false,
      width: 'auto',
    },
  ],
  height: 'auto',
  keepSource: true,
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues = {}) => {
        return await roleUsers({
          id: roleId,
          page: page.currentPage,
          size: page.pageSize,
          ...formValues,
        });
      },
    },
  },
  rowConfig: {
    keyField: 'id',
  },
  id: 'system-role-user-index',
};

const [BasicTable, tableApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
});

// 加载字典数据
const dictStore = useDictStore();
const statusLabel = ref({
  checked: $t('pages.status.enabled'),
  unchecked: $t('pages.status.disabled'),
});

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
  const checked = statusOptions.find((d) => d.value === '1');
  const unchecked = statusOptions.find((d) => d.value === '0');
  statusLabel.value = {
    checked: checked?.label || $t('pages.status.enabled'),
    unchecked: unchecked?.label || $t('pages.status.disabled'),
  };
  tableApi.formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: statusOptions.map((d) => ({
          label: d.label,
          value: Number(d.value),
        })),
      },
    },
  ]);
});

async function handleUnassignUser(row: RoleUser) {
  await roleUnassignUser(roleId, row.id);
  message.success($t('pages.system.role.roleAuth.removeSuccess'));
  await tableApi.query();
}

function handleMultiUnassignUsers() {
  const rows = (tableApi.grid?.getCheckboxRecords?.() ?? []) as RoleUser[];
  if (rows.length === 0) return;
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.role.roleAuth.removeSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await roleUnassignUsers(roleId, ids);
      message.success($t('pages.system.role.roleAuth.removeSelectedSuccess'));
      await tableApi.query();
      tableApi.grid?.clearCheckboxRow?.();
    },
  });
}

function handleBack() {
  router.push('/system/role');
}
</script>

<template>
  <Page :auto-content-height="true">
    <BasicTable :table-title="$t('pages.system.role.roleAuth.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!vxeCheckboxChecked(tableApi)"
            danger
            type="primary"
            @click="handleMultiUnassignUsers"
          >
            {{ $t('pages.system.role.roleAuth.removeAssignment') }}
          </a-button>
          <a-button @click="handleBack">
            {{ $t('pages.system.role.roleAuth.back') }}
          </a-button>
        </Space>
      </template>
      <template #status="{ row }">
        <Tag :color="row.status === 1 ? 'success' : 'error'">
          {{ row.status === 1 ? statusLabel.checked : statusLabel.unchecked }}
        </Tag>
      </template>
      <template #action="{ row }">
        <Popconfirm
          :get-popup-container="getPopupContainer"
          placement="left"
          :title="$t('pages.system.role.roleAuth.removeConfirm', {
            username: row.username,
            nickname: row.nickname,
          })"
          @confirm="handleUnassignUser(row)"
        >
          <ghost-button danger @click.stop="">
            {{ $t('pages.system.role.roleAuth.removeAssignment') }}
          </ghost-button>
        </Popconfirm>
      </template>
    </BasicTable>
  </Page>
</template>
