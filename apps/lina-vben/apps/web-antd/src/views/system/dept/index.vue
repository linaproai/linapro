<script setup lang="ts">
import type { VbenFormProps } from '@vben/common-ui';

import type { VxeGridProps } from '#/adapter/vxe-table';
import type { Dept } from '#/api/system/dept/model';

import { nextTick, onMounted } from 'vue';

import { Page, useVbenDrawer } from '@vben/common-ui';

import { message, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deptDelete, deptList } from '#/api/system/dept';
import { useDictStore } from '#/store/dict';

import { columns, querySchema } from './data';
import deptDrawer from './dept-drawer.vue';

// 加载字典数据
const dictStore = useDictStore();

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_normal_disable');
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

/** 遍历树形数据 */
function eachTreeData(
  data: any[],
  callback: (item: any) => void,
  childrenField = 'children',
) {
  for (const item of data) {
    callback(item);
    if (item[childrenField] && item[childrenField].length > 0) {
      eachTreeData(item[childrenField], callback, childrenField);
    }
  }
}

const gridOptions: VxeGridProps = {
  columns,
  height: 'auto',
  keepSource: true,
  pagerConfig: {
    enabled: false,
  },
  proxyConfig: {
    ajax: {
      query: async (_: any, formValues: Record<string, any> = {}) => {
        const resp = await deptList(formValues);
        return { items: resp };
      },
      // 默认请求接口后展开全部
      querySuccess: () => {
        eachTreeData(
          tableApi.grid.getData(),
          (item) => (item.expand = true),
        );
        nextTick(() => {
          setExpandOrCollapse(true);
        });
      },
    },
  },
  scrollY: {
    enabled: false,
    gt: 0,
  },
  rowConfig: {
    keyField: 'id',
  },
  treeConfig: {
    parentField: 'parentId',
    rowField: 'id',
    transform: true,
  },
  id: 'system-dept-index',
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
    // 需要监听使用箭头展开的情况 否则展开/折叠的数据不一致
    toggleTreeExpand: (e: any) => {
      const { row = {}, expanded } = e;
      row.expand = expanded;
    },
  },
});

const [DeptDrawer, drawerApi] = useVbenDrawer({
  connectedComponent: deptDrawer,
});

function handleAdd() {
  drawerApi.setData({ update: false });
  drawerApi.open();
}

function handleSubAdd(row: Dept) {
  const { id } = row;
  drawerApi.setData({ id, update: false });
  drawerApi.open();
}

async function handleEdit(record: Dept) {
  drawerApi.setData({ id: record.id, update: true });
  drawerApi.open();
}

async function handleDelete(row: Dept) {
  await deptDelete(row.id);
  message.success('删除成功');
  await tableApi.query();
}

/**
 * 全部展开/折叠
 * @param expand 是否展开
 */
function setExpandOrCollapse(expand: boolean) {
  eachTreeData(tableApi.grid.getData(), (item) => (item.expand = expand));
  tableApi.grid?.setAllTreeExpand(expand);
}
</script>

<template>
  <Page :auto-content-height="true">
    <BasicTable table-title="部门列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="setExpandOrCollapse(false)">折 叠</a-button>
          <a-button @click="setExpandOrCollapse(true)">展 开</a-button>
          <a-button type="primary" @click="handleAdd">新增</a-button>
        </Space>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button @click="handleEdit(row)">编辑</ghost-button>
          <ghost-button class="btn-success" @click="handleSubAdd(row)">
            新增
          </ghost-button>
          <Popconfirm
            placement="left"
            title="确认删除？"
            @confirm="handleDelete(row)"
          >
            <ghost-button danger @click.stop="">删除</ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </BasicTable>
    <DeptDrawer @reload="tableApi.query()" />
  </Page>
</template>
