<script setup lang="ts">
import type { FileDetail } from '#/api/system/file/model';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { Descriptions, DescriptionsItem, Spin, Tag } from 'ant-design-vue';

import { fileDetail } from '#/api/system/file';

const loading = ref(false);
const detail = ref<FileDetail | null>(null);

const [Modal, modalApi] = useVbenModal({
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) {
      detail.value = null;
      return;
    }
    const { id } = modalApi.getData() as { id: number };
    loading.value = true;
    try {
      detail.value = await fileDetail(id);
    } finally {
      loading.value = false;
    }
  },
});

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${Number.parseFloat((bytes / k ** i).toFixed(2))} ${sizes[i]}`;
}
</script>

<template>
  <Modal :footer="false" :title="$t('pages.system.file.detail.title')" class="w-[650px]">
    <Spin :spinning="loading">
      <template v-if="detail">
        <Descriptions :column="2" bordered size="middle" :label-style="{ minWidth: '120px' }" :content-style="{ minWidth: '120px' }">
          <DescriptionsItem :label="$t('pages.system.file.detail.fileId')">{{ detail.id }}</DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.detail.engine')">{{ detail.engine }}</DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.originalName')" :span="2">
            {{ detail.original }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.detail.storedName')" :span="2">
            {{ detail.name }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.fileType')">{{ detail.suffix }}</DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.size')">
            {{ formatFileSize(detail.size) }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.scene')" :span="2">
            <Tag color="blue">{{ detail.sceneLabel }}</Tag>
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.detail.url')" :span="2">
            <a :href="detail.url" target="_blank" rel="noopener noreferrer">
              {{ detail.url }}
            </a>
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.detail.hash')" :span="2">
            {{ detail.hash }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.uploader')">
            {{ detail.createdByName || '-' }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.system.file.fields.uploadedAt')">
            {{ detail.createdAt }}
          </DescriptionsItem>
        </Descriptions>
      </template>
    </Spin>
  </Modal>
</template>
