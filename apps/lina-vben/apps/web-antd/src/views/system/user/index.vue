<script setup lang="ts">
import { Page, useVbenDrawer, useVbenModal } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { computed, onBeforeUnmount, onMounted, ref } from 'vue';

import {
  Avatar,
  Dropdown,
  Menu,
  MenuItem,
  message,
  Modal,
  Popconfirm,
  Space,
  Switch,
} from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  getDeptTree,
  userDelete,
  userExport,
  userList,
  userStatusChange,
} from '#/api/system/user';
import {
  getPluginStateMap,
  onPluginRegistryChanged,
} from '#/plugins/slot-registry';
import { useDictStore } from '#/store/dict';
import { downloadBlob } from '#/utils/download';

import { buildColumns, querySchema } from './data';
import DeptTree from './dept-tree.vue';
import UserDrawer from './user-drawer.vue';
import UserImportModal from './user-import-modal.vue';
import UserResetPwdModal from './user-reset-pwd-modal.vue';

const orgManagementPluginId = 'org-management';

const [UserDrawerRef, userDrawerApi] = useVbenDrawer({
  connectedComponent: UserDrawer,
});

const [UserImportModalRef, userImportModalApi] = useVbenModal({
  connectedComponent: UserImportModal,
});

const [UserResetPwdModalRef, userResetPwdModalApi] = useVbenModal({
  connectedComponent: UserResetPwdModal,
});

const userStore = useUserStore();
const dictStore = useDictStore();

const orgEnabled = ref(false);
const selectDeptId = ref<string[]>([]);
const deptTreeRef = ref<InstanceType<typeof DeptTree>>();
const checkedRows = ref<any[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);
const statusLabel = computed(() => {
  const opts = dictStore.dictOptionsMap.get('sys_normal_disable') || [];
  const checked = opts.find((d) => d.value === '1');
  const unchecked = opts.find((d) => d.value === '0');
  return {
    checked: checked?.label || '正常',
    unchecked: unchecked?.label || '停用',
  };
});

let disposePluginRegistryListener: null | (() => void) = null;

function isSelf(row: any) {
  return row.id === Number(userStore.userInfo?.userId);
}

function isPluginEnabled(pluginId: string, pluginStateMap: Map<string, any>) {
  const pluginState = pluginStateMap.get(pluginId);
  return pluginState?.installed === 1 && pluginState?.enabled === 1;
}

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: querySchema,
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
    handleReset: async () => {
      selectDeptId.value = [];

      const { formApi, reload } = gridApi;
      await formApi.resetForm();
      const formValues = formApi.form.values;
      formApi.setLatestSubmissionValues(formValues);
      await reload(formValues);
    },
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
      checkMethod: ({ row }: any) => !isSelf(row),
    },
    columns: buildColumns(orgEnabled.value),
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    sortConfig: {
      remote: true,
      trigger: 'cell',
    },
    proxyConfig: {
      sort: true,
      ajax: {
        query: async (
          { page, sorts }: any,
          formValues: Record<string, any> = {},
        ) => {
          const sortParams: Record<string, string> = {};
          if (sorts && sorts.length > 0) {
            const sort = sorts[0];
            if (sort && sort.order) {
              sortParams.orderBy = sort.field;
              sortParams.orderDirection = sort.order;
            }
          }

          if (orgEnabled.value && selectDeptId.value.length === 1) {
            formValues.deptId = selectDeptId.value[0];
          } else {
            Reflect.deleteProperty(formValues, 'deptId');
          }

          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          };
          if (params.createdAt && Array.isArray(params.createdAt)) {
            params.beginTime = params.createdAt[0];
            params.endTime = params.createdAt[1];
            delete params.createdAt;
          }
          return await userList(params);
        },
      },
    },
    headerCellConfig: {
      height: 44,
    },
    cellConfig: {
      height: 48,
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-user-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = gridApi.grid?.getCheckboxRecords() || [];
    },
    checkboxAll: () => {
      checkedRows.value = gridApi.grid?.getCheckboxRecords() || [];
    },
  },
});

async function syncOrgCapability(force = false) {
  const pluginStateMap = await getPluginStateMap(force);
  const nextOrgEnabled = isPluginEnabled(orgManagementPluginId, pluginStateMap);
  const capabilityChanged = orgEnabled.value !== nextOrgEnabled;

  orgEnabled.value = nextOrgEnabled;
  if (!nextOrgEnabled) {
    selectDeptId.value = [];
  }

  gridApi.setGridOptions({
    columns: buildColumns(nextOrgEnabled),
  });

  if (capabilityChanged) {
    checkedRows.value = [];
    await gridApi.reload();
  }
}

