<script lang="ts">
export const pluginPageMeta = {
  routePath: '/monitor/loginlog',
  title: 'Login History',
};
</script>

<script setup lang="ts">
import type { LoginLog } from './loginlog-client';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { message, Modal, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  loginLogClean,
  loginLogDelete,
  loginLogExport,
  loginLogList,
} from './loginlog-client';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';
import { downloadBlob } from '#/utils/download';

import { buildColumns, buildQuerySchema } from './data';
import LoginlogDetailModal from './loginlog-detail-modal.vue';

const dictStore = useDictStore();

onMounted(async () => {
  const statusOptions = await dictStore.getDictOptionsAsync('sys_login_status');
  gridApi.formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: statusOptions.map((d: any) => ({
          label: d.label,
          value: d.value,
        })),
      },
    },
  ]);
});

const [DetailModalRef, detailModalApi] = useVbenModal({
  connectedComponent: LoginlogDetailModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: buildQuerySchema(),
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: 'grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4',
  },
  gridOptions: {
    checkboxConfig: {
      highlight: true,
      reserve: true,
    },
    columns: buildColumns(),
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

          const params: Record<string, any> = {
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          };

          if (params.loginTime && Array.isArray(params.loginTime)) {
            params.beginTime = params.loginTime[0];
            params.endTime = params.loginTime[1];
            delete params.loginTime;
          }

          return await loginLogList(params);
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
    id: 'monitor-loginlog-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
        []) as LoginLog[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() ||
        []) as LoginLog[];
    },
  },
});

const checkedRows = ref<LoginLog[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function handlePreview(row: LoginLog) {
  detailModalApi.setData(row);
  detailModalApi.open();
}

function handleClean() {
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('plugin.monitor-loginlog.messages.clearConfirm'),
    onOk: async () => {
      await loginLogClean();
      message.success($t('plugin.monitor-loginlog.messages.clearSuccess'));
      await gridApi.reload();
    },
  });
}

function handleDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as LoginLog[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    okType: 'danger',
    content: $t('plugin.monitor-loginlog.messages.deleteSelectedConfirm', {
      count: ids.length,
    }),
    onOk: async () => {
      await loginLogDelete(ids);
      message.success($t('pages.common.deleteSuccess'));
      await gridApi.query();
    },
  });
}

async function handleExport() {
  const content =
    checkedRows.value.length > 0
      ? $t('pages.exportConfirm.selected')
      : $t('pages.exportConfirm.all');

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

        if (params.loginTime && Array.isArray(params.loginTime)) {
          params.beginTime = params.loginTime[0];
          params.endTime = params.loginTime[1];
          delete params.loginTime;
        }

        if (checkedRows.value.length > 0) {
          params.ids = checkedRows.value.map((row) => row.id);
        }

        const data = await loginLogExport(params);
        downloadBlob(data, $t('plugin.monitor-loginlog.exportFileName'));
        message.success($t('pages.common.exportSuccess'));
      } catch {
        message.error($t('pages.common.exportFailed'));
      }
    },
  });
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid :table-title="$t('plugin.monitor-loginlog.tableTitle')">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleClean">{{ $t('pages.common.clear') }}</a-button>
          <a-button @click="handleExport">{{ $t('pages.common.export') }}</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleDelete"
          >
            {{ $t('pages.common.delete') }}
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <ghost-button @click.stop="handlePreview(row)">
          {{ $t('pages.common.detail') }}
        </ghost-button>
      </template>
    </Grid>

    <DetailModalRef />
  </Page>
</template>
