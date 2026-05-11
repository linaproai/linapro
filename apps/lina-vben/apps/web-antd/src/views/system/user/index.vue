<script setup lang="ts">
import { Page, useVbenDrawer, useVbenModal } from '@vben/common-ui';
import { preferences } from '@vben/preferences';
import { useUserStore } from '@vben/stores';

import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useRoute } from 'vue-router';

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
  userBatchDelete,
  userDelete,
  userExport,
  userList,
  userStatusChange,
} from '#/api/system/user';
import { $t } from '#/locales';
import { pluginCapabilityKeys } from '#/plugins/plugin-capabilities';
import {
  getPluginCapabilityStateMap,
  onPluginRegistryChanged,
} from '#/plugins/slot-registry';
import { useDictStore } from '#/store/dict';
import { useTenantStore } from '#/store/tenant';
import { downloadBlob } from '#/utils/download';

import { buildColumns, querySchema } from './data';
import DeptTree from './dept-tree.vue';
import UserDrawer from './user-drawer.vue';
import UserBatchEditModal from './user-batch-edit-modal.vue';
import UserImportModal from './user-import-modal.vue';
import UserResetPwdModal from './user-reset-pwd-modal.vue';
import { loadUserTenantOptions } from './tenant-options';

const [UserDrawerRef, userDrawerApi] = useVbenDrawer({
  connectedComponent: UserDrawer,
});

const [UserImportModalRef, userImportModalApi] = useVbenModal({
  connectedComponent: UserImportModal,
});

const [UserBatchEditModalRef, userBatchEditModalApi] = useVbenModal({
  connectedComponent: UserBatchEditModal,
});

const [UserResetPwdModalRef, userResetPwdModalApi] = useVbenModal({
  connectedComponent: UserResetPwdModal,
});

const userStore = useUserStore();
const dictStore = useDictStore();
const tenantStore = useTenantStore();
const route = useRoute();

const orgEnabled = ref(false);
const tenantEnabled = ref(false);
const selectDeptId = ref<string[]>([]);
const tenantOptions = ref<Array<{ label: string; value: number }>>([]);
const userStatusOptions = ref<Array<{ label: string; value: number }>>([]);
const deptTreeRef = ref<InstanceType<typeof DeptTree>>();
const checkedRows = ref<any[]>([]);
const managementCapabilitiesReady = ref(false);
const hasChecked = computed(() => checkedRows.value.length > 0);
const statusLabel = computed(() => {
  const opts = dictStore.dictOptionsMap.get('sys_normal_disable') || [];
  const checked = opts.find((d) => d.value === '1');
  const unchecked = opts.find((d) => d.value === '0');
  return {
    checked: checked?.label || $t('pages.status.enabled'),
    unchecked: unchecked?.label || $t('pages.status.disabled'),
  };
});

let disposePluginRegistryListener: null | (() => void) = null;

function parseRouteTenantId() {
  const rawTenantId = Array.isArray(route.query.tenantId)
    ? route.query.tenantId[0]
    : route.query.tenantId;
  const tenantId = Number(rawTenantId);
  return Number.isFinite(tenantId) && tenantId > 0 ? tenantId : undefined;
}

function isSelf(row: any) {
  return row.id === Number(userStore.userInfo?.userId);
}

async function loadTenantOptions(force = false) {
  if (!force && tenantOptions.value.length > 0) {
    return tenantOptions.value;
  }
  if (!tenantEnabled.value || !tenantStore.isPlatform) {
    tenantOptions.value = [];
    return tenantOptions.value;
  }
  tenantOptions.value = await loadUserTenantOptions({
    currentTenant: tenantStore.currentTenant,
    isPlatform: tenantStore.isPlatform,
    tenants: tenantStore.tenants,
    userId: Number(userStore.userInfo?.userId || 0),
  });
  return tenantOptions.value;
}

