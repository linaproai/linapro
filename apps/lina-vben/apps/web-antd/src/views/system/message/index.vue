<script setup lang="ts">
import type { UserMessage } from '#/api/system/message/model';

import { computed, onMounted, ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';

import {
  Badge,
  Button,
  Card,
  List,
  ListItem,
  ListItemMeta,
  Modal,
  Pagination,
  Space,
  Tag,
} from 'ant-design-vue';

import {
  messageClear,
  messageDelete,
  messageList,
  messageRead,
  messageReadAll,
} from '#/api/system/message';
import { $t } from '#/locales';
import { useMessageStore } from '#/store/message';

import NoticePreviewModal from './notice-preview-modal.vue';

const messageStore = useMessageStore();

const messages = ref<UserMessage[]>([]);
const total = ref(0);
const pageNum = ref(1);
const pageSize = ref(10);
const loading = ref(false);

const hasMessages = computed(() => messages.value.length > 0);

const [PreviewModal, previewModalApi] = useVbenModal({
  connectedComponent: NoticePreviewModal,
});

async function fetchData() {
  loading.value = true;
  try {
    const res = await messageList({
      pageNum: pageNum.value,
      pageSize: pageSize.value,
    });
    messages.value = res.items;
    total.value = res.total;
  } finally {
    loading.value = false;
  }
}

function handlePageChange(page: number, size: number) {
  pageNum.value = page;
  pageSize.value = size;
  fetchData();
}

async function handleRead(item: UserMessage) {
  if (item.isRead === 0) {
    await messageRead(item.id);
    item.isRead = 1;
    await messageStore.fetchUnreadCount();
  }
  if (item.sourceType === 'notice' && item.sourceId) {
    previewModalApi.setData({ messageId: item.id });
    previewModalApi.open();
  }
}

async function handleDelete(item: UserMessage) {
  await messageDelete(item.id);
  await messageStore.fetchUnreadCount();
  await fetchData();
}

async function handleMarkAllRead() {
  await messageReadAll();
  messages.value.forEach((m) => (m.isRead = 1));
  await messageStore.fetchUnreadCount();
}

function handleClearAll() {
  Modal.confirm({
    title: $t('pages.common.confirmTitle'),
    content: $t('pages.system.message.messages.clearAllConfirm'),
    onOk: async () => {
      await messageClear();
      messageStore.unreadCount = 0;
      await fetchData();
    },
  });
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <Page :auto-content-height="true">
    <Card :title="$t('pages.system.message.tableTitle')">
      <template #extra>
        <Space>
          <Button :disabled="!hasMessages" @click="handleMarkAllRead">
            {{ $t('pages.system.message.actions.markAllRead') }}
          </Button>
          <Button :disabled="!hasMessages" danger @click="handleClearAll">
            {{ $t('pages.system.message.actions.clearAll') }}
          </Button>
        </Space>
      </template>

      <List
        :data-source="messages"
        :loading="loading"
        item-layout="horizontal"
      >
        <template #renderItem="{ item }">
          <ListItem class="cursor-pointer" @click="handleRead(item)">
            <ListItemMeta>
              <template #title>
                <Space>
                  <Badge
                    v-if="item.isRead === 0"
                    status="processing"
                    color="blue"
                  />
                  <span :class="{ 'font-semibold': item.isRead === 0 }">
                    {{ item.title }}
                  </span>
                  <Tag :color="item.typeColor">
                    {{ item.typeLabel }}
                  </Tag>
                </Space>
              </template>
              <template #description>
                {{ item.createdAt }}
              </template>
            </ListItemMeta>
            <template #actions>
              <Button
                danger
                size="small"
                @click.stop="handleDelete(item)"
              >
                {{ $t('pages.common.delete') }}
              </Button>
            </template>
          </ListItem>
        </template>
      </List>

      <div v-if="total > 0" class="mt-4 flex justify-end">
        <Pagination
          :current="pageNum"
          :page-size="pageSize"
          :total="total"
          show-size-changer
          show-quick-jumper
          :show-total="
            (t: number) => $t('pages.system.message.pagination.total', { total: t })
          "
          @change="handlePageChange"
        />
      </div>
    </Card>
    <PreviewModal />
  </Page>
</template>
