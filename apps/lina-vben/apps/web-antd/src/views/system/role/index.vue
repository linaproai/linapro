<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Role } from '#/api/system/role';

import { onBeforeUnmount, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { Modal, Popconfirm, Space, Switch, message } from 'ant-design-vue';

import { useVbenVxeGrid, vxeCheckboxChecked } from '#/adapter/vxe-table';
import {
  roleBatchDelete,
  roleList,
  roleRemove,
  roleStatusChange,
} from '#/api/system/role';
import { $t } from '#/locales';
import { resolveManagementCapabilityState } from '#/plugins/management-capabilities';
import { onPluginRegistryChanged } from '#/plugins/slot-registry';
import { useDictStore } from '#/store/dict';

import { DATA_SCOPE_DICT_TYPE, columns, querySchema } from './data';
import RoleDrawer from './role-drawer.vue';

const router = useRouter();

const formOptions: VbenFormProps = {
  commonConfig: {
    labelWidth: 80,
    componentProps: {
      allowClear: true,
    },
  },
  schema: querySchema(),
  wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
};

const gridOptions: VxeGridProps = {
  checkboxConfig: {
    highlight: true,
    reserve: true,
    checkMethod: ({ row }) => row.id !== 1,
  },
  columns: columns(true, false),
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
let disposePluginRegistryListener: null | (() => void) = null;

// 加载字典数据
const dictStore = useDictStore();
const statusLabel = ref({
  checked: $t('pages.status.enabled'),
  unchecked: $t('pages.status.disabled'),
});

async function syncRoleCapabilities(force = false) {
  const capabilities = await resolveManagementCapabilityState(force);
  tableApi.setGridOptions({
    columns: columns(
      capabilities.organizationEnabled,
      capabilities.tenantEnabled,
    ),
  });
}

onMounted(async () => {
  const statusOptions =
    await dictStore.getDictOptionsAsync('sys_normal_disable');
  await dictStore.getDictOptionsAsync(DATA_SCOPE_DICT_TYPE);
  await syncRoleCapabilities();
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
  disposePluginRegistryListener = onPluginRegistryChanged(async () => {
    await syncRoleCapabilities(true);
  });
});

onBeforeUnmount(() => {
  disposePluginRegistryListener?.();
  disposePluginRegistryListener = null;
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
  message.success($t('pages.common.deleteSuccess'));
  await tableApi.query();
}

function handleMultiDelete() {
  const rows = (tableApi.grid?.getCheckboxRecords?.() ?? []) as Role[];
  if (rows.length === 0) return;
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.role.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await roleBatchDelete(ids);
      message.success($t('pages.common.deleteSuccess'));
      await tableApi.query();
    },
  });
}

/** Row IDs currently submitting a status switch request. */
const statusChangingIds = ref<Record<number, boolean>>({});

function isStatusChanging(row: Role) {
  return statusChangingIds.value[row.id] === true;
}

function setStatusChanging(id: number, changing: boolean) {
  const next = { ...statusChangingIds.value };
  if (changing) {
    next[id] = true;
  } else {
    delete next[id];
  }
  statusChangingIds.value = next;
}

async function handleStatusChange(row: Role, checked: boolean) {
  if (row.id === 1 || isStatusChanging(row)) {
    return;
  }
  const previous = row.status;
  const next = checked ? 1 : 0;
  if (previous === next) {
    return;
  }
  setStatusChanging(row.id, true);
  row.status = next;
  try {
    await roleStatusChange(row.id, next);
    message.success($t('pages.system.role.messages.statusUpdated'));
  } catch {
    row.status = previous;
    await tableApi.query();
  } finally {
    setStatusChanging(row.id, false);
  }
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
    <BasicTable :table-title="$t('pages.system.role.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button
            data-testid="role-batch-delete-button"
            :disabled="!vxeCheckboxChecked(tableApi)"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button type="primary" @click="handleAdd">
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>
      <template #status="{ row }">
        <Switch
          :checked="row.status === 1"
          :checked-children="statusLabel.checked"
          :un-checked-children="statusLabel.unchecked"
          :loading="isStatusChanging(row)"
          :disabled="row.id === 1 || isStatusChanging(row)"
          @change="(checked) => handleStatusChange(row, !!checked)"
        />
      </template>
      <template #action="{ row }">
        <template v-if="row.id !== 1">
          <Space>
            <ghost-button @click.stop="handleEdit(row)">
              {{ $t('pages.common.edit') }}
            </ghost-button>
            <ghost-button @click.stop="handleAssignRole(row)">
              {{ $t('pages.system.role.actions.assign') }}
            </ghost-button>
            <Popconfirm
              :get-popup-container="getPopupContainer"
              placement="left"
              :title="$t('pages.system.role.messages.deleteConfirm')"
              @confirm="handleDelete(row)"
            >
              <ghost-button danger @click.stop="">
                {{ $t('pages.common.delete') }}
              </ghost-button>
            </Popconfirm>
          </Space>
        </template>
      </template>
    </BasicTable>
    <RoleDrawerRef @reload="onReload" />
  </Page>
</template>
