<script setup lang="ts">
import type { SysConfig } from '#/api/system/config/model';

import { computed, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  configDelete,
  configExport,
  configList,
} from '#/api/system/config';
import { downloadBlob } from '#/utils/download';

import { columns, querySchema } from './data';
import ConfigImportModal from './config-import-modal.vue';
import ConfigModal from './config-modal.vue';

const [ConfigModalRef, modalApi] = useVbenModal({
  connectedComponent: ConfigModal,
});

const [ImportModalRef, importModalApi] = useVbenModal({
  connectedComponent: ConfigImportModal,
});

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
    fieldMappingTime: [
      ['createTime', ['beginTime', 'endTime'], ['YYYY-MM-DD', 'YYYY-MM-DD']],
    ],
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns,
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }, formValues = {}) => {
          return await configList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-config-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as SysConfig[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as SysConfig[];
    },
  },
});

const checkedRows = ref<SysConfig[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function handleAdd() {
  modalApi.setData({});
  modalApi.open();
}

function handleEdit(row: SysConfig) {
  modalApi.setData({ id: row.id });
  modalApi.open();
}

async function handleDelete(row: SysConfig) {
  await configDelete(row.id);
  message.success('删除成功');
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as SysConfig[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await configDelete(id);
      }
      checkedRows.value = [];
      await gridApi.query();
    },
  });
}

async function handleExport() {
  const content = checkedRows.value.length > 0
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
        const formValues = gridApi.formApi.form.values;
        const params: Record<string, any> = { ...formValues };
        if (params.createTime) {
          delete params.createTime;
        }
        if (checkedRows.value.length > 0) {
          params.ids = checkedRows.value.map((row) => row.id);
        }
        const data = await configExport(params);
        downloadBlob(data, '参数设置导出.xlsx');
        message.success('导出成功');
      } catch {
        message.error('导出失败');
      }
    },
  });
}

function onReload() {
  gridApi.query();
}

function handleImport() {
  importModalApi.open();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="参数设置列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">导 出</a-button>
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

      <template #action="{ row }">
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
      </template>
    </Grid>

    <ConfigModalRef @reload="onReload" />
    <ImportModalRef @reload="onReload" />
  </Page>
</template>
