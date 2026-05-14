<script setup lang="ts">
import type {
  MediaNode,
  MediaTenantStreamConfigInput,
} from "../media-client";

import { computed, reactive, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import {
  Form,
  FormItem,
  Input,
  InputNumber,
  message,
  RadioGroup,
  Select,
} from "ant-design-vue";

import {
  createMediaTenantStreamConfig,
  getMediaTenantStreamConfig,
  listMediaNodes,
  updateMediaTenantStreamConfig,
} from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface TenantStreamConfigModalData {
  tenantId?: string;
}

interface TenantStreamConfigFormData extends MediaTenantStreamConfigInput {
  oldTenantId?: string;
}

const defaultValues: TenantStreamConfigFormData = {
  tenantId: "",
  maxConcurrent: 0,
  nodeNum: 0,
  enable: 1,
};

const formData = reactive<TenantStreamConfigFormData>({ ...defaultValues });
const nodeOptions = ref<{ label: string; value: number }[]>([]);
const isEdit = computed(() => Boolean(formData.oldTenantId));
const modalTitle = computed(() =>
  isEdit.value ? "编辑租户流配置" : "新增租户流配置",
);

const formRules = reactive({
  tenantId: [{ message: "请输入租户 ID", required: true }],
  maxConcurrent: [{ message: "请输入最大并发数", required: true }],
  nodeNum: [{ message: "请选择节点", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<TenantStreamConfigFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.oldTenantId === undefined) {
    delete formData.oldTenantId;
  }
}

async function loadNodeOptions() {
  const nodes = await listMediaNodes({ pageNum: 1, pageSize: 100 });
  nodeOptions.value = nodes.items.map((item: MediaNode) => ({
    label: `${item.name} #${item.nodeNum}`,
    value: item.nodeNum,
  }));
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[620px]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = (modalApi.getData() || {}) as TenantStreamConfigModalData;
    modalApi.setState({ confirmLoading: true });
    try {
      resetFields();
      setFormData();
      await loadNodeOptions();
      if (data.tenantId) {
        const record = await getMediaTenantStreamConfig(data.tenantId);
        setFormData({
          oldTenantId: record.tenantId,
          tenantId: record.tenantId,
          maxConcurrent: record.maxConcurrent,
          nodeNum: record.nodeNum,
          enable: record.enable,
        });
      } else {
        setFormData({ nodeNum: nodeOptions.value[0]?.value ?? 0 });
      }
    } finally {
      modalApi.setState({ confirmLoading: false });
    }
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();
    const values: MediaTenantStreamConfigInput = {
      tenantId: formData.tenantId.trim(),
      maxConcurrent: Number(formData.maxConcurrent),
      nodeNum: Number(formData.nodeNum),
      enable: Number(formData.enable),
    };
    if (isEdit.value && formData.oldTenantId) {
      await updateMediaTenantStreamConfig(formData.oldTenantId, values);
      message.success("租户流配置已更新");
    } else {
      await createMediaTenantStreamConfig(values);
      message.success("租户流配置已创建");
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
      <FormItem label="租户 ID" v-bind="validateInfos.tenantId">
        <Input
          data-testid="media-tenant-stream-tenant-id"
          v-model:value="formData.tenantId"
          allow-clear
          :maxlength="64"
          placeholder="例如：tenant-a"
        />
      </FormItem>
      <div class="grid gap-4 md:grid-cols-2">
        <FormItem label="最大并发数" v-bind="validateInfos.maxConcurrent">
          <InputNumber
            data-testid="media-tenant-stream-max-concurrent"
            v-model:value="formData.maxConcurrent"
            class="w-full"
            :min="0"
          />
        </FormItem>
        <FormItem label="节点" v-bind="validateInfos.nodeNum">
          <Select
            data-testid="media-tenant-stream-node"
            v-model:value="formData.nodeNum"
            :options="nodeOptions"
            placeholder="请选择节点"
            show-search
          />
        </FormItem>
      </div>
      <FormItem label="启用状态">
        <RadioGroup
          data-testid="media-tenant-stream-enable"
          v-model:value="formData.enable"
          button-style="solid"
          option-type="button"
          :options="[
            { label: '开启', value: 1 },
            { label: '关闭', value: 0 },
          ]"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
