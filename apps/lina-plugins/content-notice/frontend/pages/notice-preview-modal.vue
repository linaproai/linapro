<script setup lang="ts">
import type { Notice } from './notice-client';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Descriptions, DescriptionsItem } from 'ant-design-vue';

import { noticeInfo } from './notice-client';
import { DictTag } from '#/components/dict';
import { $t } from '#/locales';
import { useDictStore } from '#/store/dict';

const notice = ref<Notice | null>(null);
const dictStore = useDictStore();
const noticeTypeDicts = ref<any[]>([]);
const title = computed(() =>
  notice.value
    ? notice.value.title
    : $t('plugin.content-notice.preview.title'),
);

const [Modal, modalApi] = useVbenModal({
  class: 'w-[800px]',
  fullscreenButton: true,
  footer: false,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    if (data?.id) {
      modalApi.setState({ loading: true });
      try {
        noticeTypeDicts.value =
          await dictStore.getDictOptionsAsync('sys_notice_type');
        const record = await noticeInfo(data.id);
        notice.value = record;
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
        <DescriptionsItem :label="$t('plugin.content-notice.fields.type')">
          <DictTag :dicts="noticeTypeDicts" :value="String(notice.type)" />
        </DescriptionsItem>
        <DescriptionsItem :label="$t('plugin.content-notice.fields.createdBy')">
          {{ notice.createdByName || '-' }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.createdAt')">
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
