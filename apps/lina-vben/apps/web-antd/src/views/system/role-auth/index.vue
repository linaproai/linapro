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
        placeholder: '请输入用户账号',
      },
      fieldName: 'username',
      label: '用户账号',
    },
    {
      component: 'Input',
      componentProps: {
        placeholder: '请输入手机号码',
      },
      fieldName: 'phone',
      label: '手机号码',
    },
    {
      component: 'Select',
      componentProps: {
        placeholder: '请选择状态',
        options: [],
      },
      fieldName: 'status',
      label: '状态',
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
      title: '用户账号',
      field: 'username',
      minWidth: 100,
    },
    {
      title: '用户昵称',
      field: 'nickname',
      minWidth: 100,
    },
    {
      title: '邮箱',
      field: 'email',
      minWidth: 120,
    },
    {
      title: '手机号码',
      field: 'phone',
      minWidth: 110,
    },
    {
      title: '状态',
      field: 'status',
      width: 80,
      slots: { default: 'status' },
    },
    {
      title: '创建时间',
      field: 'createdAt',
      width: 170,
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
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
const statusLabel = ref({ checked: '正常', unchecked: '停用' });

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
  const checked = statusOptions.find((d) => d.value === '1');
  const unchecked = statusOptions.find((d) => d.value === '0');
  statusLabel.value = {
    checked: checked?.label || '正常',
    unchecked: unchecked?.label || '停用',
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
  message.success('取消授权成功');
  await tableApi.query();
}

function handleMultiUnassignUsers() {
  const rows = (tableApi.grid?.getCheckboxRecords?.() ?? []) as RoleUser[];
  if (rows.length === 0) return;
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认取消选中的${ids.length}条授权记录吗？`,
    onOk: async () => {
      await roleUnassignUsers(roleId, ids);
      message.success('批量取消授权成功');
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
    <BasicTable :table-title="`已分配的用户列表`">
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!vxeCheckboxChecked(tableApi)"
            danger
            type="primary"
            @click="handleMultiUnassignUsers"
          >
            取消授权
          </a-button>
          <a-button @click="handleBack">返 回</a-button>
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
          :title="`确认取消授权用户[${row.username} - ${row.nickname}]?`"
          @confirm="handleUnassignUser(row)"
        >
          <ghost-button danger @click.stop="">
            取消授权
          </ghost-button>
        </Popconfirm>
      </template>
    </BasicTable>
  </Page>
</template>
