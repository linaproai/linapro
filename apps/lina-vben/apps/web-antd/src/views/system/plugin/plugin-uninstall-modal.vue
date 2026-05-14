<script setup lang="ts">
import type {
  PluginDependencyCheckResult,
  SystemPlugin,
} from '#/api/system/plugin/model';

import { computed, ref, watch } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  Alert,
  Checkbox,
  Descriptions,
  DescriptionsItem,
  Tag,
  message,
} from 'ant-design-vue';

import { pluginDependencyCheck, pluginUninstall } from '#/api/system/plugin';
import { $t } from '#/locales';

import PluginDependencySummary from './plugin-dependency-summary.vue';

const emit = defineEmits<{
  lifecycleGuard: [
    payload: {
      force: () => Promise<void>;
      pluginId: string;
      reasons: string[];
    },
  ];
  reload: [];
}>();

const currentPlugin = ref<SystemPlugin | null>(null);
const dependencyCheck = ref<null | PluginDependencyCheckResult>(null);
const dependencyLoading = ref(false);
const purgeStorageData = ref(true);

const [BasicModal, modalApi] = useVbenModal({
  onClosed: handleClosed,
  onConfirm: handleConfirm,
  onOpenChange: handleOpenChange,
});

const isSourcePlugin = computed(() => currentPlugin.value?.type === 'source');
const isDynamicPlugin = computed(() => currentPlugin.value?.type === 'dynamic');
const isAutoEnableManaged = computed(
  () => currentPlugin.value?.autoEnableManaged === 1,
);
const supportsPurgeStorageData = computed(
  () => isSourcePlugin.value || isDynamicPlugin.value,
);
const reverseDependencyBlocked = computed(() => {
  return (
    dependencyLoading.value ||
    (dependencyCheck.value?.reverseDependents ?? []).length > 0 ||
    (dependencyCheck.value?.reverseBlockers ?? []).length > 0
  );
});

watch(reverseDependencyBlocked, updateConfirmDisabled);

async function handleOpenChange(open: boolean) {
  if (!open) {
    return;
  }
  const data = modalApi.getData<{ row: SystemPlugin }>();
  currentPlugin.value = data?.row ?? null;
  dependencyCheck.value = currentPlugin.value?.dependencyCheck ?? null;
  purgeStorageData.value = supportsPurgeStorageData.value;
  await refreshDependencyCheck();
  updateConfirmDisabled();
}

async function handleConfirm() {
  if (reverseDependencyBlocked.value) {
    message.warning(
      $t('pages.system.plugin.dependency.resolveBeforeUninstall'),
    );
    return;
  }
  await submitUninstall(false);
}

async function forceUninstall() {
  await submitUninstall(true);
}

async function forceUninstallByState(
  pluginId: string,
  purgeStorageDataValue?: boolean,
) {
  await submitUninstallByState(pluginId, purgeStorageDataValue, true);
}

async function submitUninstall(force: boolean) {
  if (!currentPlugin.value) {
    return;
  }

  const pluginId = currentPlugin.value.id;
  const purgeStorageDataValue = supportsPurgeStorageData.value
    ? purgeStorageData.value
    : undefined;

  await submitUninstallByState(pluginId, purgeStorageDataValue, force);
}

async function submitUninstallByState(
  pluginId: string,
  purgeStorageDataValue: boolean | undefined,
  force: boolean,
) {
  try {
    modalApi.lock(true);
    try {
      await pluginUninstall(pluginId, {
        force,
        purgeStorageData: purgeStorageDataValue,
      });
      message.success($t('pages.system.plugin.messages.uninstalled'));
      emit('reload');
      handleClosed();
    } catch (error) {
      if (
        !force &&
        handleLifecycleGuardVeto(error, pluginId, purgeStorageDataValue)
      ) {
        return;
      }
      message.error(resolveRuntimeErrorMessage(error));
      if (force) {
        throw error;
      }
    }
  } finally {
    modalApi.lock(false);
  }
}

function handleLifecycleGuardVeto(
  error: unknown,
  pluginId: string,
  purgeStorageDataValue: boolean | undefined,
) {
  const reasons = extractLifecycleGuardReasons(error);
  if (!reasons) {
    return false;
  }
  emit('lifecycleGuard', {
    force: () => forceUninstallByState(pluginId, purgeStorageDataValue),
    pluginId,
    reasons,
  });
  handleClosed();
  return true;
}

async function refreshDependencyCheck() {
  if (!currentPlugin.value?.id) {
    return;
  }
  dependencyLoading.value = true;
  updateConfirmDisabled();
  try {
    dependencyCheck.value = await pluginDependencyCheck(currentPlugin.value.id);
  } catch {
    message.warning($t('pages.system.plugin.dependency.checkFailed'));
  } finally {
    dependencyLoading.value = false;
    updateConfirmDisabled();
  }
}

