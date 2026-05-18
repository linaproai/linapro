<script setup lang="ts">
import type { DictData } from '#/api/system/dict/dict-data-model';

import { computed, ref, watch } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { preferences } from '@vben/preferences';

import { message, Modal, Popconfirm, Space, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { dictDataDelete, dictDataList } from '#/api/system/dict/dict-data';
import { useDictStore } from '#/store/dict';

import { emitter } from '../mitt';
import { columns, querySchema } from './data';
import dictDataDrawer from './dict-data-drawer.vue';

const dictType = ref('');
const dictStore = useDictStore();

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
      checkMethod: ({ row }: { row: DictData }) => !isBuiltInRecord(row),
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
          if (!dictType.value) {
            return { items: [], total: 0 };
          }
          return await dictDataList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            dictType: dictType.value,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    id: 'system-dict-data-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() ||
        []) as DictData[];
    },
    checkboxAll: () => {
      checkedRows.value = (tableApi.grid?.getCheckboxRecords() ||
        []) as DictData[];
    },
  },
});

const checkedRows = ref<DictData[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

const [DictDataDrawer, drawerApi] = useVbenDrawer({
  connectedComponent: dictDataDrawer,
});

function handleAdd() {
  drawerApi.setData({ dictType: dictType.value });
  drawerApi.open();
}

function handleEdit(row: DictData) {
  drawerApi.setData({ dictType: dictType.value, id: row.id });
  drawerApi.open();
}

function isBuiltInRecord(row: DictData) {
  return row.isBuiltin === 1;
}

async function handleDelete(row: DictData) {
  if (isBuiltInRecord(row)) {
    return;
  }
  await dictDataDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  dictStore.resetCache();
  await tableApi.query();
}

function handleMultiDelete() {
  const rows = (tableApi.grid.getCheckboxRecords() as DictData[]).filter(
    (row) => !isBuiltInRecord(row),
  );
  const ids = rows.map((row) => row.id);
  if (ids.length === 0) {
    return;
  }
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.dict.data.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      for (const id of ids) {
        await dictDataDelete(id);
      }
      checkedRows.value = [];
      dictStore.resetCache();
      await tableApi.query();
    },
  });
}

function onReload() {
  dictStore.resetCache();
  tableApi.query();
}

emitter.on('rowClick', async (value: string) => {
  dictType.value = value;
  await tableApi.query();
});

watch(
  () => preferences.app.locale,
  async () => {
    if (!dictType.value) {
      return;
    }
    dictStore.resetCache();
    await tableApi.query();
  },
);
</script>

<template>
  <div>
    <BasicTable
      id="dict-data"
      :table-title="$t('pages.system.dict.data.tableTitle')"
    >
      <template #toolbar-tools>
        <Space>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button
            :disabled="dictType === ''"
            type="primary"
            @click="handleAdd"
          >
            {{ $t('pages.common.add') }}
          </a-button>
        </Space>
      </template>
      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleEdit(row)">{{
            $t('pages.common.edit')
          }}</ghost-button>
          <Tooltip
            v-if="isBuiltInRecord(row)"
            :title="$t('pages.common.builtinDeleteDisabled')"
          >
            <span
              class="inline-flex"
              :data-testid="`dict-data-delete-${row.id}`"
              @click.stop
            >
              <ghost-button danger disabled>
                {{ $t('pages.common.delete') }}
              </ghost-button>
            </span>
          </Tooltip>
          <Popconfirm
            v-else
            placement="left"
            :title="$t('pages.common.deleteConfirm')"
            @confirm="handleDelete(row)"
          >
            <ghost-button
              danger
              :data-testid="`dict-data-delete-${row.id}`"
              @click.stop=""
            >
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </BasicTable>
    <DictDataDrawer @reload="onReload" />
  </div>
</template>
