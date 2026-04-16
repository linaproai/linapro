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
import { useMessageStore } from '#/store/message';

import NoticePreviewModal from '../notice/notice-preview-modal.vue';

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
    previewModalApi.setData({ id: item.sourceId });
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
    title: '提示',
    content: '确认清空所有消息？清空后不可恢复。',
    onOk: async () => {
      await messageClear();
      messageStore.unreadCount = 0;
      await fetchData();
    },
  });
}

function getTypeLabel(type: number) {
  return type === 1 ? '通知' : '公告';
}

function getTypeColor(type: number) {
  return type === 1 ? 'blue' : 'green';
}

onMounted(() => {
  fetchData();
});
</script>

<template>
  <Page :auto-content-height="true">
    <Card title="消息列表">
      <template #extra>
        <Space>
          <Button :disabled="!hasMessages" @click="handleMarkAllRead">
            全部已读
          </Button>
          <Button :disabled="!hasMessages" danger @click="handleClearAll">
            清空消息
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
                  <Tag :color="getTypeColor(item.type)">
                    {{ getTypeLabel(item.type) }}
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
                删除
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
          :show-total="(t: number) => `共 ${t} 条`"
          @change="handlePageChange"
        />
      </div>
    </Card>
    <PreviewModal />
  </Page>
</template>
