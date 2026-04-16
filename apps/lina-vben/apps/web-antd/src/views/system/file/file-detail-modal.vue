<script setup lang="ts">
import type { FileDetail } from '#/api/system/file/model';

import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

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
  <Modal :footer="false" title="文件详情" class="w-[650px]">
    <Spin :spinning="loading">
      <template v-if="detail">
        <Descriptions :column="2" bordered size="middle" :label-style="{ minWidth: '120px' }" :content-style="{ minWidth: '120px' }">
          <DescriptionsItem label="文件ID">{{ detail.id }}</DescriptionsItem>
          <DescriptionsItem label="存储引擎">{{ detail.engine }}</DescriptionsItem>
          <DescriptionsItem label="原始文件名" :span="2">
            {{ detail.original }}
          </DescriptionsItem>
          <DescriptionsItem label="存储文件名" :span="2">
            {{ detail.name }}
          </DescriptionsItem>
          <DescriptionsItem label="文件类型">{{ detail.suffix }}</DescriptionsItem>
          <DescriptionsItem label="文件大小">
            {{ formatFileSize(detail.size) }}
          </DescriptionsItem>
          <DescriptionsItem label="使用场景" :span="2">
            <Tag color="blue">{{ detail.sceneLabel }}</Tag>
          </DescriptionsItem>
          <DescriptionsItem label="文件URL" :span="2">
            <a :href="detail.url" target="_blank" rel="noopener noreferrer">
              {{ detail.url }}
            </a>
          </DescriptionsItem>
          <DescriptionsItem label="文件哈希" :span="2">
            {{ detail.hash }}
          </DescriptionsItem>
          <DescriptionsItem label="上传者">
            {{ detail.createdByName || '-' }}
          </DescriptionsItem>
          <DescriptionsItem label="上传时间">
            {{ detail.createdAt }}
          </DescriptionsItem>
        </Descriptions>
      </template>
    </Spin>
  </Modal>
</template>
