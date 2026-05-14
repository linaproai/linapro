<script lang="ts">
export const pluginPageMeta = {
  routePath: "/water",
  title: "水印服务",
};
</script>

<script setup lang="ts">
import type { WaterPreviewResult, WaterTask } from "./water-client";

import { reactive, ref } from "vue";

import { Page } from "@vben/common-ui";

import {
  Button,
  Card,
  Col,
  Descriptions,
  DescriptionsItem,
  Form,
  FormItem,
  Image,
  Input,
  Row,
  Space,
  Statistic,
  Tag,
  message,
} from "ant-design-vue";

import {
  getWaterTask,
  previewWatermark,
  submitWaterSnap,
} from "./water-client";

const form = reactive({
  tenant: "tenant-a",
  deviceType: "gb",
  deviceId: "34020000001320000001",
  deviceCode: "34020000001320000001",
  channelCode: "34020000001320000001",
  deviceIdx: "1",
  callbackUrl: "",
  image: "",
});

const taskQuery = ref("");
const submitting = ref(false);
const previewing = ref(false);
const querying = ref(false);
const lastTaskId = ref("");
const previewResult = ref<WaterPreviewResult | null>(null);
const taskResult = ref<WaterTask | null>(null);

function statusColor(status?: string) {
  if (status === "success") return "success";
  if (status === "failed") return "error";
  if (status === "skipped") return "warning";
  if (status === "processing") return "processing";
  return "default";
}

async function handlePreview() {
  previewing.value = true;
  try {
    previewResult.value = await previewWatermark({
      tenant: form.tenant,
      deviceId: form.deviceId,
      deviceCode: form.deviceCode,
      channelCode: form.channelCode,
      image: form.image,
    });
    message.success("预览完成");
  } finally {
    previewing.value = false;
  }
}

async function handleSubmit() {
  submitting.value = true;
  try {
    const result = await submitWaterSnap({ ...form });
    lastTaskId.value = result.taskId;
    taskQuery.value = result.taskId;
    message.success("任务已提交");
  } finally {
    submitting.value = false;
  }
}

async function handleQueryTask() {
  if (!taskQuery.value) {
    message.warning("请输入任务ID");
    return;
  }
  querying.value = true;
  try {
    taskResult.value = await getWaterTask(taskQuery.value);
  } finally {
    querying.value = false;
  }
}
</script>

<template>
  <Page auto-content-height>
    <div class="water-page">
      <section class="water-header">
        <div>
          <h1>水印服务</h1>
          <p>基于媒体策略表解析租户、设备和全局策略，处理截图水印。</p>
        </div>
        <Space>
          <Statistic title="最近任务" :value="lastTaskId || '未提交'" />
        </Space>
      </section>

      <Row :gutter="[16, 16]">
        <Col :lg="10" :xs="24">
          <Card :bordered="false" class="water-panel" title="任务参数">
            <Form layout="vertical">
              <Row :gutter="12">
                <Col :span="12">
                  <FormItem label="媒体租户ID" required>
                    <Input v-model:value="form.tenant" />
                  </FormItem>
                </Col>
                <Col :span="12">
                  <FormItem label="设备类型" required>
                    <Input v-model:value="form.deviceType" />
                  </FormItem>
                </Col>
                <Col :span="12">
                  <FormItem label="设备ID" required>
                    <Input v-model:value="form.deviceId" />
                  </FormItem>
                </Col>
                <Col :span="12">
                  <FormItem label="通道编码">
                    <Input v-model:value="form.channelCode" />
                  </FormItem>
                </Col>
              </Row>
              <FormItem label="回调地址">
                <Input v-model:value="form.callbackUrl" allow-clear />
              </FormItem>
              <FormItem label="图片 Base64 / Data URL" required>
                <Input.TextArea
                  v-model:value="form.image"
                  :auto-size="{ minRows: 8, maxRows: 14 }"
                  placeholder="粘贴 data:image/png;base64,... 或纯 base64"
                />
              </FormItem>
              <Space>
                <Button
                  type="primary"
                  :loading="previewing"
                  @click="handlePreview"
                >
                  同步预览
                </Button>
                <Button :loading="submitting" @click="handleSubmit">
                  提交异步任务
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>

        <Col :lg="14" :xs="24">
          <Card :bordered="false" class="water-panel" title="预览结果">
            <div v-if="previewResult" class="result-layout">
              <Descriptions :column="2" size="small" bordered>
                <DescriptionsItem label="状态">
                  <Tag :color="statusColor(previewResult.status)">
                    {{ previewResult.message }}
                  </Tag>
                </DescriptionsItem>
                <DescriptionsItem label="耗时">
                  {{ previewResult.durationMs }} ms
                </DescriptionsItem>
                <DescriptionsItem label="策略来源">
                  {{ previewResult.sourceLabel }}
                </DescriptionsItem>
                <DescriptionsItem label="策略名称">
                  {{ previewResult.strategyName || "-" }}
                </DescriptionsItem>
              </Descriptions>
              <Image
                v-if="previewResult.image"
                :src="previewResult.image"
                class="preview-image"
              />
            </div>
            <div v-else class="empty-result">暂无预览结果</div>
          </Card>

          <Card :bordered="false" class="water-panel mt-4" title="任务查询">
            <Space.Compact class="task-search">
              <Input v-model:value="taskQuery" placeholder="输入任务ID" />
              <Button :loading="querying" type="primary" @click="handleQueryTask">
                查询
              </Button>
            </Space.Compact>
            <Descriptions
              v-if="taskResult"
              :column="2"
              class="mt-4"
              size="small"
              bordered
            >
              <DescriptionsItem label="任务ID">
                {{ taskResult.taskId }}
              </DescriptionsItem>
              <DescriptionsItem label="状态">
                <Tag :color="statusColor(taskResult.status)">
                  {{ taskResult.message }}
                </Tag>
              </DescriptionsItem>
              <DescriptionsItem label="租户">
                {{ taskResult.tenant }}
              </DescriptionsItem>
              <DescriptionsItem label="设备">
                {{ taskResult.deviceId }}
              </DescriptionsItem>
              <DescriptionsItem label="策略来源">
                {{ taskResult.sourceLabel }}
              </DescriptionsItem>
              <DescriptionsItem label="耗时">
                {{ taskResult.durationMs }} ms
              </DescriptionsItem>
              <DescriptionsItem v-if="taskResult.error" label="错误">
                {{ taskResult.error }}
              </DescriptionsItem>
            </Descriptions>
          </Card>
        </Col>
      </Row>
    </div>
  </Page>
</template>

<style scoped>
.water-page {
  padding: 16px;
}

.water-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.water-header h1 {
  margin: 0;
  font-size: 24px;
  font-weight: 650;
}

.water-header p {
  margin: 4px 0 0;
  color: rgb(100 116 139);
}

.water-panel {
  border-radius: 8px;
}

.result-layout {
  display: grid;
  gap: 16px;
}

.preview-image {
  max-width: 100%;
  max-height: 420px;
  object-fit: contain;
}

.empty-result {
  display: grid;
  min-height: 220px;
  place-items: center;
  color: rgb(148 163 184);
}

.task-search {
  width: 100%;
}
</style>
