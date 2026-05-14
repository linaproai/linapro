<script setup lang="ts">
import type { MediaDeviceNodeInput, MediaNode } from "../media-client";

import { computed, reactive, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, message, Select } from "ant-design-vue";

import {
  createMediaDeviceNode,
  getMediaDeviceNode,
  listMediaNodes,
  updateMediaDeviceNode,
} from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface DeviceNodeModalData {
  deviceId?: string;
}

interface DeviceNodeFormData extends MediaDeviceNodeInput {
  oldDeviceId?: string;
}

const defaultValues: DeviceNodeFormData = {
  deviceId: "",
  nodeNum: 0,
};

const formData = reactive<DeviceNodeFormData>({ ...defaultValues });
const nodeOptions = ref<{ label: string; value: number }[]>([]);
const isEdit = computed(() => Boolean(formData.oldDeviceId));
const modalTitle = computed(() =>
  isEdit.value ? "编辑设备节点" : "新增设备节点",
);

const formRules = reactive({
  deviceId: [{ message: "请输入设备国标 ID", required: true }],
  nodeNum: [{ message: "请选择节点", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<DeviceNodeFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.oldDeviceId === undefined) {
    delete formData.oldDeviceId;
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
    const data = (modalApi.getData() || {}) as DeviceNodeModalData;
    modalApi.setState({ confirmLoading: true });
    try {
      resetFields();
      setFormData();
      await loadNodeOptions();
      if (data.deviceId) {
        const record = await getMediaDeviceNode(data.deviceId);
        setFormData({
          oldDeviceId: record.deviceId,
          deviceId: record.deviceId,
          nodeNum: record.nodeNum,
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
    const values: MediaDeviceNodeInput = {
      deviceId: formData.deviceId.trim(),
      nodeNum: Number(formData.nodeNum),
    };
    if (isEdit.value && formData.oldDeviceId) {
      await updateMediaDeviceNode(formData.oldDeviceId, values);
      message.success("设备节点已更新");
    } else {
      await createMediaDeviceNode(values);
      message.success("设备节点已创建");
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
      <FormItem label="设备国标 ID" v-bind="validateInfos.deviceId">
        <Input
          data-testid="media-device-node-device-id"
          v-model:value="formData.deviceId"
          allow-clear
          :maxlength="64"
          placeholder="例如：34020000001320000001"
        />
      </FormItem>
      <FormItem label="节点" v-bind="validateInfos.nodeNum">
        <Select
          data-testid="media-device-node-node"
          v-model:value="formData.nodeNum"
          :options="nodeOptions"
          placeholder="请选择节点"
          show-search
        />
      </FormItem>
    </Form>
  </Modal>
</template>
