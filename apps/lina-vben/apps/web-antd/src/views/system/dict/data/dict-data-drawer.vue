<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import {
  dictDataAdd,
  dictDataInfo,
  dictDataUpdate,
} from '#/api/system/dict/dict-data';
import { tagTypes } from '#/components/dict';
import { useDictStore } from '#/store/dict';

import { drawerSchema } from './data';
import TagStylePicker from './tag-style-picker.vue';

const emit = defineEmits<{ reload: [] }>();

const dictStore = useDictStore();

interface DrawerProps {
  dictType: string;
  id?: number;
}

const isEdit = ref(false);
const editId = ref<number>(0);
const title = computed(() => (isEdit.value ? '编辑字典数据' : '新增字典数据'));

/**
 * 标签样式选择器
 * default: 预设标签样式
 * custom: 自定义标签样式
 */
const selectType = ref<'custom' | 'default'>('default');

/**
 * 根据标签样式判断是自定义还是默认
 */
function setupSelectType(tagStyle: string) {
  const isDefault = Reflect.has(tagTypes, tagStyle);
  selectType.value = isDefault ? 'default' : 'custom';
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    componentProps: {
      class: 'w-full',
    },
    formItemClass: 'col-span-2',
    labelWidth: 80,
  },
  schema: drawerSchema,
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      return;
    }
    drawerApi.setState({ loading: true });

    const { dictType, id } = drawerApi.getData() as DrawerProps;
    isEdit.value = !!id;
    editId.value = id ?? 0;
    await formApi.setFieldValue('dictType', dictType);

    if (id && isEdit.value) {
      const record = await dictDataInfo(id);
      setupSelectType(record.tagStyle ?? '');
      await formApi.setValues(record);
    }

    drawerApi.setState({ loading: false });
  },
  async onConfirm() {
    try {
      drawerApi.lock(true);
      const { valid } = await formApi.validate();
      if (!valid) {
        return;
      }
      const data = await formApi.getValues();
      // 确保 tagStyle 为空字符串而非 undefined
      if (!data.tagStyle) {
        data.tagStyle = '';
      }

      if (isEdit.value) {
        await dictDataUpdate(editId.value, data);
        message.success('更新成功');
      } else {
        await dictDataAdd(data);
        message.success('创建成功');
      }

      // 清除字典缓存，确保其他页面读取最新数据
      dictStore.resetCache();

      emit('reload');
      drawerApi.close();
    } catch (error) {
      console.error(error);
    } finally {
      drawerApi.lock(false);
    }
  },
  onClosed() {
    formApi.resetForm();
    selectType.value = 'default';
  },
});

/**
 * 取消标签选中 必须设置为undefined才行
 */
async function handleDeSelect() {
  await formApi.setFieldValue('tagStyle', undefined);
}
</script>

<template>
  <Drawer :title="title" class="w-[600px]">
    <Form>
      <template #tagStyle="slotProps">
        <TagStylePicker
          v-bind="slotProps"
          v-model:select-type="selectType"
          @deselect="handleDeSelect"
        />
      </template>
    </Form>
  </Drawer>
</template>
