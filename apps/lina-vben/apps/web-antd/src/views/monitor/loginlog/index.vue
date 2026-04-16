<script setup lang="ts">
import type { LoginLog } from '#/api/monitor/loginlog/model';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { message, Modal, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import {
  loginLogClean,
  loginLogDelete,
  loginLogExport,
  loginLogList,
} from '#/api/monitor/loginlog';
import { downloadBlob } from '#/utils/download';
import { useDictStore } from '#/store/dict';

import { columns, querySchema } from './data';
import LoginlogDetailModal from './loginlog-detail-modal.vue';

const dictStore = useDictStore();

onMounted(async () => {
  // Use the async helper so the select is populated after the dictionary API resolves.
  const statusOptions = await dictStore.getDictOptionsAsync('sys_oper_status');
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
    schema: querySchema,
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
    columns,
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
        query: async ({ page, sorts }: any, formValues: Record<string, any> = {}) => {
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

          // Handle loginTime date range
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
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as LoginLog[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as LoginLog[];
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
    title: '提示',
    okType: 'danger',
    content: '确认要清空所有登录日志数据吗？',
    onOk: async () => {
      await loginLogClean();
      message.success('清空成功');
      await gridApi.reload();
    },
  });
}

function handleDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as LoginLog[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条登录日志吗？`,
    onOk: async () => {
      await loginLogDelete(ids);
      message.success('删除成功');
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

        if (params.loginTime && Array.isArray(params.loginTime)) {
          params.beginTime = params.loginTime[0];
          params.endTime = params.loginTime[1];
          delete params.loginTime;
        }

        if (checkedRows.value.length > 0) {
          params.ids = checkedRows.value.map((row) => row.id);
        }

        const data = await loginLogExport(params);
        downloadBlob(data, '登录日志导出.xlsx');
        message.success('导出成功');
      } catch {
        message.error('导出失败');
      }
    },
  });
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="登录日志列表">
      <template #toolbar-tools>
        <Space>
          <a-button @click="handleClean">清 空</a-button>
          <a-button @click="handleExport">导 出</a-button>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleDelete"
          >
            删 除
          </a-button>
        </Space>
      </template>

      <template #action="{ row }">
        <ghost-button @click.stop="handlePreview(row)">详情</ghost-button>
      </template>
    </Grid>

    <DetailModalRef />
  </Page>
</template>
