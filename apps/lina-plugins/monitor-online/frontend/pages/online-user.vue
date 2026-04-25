<script lang="ts">
export const pluginPageMeta = {
  routePath: '/monitor/online',
  title: 'Sessions',
};
</script>

<script setup lang="ts">
import type { OnlineUser } from './online-client';

import { ref } from 'vue';

import { Page } from '@vben/common-ui';

import { Popconfirm } from 'ant-design-vue';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { $t } from '#/locales';
import { forceLogout, onlineList } from './online-client';

import { buildColumns, buildQuerySchema } from './data';

const onlineCount = ref(0);

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
    columns: buildColumns(),
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
            {{ $t('plugin.monitor-online.page.tableTitlePrefix') }}
            <span class="text-primary font-bold">{{ onlineCount }}</span>
            {{ $t('plugin.monitor-online.page.tableTitleSuffix') }}
          </div>
        </div>
      </template>
      <template #action="{ row }">
        <Popconfirm
          :title="
            $t('plugin.monitor-online.page.messages.forceLogoutConfirm', {
              username: row.username,
            })
          "
          placement="left"
          @confirm="handleForceOffline(row)"
        >
          <ghost-button danger>
            {{ $t('plugin.monitor-online.page.actions.forceLogout') }}
          </ghost-button>
        </Popconfirm>
      </template>
    </Grid>
  </Page>
</template>
