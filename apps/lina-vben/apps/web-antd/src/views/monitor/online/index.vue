<script setup lang="ts">
import type { OnlineUser } from '#/api/monitor/online/model';

import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Popconfirm } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { forceLogout, onlineList } from '#/api/monitor/online';

import { columns, querySchema } from './data';

const onlineCount = ref(0);

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
    columns,
    height: 'auto',
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async ({ page }: any, formValues: Record<string, any> = {}) => {
          const { currentPage, pageSize } = page;
          const resp = await onlineList({
            ...formValues,
            pageNum: currentPage,
            pageSize,
          });
          onlineCount.value = resp.total;
          return { items: resp.items, total: resp.total };
        },
      },
    },
    scrollY: {
      enabled: true,
      gt: 0,
    },
    rowConfig: {
      keyField: 'tokenId',
    },
    id: 'monitor-online-index',
  },
});

async function handleForceOffline(row: OnlineUser) {
  await forceLogout(row.tokenId);
  await gridApi.query();
}
</script>

<template>
  <Page :auto-content-height="true">
    <Grid>
      <template #toolbar-actions>
        <div class="mr-1 pl-1 text-[1rem]">
          <div>
            在线用户列表 (共
            <span class="text-primary font-bold">{{ onlineCount }}</span>
            人在线)
          </div>
        </div>
      </template>
      <template #action="{ row }">
        <Popconfirm
          :title="`确认强制下线[${row.username}]?`"
          placement="left"
          @confirm="handleForceOffline(row)"
        >
          <ghost-button danger>强制下线</ghost-button>
        </Popconfirm>
      </template>
    </Grid>
  </Page>
</template>
