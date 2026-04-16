<script setup lang="ts">
import type { DictType } from '#/api/system/dict/dict-type-model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { message, Modal, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  dictExport,
  dictTypeDelete,
  dictTypeList,
} from '#/api/system/dict/dict-type';
import { downloadBlob } from '#/utils/download';

import { emitter } from '../mitt';
import { columns, querySchema } from './data';
import DictTypeImportModal from './dict-type-import-modal.vue';
import dictTypeModal from './dict-type-modal.vue';

const [DictTypeModal, modalApi] = useVbenModal({
  connectedComponent: dictTypeModal,
});

const [ImportModal, importModalApi] = useVbenModal({
  connectedComponent: DictTypeImportModal,
});

const lastDictType = ref('');

const [BasicTable, tableApi] = useVbenVxeGrid({
  formOptions: {
    schema: querySchema,
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3',
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
          return await dictTypeList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
      isCurrent: true,
    },
    id: 'system-dict-type-index',
    rowClassName: 'hover:cursor-pointer',
  },
  gridEvents: {
    cellClick: (e: any) => {
      const { row } = e;
      if (lastDictType.value === row.type) {
        return;
      }
      emitter.emit('rowClick', row.type);
      lastDictType.value = row.type;
    },
    checkboxChange: () => {
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() || []) as DictType[];
    },
    checkboxAll: () => {
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() || []) as DictType[];
    },
  },
});

const checkedRows = ref<DictType[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function handleAdd() {
  modalApi.setData({});
  modalApi.open();
}

function handleEdit(record: DictType) {
  modalApi.setData({ id: record.id });
  modalApi.open();
}

async function handleDelete(row: DictType) {
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: '删除字典类型将同时删除该类型下的所有字典数据，确认删除？',
    onOk: async () => {
      try {
        await dictTypeDelete(row.id);
        message.success('删除成功');
        await tableApi.query();
        // Refresh dict data panel if the deleted type was selected
        if (lastDictType.value === row.type) {
          emitter.emit('rowClick', '');
          lastDictType.value = '';
        }
      } catch (error) {
        console.error('Delete failed:', error);
        message.error('删除失败');
      }
    },
  });
}

function handleMultiDelete() {
  const rows = tableApi.grid.getCheckboxRecords() as DictType[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `删除字典类型将同时删除该类型下的所有字典数据，确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      for (const id of ids) {
        await dictTypeDelete(id);
      }
      checkedRows.value = [];
      await tableApi.query();
      // Clear dict data panel selection
      emitter.emit('rowClick', '');
      lastDictType.value = '';
    },
  });
}

function onReload() {
  tableApi.query();
}

function onImportReload() {
  tableApi.query();
}

async function handleExport() {
  const content = checkedRows.value.length > 0
    ? '是否导出选中的字典类型及其关联的字典数据？'
    : '是否导出全部字典类型和字典数据？';

  Modal.confirm({
    title: '提示',
    okType: 'primary',
    content,
    okText: '确认',
    cancelText: '取消',
    onOk: async () => {
      try {
        const formValues = tableApi.formApi.form.values;
        const params: Record<string, any> = { ...formValues };
        if (checkedRows.value.length > 0) {
          params.ids = checkedRows.value.map((row) => row.id);
        }
        const data = await dictExport(params);
        downloadBlob(data, '字典管理导出.xlsx');
        message.success('导出成功');
      } catch {
        message.error('导出失败');
      }
    },
  });
}

function handleImport() {
  importModalApi.open();
}
</script>

<template>
  <div>
    <BasicTable id="dict-type" table-title="字典类型列表">
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
          <ghost-button danger @click.stop="handleDelete(row)">删除</ghost-button>
        </Space>
      </template>
    </BasicTable>
    <DictTypeModal @reload="onReload" />
    <ImportModal @reload="onImportReload" />
  </div>
</template>

<style lang="scss">
div#dict-type {
  .vxe-body--row {
    &.row--current {
      @apply font-semibold;
    }
  }
}
</style>
