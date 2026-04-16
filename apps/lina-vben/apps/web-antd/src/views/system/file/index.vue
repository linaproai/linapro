<script setup lang="ts">
import type { FileInfo } from '#/api/system/file/model';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import { Image, Modal, Popconfirm, Space, Spin, Switch, Tag, Tooltip } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { fileList, fileRemove, fileSuffixes, fileUsageScenes } from '#/api/system/file';
import { requestClient } from '#/api/request';

import { columns, querySchema, supportImageList } from './data';
import FileDetailModal from './file-detail-modal.vue';
import FileUploadModal from './file-upload-modal.vue';
import ImageUploadModal from './image-upload-modal.vue';

const preview = ref(true);

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
      [
        'createTime',
        ['beginTime', 'endTime'],
        ['YYYY-MM-DD 00:00:00', 'YYYY-MM-DD 23:59:59'],
      ],
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
          return await fileList({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
            ...sortParams,
          });
        },
      },
    },
    rowConfig: {
      keyField: 'id',
    },
    headerCellConfig: {
      height: 44,
    },
    cellConfig: {
      height: 65,
    },
    sortConfig: {
      remote: true,
      trigger: 'cell',
    },
    id: 'system-file-index',
  },
  gridEvents: {
    sortChange: () => gridApi.query(),
    checkboxChange: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as FileInfo[];
    },
    checkboxAll: () => {
      checkedRows.value = (gridApi.grid?.getCheckboxRecords() || []) as FileInfo[];
    },
  },
});

const checkedRows = ref<FileInfo[]>([]);
const hasChecked = computed(() => checkedRows.value.length > 0);
const sceneLabelMap = ref<Record<string, string>>({});

onMounted(async () => {
  try {
    const [scenes, suffixes] = await Promise.all([
      fileUsageScenes(),
      fileSuffixes(),
    ]);
    // Build scene label map for table column display
    for (const s of scenes) {
      sceneLabelMap.value[s.value] = s.label;
    }
    gridApi.formApi.updateSchema([
      {
        fieldName: 'scene',
        componentProps: {
          options: scenes.map((s) => ({
            label: s.label,
            value: s.value,
          })),
        },
      },
      {
        fieldName: 'suffix',
        componentProps: {
          options: suffixes.map((s) => ({
            label: s.label,
            value: s.value,
          })),
        },
      },
    ]);
  } catch {
    // ignore error if API fails
  }
});

function isImageFile(url: string) {
  if (!url) return false;
  const ext = url.split('.').pop()?.toLowerCase() || '';
  return supportImageList.includes(ext);
}

function isPdfFile(url: string) {
  if (!url) return false;
  return url.toLowerCase().endsWith('.pdf');
}

function pdfPreview(url: string) {
  window.open(url);
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`;
}

async function handleDownload(row: FileInfo) {
  try {
    const data = await requestClient.download(`/file/download/${row.id}`, {
      timeout: 30_000,
    });
    const url = window.URL.createObjectURL(data);
    const link = document.createElement('a');
    link.href = url;
    link.download = row.original;
    link.click();
    window.URL.revokeObjectURL(url);
  } catch (error) {
    console.error('Download failed:', error);
  }
}

async function handleDelete(row: FileInfo) {
  await fileRemove([row.id]);
  await gridApi.query();
}

function handleMultiDelete() {
  const rows = gridApi.grid.getCheckboxRecords() as FileInfo[];
  const ids = rows.map((row) => row.id);
  Modal.confirm({
    title: '提示',
    okType: 'danger',
    content: `确认删除选中的${ids.length}条记录吗？`,
    onOk: async () => {
      await fileRemove(ids);
      checkedRows.value = [];
      await gridApi.query();
    },
  });
}

const [FileUploadModalRef, fileUploadApi] = useVbenModal({
  connectedComponent: FileUploadModal,
});

const [ImageUploadModalRef, imageUploadApi] = useVbenModal({
  connectedComponent: ImageUploadModal,
});

const [FileDetailModalRef, fileDetailApi] = useVbenModal({
  connectedComponent: FileDetailModal,
});

function handleDetail(row: FileInfo) {
  fileDetailApi.setData({ id: row.id });
  fileDetailApi.open();
}

function onReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="文件列表">
      <template #toolbar-tools>
        <Space>
          <Tooltip title="预览图片">
            <Switch v-model:checked="preview" />
          </Tooltip>
          <a-button
            :disabled="!hasChecked"
            danger
            type="primary"
            @click="handleMultiDelete"
          >
            删 除
          </a-button>
          <a-button @click="fileUploadApi.open">文件上传</a-button>
          <a-button @click="imageUploadApi.open">图片上传</a-button>
        </Space>
      </template>

      <template #url="{ row }">
        <Image
          :key="row.id"
          v-if="preview && isImageFile(row.url)"
          :src="row.url"
          height="50px"
        >
          <template #placeholder>
            <div class="flex size-full items-center justify-center">
              <Spin />
            </div>
          </template>
        </Image>
        <span
          v-else-if="preview && isPdfFile(row.url)"
          class="cursor-pointer text-primary"
          @click.stop="pdfPreview(row.url)"
        >
          PDF 预览
        </span>
        <span v-else>
          <Tooltip :title="row.url">
            <span class="block max-w-[300px] truncate">{{ row.url }}</span>
          </Tooltip>
        </span>
      </template>

      <template #scene="{ row }">
        <Tag color="blue">{{ sceneLabelMap[row.scene] || row.scene }}</Tag>
      </template>

      <template #size="{ row }">
        {{ formatFileSize(row.size) }}
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button @click.stop="handleDetail(row)">详情</ghost-button>
          <ghost-button @click.stop="handleDownload(row)">下载</ghost-button>
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

    <FileDetailModalRef />
    <FileUploadModalRef @reload="onReload" />
    <ImageUploadModalRef @reload="onReload" />
  </Page>
</template>