async function syncTenantQuerySchema() {
  const schema = querySchema(tenantEnabled.value && tenantStore.isPlatform);
  gridApi.formApi.setState({ schema });
  syncStatusQueryOptions();
  if (!tenantEnabled.value || !tenantStore.isPlatform) {
    const values = gridApi.formApi.form.values;
    Reflect.deleteProperty(values, 'tenantId');
    gridApi.formApi.setLatestSubmissionValues(values);
    return;
  }
  const options = await loadTenantOptions();
  gridApi.formApi.updateSchema([
    {
      fieldName: 'tenantId',
      componentProps: {
        options,
      },
    },
  ]);
}

async function applyRouteTenantFilter({
  clearWhenMissing = true,
}: { clearWhenMissing?: boolean } = {}) {
  const tenantId = parseRouteTenantId();
  if (!tenantEnabled.value || !tenantStore.isPlatform) {
    return;
  }
  if (!tenantId) {
    if (clearWhenMissing && gridApi.formApi.form.values.tenantId) {
      await gridApi.formApi.setFieldValue('tenantId', undefined);
      const values = { ...gridApi.formApi.form.values };
      Reflect.deleteProperty(values, 'tenantId');
      gridApi.formApi.setLatestSubmissionValues(values);
      await gridApi.reload(values);
    }
    return;
  }
  const options = await loadTenantOptions();
  if (!options.some((item) => item.value === tenantId)) {
    return;
  }
  await gridApi.formApi.setFieldValue('tenantId', tenantId);
  const values = { ...gridApi.formApi.form.values, tenantId };
  gridApi.formApi.setLatestSubmissionValues(values);
  await gridApi.reload(values);
}

function syncStatusQueryOptions() {
  if (userStatusOptions.value.length === 0) {
    return;
  }
  gridApi.formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: userStatusOptions.value,
      },
    },
  ]);
}

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: querySchema(),
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
    columns: buildColumns(orgEnabled.value, tenantEnabled.value),
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

async function syncManagementCapabilities(force = false) {
  const capabilityMap = await getPluginCapabilityStateMap(force);
  const nextOrgEnabled =
    capabilityMap.get(pluginCapabilityKeys.organizationManagement)?.enabled ===
    true;
  const nextTenantEnabled =
    capabilityMap.get(pluginCapabilityKeys.tenantManagement)?.enabled === true;
  const capabilityChanged =
    orgEnabled.value !== nextOrgEnabled ||
    tenantEnabled.value !== nextTenantEnabled;

  orgEnabled.value = nextOrgEnabled;
  tenantEnabled.value = nextTenantEnabled;
  if (!nextOrgEnabled) {
    selectDeptId.value = [];
  }

  gridApi.setGridOptions({
    columns: buildColumns(nextOrgEnabled, nextTenantEnabled),
  });
  await syncTenantQuerySchema();

  if (capabilityChanged) {
    checkedRows.value = [];
    await gridApi.reload();
  }
}

onMounted(async () => {
  const statusOptions =
    await dictStore.getDictOptionsAsync('sys_normal_disable');
  userStatusOptions.value = statusOptions.map((d) => ({
    label: d.label,
    value: Number(d.value),
  }));
  syncStatusQueryOptions();

  await syncManagementCapabilities();
  managementCapabilitiesReady.value = true;
  await applyRouteTenantFilter({ clearWhenMissing: false });
  disposePluginRegistryListener = onPluginRegistryChanged(async () => {
    await syncManagementCapabilities(true);
    await applyRouteTenantFilter({ clearWhenMissing: false });
  });
});

watch(
  () => route.query.tenantId,
  async () => {
    if (!managementCapabilitiesReady.value) {
      return;
    }
    await applyRouteTenantFilter();
  },
);

onBeforeUnmount(() => {
  disposePluginRegistryListener?.();
  disposePluginRegistryListener = null;
});

function handleAdd() {
  userDrawerApi.setData({
    isEdit: false,
    orgEnabled: orgEnabled.value,
    tenantEnabled: tenantEnabled.value,
  });
  userDrawerApi.open();
}

