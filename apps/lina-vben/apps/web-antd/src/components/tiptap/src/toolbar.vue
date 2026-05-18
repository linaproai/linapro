<script setup lang="ts">
import { computed, ref } from 'vue';

import { IconifyIcon } from '@vben/icons';

import type { Editor } from '@tiptap/vue-3';
import { Button, Input, message, Modal, Space, Tooltip } from 'ant-design-vue';

import { uploadApi } from '#/api';
import { $t } from '#/locales';

const props = defineProps<{
  editor: Editor | undefined;
  disabled?: boolean;
  uploadHandler?: (file: File) => Promise<string>;
  /** 使用场景标识，用于记录文件用途 */
  scene?: string;
}>();

const linkUrl = ref('');
const linkModalVisible = ref(false);

const isActive = (name: string, attrs?: Record<string, any>) => {
  return props.editor?.isActive(name, attrs) ?? false;
};

function toggleBold() {
  props.editor?.chain().focus().toggleBold().run();
}
function toggleItalic() {
  props.editor?.chain().focus().toggleItalic().run();
}
function toggleUnderline() {
  props.editor?.chain().focus().toggleUnderline().run();
}
function toggleStrike() {
  props.editor?.chain().focus().toggleStrike().run();
}
function toggleHeading(level: 1 | 2 | 3) {
  props.editor?.chain().focus().toggleHeading({ level }).run();
}
function toggleBulletList() {
  props.editor?.chain().focus().toggleBulletList().run();
}
function toggleOrderedList() {
  props.editor?.chain().focus().toggleOrderedList().run();
}
function toggleBlockquote() {
  props.editor?.chain().focus().toggleBlockquote().run();
}
function toggleCodeBlock() {
  props.editor?.chain().focus().toggleCodeBlock().run();
}
function setHorizontalRule() {
  props.editor?.chain().focus().setHorizontalRule().run();
}
function undo() {
  props.editor?.chain().focus().undo().run();
}
function redo() {
  props.editor?.chain().focus().redo().run();
}

function handleImageUpload() {
  const input = document.createElement('input');
  input.type = 'file';
  input.accept = 'image/*';
  input.onchange = async () => {
    const file = input.files?.[0];
    if (!file) return;

    if (props.uploadHandler) {
      const url = await props.uploadHandler(file);
      props.editor?.chain().focus().setImage({ src: url }).run();
    } else {
      // Default: upload via file upload API with scene parameter
      try {
        const result = await uploadApi(file, { scene: props.scene || 'other' });
        props.editor?.chain().focus().setImage({ src: result.url }).run();
      } catch {
        message.error($t('pages.editor.messages.imageUploadFailed'));
      }
    }
  };
  input.click();
}

function openLinkModal() {
  const previousUrl = props.editor?.getAttributes('link').href || '';
  linkUrl.value = previousUrl;
  linkModalVisible.value = true;
}

function confirmLink() {
  if (linkUrl.value) {
    props.editor
      ?.chain()
      .focus()
      .extendMarkRange('link')
      .setLink({ href: linkUrl.value })
      .run();
  } else {
    props.editor?.chain().focus().extendMarkRange('link').unsetLink().run();
  }
  linkModalVisible.value = false;
}

const disabled = computed(() => props.disabled);
</script>

