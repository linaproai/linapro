<script setup lang="ts">
import type { Notice } from '#/api/system/notice/model';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { message, Modal, Popconfirm, Space } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { noticeDelete, noticeList } from '#/api/system/notice';
import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

import { columns, querySchema } from './data';
import NoticeModal from './notice-modal.vue';
import NoticePreviewModal from './notice-preview-modal.vue';

const dictStore = useDictStore();
const noticeTypeDicts = ref<any[]>([]);
const noticeStatusDicts = ref<any[]>([]);

onMounted(async () => {
  [noticeTypeDicts.value, noticeStatusDicts.value] = await Promise.all([
    dictStore.getDictOptions('sys_notice_type'),
    dictStore.getDictOptions('sys_notice_status'),
  ]);
});

const [NoticeModalRef, noticeModalApi] = useVbenModal({
  connectedComponent: NoticeModal,
});

const [NoticePreviewModalRef, noticePreviewModalApi] = useVbenModal({
  connectedComponent: NoticePreviewModal,
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
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await noticeList({
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
    id: 'system-notice-index',
  },
  gridEvents: {
    checkboxChange: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as Notice[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as Notice[];
    },
  },
});

const checkedRows = ref<Notice[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);

function handleAdd() {
  noticeModalApi.setData({});
  noticeModalApi.open();
}

function handlePreview(row: Notice) {
  noticePreviewModalApi.setData({ id: row.id });
  noticePreviewModalApi.open();
}

function handleEdit(row: Notice) {
  noticeModalApi.setData({ id: row.id });
  noticeModalApi.open();
}

async function handleDelete(row: Notice) {
  await noticeDelete(String(row.id));
  message.success('删除成功');
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as Notice[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      await noticeDelete(ids.join(','));
      checkedRows.value = [];
      await gridApi.query();
    },
  });
}

function onReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="通知公告列表">
      <template #toolbar-tools>
        <Space>
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

      <template #type="{ row }">
        <DictTag :dicts="noticeTypeDicts" :value="String(row.type)" />
      </template>

      <template #status="{ row }">
        <DictTag :dicts="noticeStatusDicts" :value="String(row.status)" />
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handlePreview(row)">预览</ghost-button>
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

    <NoticeModalRef @reload="onReload" />
    <NoticePreviewModalRef />
  </Page>
</template>
