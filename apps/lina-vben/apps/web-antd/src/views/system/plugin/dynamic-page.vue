<script setup lang="ts">
import type { RouteLocationNormalizedLoaded } from 'vue-router';

import { computed, onBeforeUnmount, ref, shallowRef, watch } from 'vue';
import { useRoute } from 'vue-router';

import { Result as AResult, Spin as ASpin } from 'ant-design-vue';
import { preferences } from '@vben/preferences';
import { useAccessStore } from '@vben/stores';

import { $t } from '#/locales';
import { getPluginPageByRoute } from '#/plugins/page-registry';
import {
  getRuntimeLocaleMessagesSnapshot,
  lookupRuntimeMessageString,
  runtimeI18nVersion,
} from '#/runtime/runtime-i18n';

const dynamicEmbeddedMountMode = 'embedded-mount';
const dynamicEmbeddedHostTestId = 'plugin-dynamic-embedded-host';
const dynamicEmbeddedSourceQueryKey = 'embeddedSrc';
const dynamicEmbeddedAccessModeQueryKey = 'pluginAccessMode';

type DynamicEmbeddedRouteQuery = Record<string, string>;

type DynamicEmbeddedMountContext = {
  accessToken: string;
  assetURL: string;
  baseURL: string;
  container: HTMLElement;
  locale: string;
  messages: Record<string, any>;
  query: DynamicEmbeddedRouteQuery;
  route: RouteLocationNormalizedLoaded;
  routePath: string;
  t: (key: string, fallback?: string) => string;
  title: string;
};

type DynamicEmbeddedMountInstance = {
  unmount?: (context: DynamicEmbeddedMountContext) => Promise<void> | void;
  update?: (context: DynamicEmbeddedMountContext) => Promise<void> | void;
};

type DynamicEmbeddedMountResult =
  | DynamicEmbeddedMountInstance
  | ((context: DynamicEmbeddedMountContext) => Promise<void> | void)
  | null
  | undefined;

type DynamicEmbeddedMountFunction = (
  context: DynamicEmbeddedMountContext,
) => Promise<DynamicEmbeddedMountResult> | DynamicEmbeddedMountResult;

type DynamicEmbeddedModule = {
  mount?: DynamicEmbeddedMountFunction;
  unmount?: (context: DynamicEmbeddedMountContext) => Promise<void> | void;
  update?: (context: DynamicEmbeddedMountContext) => Promise<void> | void;
};

type MountedDynamicEmbeddedModule = {
  context: DynamicEmbeddedMountContext;
  instance: null | DynamicEmbeddedMountInstance;
  module: DynamicEmbeddedModule;
};