<template>
  <div class="tiptap-toolbar" v-if="!disabled">
    <Space :size="2" wrap>
      <Tooltip :title="$t('pages.editor.actions.undo')">
        <Button size="small" type="text" @click="undo">
          <template #icon><IconifyIcon icon="ant-design:undo-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.redo')">
        <Button size="small" type="text" @click="redo">
          <template #icon><IconifyIcon icon="ant-design:redo-outlined" /></template>
        </Button>
      </Tooltip>

      <span class="toolbar-divider" />

      <Tooltip :title="$t('pages.editor.actions.bold')">
        <Button
          size="small"
          :type="isActive('bold') ? 'primary' : 'text'"
          @click="toggleBold"
        >
          <template #icon><IconifyIcon icon="ant-design:bold-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.italic')">
        <Button
          size="small"
          :type="isActive('italic') ? 'primary' : 'text'"
          @click="toggleItalic"
        >
          <template #icon><IconifyIcon icon="ant-design:italic-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.underline')">
        <Button
          size="small"
          :type="isActive('underline') ? 'primary' : 'text'"
          @click="toggleUnderline"
        >
          <template #icon><IconifyIcon icon="ant-design:underline-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.strike')">
        <Button
          size="small"
          :type="isActive('strike') ? 'primary' : 'text'"
          @click="toggleStrike"
        >
          <template #icon><IconifyIcon icon="ant-design:strikethrough-outlined" /></template>
        </Button>
      </Tooltip>

      <span class="toolbar-divider" />

      <Tooltip :title="$t('pages.editor.actions.heading1')">
        <Button
          size="small"
          :type="isActive('heading', { level: 1 }) ? 'primary' : 'text'"
          @click="toggleHeading(1)"
        >
          H1
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.heading2')">
        <Button
          size="small"
          :type="isActive('heading', { level: 2 }) ? 'primary' : 'text'"
          @click="toggleHeading(2)"
        >
          H2
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.heading3')">
        <Button
          size="small"
          :type="isActive('heading', { level: 3 }) ? 'primary' : 'text'"
          @click="toggleHeading(3)"
        >
          H3
        </Button>
      </Tooltip>

      <span class="toolbar-divider" />

      <Tooltip :title="$t('pages.editor.actions.bulletList')">
        <Button
          size="small"
          :type="isActive('bulletList') ? 'primary' : 'text'"
          @click="toggleBulletList"
        >
          <template #icon><IconifyIcon icon="ant-design:unordered-list-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.orderedList')">
        <Button
          size="small"
          :type="isActive('orderedList') ? 'primary' : 'text'"
          @click="toggleOrderedList"
        >
          <template #icon><IconifyIcon icon="ant-design:ordered-list-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.blockquote')">
        <Button
          size="small"
          :type="isActive('blockquote') ? 'primary' : 'text'"
          @click="toggleBlockquote"
        >
          &ldquo;
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.codeBlock')">
        <Button
          size="small"
          :type="isActive('codeBlock') ? 'primary' : 'text'"
          @click="toggleCodeBlock"
        >
          <template #icon><IconifyIcon icon="ant-design:code-outlined" /></template>
        </Button>
      </Tooltip>

      <span class="toolbar-divider" />

      <Tooltip :title="$t('pages.editor.actions.horizontalRule')">
        <Button size="small" type="text" @click="setHorizontalRule">
          <template #icon><IconifyIcon icon="ant-design:minus-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.image')">
        <Button size="small" type="text" @click="handleImageUpload">
          <template #icon><IconifyIcon icon="ant-design:picture-outlined" /></template>
        </Button>
      </Tooltip>
      <Tooltip :title="$t('pages.editor.actions.link')">
        <Button
          size="small"
          :type="isActive('link') ? 'primary' : 'text'"
          @click="openLinkModal"
        >
          <template #icon><IconifyIcon icon="ant-design:link-outlined" /></template>
        </Button>
      </Tooltip>
    </Space>

    <Modal
      v-model:open="linkModalVisible"
      :title="$t('pages.editor.link.title')"
      :width="400"
      @ok="confirmLink"
    >
      <Input
        v-model:value="linkUrl"
        :placeholder="$t('pages.editor.link.placeholder')"
        @press-enter="confirmLink"
      />
    </Modal>
  </div>
</template>

<style scoped>
.tiptap-toolbar {
  padding: 6px 8px;
  border-bottom: 1px solid #d9d9d9;
  background: #fafafa;
}

.toolbar-divider {
  display: inline-block;
  width: 1px;
  height: 16px;
  margin: 0 4px;
  vertical-align: middle;
  background: #d9d9d9;
}
</style>
