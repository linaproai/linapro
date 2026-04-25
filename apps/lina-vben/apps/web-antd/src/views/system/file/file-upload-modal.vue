<script setup lang="ts">
import { ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';

import { FileUpload } from '#/components/upload';

const emit = defineEmits<{ reload: [] }>();

const fileList = ref<string[]>([]);
const [BasicModal, modalApi] = useVbenModal({
  onOpenChange: (isOpen) => {
    if (!isOpen && fileList.value.length > 0) {
      fileList.value = [];
      emit('reload');
      modalApi.close();
    }
  },
});
</script>

<template>
  <BasicModal
    :close-on-click-modal="false"
    :footer="false"
    :fullscreen-button="false"
    :title="$t('pages.system.file.actions.fileUpload')"
  >
    <div class="flex flex-col gap-4">
      <FileUpload
        v-model:value="fileList"
        :enable-drag-upload="true"
        :max-count="3"
        scene="other"
      />
    </div>
  </BasicModal>
</template>
