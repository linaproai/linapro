<script lang="ts">
export const pluginPageMeta = {
  routePath: "plugin-demo-source-sidebar-entry",
  title: "源码插件示例",
};
</script>

<script setup lang="ts">
import type { DemoRecordItem } from "./demo-record-client";

import { Popconfirm, Space } from "ant-design-vue";

import { useAccess } from "@vben/access";
import { useVbenModal } from "@vben/common-ui";

import { useVbenVxeGrid } from "#/adapter/vxe-table";
import { Page } from "#/plugins/dynamic";
import { downloadBlob } from "#/utils/download";

import DemoRecordModal from "./components/demo-record-modal.vue";
import {
  deleteDemoRecord,
  downloadDemoRecordAttachment,
  listDemoRecords,
} from "./demo-record-client";

const { hasAccessByCodes } = useAccess();

const pluginAccessCodes = {
  create: "plugin-demo-source:example:create",
  update: "plugin-demo-source:example:update",
  delete: "plugin-demo-source:example:delete",
} as const;

const [RecordModal, recordModalApi] = useVbenModal({
  connectedComponent: DemoRecordModal,
});

const [Grid, gridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "记录标题",
      },
    ],
    commonConfig: {
      labelWidth: 80,
      componentProps: {
        allowClear: true,
      },
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      { field: "title", minWidth: 180, title: "记录标题" },
      {
        field: "content",
        minWidth: 260,
        showOverflow: "ellipsis",
        title: "记录内容",
      },
      {
        field: "attachmentName",
        minWidth: 180,
        slots: { default: "attachment" },
        title: "附件",
      },
      { field: "updatedAt", title: "更新时间", width: 180 },
      {
        field: "action",
        fixed: "right",
        slots: { default: "action" },
        title: "操作",
        width: 180,
      },
    ],
    height: "auto",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues = {},
        ) => {
          return await listDemoRecords({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "id",
    },
    id: "plugin-demo-source-record-grid",
  },
});

function canCreateRecord() {
  return hasAccessByCodes([pluginAccessCodes.create]);
}

function canUpdateRecord() {
  return hasAccessByCodes([pluginAccessCodes.update]);
}

function canDeleteRecord() {
  return hasAccessByCodes([pluginAccessCodes.delete]);
}

function handleAddRecord() {
  recordModalApi.setData({});
  recordModalApi.open();
}

function handleEditRecord(row: DemoRecordItem) {
  recordModalApi.setData({ id: row.id });
  recordModalApi.open();
}

async function handleDeleteRecord(row: DemoRecordItem) {
  await deleteDemoRecord(row.id);
  await gridApi.query();
}

async function handleDownloadAttachment(row: DemoRecordItem) {
  const data = await downloadDemoRecordAttachment(row.id);
  downloadBlob(data, row.attachmentName || "plugin-demo-source-attachment");
}

function handleReload() {
  gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid table-title="示例记录">
      <template #toolbar-tools>
        <Space>
          <a-button
            v-if="canCreateRecord()"
            data-testid="plugin-demo-source-record-add"
            type="primary"
            @click="handleAddRecord"
          >
            新 增
          </a-button>
        </Space>
      </template>

      <template #attachment="{ row }">
        <a-button
          v-if="row.hasAttachment === 1"
          type="link"
          @click="handleDownloadAttachment(row)"
        >
          {{ row.attachmentName }}
        </a-button>
        <span v-else>-</span>
      </template>

      <template #action="{ row }">
        <Space>
          <ghost-button
            v-if="canUpdateRecord()"
            :data-testid="`plugin-demo-source-record-edit-${row.id}`"
            @click.stop="handleEditRecord(row)"
          >
            编辑
          </ghost-button>
          <Popconfirm
            v-if="canDeleteRecord()"
            title="确认删除该记录？"
            @confirm="handleDeleteRecord(row)"
          >
            <ghost-button
              danger
              :data-testid="`plugin-demo-source-record-delete-${row.id}`"
              @click.stop=""
            >
              删除
            </ghost-button>
          </Popconfirm>
        </Space>
      </template>
    </Grid>
    <RecordModal @reload="handleReload" />
  </Page>
</template>
