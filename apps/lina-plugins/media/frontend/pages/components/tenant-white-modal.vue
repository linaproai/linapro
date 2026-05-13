<script setup lang="ts">
import type { MediaTenantWhiteInput } from "../media-client";

import { computed, reactive } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, message, RadioGroup } from "ant-design-vue";

import {
  createMediaTenantWhite,
  getMediaTenantWhite,
  updateMediaTenantWhite,
} from "../media-client";

const emit = defineEmits<{ reload: [] }>();

interface TenantWhiteModalData {
  tenantId?: string;
  ip?: string;
}

interface TenantWhiteFormData extends MediaTenantWhiteInput {
  oldIp?: string;
  oldTenantId?: string;
}

const defaultValues: TenantWhiteFormData = {
  tenantId: "",
  ip: "",
  description: "",
  enable: 1,
};

const formData = reactive<TenantWhiteFormData>({ ...defaultValues });
const isEdit = computed(() => Boolean(formData.oldTenantId && formData.oldIp));
const modalTitle = computed(() =>
  isEdit.value ? "编辑租户白名单" : "新增租户白名单",
);

const formRules = reactive({
  tenantId: [{ message: "请输入租户 ID", required: true }],
  ip: [
    { message: "请输入白名单地址", required: true },
    { validator: validateIpAddress },
  ],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<TenantWhiteFormData> = {}) {
  Object.assign(formData, defaultValues, values);
  if (values.oldTenantId === undefined) {
    delete formData.oldTenantId;
  }
  if (values.oldIp === undefined) {
    delete formData.oldIp;
  }
}

function isValidIpv4(value: string) {
  const parts = value.split(".");
  if (parts.length !== 4) {
    return false;
  }
  return parts.every((part) => {
    if (!/^\d+$/.test(part)) {
      return false;
    }
    if (part.length > 1 && part.startsWith("0")) {
      return false;
    }
    const numberValue = Number(part);
    return numberValue >= 0 && numberValue <= 255;
  });
}

function isValidIpv6(value: string) {
  try {
    const url = new URL(`http://[${value}]/`);
    return url.hostname.length > 2;
  } catch {
    return false;
  }
}

function isValidIpAddress(value: string) {
  const normalized = value.trim();
  if (!normalized) {
    return false;
  }
  if (isValidIpv4(normalized)) {
    return true;
  }
  return normalized.includes(":") && isValidIpv6(normalized);
}

async function validateIpAddress(_rule: unknown, value?: string) {
  if (!value || isValidIpAddress(value)) {
    return Promise.resolve();
  }
  return Promise.reject(new Error("白名单地址必须是有效的 IPv4 或 IPv6 地址"));
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[620px]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    const data = (modalApi.getData() || {}) as TenantWhiteModalData;
    resetFields();
    setFormData();
    if (data.tenantId && data.ip) {
      modalApi.setState({ confirmLoading: true });
      try {
        const record = await getMediaTenantWhite(data.tenantId, data.ip);
        setFormData({
          oldTenantId: record.tenantId,
          oldIp: record.ip,
          tenantId: record.tenantId,
          ip: record.ip,
          description: record.description || "",
          enable: record.enable,
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
    const values: MediaTenantWhiteInput = {
      tenantId: formData.tenantId.trim(),
      ip: formData.ip.trim(),
      description: formData.description.trim(),
      enable: formData.enable,
    };
    if (isEdit.value && formData.oldTenantId && formData.oldIp) {
      await updateMediaTenantWhite(formData.oldTenantId, formData.oldIp, values);
      message.success("租户白名单已更新");
    } else {
      await createMediaTenantWhite(values);
      message.success("租户白名单已创建");
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
          data-testid="media-tenant-white-tenant-id"
          v-model:value="formData.tenantId"
          allow-clear
          :maxlength="64"
          placeholder="例如：tenant-a"
        />
      </FormItem>
      <FormItem label="白名单地址" v-bind="validateInfos.ip">
        <Input
          data-testid="media-tenant-white-ip"
          v-model:value="formData.ip"
          allow-clear
          :maxlength="32"
          placeholder="例如：192.168.1.10"
        />
      </FormItem>
      <FormItem label="白名单描述">
        <Input
          data-testid="media-tenant-white-description"
          v-model:value="formData.description"
          allow-clear
          :maxlength="32"
          placeholder="例如：总部出口"
        />
      </FormItem>
      <FormItem label="启用状态">
        <RadioGroup
          data-testid="media-tenant-white-enable"
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