onMounted(async () => {
  const statusOptions =
    await dictStore.getDictOptionsAsync('sys_normal_disable');
  gridApi.formApi.updateSchema([
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

  await syncOrgCapability();
  disposePluginRegistryListener = onPluginRegistryChanged(async () => {
    await syncOrgCapability(true);
  });
});

onBeforeUnmount(() => {
  disposePluginRegistryListener?.();
  disposePluginRegistryListener = null;
});

function handleAdd() {
  userDrawerApi.setData({ isEdit: false, orgEnabled: orgEnabled.value });
  userDrawerApi.open();
}

function handleEdit(row: any) {
  userDrawerApi.setData({ isEdit: true, orgEnabled: orgEnabled.value, row });
  userDrawerApi.open();
}

async function handleDelete(row: any) {
  await userDelete(row.id);
  message.success('删除成功');
  await gridApi.query();
  deptTreeRef.value?.refreshTree();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  const ids = rows.map((row: any) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await userDelete(id);
      }
      checkedRows.value = [];
      await gridApi.query();
      deptTreeRef.value?.refreshTree();
    },
  });
}

async function handleStatusChange(row: any) {
  await userStatusChange(row.id, row.status);
}

function onReload() {
  void gridApi.query();
  deptTreeRef.value?.refreshTree();
}

async function handleExport() {
  const content =
    checkedRows.value.length > 0
      ? '是否导出选中的记录？'
      : '是否导出全部数据？';

  Modal.confirm({
    title: '提示',
    okType: 'primary',
    content,
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        const ids = checkedRows.value.map((row: any) => row.id);
        const data = await userExport({ ids });
        downloadBlob(data, '用户数据导出.xlsx');
        message.success('导出成功');
      } catch {
        message.error('导出失败');
      }
    },
  });
}

function handleImport() {
  userImportModalApi.open();
}

function handleResetPwd(row: any) {
  userResetPwdModalApi.setData({ record: row });
  userResetPwdModalApi.open();
}
</script>

<template>
  <Page :auto-content-height="true">
    <div class="flex h-full gap-[8px]">
      <DeptTree
        v-if="orgEnabled"
        ref="deptTreeRef"
        v-model:select-dept-id="selectDeptId"
        :api="getDeptTree"
        class="w-[260px] shrink-0"
        @reload="() => gridApi.reload()"
        @select="() => gridApi.reload()"
      />
      <Grid class="flex-1 overflow-hidden" table-title="用户列表">
        <template #toolbar-tools>
          <Space>
            <a-button @click="handleExport"> 导 出 </a-button>
            <a-button @click="handleImport">导 入</a-button>
            <a-button
              :disabled="!hasChecked"
              danger
              type="primary"
              @click="handleMultiDelete"
            >
              删 除
            </a-button>
            <a-button type="primary" @click="handleAdd">新 增</a-button>
          </Space>
        </template>

        <template #avatar="{ row }">
          <Avatar :src="row.avatar || preferences.app.defaultAvatar" />
        </template>

        <template #status="{ row }">
          <Switch
            v-model:checked="row.status"
            :checked-value="1"
            :disabled="isSelf(row)"
            :un-checked-value="0"
            :checked-children="statusLabel.checked"
            :un-checked-children="statusLabel.unchecked"
            @change="() => handleStatusChange(row)"
          />
        </template>

        <template #action="{ row }">
          <template v-if="!isSelf(row)">
            <Space>
              <ghost-button @click.stop="handleEdit(row)">编辑</ghost-button>
              <Popconfirm
                placement="left"
                title="确认删除？"
                @confirm="handleDelete(row)"
              >
                <ghost-button danger @click.stop="">删除</ghost-button>
              </Popconfirm>
            </Space>
            <Dropdown placement="bottomRight">
              <template #overlay>
                <Menu>
                  <MenuItem key="resetPwd" @click="handleResetPwd(row)">
                    重置密码
                  </MenuItem>
                </Menu>
              </template>
              <a-button size="small" type="link">更多</a-button>
            </Dropdown>
          </template>
        </template>
      </Grid>
    </div>

    <UserDrawerRef @success="onReload" />
    <UserImportModalRef @reload="onReload" />
    <UserResetPwdModalRef @reload="onReload" />
  </Page>
</template>
