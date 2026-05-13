<script setup lang="ts">
import type { MediaBindingKind, MediaStrategy } from "../media-client";

import { computed, reactive, ref } from "vue";

import { useVbenModal } from "@vben/common-ui";

import { Form, FormItem, Input, message, Select } from "ant-design-vue";

import {
  listMediaStrategies,
  saveMediaDeviceBinding,
  saveMediaTenantBinding,
  saveMediaTenantDeviceBinding,
} from "../media-client";

const emit = defineEmits<{ reload: [kind: MediaBindingKind] }>();

interface BindingModalData {
  tenantId?: string;
  deviceId?: string;
  kind?: MediaBindingKind;
  strategyId?: number;
}

interface BindingFormData {
  tenantId: string;
  deviceId: string;
  kind: MediaBindingKind;
  strategyId: number;
}

const formData = reactive<BindingFormData>({
  tenantId: "",
  deviceId: "",
  kind: "device",
  strategyId: 0,
});
const strategyOptions = ref<{ label: string; value: number }[]>([]);
const editing = ref(false);

const modalTitle = computed(() => {
  const action = editing.value ? "编辑" : "新增";
  if (formData.kind === "tenant") return `${action}租户策略绑定`;
  if (formData.kind === "tenantDevice")
    return `${action}租户设备策略绑定`;
  return `${action}设备策略绑定`;
});

const needTenant = computed(
  () =>
    formData.kind === "tenant" || formData.kind === "tenantDevice",
);
const needDevice = computed(
  () => formData.kind === "device" || formData.kind === "tenantDevice",
);

const formRules = reactive({
  strategyId: [{ message: "请选择媒体策略", required: true }],
});

const { resetFields, validate, validateInfos } = Form.useForm(
  formData,
  formRules,
);

function setFormData(values: Partial<BindingFormData> = {}) {
  Object.assign(formData, {
    tenantId: "",
    deviceId: "",
    kind: "device",
    strategyId: 0,
    ...values,
  });
}

const [Modal, modalApi] = useVbenModal({
  class: "w-[620px]",
  onConfirm: handleConfirm,
  onOpenChange: async (isOpen: boolean) => {
    if (!isOpen) return;
    modalApi.setState({ confirmLoading: true });
    try {
      const data = (modalApi.getData() || {}) as BindingModalData;
      const kind = data.kind || "device";
      resetFields();
      const strategies = await listMediaStrategies({
        enable: 1,
        pageNum: 1,
        pageSize: 100,
      });
      strategyOptions.value = strategies.items.map((item: MediaStrategy) => ({
        label: `${item.name} #${item.id}`,
        value: item.id,
      }));
      editing.value = Boolean(data.tenantId || data.deviceId);
      setFormData({
        tenantId: data.tenantId || "",
        deviceId: data.deviceId || "",
        kind,
        strategyId: data.strategyId || strategyOptions.value[0]?.value || 0,
      });
    } finally {
      modalApi.setState({ confirmLoading: false });
    }
  },
});

async function handleConfirm() {
  try {
    modalApi.lock(true);
    await validate();
    const tenantId = formData.tenantId.trim();
    const deviceId = formData.deviceId.trim();
    if (needTenant.value && !tenantId) {
      message.warning("请输入租户 ID");
      return;
    }
    if (needDevice.value && !deviceId) {
      message.warning("请输入设备国标 ID");
      return;
    }

    if (formData.kind === "tenant") {
      await saveMediaTenantBinding(tenantId, formData.strategyId);
    } else if (formData.kind === "tenantDevice") {
      await saveMediaTenantDeviceBinding(
        tenantId,
        deviceId,
        formData.strategyId,
      );
    } else {
      await saveMediaDeviceBinding(deviceId, formData.strategyId);
    }

    message.success("策略绑定已保存");
    emit("reload", formData.kind);
    modalApi.close();
  } finally {
    modalApi.lock(false);
  }
}
</script>

<template>
  <Modal :title="modalTitle">
    <Form layout="vertical">
      <FormItem v-if="needTenant" label="租户 ID">
        <Input
          data-testid="media-binding-tenant-id"
          v-model:value="formData.tenantId"
          allow-clear
          :disabled="editing"
          placeholder="例如：tenant-a"
        />
      </FormItem>
      <FormItem v-if="needDevice" label="设备国标 ID">
        <Input
          data-testid="media-binding-device-id"
          v-model:value="formData.deviceId"
          allow-clear
          :disabled="editing"
          placeholder="例如：34020000001320000001"
        />
      </FormItem>
      <FormItem label="媒体策略" v-bind="validateInfos.strategyId">
        <Select
          data-testid="media-binding-strategy"
          v-model:value="formData.strategyId"
          :options="strategyOptions"
          placeholder="请选择媒体策略"
          show-search
        />
      </FormItem>
    </Form>
  </Modal>
</template>
