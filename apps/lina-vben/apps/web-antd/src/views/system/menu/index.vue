<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Menu } from '#/api/system/menu';

import { ref } from 'vue';
import { useRouter } from 'vue-router';

import { Page, useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { getPopupContainer } from '@vben/utils';

import { Popconfirm, Space, Switch, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { menuList, menuRemove } from '#/api/system/menu';
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
    cellDblclick: (e: any) => {
      const { row = {} } = e;
      if (!row?.children) {
        return;
      }
      const isExpanded = row?.expand;
      tableApi.grid.setTreeExpand(row, !isExpanded);
      row.expand = !isExpanded;
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
    return `是否确认删除 [${menuName}] ?`;
  }
  const menuAndChildren = treeToList([row], { childProp: 'children' });
  if (menuAndChildren.length === 1) {
    return `是否确认删除 [${menuName}] ?`;
  }
  return `是否确认删除 [${menuName}] 及 [${menuAndChildren.length - 1}]个子项目 ?`;
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
      table-title="菜单列表"
      table-title-help="双击展开/收起子菜单"
    >
      <template #toolbar-tools>
        <Space>
          <Tooltip title="删除菜单以及子菜单">
            <div class="mr-2 flex items-center">
              <span class="mr-2 text-sm text-[#666666]">级联删除</span>
              <Switch v-model:checked="cascadingDeletion" />
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
}
</style>
