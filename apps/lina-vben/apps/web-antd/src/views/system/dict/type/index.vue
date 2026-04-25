<script setup lang="ts">
import type { DictType } from '#/api/system/dict/dict-type-model';

import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { preferences } from '@vben/preferences';

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
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
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
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() ||
        []) as DictType[];
    },
    checkboxAll: () => {
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() ||
        []) as DictType[];
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
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.dict.type.messages.deleteConfirm'),
    onOk: async () => {
      try {
        await dictTypeDelete(row.id);
        message.success($t('pages.common.deleteSuccess'));
        await tableApi.query();
        // Refresh dict data panel if the deleted type was selected
        if (lastDictType.value === row.type) {
          emitter.emit('rowClick', '');
          lastDictType.value = '';
        }
      } catch (error) {
        console.error('Delete failed:', error);
        message.error($t('pages.system.dict.type.messages.deleteFailed'));
      }
    },
  });
}

function handleMultiDelete() {
  const rows = tableApi.grid.getCheckboxRecords() as DictType[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.dict.type.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
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
  const content =
    checkedRows.value.length > 0
      ? $t('pages.system.dict.type.messages.exportSelectedConfirm')
      : $t('pages.system.dict.type.messages.exportAllConfirm');

  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'primary',
    content,
    okText: $t('pages.common.confirm'),
    cancelText: $t('pages.common.cancel'),
    onOk: async () => {
      try {
        const formValues = tableApi.formApi.form.values;
        const params: Record<string, any> = { ...formValues };
        if (checkedRows.value.length > 0) {
          params.ids = checkedRows.value.map((row) => row.id);
        }
        const data = await dictExport(params);
        downloadBlob(data, $t('pages.system.dict.type.exportFileName'));
        message.success($t('pages.common.exportSuccess'));
      } catch {
        message.error($t('pages.common.exportFailed'));
      }
    },
  });
}

function handleImport() {
  importModalApi.open();
}

watch(
  () => preferences.app.locale,
  async () => {
    await tableApi.query();
    if (lastDictType.value) {
      emitter.emit('rowClick', lastDictType.value);
    }
  },
);
</script>

<template>
  <div>
    <BasicTable id="dict-type" :table-title="$t('pages.system.dict.type.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">{{ $t('pages.common.export') }}</a-button>
          <a-button @click="handleImport">{{ $t('pages.common.import') }}</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button type="primary" @click="handleAdd">{{ $t('pages.common.add') }}</a-button>
        </Space>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">{{ $t('pages.common.edit') }}</ghost-button>
          <ghost-button danger @click.stop="handleDelete(row)">
            {{ $t('pages.common.delete') }}
          </ghost-button>
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
