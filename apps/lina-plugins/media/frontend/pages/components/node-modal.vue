<script setup lang="ts">
import type { MediaNodeInput } from "../media-client";

import { computed, reactive } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, InputNumber, message } from "ant-design-vue";

import { createMediaNode, getMediaNode, updateMediaNode } from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface NodeModalData {
  nodeNum?: number;
}

interface NodeFormData extends MediaNodeInput {
  oldNodeNum?: number;
}

const defaultValues: NodeFormData = {
  nodeNum: 1,
  name: "",
  qnUrl: "",
  basicUrl: "",
  dnUrl: "",
};

const formData = reactive<NodeFormData>({ ...defaultValues });
const isEdit = computed(() => formData.oldNodeNum !== undefined);
const modalTitle = computed(() => (isEdit.value ? "编辑节点" : "新增节点"));

const formRules = reactive({
  nodeNum: [{ message: "请输入节点编号", required: true }],
  name: [{ message: "请输入节点名称", required: true }],
  qnUrl: [{ message: "请输入节点网关地址", required: true }],
  basicUrl: [{ message: "请输入基础平台网关地址", required: true }],
  dnUrl: [{ message: "请输入属地网关地址", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<NodeFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.oldNodeNum === undefined) {
    delete formData.oldNodeNum;
  }
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[720px]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = (modalApi.getData() || {}) as NodeModalData;
    resetFields();
    setFormData();
    if (data.nodeNum !== undefined) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await getMediaNode(data.nodeNum);
        setFormData({
          oldNodeNum: record.nodeNum,
          nodeNum: record.nodeNum,
          name: record.name,
          qnUrl: record.qnUrl,
          basicUrl: record.basicUrl,
          dnUrl: record.dnUrl,
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
    const values: MediaNodeInput = {
      nodeNum: Number(formData.nodeNum),
      name: formData.name.trim(),
      qnUrl: formData.qnUrl.trim(),
      basicUrl: formData.basicUrl.trim(),
      dnUrl: formData.dnUrl.trim(),
    };
    if (isEdit.value && formData.oldNodeNum !== undefined) {
      await updateMediaNode(formData.oldNodeNum, values);
      message.success("节点已更新");
    } else {
      await createMediaNode(values);
      message.success("节点已创建");
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
      <div class="grid gap-4 md:grid-cols-2">
        <FormItem label="节点编号" v-bind="validateInfos.nodeNum">
          <InputNumber
            data-testid="media-node-num"
            v-model:value="formData.nodeNum"
            class="w-full"
            :min="0"
            :max="255"
          />
        </FormItem>
        <FormItem label="节点名称" v-bind="validateInfos.name">
          <Input
            data-testid="media-node-name"
            v-model:value="formData.name"
            allow-clear
            :maxlength="32"
            placeholder="例如：华东节点"
          />
        </FormItem>
      </div>
      <FormItem label="节点网关地址" v-bind="validateInfos.qnUrl">
        <Input
          data-testid="media-node-qn-url"
          v-model:value="formData.qnUrl"
          allow-clear
          :maxlength="255"
          placeholder="例如：https://qn.example.com"
        />
      </FormItem>
      <FormItem label="基础平台网关地址" v-bind="validateInfos.basicUrl">
        <Input
          data-testid="media-node-basic-url"
          v-model:value="formData.basicUrl"
          allow-clear
          :maxlength="255"
          placeholder="例如：https://basic.example.com"
        />
      </FormItem>
      <FormItem label="属地网关地址" v-bind="validateInfos.dnUrl">
        <Input
          data-testid="media-node-dn-url"
          v-model:value="formData.dnUrl"
          allow-clear
          :maxlength="255"
          placeholder="例如：https://dn.example.com"
        />
      </FormItem>
    </Form>
  </Modal>
</template>
