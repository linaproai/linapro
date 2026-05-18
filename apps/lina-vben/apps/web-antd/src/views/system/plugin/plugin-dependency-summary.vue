<script setup lang="ts">
import type {
  PluginDependencyAutoInstallItem,
  PluginDependencyBlocker,
  PluginDependencyCheckResult,
  PluginDependencyItem,
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
const autoInstallPlan = computed(() => props.check?.autoInstallPlan ?? []);
const autoInstalled = computed(() => props.check?.autoInstalled ?? []);
const manualInstallRequired = computed(() => {
  return props.check?.manualInstallRequired ?? [];
});
const softUnsatisfied = computed(() => props.check?.softUnsatisfied ?? []);
const reverseDependents = computed(() => props.check?.reverseDependents ?? []);
const reverseBlockers = computed(() => props.check?.reverseBlockers ?? []);
const cycle = computed(() => props.check?.cycle ?? []);

const hasInstallContent = computed(() => {
  return (
    props.loading ||
    blockers.value.length > 0 ||
    autoInstallPlan.value.length > 0 ||
    autoInstalled.value.length > 0 ||
    manualInstallRequired.value.length > 0 ||
    softUnsatisfied.value.length > 0 ||
    cycle.value.length > 0 ||
    props.check?.framework?.status === 'unsatisfied'
  );
});

const hasUninstallContent = computed(() => {
  return (
    props.loading ||
    reverseDependents.value.length > 0 ||
    reverseBlockers.value.length > 0
  );
});

const shouldRender = computed(() => {
  return props.mode === 'install'
    ? hasInstallContent.value
    : hasUninstallContent.value;
});

function formatAutoInstallItem(item: PluginDependencyAutoInstallItem) {
  const name = item.name || item.pluginId;
  return item.version ? `${name}@${item.version}` : name;
}

function formatDependencyItem(item: PluginDependencyItem) {
  const name = item.dependencyName || item.dependencyId;
  return item.requiredVersion ? `${name} ${item.requiredVersion}` : name;
}

function formatReverseDependent(item: PluginDependencyReverseDependent) {
  const name = item.name || item.pluginId;
  return item.requiredVersion ? `${name} ${item.requiredVersion}` : name;
}

function formatBlocker(blocker: PluginDependencyBlocker) {
  const label = $t(`pages.system.plugin.dependency.blocker.${blocker.code}`);
  const normalizedLabel =
    label === `pages.system.plugin.dependency.blocker.${blocker.code}`
      ? blocker.code
      : label;
  const dependency = blocker.dependencyId || blocker.pluginId || '';
  const version = blocker.requiredVersion ? ` ${blocker.requiredVersion}` : '';
  return [normalizedLabel, dependency, version].filter(Boolean).join(' ');
}
</script>

<template>
  <div
    v-if="shouldRender"
    class="flex flex-col gap-3 rounded-md border border-[var(--ant-color-border)] bg-[var(--ant-color-fill-quaternary)] p-3"
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
      v-if="props.mode === 'install' && autoInstallPlan.length > 0"
      data-testid="plugin-dependency-auto-install-plan"
      show-icon
      type="info"
      :message="$t('pages.system.plugin.dependency.autoInstallPlan')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="item in autoInstallPlan"
            :key="`${item.pluginId}-${item.requiredBy}`"
            color="blue"
          >
            {{ formatAutoInstallItem(item) }}
          </Tag>
        </div>
      </template>
    </Alert>

    <Alert
      v-if="props.mode === 'install' && autoInstalled.length > 0"
      data-testid="plugin-dependency-auto-installed"
      show-icon
      type="success"
      :message="$t('pages.system.plugin.dependency.autoInstalled')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="item in autoInstalled"
            :key="`${item.pluginId}-${item.version}`"
            color="green"
          >
            {{ formatAutoInstallItem(item) }}
          </Tag>
        </div>
      </template>
    </Alert>

    <Alert
      v-if="props.mode === 'install' && manualInstallRequired.length > 0"
      data-testid="plugin-dependency-manual-required"
      show-icon
      type="warning"
      :message="$t('pages.system.plugin.dependency.manualRequired')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="item in manualInstallRequired"
            :key="`${item.ownerId}-${item.dependencyId}`"
            color="gold"
          >
            {{ formatDependencyItem(item) }}
          </Tag>
        </div>
      </template>
    </Alert>

    <Alert
      v-if="props.mode === 'install' && softUnsatisfied.length > 0"
      data-testid="plugin-dependency-soft-unsatisfied"
      show-icon
      type="info"
      :message="$t('pages.system.plugin.dependency.softUnsatisfied')"
    >
      <template #description>
        <div class="mt-2 flex flex-wrap gap-2">
          <Tag
            v-for="item in softUnsatisfied"
            :key="`${item.ownerId}-${item.dependencyId}`"
          >
            {{ formatDependencyItem(item) }}
          </Tag>
        </div>
      </template>
    </Alert>

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
        (reverseDependents.length > 0 || reverseBlockers.length > 0)
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
            v-for="(blocker, index) in reverseBlockers"
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
