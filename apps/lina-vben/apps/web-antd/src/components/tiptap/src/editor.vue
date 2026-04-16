<script setup lang="ts">
import { onBeforeUnmount, watch } from 'vue';

import { EditorContent, useEditor } from '@tiptap/vue-3';

import { getExtensions } from './extensions';
import Toolbar from './toolbar.vue';

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    disabled?: boolean;
    height?: number | string;
    placeholder?: string;
    uploadHandler?: (file: File) => Promise<string>;
    /** 使用场景标识，用于记录文件用途 */
    scene?: string;
  }>(),
  {
    modelValue: '',
    disabled: false,
    height: 300,
    placeholder: '请输入内容...',
  },
);

const emit = defineEmits<{
  'update:modelValue': [value: string];
}>();

const editor = useEditor({
  content: props.modelValue,
  extensions: getExtensions(props.placeholder),
  editable: !props.disabled,
  onUpdate: ({ editor: e }) => {
    emit('update:modelValue', e.getHTML());
  },
});

// Sync external modelValue changes
watch(
  () => props.modelValue,
  (val) => {
    if (editor.value && val !== editor.value.getHTML()) {
      editor.value.commands.setContent(val, false);
    }
  },
);

// Sync disabled state
watch(
  () => props.disabled,
  (val) => {
    editor.value?.setEditable(!val);
  },
);

onBeforeUnmount(() => {
  editor.value?.destroy();
});

const heightStyle =
  typeof props.height === 'number' ? `${props.height}px` : props.height;
</script>

<template>
  <div class="tiptap-editor" :class="{ 'tiptap-disabled': disabled }">
    <Toolbar
      :editor="editor"
      :disabled="disabled"
      :upload-handler="uploadHandler"
      :scene="scene"
    />
    <EditorContent
      :editor="editor"
      class="tiptap-content"
      :style="{ minHeight: heightStyle }"
    />
  </div>
</template>

<style scoped>
.tiptap-editor {
  border: 1px solid #d9d9d9;
  border-radius: 6px;
  overflow: hidden;
}

.tiptap-editor:focus-within {
  border-color: #1677ff;
  box-shadow: 0 0 0 2px rgba(5, 145, 255, 0.1);
}

.tiptap-disabled {
  background: #f5f5f5;
  cursor: not-allowed;
}

.tiptap-content {
  padding: 12px 16px;
  overflow-y: auto;
}

.tiptap-content :deep(.tiptap) {
  outline: none;
  min-height: inherit;
}

.tiptap-content :deep(.tiptap p.is-editor-empty:first-child::before) {
  content: attr(data-placeholder);
  float: left;
  color: #adb5bd;
  pointer-events: none;
  height: 0;
}

.tiptap-content :deep(.tiptap h1) {
  font-size: 2em;
  font-weight: bold;
  margin: 0.67em 0;
}

.tiptap-content :deep(.tiptap h2) {
  font-size: 1.5em;
  font-weight: bold;
  margin: 0.83em 0;
}

.tiptap-content :deep(.tiptap h3) {
  font-size: 1.17em;
  font-weight: bold;
  margin: 1em 0;
}

.tiptap-content :deep(.tiptap ul),
.tiptap-content :deep(.tiptap ol) {
  padding-left: 1.5em;
  margin: 0.5em 0;
}

.tiptap-content :deep(.tiptap ul) {
  list-style-type: disc;
}

.tiptap-content :deep(.tiptap ol) {
  list-style-type: decimal;
}

.tiptap-content :deep(.tiptap blockquote) {
  border-left: 3px solid #d9d9d9;
  padding-left: 1em;
  margin: 0.5em 0;
  color: #666;
}

.tiptap-content :deep(.tiptap pre) {
  background: #f5f5f5;
  padding: 0.75em 1em;
  border-radius: 4px;
  overflow-x: auto;
}

.tiptap-content :deep(.tiptap code) {
  background: #f5f5f5;
  padding: 0.2em 0.4em;
  border-radius: 3px;
  font-size: 0.9em;
}

.tiptap-content :deep(.tiptap img) {
  max-width: 100%;
  height: auto;
}

.tiptap-content :deep(.tiptap hr) {
  border: none;
  border-top: 1px solid #d9d9d9;
  margin: 1em 0;
}

.tiptap-content :deep(.tiptap a) {
  color: #1677ff;
  text-decoration: underline;
}
</style>