const route = useRoute();
const currentRoutePath = computed(() => route.path.replace(/^\//, ''));
const pageEntry = computed(() => getPluginPageByRoute(currentRoutePath.value));
const dynamicEmbeddedHost = ref<HTMLElement>();
const dynamicEmbeddedLoading = ref(false);
const dynamicEmbeddedError = ref('');
const mountedDynamicEmbeddedModule =
  shallowRef<MountedDynamicEmbeddedModule | null>(null);
const accessStore = useAccessStore();

let dynamicEmbeddedMountToken = 0;

const normalizedRouteQuery = computed<DynamicEmbeddedRouteQuery>(() => {
  const mergedQuery = {
    ...((route.meta.query ?? {}) as Record<string, unknown>),
    ...(route.query as Record<string, unknown>),
  };

  const query: DynamicEmbeddedRouteQuery = {};
  for (const [key, value] of Object.entries(mergedQuery)) {
    if (Array.isArray(value)) {
      const firstValue = value.at(0);
      if (firstValue != null) {
        query[key] = String(firstValue);
      }
      continue;
    }
    if (value != null) {
      query[key] = String(value);
    }
  }
  return query;
});

const dynamicEmbeddedSource = computed(() => {
  return (
    normalizedRouteQuery.value[dynamicEmbeddedSourceQueryKey]?.trim() ?? ''
  );
});

const isDynamicEmbeddedMountMode = computed(() => {
  return (
    normalizedRouteQuery.value[dynamicEmbeddedAccessModeQueryKey] ===
      dynamicEmbeddedMountMode && !!dynamicEmbeddedSource.value
  );
});

function toAbsoluteDynamicEmbeddedAssetURL(source: string) {
  return new URL(source, window.location.origin).toString();
}

function normalizeDynamicEmbeddedMountResult(
  result: DynamicEmbeddedMountResult,
): null | DynamicEmbeddedMountInstance {
  if (!result) {
    return null;
  }
  if (typeof result === 'function') {
    return {
      unmount: result,
    };
  }
  return result;
}

function resolveDynamicEmbeddedModule(
  candidate: unknown,
): DynamicEmbeddedModule {
  const moduleCandidate = candidate as Record<string, unknown> | undefined;
  const defaultExport =
    (moduleCandidate?.default as Record<string, unknown> | undefined) ?? {};
  const defaultMount =
    typeof moduleCandidate?.default === 'function'
      ? (moduleCandidate.default as DynamicEmbeddedMountFunction)
      : (defaultExport.mount as DynamicEmbeddedMountFunction | undefined);

  return {
    mount:
      (moduleCandidate?.mount as DynamicEmbeddedMountFunction | undefined) ??
      defaultMount,
    unmount:
      (moduleCandidate?.unmount as DynamicEmbeddedModule['unmount']) ??
      (defaultExport.unmount as DynamicEmbeddedModule['unmount']),
    update:
      (moduleCandidate?.update as DynamicEmbeddedModule['update']) ??
      (defaultExport.update as DynamicEmbeddedModule['update']),
  };
}

function buildDynamicEmbeddedMountContext(
  assetURL: string,
): DynamicEmbeddedMountContext {
  const container = dynamicEmbeddedHost.value;
  if (!container) {
    throw new Error('Dynamic embedded mount container is not ready.');
  }
  const locale = preferences.app.locale;
  const messages = getRuntimeLocaleMessagesSnapshot();

  return {
    accessToken: accessStore.accessToken ?? '',
    assetURL,
    baseURL: assetURL.slice(0, assetURL.lastIndexOf('/') + 1),
    container,
    locale,
    messages,
    query: normalizedRouteQuery.value,
    route,
    routePath: currentRoutePath.value,
    t: (key: string, fallback = key) =>
      lookupRuntimeMessageString(messages, key) || fallback,
    title: String(route.meta.title ?? currentRoutePath.value),
  };
}

async function cleanupMountedDynamicEmbeddedModule() {
  const mounted = mountedDynamicEmbeddedModule.value;
  mountedDynamicEmbeddedModule.value = null;

  if (!mounted) {
    dynamicEmbeddedHost.value?.replaceChildren();
    return;
  }

  try {
    if (mounted.instance?.unmount) {
      await mounted.instance.unmount(mounted.context);
    } else if (mounted.module.unmount) {
      await mounted.module.unmount(mounted.context);
    }
  } finally {
    mounted.context.container.replaceChildren();
  }
}

async function mountDynamicEmbeddedModule() {
  const hostElement = dynamicEmbeddedHost.value;
  dynamicEmbeddedMountToken += 1;
  const currentMountToken = dynamicEmbeddedMountToken;

  await cleanupMountedDynamicEmbeddedModule();

  if (!hostElement || !isDynamicEmbeddedMountMode.value) {
    dynamicEmbeddedLoading.value = false;
    dynamicEmbeddedError.value = '';
    return;
  }

  dynamicEmbeddedLoading.value = true;
  dynamicEmbeddedError.value = '';

  try {
    const assetURL = toAbsoluteDynamicEmbeddedAssetURL(
      dynamicEmbeddedSource.value,
    );

    // Dynamic embedded modules are delivered as hosted ESM assets. The host
    // imports them lazily so the plugin can use its own frontend stack while
    // still being mounted inside the LinaPro content container.
    const importedModule = await import(/* @vite-ignore */ assetURL);
    if (currentMountToken !== dynamicEmbeddedMountToken) {
      return;
    }

    const dynamicEmbeddedModule = resolveDynamicEmbeddedModule(importedModule);
    if (!dynamicEmbeddedModule.mount) {
      throw new Error(
        'Dynamic embedded entry must export a mount(context) function.',
      );
    }

    const mountContext = buildDynamicEmbeddedMountContext(assetURL);
    const mountResult = await dynamicEmbeddedModule.mount(mountContext);
    if (currentMountToken !== dynamicEmbeddedMountToken) {
      return;
    }

    mountedDynamicEmbeddedModule.value = {
      context: mountContext,
      instance: normalizeDynamicEmbeddedMountResult(mountResult),
      module: dynamicEmbeddedModule,
    };
  } catch (error) {
    dynamicEmbeddedError.value =
      error instanceof Error
        ? error.message
        : 'Dynamic embedded plugin mount failed.';
    dynamicEmbeddedHost.value?.replaceChildren();
  } finally {
    if (currentMountToken === dynamicEmbeddedMountToken) {
      dynamicEmbeddedLoading.value = false;
    }
  }
}

async function refreshMountedDynamicEmbeddedModule() {
  const mounted = mountedDynamicEmbeddedModule.value;
  if (!mounted || !isDynamicEmbeddedMountMode.value) {
    await mountDynamicEmbeddedModule();
    return;
  }

  try {
    const assetURL = toAbsoluteDynamicEmbeddedAssetURL(
      dynamicEmbeddedSource.value,
    );
    if (assetURL !== mounted.context.assetURL) {
      await mountDynamicEmbeddedModule();
      return;
    }

    const nextContext = buildDynamicEmbeddedMountContext(assetURL);
    const updateHandler =
      mounted.instance?.update ?? mounted.module.update ?? null;
    if (!updateHandler) {
      await mountDynamicEmbeddedModule();
      return;
    }

    mounted.context = nextContext;
    await updateHandler(nextContext);
  } catch (error) {
    dynamicEmbeddedError.value =
      error instanceof Error
        ? error.message
        : 'Dynamic embedded plugin mount failed.';
    await mountDynamicEmbeddedModule();
  }
}

watch(
  [
    currentRoutePath,
    () => route.fullPath,
    dynamicEmbeddedSource,
    isDynamicEmbeddedMountMode,
    dynamicEmbeddedHost,
  ],
  async () => {
    if (pageEntry.value) {
      dynamicEmbeddedError.value = '';
      dynamicEmbeddedLoading.value = false;
      await cleanupMountedDynamicEmbeddedModule();
      return;
    }
    await mountDynamicEmbeddedModule();
  },
  { immediate: true },
);

watch(
  [
    () => route.meta.title,
    () => preferences.app.locale,
    () => runtimeI18nVersion.value,
  ],
  async () => {
    if (pageEntry.value || !isDynamicEmbeddedMountMode.value) {
      return;
    }
    await refreshMountedDynamicEmbeddedModule();
  },
);

onBeforeUnmount(() => {
  dynamicEmbeddedMountToken += 1;
  void cleanupMountedDynamicEmbeddedModule();
});
</script>

<template>
  <component :is="pageEntry.component" v-if="pageEntry" />
  <section v-else-if="isDynamicEmbeddedMountMode" class="dynamic-embedded-page">
    <div class="dynamic-embedded-page__body">
      <div
        :data-testid="dynamicEmbeddedHostTestId"
        class="dynamic-embedded-page__host"
        ref="dynamicEmbeddedHost"
      />

      <div class="dynamic-embedded-page__overlay" v-if="dynamicEmbeddedLoading">
        <a-spin size="large" />
      </div>

      <div
        class="dynamic-embedded-page__overlay"
        v-else-if="dynamicEmbeddedError"
      >
        <a-result
          status="error"
          :title="$t('page.plugin.dynamicPage.mountFailedTitle')"
          :sub-title="dynamicEmbeddedError"
        />
      </div>
    </div>
  </section>
  <a-result
    v-else
    status="404"
    :title="$t('page.plugin.dynamicPage.notFoundTitle')"
    :sub-title="$t('page.plugin.dynamicPage.notFoundSubtitle')"
  />
</template>

<style scoped>
.dynamic-embedded-page {
  height: 100%;
  min-height: 460px;
}

.dynamic-embedded-page__body {
  position: relative;
  height: 100%;
  min-height: 460px;
  border-radius: 20px;
  background: transparent;
  overflow: hidden;
}

.dynamic-embedded-page__host {
  height: 100%;
  min-height: 460px;
}

.dynamic-embedded-page__overlay {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgb(255 255 255 / 88%);
}
</style>
