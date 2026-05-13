<script setup lang="ts">
import type { MediaStrategyInput } from "../media-client";

import { computed, reactive } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, message, RadioGroup } from "ant-design-vue";

import {
  createMediaStrategy,
  getMediaStrategy,
  updateMediaStrategy,
} from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface StrategyFormData extends MediaStrategyInput {
  id?: number;
}

const defaultValues: StrategyFormData = {
  enable: 1,
  global: 2,
  name: "",
  strategy: "",
};

const formData = reactive<StrategyFormData>({ ...defaultValues });
const isEdit = computed(() => !!formData.id);
const modalTitle = computed(() =>
  isEdit.value ? "编辑媒体策略" : "新增媒体策略",
);

const formRules = reactive({
  name: [{ message: "请输入策略名称", required: true }],
  strategy: [{ message: "请输入 YAML 策略内容", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<StrategyFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.id === undefined) {
    delete formData.id;
  }
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[760px]",
  fullscreenButton: true,
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = modalApi.getData();
    resetFields();
    setFormData();
    if (data?.id) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await getMediaStrategy(data.id);
        setFormData({
          id: record.id,
          enable: record.enable,
          global: record.global,
          name: record.name,
          strategy: record.strategy,
        });
      } finally {
        modalApi.setState({ confirmLoading: false });
      }
    }
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();
    const { id, ...values } = { ...formData };
    if (isEdit.value && id) {
      await updateMediaStrategy(id, values);
      message.success("媒体策略已更新");
    } else {
      await createMediaStrategy(values);
      message.success("媒体策略已创建");
    }
    emit("reload");
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal :title="modalTitle">
    <Form layout="vertical">
      <FormItem label="策略名称" v-bind="validateInfos.name">
        <Input
          data-testid="media-strategy-name"
          v-model:value="formData.name"
          allow-clear
          placeholder="例如：默认直播策略"
        />
      </FormItem>
      <div class="grid gap-4 md:grid-cols-2">
        <FormItem label="启用状态">
          <RadioGroup
            data-testid="media-strategy-enable"
            v-model:value="formData.enable"
            button-style="solid"
            option-type="button"
            :options="[
              { label: '开启', value: 1 },
              { label: '关闭', value: 2 },
            ]"
          />
        </FormItem>
        <FormItem label="全局策略">
          <RadioGroup
            data-testid="media-strategy-global"
            v-model:value="formData.global"
            button-style="solid"
            option-type="button"
            :options="[
              { label: '是', value: 1 },
              { label: '否', value: 2 },
            ]"
          />
        </FormItem>
      </div>
      <FormItem label="YAML 策略内容" v-bind="validateInfos.strategy">
        <Input.TextArea
          data-testid="media-strategy-body"
          v-model:value="formData.strategy"
          :auto-size="{ minRows: 10, maxRows: 18 }"
          placeholder="record: true&#10;stream: live"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
