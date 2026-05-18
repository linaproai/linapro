<script setup lang="ts">
import type { SysConfig } from '#/api/system/config/model';

import { computed, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { message, Modal, Popconfirm, Space, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { configDelete, configExport, configList } from '#/api/system/config';
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
      checkMethod: ({ row }: { row: SysConfig }) => !isBuiltInRecord(row),
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
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
        []) as SysConfig[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
        []) as SysConfig[];
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

function isBuiltInRecord(row: SysConfig) {
  return row.isBuiltin === 1;
}

async function handleDelete(row: SysConfig) {
  if (isBuiltInRecord(row)) {
    return;
  }
  await configDelete(row.id);
  message.success($t('pages.common.deleteSuccess'));
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = (gridApi.grid.getCheckboxRecords() as SysConfig[]).filter(
    (row) => !isBuiltInRecord(row),
  );
  const ids = rows.map((row) => row.id);
  if (ids.length === 0) {
    return;
  }
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('pages.system.config.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
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
  const content =
    checkedRows.value.length > 0
      ? $t('pages.system.config.messages.exportSelectedConfirm')
      : $t('pages.system.config.messages.exportAllConfirm');

  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'primary',
    content,
    okText: $t('pages.common.confirm'),
    cancelText: $t('pages.common.cancel'),
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
        downloadBlob(data, $t('pages.system.config.exportFileName'));
        message.success($t('pages.common.exportSuccess'));
      } catch {
        message.error($t('pages.common.exportFailed'));
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
    <Grid :table-title="$t('pages.system.config.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleExport">{{
            $t('pages.common.export')
          }}</a-button>
          <a-button @click="handleImport">{{
            $t('pages.common.import')
          }}</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
          <a-button type="primary" @click="handleAdd">{{
            $t('pages.common.add')
          }}</a-button>
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
              :data-testid="`config-delete-${row.id}`"
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
              :data-testid="`config-delete-${row.id}`"
              @click.stop=""
            >
              {{ $t('pages.common.delete') }}
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>

    <ConfigModalRef @reload="onReload" />
    <ImportModalRef @reload="onReload" />
  </Page>
</template>
