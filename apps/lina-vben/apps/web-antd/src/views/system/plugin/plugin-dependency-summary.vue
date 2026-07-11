<script setup lang="ts">
import type {
  PluginDependencyBlocker,
  PluginDependencyCheckResult,
  PluginDependencyReverseDependent,
} from '#/api/system/plugin/model';

import { computed } from 'vue';

import { Alert, Tag } from 'ant-design-vue';

import { $t } from '#/locales';

interface Props {
  check?: null | PluginDependencyCheckResult;
  loading?: boolean;
  mode: 'install' | 'uninstall';
}

const props = withDefaults(defineProps<Props>(), {
  check: null,
  loading: false,
});

const blockers = computed(() => props.check?.blockers ?? []);
const framework = computed(() => props.check?.framework);
const frameworkUnsatisfied = computed(() => {
  return framework.value?.status === 'unsatisfied';
});
const reverseDependents = computed(() => props.check?.reverseDependents ?? []);
const reverseBlockers = computed(() => props.check?.reverseBlockers ?? []);
// reverse_dependency edges already appear in reverseDependents; keep only
// blockers that add information (unknown snapshot, version mismatch, etc.).
const visibleReverseBlockers = computed(() => {
  if (reverseDependents.value.length === 0) {
    return reverseBlockers.value;
  }
  return reverseBlockers.value.filter(
    (blocker) => blocker.code !== 'reverse_dependency',
  );
});
const cycle = computed(() => props.check?.cycle ?? []);

const hasInstallContent = computed(() => {
  return (
    props.loading ||
    blockers.value.length > 0 ||
    cycle.value.length > 0 ||
    frameworkUnsatisfied.value
  );
});

const hasUninstallContent = computed(() => {
  return (
    props.loading ||
    reverseDependents.value.length > 0 ||
    visibleReverseBlockers.value.length > 0
  );
});

const shouldRender = computed(() => {
  return props.mode === 'install'
    ? hasInstallContent.value
    : hasUninstallContent.value;
});

function formatReverseDependent(item: PluginDependencyReverseDependent) {
  const name = item.name || item.pluginId;
  return item.requiredVersion ? `${name} ${item.requiredVersion}` : name;
}

function formatFrameworkMismatch() {
  return $t('pages.system.plugin.dependency.frameworkUnsatisfiedDescription', {
    current: framework.value?.currentVersion || '-',
    required: framework.value?.requiredVersion || '-',
  });
}

// Reverse blockers identify the downstream consumer in pluginId; install
// blockers identify the missing/unsatisfied dependency in dependencyId.
function isReverseBlockerCode(code?: string) {
  return (
    code === 'reverse_dependency' ||
    code === 'reverse_dependency_version' ||
    code === 'dependency_snapshot_unknown'
  );
}

function formatBlocker(blocker: PluginDependencyBlocker) {
  const label = $t(`pages.system.plugin.dependency.blocker.${blocker.code}`);
  const normalizedLabel =
    label === `pages.system.plugin.dependency.blocker.${blocker.code}`
      ? blocker.code
      : label;
  const subject = isReverseBlockerCode(blocker.code)
    ? blocker.pluginId || blocker.dependencyId || ''
    : blocker.dependencyId || blocker.pluginId || '';
  const version = blocker.requiredVersion ? ` ${blocker.requiredVersion}` : '';
  return [normalizedLabel, subject, version].filter(Boolean).join(' ');
}
</script>

<template>
  <div
    v-if="shouldRender"
    class="flex flex-col gap-3 rounded-xl border border-[var(--ant-color-border)] bg-[var(--ant-color-fill-quaternary)] p-3"
    data-testid="plugin-dependency-summary"
  >
    <Alert
      v-if="loading"
      show-icon
      type="info"
      :message="$t('pages.system.plugin.dependency.loading')"
    />

    <Alert
      v-if="props.mode === 'install' && blockers.length > 0"
      data-testid="plugin-dependency-blockers"
      show-icon
      type="error"
      :message="$t('pages.system.plugin.dependency.installBlocked')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="(blocker, index) in blockers"
            :key="`${blocker.code}-${blocker.dependencyId}-${index}`"
            color="red"
          >
            {{ formatBlocker(blocker) }}
          </Tag>
        </div>
      </template>
    </Alert>

    <Alert
      v-if="props.mode === 'install' && frameworkUnsatisfied && blockers.length === 0"
      data-testid="plugin-dependency-framework-blocker"
      show-icon
      type="error"
      :message="$t('pages.system.plugin.dependency.frameworkUnsatisfied')"
      :description="formatFrameworkMismatch()"
    />

    <Alert
      v-if="props.mode === 'install' && cycle.length > 0"
      data-testid="plugin-dependency-cycle"
      show-icon
      type="error"
      :message="$t('pages.system.plugin.dependency.cycle')"
      :description="cycle.join(' -> ')"
    />

    <Alert
      v-if="
        props.mode === 'uninstall' &&
        (reverseDependents.length > 0 || visibleReverseBlockers.length > 0)
      "
      data-testid="plugin-dependency-reverse-blockers"
      show-icon
      type="error"
      :message="$t('pages.system.plugin.dependency.uninstallBlocked')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="item in reverseDependents"
            :key="`${item.pluginId}-${item.requiredVersion}`"
            color="red"
          >
            {{ formatReverseDependent(item) }}
          </Tag>
          <Tag
            v-for="(blocker, index) in visibleReverseBlockers"
            :key="`${blocker.code}-${blocker.pluginId}-${index}`"
            color="red"
          >
            {{ formatBlocker(blocker) }}
          </Tag>
        </div>
      </template>
    </Alert>
  </div>
</template>

<style scoped>
/*
 * Ant Design Alert 在带 description 时会把 message 渲染成标题字号（约 16px）。
 * 依赖阻断提示属于说明文案，与正文保持 14px / 常规字重更合适。
 */
:deep(.ant-alert-with-description .ant-alert-message) {
  font-size: 14px;
  font-weight: 400;
  line-height: 1.5715;
}
</style>