function handleEdit(row: any) {
  userDrawerApi.setData({
    isEdit: true,
    orgEnabled: orgEnabled.value,
    row,
    tenantEnabled: tenantEnabled.value,
  });
  userDrawerApi.open();
}

async function handleDelete(row: any) {
  await userDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  await gridApi.query();
  deptTreeRef.value?.refreshTree();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords();
  const ids = rows.map((row: any) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.user.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await userBatchDelete(ids);
      checkedRows.value = [];
      await gridApi.query();
      deptTreeRef.value?.refreshTree();
    },
  });
}

function handleBatchEdit() {
  userBatchEditModalApi.setData({
    rows: gridApi.grid.getCheckboxRecords(),
    tenantEnabled: tenantEnabled.value,
  });
  userBatchEditModalApi.open();
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
      ? $t('pages.system.user.messages.exportSelectedConfirm')
      : $t('pages.system.user.messages.exportAllConfirm');

  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'primary',
    content,
    okText: $t('pages.common.confirm'),
    cancelText: $t('pages.common.cancel'),
    onOk: async () => {
      try {
        const ids = checkedRows.value.map((row: any) => row.id);
        const data = await userExport({ ids });
        downloadBlob(data, $t('pages.system.user.exportFileName'));
        message.success($t('pages.common.exportSuccess'));
      } catch {
        message.error($t('pages.common.exportFailed'));
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
    <div class="flex h-full flex-col gap-[8px] 2xl:flex-row">
      <DeptTree
        v-if="orgEnabled"
        ref="deptTreeRef"
        v-model:select-dept-id="selectDeptId"
        class="w-full shrink-0 2xl:w-[240px]"
        @reload="() => gridApi.reload()"
        @select="() => gridApi.reload()"
      />
      <Grid
        class="flex-1 overflow-hidden"
        :table-title="$t('pages.system.user.tableTitle')"
      >
        <template #toolbar-tools>
          <Space>
            <a-button @click="handleExport">
              {{ $t('pages.common.export') }}
            </a-button>
            <a-button @click="handleImport">
              {{ $t('pages.common.import') }}
            </a-button>
            <a-button
              data-testid="user-batch-edit-button"
              :disabled="!hasChecked"
              @click="handleBatchEdit"
            >
              {{ $t('pages.system.user.actions.batchEdit') }}
            </a-button>
            <a-button
              data-testid="user-batch-delete-button"
              :disabled="!hasChecked"
              danger
              type="primary"
              @click="handleMultiDelete"
            >
              {{ $t('pages.common.delete') }}
            </a-button>
            <a-button
              data-testid="user-create-button"
              type="primary"
              @click="handleAdd"
            >
              {{ $t('pages.common.add') }}
            </a-button>
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
              <ghost-button @click.stop="handleEdit(row)">
                {{ $t('pages.common.edit') }}
              </ghost-button>
              <Popconfirm
                placement="left"
                :title="$t('pages.system.user.messages.deleteConfirm')"
                @confirm="handleDelete(row)"
              >
                <ghost-button danger @click.stop="">
                  {{ $t('pages.common.delete') }}
                </ghost-button>
              </Popconfirm>
            </Space>
            <Dropdown placement="bottomRight">
              <template #overlay>
                <Menu>
                  <MenuItem key="resetPwd" @click="handleResetPwd(row)">
                    {{ $t('pages.system.user.actions.resetPassword') }}
                  </MenuItem>
                </Menu>
              </template>
              <a-button size="small" type="link">
                {{ $t('pages.common.more') }}
              </a-button>
            </Dropdown>
          </template>
        </template>
      </Grid>
    </div>

    <UserDrawerRef @success="onReload" />
    <UserBatchEditModalRef @success="onReload" />
    <UserImportModalRef @reload="onReload" />
    <UserResetPwdModalRef @reload="onReload" />
  </Page>
</template>
