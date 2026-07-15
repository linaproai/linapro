<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Menu } from '#/api/system/menu';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { getPopupContainer } from '@vben/utils';

import { message, Popconfirm, Space, Switch, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { menuList, menuRemove, menuUpdate } from '#/api/system/menu';
import { $t } from '#/locales';
import { refreshAccessibleState } from '#/router/access-refresh';
import { eachTree, treeToList } from '#/utils/tree';

import { columns, querySchema } from './data';
import MenuDrawer from './menu-drawer.vue';

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

const router = useRouter();

const gridOptions: VxeGridProps = {
  columns,
  height: 'auto',
  keepSource: true,
  pagerConfig: {
    enabled: false,
  },
  proxyConfig: {
    ajax: {
      query: async (_, formValues = {}) => {
        const resp = await menuList({
          ...formValues,
        });
        // Backend already returns tree structure, use directly
        const treeData = resp;
        // Add hasChildren field for VXE-Grid lazy tree
        eachTree(treeData, (item: any) => {
          item.hasChildren = !!(item.children && item.children.length > 0);
        });
        return { items: treeData };
      },
    },
  },
  rowConfig: {
    keyField: 'id',
    isCurrent: true,
  },
  rowClassName: ({ row }: any) =>
    hasExpandableChildren(row) ? 'system-menu-row--expandable' : '',
  treeConfig: {
    parentField: 'parentId',
    rowField: 'id',
    transform: false,
    reserve: true,
    hasChildField: 'hasChildren',
    lazy: true,
    loadMethod: ({ row }: any) => row.children ?? [],
  },
  id: 'system-menu-index',
};

const [BasicTable, tableApi] = useVbenVxeGrid({
  formOptions,
  gridOptions,
  gridEvents: {
    cellClick: (e: any) => {
      const { column = {}, row = {} } = e;
      if (column.field !== 'name') {
        return;
      }
      toggleTreeRow(row);
    },
    toggleTreeExpand: (e: any) => {
      const { row = {}, expanded } = e;
      row.expand = expanded;
    },
  },
});
const [MenuDrawerRef, drawerApi] = useVbenDrawer({
  connectedComponent: MenuDrawer,
});

function hasExpandableChildren(row: any) {
  return Array.isArray(row?.children) && row.children.length > 0;
}

function toggleTreeRow(row: any) {
  if (!hasExpandableChildren(row)) {
    return;
  }
  const isExpanded = row?.expand;
  tableApi.grid.setTreeExpand(row, !isExpanded);
  row.expand = !isExpanded;
}

function handleAdd() {
  drawerApi.setData({ isEdit: false });
  drawerApi.open();
}

function handleSubAdd(row: Menu) {
  drawerApi.setData({ parentId: row.id, isEdit: false });
  drawerApi.open();
}

async function handleEdit(record: Menu) {
  drawerApi.setData({ id: record.id, update: true });
  drawerApi.open();
}

/**
 * 是否级联删除
 */
const cascadingDeletion = ref(false);
async function handleDelete(row: Menu) {
  await menuRemove(row.id, cascadingDeletion.value);
  await Promise.all([
    tableApi.query(),
    refreshAccessibleState(router, { showLoadingToast: false }),
  ]);
}

function removeConfirmTitle(row: Menu) {
  const menuName = row.name;
  if (!cascadingDeletion.value) {
    return $t('pages.system.menu.messages.deleteConfirm', { menuName });
  }
  const menuAndChildren = treeToList([row], { childProp: 'children' });
  if (menuAndChildren.length === 1) {
    return $t('pages.system.menu.messages.deleteConfirm', { menuName });
  }
  return $t('pages.system.menu.messages.deleteWithChildrenConfirm', {
    childCount: menuAndChildren.length - 1,
    menuName,
  });
}

/**
 * 编辑/添加成功后刷新表格
 */
async function afterEditOrAdd() {
  await Promise.all([
    tableApi.query(),
    refreshAccessibleState(router, { showLoadingToast: false }),
  ]);
}

/**
 * 全部展开/折叠
 */
function setExpandOrCollapse(expand: boolean) {
  eachTree(tableApi.grid.getData() as any[], (item: any) => (item.expand = expand));
  tableApi.grid?.setAllTreeExpand(expand);
}

const statusLabel = {
  checked: $t('pages.status.enabled'),
  unchecked: $t('pages.status.disabled'),
};
const visibleLabel = {
  checked: $t('pages.system.menu.visible.shown'),
  unchecked: $t('pages.system.menu.visible.hidden'),
};

/** Row IDs currently submitting a status or visibility switch. */
const statusChangingIds = ref<Record<number, boolean>>({});
const visibleChangingIds = ref<Record<number, boolean>>({});

function isStatusChanging(row: Menu) {
  return statusChangingIds.value[row.id] === true;
}

function isVisibleChanging(row: Menu) {
  return visibleChangingIds.value[row.id] === true;
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

function setVisibleChanging(id: number, changing: boolean) {
  const next = { ...visibleChangingIds.value };
  if (changing) {
    next[id] = true;
  } else {
    delete next[id];
  }
  visibleChangingIds.value = next;
}

async function reloadMenuSurfaces() {
  await Promise.all([
    tableApi.query(),
    refreshAccessibleState(router, { showLoadingToast: false }),
  ]);
}

/**
 * Toggle menu status from the list switch. The server cascades the target
 * status value to all descendants.
 */
async function handleStatusChange(row: Menu, checked: boolean) {
  if (isStatusChanging(row)) {
    return;
  }
  const next = checked ? 1 : 0;
  if (row.status === next) {
    return;
  }
  // Keep the controlled Switch on the current value while the request is in
  // flight; reload commits cascade results after the API succeeds.
  setStatusChanging(row.id, true);
  try {
    await menuUpdate(row.id, { status: next });
    await reloadMenuSurfaces();
    message.success($t('pages.common.updateSuccess'));
  } catch {
    await tableApi.query();
  } finally {
    setStatusChanging(row.id, false);
  }
}

/**
 * Toggle menu visibility from the list switch. The server cascades the target
 * visibility value to all descendants.
 */
async function handleVisibleChange(row: Menu, checked: boolean) {
  if (isVisibleChanging(row)) {
    return;
  }
  const next = checked ? 1 : 0;
  if (row.visible === next) {
    return;
  }
  // Keep the controlled Switch on the current value while the request is in
  // flight; reload commits cascade results after the API succeeds.
  setVisibleChanging(row.id, true);
  try {
    await menuUpdate(row.id, { visible: next });
    await reloadMenuSurfaces();
    message.success($t('pages.common.updateSuccess'));
  } catch {
    await tableApi.query();
  } finally {
    setVisibleChanging(row.id, false);
  }
}

/**
 * 菜单管理页面权限暂时不做前端校验
 * 后端API已有权限控制，前端仅展示
 */
const canAccess = ref(true);
</script>

<template>
  <Page v-if="canAccess" :auto-content-height="true">
    <BasicTable
      id="system-menu-table"
      :table-title="$t('pages.system.menu.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <Tooltip :title="$t('pages.system.menu.messages.cascadeDeleteTooltip')">
            <div class="mr-2 flex items-center">
              <span class="mr-2 text-sm text-[#666666]">
                {{ $t('pages.system.menu.fields.cascadeDelete') }}
              </span>
              <Switch
                v-model:checked="cascadingDeletion"
                data-testid="menu-cascade-delete-switch"
              />
            </div>
          </Tooltip>

          <a-button @click="setExpandOrCollapse(false)">
            {{ $t('pages.common.collapse') }}
          </a-button>

          <a-button
            type="primary"
            @click="handleAdd"
          >
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
          :disabled="isStatusChanging(row)"
          data-testid="menu-status-switch"
          @change="(checked) => handleStatusChange(row, !!checked)"
        />
      </template>
      <template #visible="{ row }">
        <Switch
          :checked="row.visible === 1"
          :checked-children="visibleLabel.checked"
          :un-checked-children="visibleLabel.unchecked"
          :loading="isVisibleChanging(row)"
          :disabled="isVisibleChanging(row)"
          data-testid="menu-visible-switch"
          @change="(checked) => handleVisibleChange(row, !!checked)"
        />
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button
            @click="handleEdit(row)"
          >
            {{ $t('pages.common.edit') }}
          </ghost-button>
          <!-- '按钮类型'无法再添加子菜单 -->
          <ghost-button
            v-if="row.type !== 'B'"
            class="btn-success"
            @click="handleSubAdd(row)"
          >
            {{ $t('pages.common.add') }}
          </ghost-button>
          <Popconfirm
            :get-popup-container="getPopupContainer"
            placement="left"
            :title="removeConfirmTitle(row)"
            @confirm="handleDelete(row)"
          >
            <ghost-button
              danger
              @click.stop=""
            >
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </BasicTable>
    <MenuDrawerRef @reload="afterEditOrAdd" />
  </Page>
</template>

<style lang="scss">
#system-menu-table > .vxe-grid {
  --vxe-ui-table-row-current-background-color: hsl(var(--primary-100));

  html.dark & {
    --vxe-ui-table-row-current-background-color: hsl(var(--primary-800));
  }

  .system-menu-row--expandable .system-menu-name-column {
    cursor: pointer;
  }
}
</style>
