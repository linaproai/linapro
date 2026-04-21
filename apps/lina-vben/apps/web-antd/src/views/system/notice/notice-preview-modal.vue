<script setup lang="ts">
import type { UserMessageDetail } from '#/api/system/message/model';
import type { Notice } from '#/api/system/notice/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { messageInfo } from '#/api/system/message';
import { noticeInfo } from '#/api/system/notice';
import { DictTag } from '#/components/dict';
import { useDictStore } from '#/store/dict';

type PreviewNotice = Pick<
  Notice,
  'createdAt' | 'createdByName' | 'content' | 'title' | 'type'
> & {
  id: number;
  sourceId?: UserMessageDetail['sourceId'];
  sourceType?: UserMessageDetail['sourceType'];
};

const notice = ref<PreviewNotice | null>(null);
const dictStore = useDictStore();
const noticeTypeDicts = ref<any[]>([]);
const title = computed(() => notice.value?.title ?? '预览通知公告');

const fallbackNoticeTypeDicts = [
  { label: '通知', value: '1' },
  { label: '公告', value: '2' },
];

async function loadNoticeTypeDicts() {
  try {
    noticeTypeDicts.value =
      await dictStore.getDictOptionsAsync('sys_notice_type');
  } catch {
    noticeTypeDicts.value = fallbackNoticeTypeDicts;
  }
}

const [Modal, modalApi] = useVbenModal({
  class: 'w-[800px]',
  fullscreenButton: true,
  footer: false,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    if (data?.id || data?.messageId) {
      modalApi.setState({ loading: true });
      try {
        await loadNoticeTypeDicts();
        if (data?.messageId) {
          notice.value = await messageInfo(data.messageId);
        } else {
          notice.value = await noticeInfo(data.id);
        }
      } finally {
        modalApi.setState({ loading: false });
      }
    }
  },
});
</script>

<template>
  <Modal :title="title">
    <div v-if="notice" class="p-2">
      <Descriptions :column="3" size="small" bordered class="mb-4">
        <DescriptionsItem label="公告类型">
          <DictTag :dicts="noticeTypeDicts" :value="String(notice.type)" />
        </DescriptionsItem>
        <DescriptionsItem label="创建人">
          {{ notice.createdByName || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="创建时间">
          {{ notice.createdAt }}
        </DescriptionsItem>
      </Descriptions>
      <div class="notice-content prose mt-6 max-w-none" v-html="notice.content" />
    </div>
  </Modal>
</template>

<style scoped>
.notice-content :deep(img) {
  max-width: 100%;
  height: auto;
}

.notice-content :deep(h1) {
  font-size: 2em;
  font-weight: bold;
  margin: 0.67em 0;
}

.notice-content :deep(h2) {
  font-size: 1.5em;
  font-weight: bold;
  margin: 0.83em 0;
}

.notice-content :deep(h3) {
  font-size: 1.17em;
  font-weight: bold;
  margin: 1em 0;
}

.notice-content :deep(ul),
.notice-content :deep(ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.notice-content :deep(ul) {
  list-style-type: disc;
}

.notice-content :deep(ol) {
  list-style-type: decimal;
}

.notice-content :deep(blockquote) {
  border-left: 3px solid #d9d9d9;
  padding-left: 1em;
  margin: 0.5em 0;
  color: #666;
}
</style>
