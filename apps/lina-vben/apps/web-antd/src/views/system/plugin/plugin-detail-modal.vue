<script setup lang="ts">
import type { SystemPlugin } from '#/api/system/plugin/model';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { Alert, Descriptions, DescriptionsItem, Tag } from 'ant-design-vue';

import PluginHostServiceCards from './plugin-host-service-cards.vue';
import { buildPluginDetailHostServiceCards } from './plugin-host-service-view';

const currentPlugin = ref<SystemPlugin | null>(null);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onOpenChange: handleOpenChange,
});

const hostServiceCards = computed(() => {
  return buildPluginDetailHostServiceCards(
    currentPlugin.value?.requestedHostServices,
    currentPlugin.value?.authorizedHostServices,
  );
});

const hasHostServiceDetails = computed(() => {
  return hostServiceCards.value.length > 0;
});

const showHostServiceSection = computed(() => {
  return currentPlugin.value?.type === 'dynamic';
});

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
}

function formatPluginType(type: string) {
  if (type === 'source') {
    return '源码插件';
  }
  if (type === 'dynamic') {
    return '动态插件';
  }
  return type || '-';
}

function formatInstalledStatus(installed: number) {
  return installed === 1 ? '已接入' : '未安装';
}

function getInstalledStatusColor(installed: number) {
  return installed === 1 ? 'green' : 'default';
}

function formatEnabledStatus(enabled: number) {
  return enabled === 1 ? '启用' : '禁用';
}

function getEnabledStatusColor(enabled: number) {
  return enabled === 1 ? 'green' : 'default';
}

function formatAuthorizationRequirement(required: number) {
  return required === 1 ? '需要确认' : '无需确认';
}

function getAuthorizationRequirementColor(required: number) {
  return required === 1 ? 'gold' : 'default';
}

function formatAuthorizationStatus(status: string) {
  switch (status) {
    case 'confirmed': {
      return '已确认';
    }
    case 'not_required': {
      return '无需确认';
    }
    case 'pending': {
      return '待确认';
    }
    default: {
      return status || '-';
    }
  }
}

function getAuthorizationStatusColor(status: string) {
  switch (status) {
    case 'confirmed': {
      return 'green';
    }
    case 'pending': {
      return 'gold';
    }
    default: {
      return 'default';
    }
  }
}
</script>

<template>
  <BasicModal
    :footer="false"
    title="插件详情"
    class="w-[860px] max-w-[calc(100vw-32px)]"
  >
    <div
      v-if="currentPlugin"
      data-testid="plugin-detail-modal"
      class="flex flex-col gap-4"
    >
      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem label="插件名称">
          {{ currentPlugin.name || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件标识">
          {{ currentPlugin.id || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件类型">
          <Tag color="blue">
            {{ formatPluginType(currentPlugin.type) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="插件版本">
          {{ currentPlugin.version || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="接入状态">
          <Tag :color="getInstalledStatusColor(currentPlugin.installed)">
            {{ formatInstalledStatus(currentPlugin.installed) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="当前状态">
          <Tag :color="getEnabledStatusColor(currentPlugin.enabled)">
            {{ formatEnabledStatus(currentPlugin.enabled) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="授权要求">
          <Tag
            :color="
              getAuthorizationRequirementColor(
                currentPlugin.authorizationRequired,
              )
            "
          >
            {{
              formatAuthorizationRequirement(currentPlugin.authorizationRequired)
            }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="授权状态">
          <Tag
            :color="getAuthorizationStatusColor(currentPlugin.authorizationStatus)"
          >
            {{ formatAuthorizationStatus(currentPlugin.authorizationStatus) }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem label="安装时间">
          {{ currentPlugin.installedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="更新时间">
          {{ currentPlugin.updatedAt || '-' }}
        </DescriptionsItem>
        <DescriptionsItem label="插件描述" :span="2">
          {{ currentPlugin.description || '-' }}
        </DescriptionsItem>
      </Descriptions>

      <template v-if="showHostServiceSection">
        <Alert
          v-if="!hasHostServiceDetails"
          data-testid="plugin-detail-empty-host-services"
          show-icon
          type="info"
          message="当前动态插件未声明额外宿主服务。"
        />

        <template v-else>
          <Alert
            show-icon
            type="info"
            message="申请清单表示插件当前版本声明的宿主服务范围；授权快照表示宿主管理员对当前 release 最终确认并实际生效的授权结果。"
          />

          <div class="text-[13px] font-medium text-[var(--ant-color-text)]">
            宿主服务信息
          </div>

          <PluginHostServiceCards :cards="hostServiceCards" />
        </template>
      </template>
    </div>
  </BasicModal>
</template>