function updateConfirmDisabled() {
  modalApi.setState({ confirmDisabled: reverseDependencyBlocked.value });
}

function extractLifecycleGuardReasons(error: unknown): null | string[] {
  const envelope = extractRuntimeErrorEnvelope(error);
  if (envelope?.errorCode !== 'PLUGIN_LIFECYCLE_GUARD_VETOED') {
    return null;
  }
  return normalizeLifecycleGuardReasons(envelope.messageParams?.reasons);
}

function extractRuntimeErrorEnvelope(error: unknown): null | {
  errorCode?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
} {
  if (!error || typeof error !== 'object') {
    return null;
  }
  // RequestClient surfaces backend errors as the bizerr envelope directly, but
  // tests and raw axios paths may still expose it under response.data.
  const response = (error as { response?: { data?: unknown } }).response;
  const envelope = (response?.data ?? error) as {
    errorCode?: string;
    message?: string;
    messageKey?: string;
    messageParams?: Record<string, unknown>;
  };
  return envelope;
}

function resolveRuntimeErrorMessage(error: unknown) {
  const envelope = extractRuntimeErrorEnvelope(error);
  if (envelope?.messageKey) {
    const localized = $t(envelope.messageKey, envelope.messageParams || {});
    if (localized && localized !== envelope.messageKey) {
      return localized;
    }
  }
  if (envelope?.message) {
    return envelope.message;
  }
  if (error instanceof Error && error.message) {
    return error.message;
  }
  return $t('ui.fallback.http.internalServerError');
}

function normalizeLifecycleGuardReasons(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value
      .map((item) => normalizeLifecycleGuardReason(String(item)))
      .filter((item) => item.length > 0);
  }
  if (typeof value !== 'string') {
    return [];
  }
  return value
    .split(';')
    .map((item) => normalizeLifecycleGuardReason(item))
    .filter((item) => item.length > 0);
}

function normalizeLifecycleGuardReason(value: string) {
  const trimmed = value.trim();
  const separatorIndex = trimmed.indexOf(':');
  if (separatorIndex < 0) {
    return trimmed;
  }
  return trimmed.slice(separatorIndex + 1).trim();
}

function handleClosed() {
  modalApi.close();
  currentPlugin.value = null;
  dependencyCheck.value = null;
  dependencyLoading.value = false;
  purgeStorageData.value = true;
  updateConfirmDisabled();
}

defineExpose({ forceUninstall });
</script>

<template>
  <BasicModal :title="$t('pages.system.plugin.uninstall.title')">
    <div
      v-if="currentPlugin"
      data-testid="plugin-uninstall-modal"
      class="flex flex-col gap-4"
    >
      <Alert
        v-if="isAutoEnableManaged"
        data-testid="plugin-auto-enable-uninstall-alert"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.messages.autoEnableUninstallAlert')"
      />
      <Alert
        v-if="isSourcePlugin"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.uninstall.sourceWarning')"
      />
      <Alert
        v-else-if="isDynamicPlugin"
        show-icon
        type="warning"
        :message="$t('pages.system.plugin.uninstall.dynamicWarning')"
      />
      <Alert
        v-else
        show-icon
        type="info"
        :message="$t('pages.system.plugin.uninstall.defaultWarning')"
      />

      <Descriptions bordered size="small" :column="2">
        <DescriptionsItem :label="$t('pages.system.plugin.fields.id')">
          {{ currentPlugin.id }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.version')">
          {{ currentPlugin.version }}
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.system.plugin.fields.type')">
          <Tag :color="isSourcePlugin ? 'blue' : 'green'">
            {{
              isSourcePlugin
                ? $t('pages.system.plugin.type.source')
                : $t('pages.system.plugin.type.dynamic')
            }}
          </Tag>
        </DescriptionsItem>
        <DescriptionsItem :label="$t('pages.common.status')">
          <Tag :color="currentPlugin.enabled === 1 ? 'green' : 'default'">
            {{
              currentPlugin.enabled === 1
                ? $t('pages.status.enabled')
                : $t('pages.status.disabled')
            }}
          </Tag>
        </DescriptionsItem>
      </Descriptions>

      <PluginDependencySummary
        :check="dependencyCheck"
        :loading="dependencyLoading"
        mode="uninstall"
      />

      <Alert
        v-if="supportsPurgeStorageData"
        data-testid="plugin-uninstall-purge-warning"
        type="error"
      >
        <template #message>
          <Checkbox
            v-model:checked="purgeStorageData"
            data-testid="plugin-uninstall-purge-checkbox"
          >
            <span class="font-semibold">
              {{ $t('pages.system.plugin.uninstall.purgeStorage') }}
            </span>
          </Checkbox>
        </template>
        <template #description>
          {{ $t('pages.system.plugin.uninstall.purgeStorageWarning') }}
        </template>
      </Alert>
    </div>
  </BasicModal>
</template>
