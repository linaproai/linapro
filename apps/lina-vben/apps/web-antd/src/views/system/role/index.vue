<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Role } from '#/api/system/role';

import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { Modal, Popconfirm, Space, Switch, message } from 'ant-design-vue';

import { useVbenVxeGrid, vxeCheckboxChecked } from '#/adapter/vxe-table';
import { roleList, roleRemove, roleStatusChange } from '#/api/system/role';
import { useDictStore } from '#/store/dict';

import { columns, querySchema } from './data';
import RoleDrawer from './role-drawer.vue';

const router = useRouter();

const formOptions: VbenFormProps = {
  commonConfig: {
    labelWidth: 80,
    componentProps: {
      allowClear: true,
    },
  },
  schema: querySchema,
  wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
};

const gridOptions: VxeGridProps = {
  checkboxConfig: {
    highlight: true,
    reserve: true,
    checkMethod: ({ row }) => row.id !== 1,
  },
  columns,
  height: 'auto',
  keepSource: true,
  pagerConfig: {},
  proxyConfig: {
    ajax: {
      query: async ({ page }, formValues = {}) => {
        return await roleList({
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
  id: 'system-role-index',
};

const [BasicTable, tableApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
});
const [RoleDrawerRef, drawerApi] = useVbenDrawer({
  connectedComponent: RoleDrawer,
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

function handleAdd() {
  drawerApi.setData({});
  drawerApi.open();
}

async function handleEdit(record: Role) {
  drawerApi.setData({ id: record.id });
  drawerApi.open();
}

async function handleDelete(row: Role) {
  await roleRemove(row.id);
  message.success('删除成功');
  await tableApi.query();
}

function handleMultiDelete() {
  const rows = (tableApi.grid?.getCheckboxRecords?.() ?? []) as Role[];
  if (rows.length === 0) return;
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await roleRemove(id);
      }
      message.success('删除成功');
      await tableApi.query();
    },
  });
}

async function handleStatusChange(row: Role) {
  await roleStatusChange(row.id, row.status);
  message.success('状态更新成功');
}

function onReload() {
  tableApi.query();
}

function handleAssignRole(record: Role) {
  router.push(`/system/role-auth/user/${record.id}`);
}
</script>

<template>
  <Page :auto-content-height="true">
    <BasicTable table-title="角色列表">
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!vxeCheckboxChecked(tableApi)"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button type="primary" @click="handleAdd">新 增</a-button>
        </Space>
      </template>
      <template #status="{ row }">
        <Switch
          v-model:checked="row.status"
          :checked-value="1"
          :un-checked-value="0"
          :checked-children="statusLabel.checked"
          :un-checked-children="statusLabel.unchecked"
          :disabled="row.id === 1"
          @change="() => handleStatusChange(row)"
        />
      </template>
      <template #action="{ row }">
        <template v-if="row.id !== 1">
          <Space>
            <ghost-button @click.stop="handleEdit(row)">
              编 辑
            </ghost-button>
            <ghost-button @click.stop="handleAssignRole(row)">
              分 配
            </ghost-button>
            <Popconfirm
              :get-popup-container="getPopupContainer"
              placement="left"
              title="确认删除？"
              @confirm="handleDelete(row)"
            >
              <ghost-button danger @click.stop="">
                删 除
              </ghost-button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </BasicTable>
    <RoleDrawerRef @reload="onReload" />
  </Page>
</template>
