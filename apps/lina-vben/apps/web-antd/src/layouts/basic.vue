<script lang="ts" setup>
import type { NotificationItem } from '@vben/layouts';
import type { PendingPluginPageRefresh } from '#/plugins/plugin-page-refresh';

import { computed, h, nextTick, onBeforeUnmount, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';

import { AuthenticationLoginExpiredModal, useVbenModal } from '@vben/common-ui';
import { useWatermark } from '@vben/hooks';
import {
  BasicLayout,
  LockScreen,
  Notification,
  UserDropdown,
} from '@vben/layouts';
import { preferences } from '@vben/preferences';
import {
  getTabKey,
  useAccessStore,
  useTabbarStore,
  useUserStore,
} from '@vben/stores';

import { Button, Modal, notification } from 'ant-design-vue';

import PluginSlotOutlet from '#/components/plugin/plugin-slot-outlet.vue';
import { $t } from '#/locales';
import {
  clearPendingPluginPageRefresh,
  detectPendingPluginPageRefresh,
  getPendingPluginPageRefresh,
  rememberPluginPageGeneration,
  resolvePluginPageId,
} from '#/plugins/plugin-page-refresh';
import { pluginSlotKeys } from '#/plugins/plugin-slots';
import {
  getPluginStateMap,
  notifyPluginRegistryChangedIfNeeded,
  onPluginRegistryChanged,
} from '#/plugins/slot-registry';
import { refreshAccessibleState } from '#/router/access-refresh';
import { useAuthStore } from '#/store';
import { useMessageStore } from '#/store/message';
import LoginForm from '#/views/_core/authentication/login.vue';
import NoticePreviewModal from '#/views/system/notice/notice-preview-modal.vue';

const router = useRouter();
const userStore = useUserStore();
const authStore = useAuthStore();
const accessStore = useAccessStore();
const tabbarStore = useTabbarStore();
const messageStore = useMessageStore();
const { destroyWatermark, updateWatermark } = useWatermark();

const [PreviewModal, previewModalApi] = useVbenModal({
  connectedComponent: NoticePreviewModal,
});

const pluginPageRefreshNotificationKey = 'plugin-page-refresh';

let disposePluginRegistryListener: (() => void) | null = null;

// Map server messages to NotificationItem format
const notifications = computed<NotificationItem[]>(() =>
  messageStore.messages.map((msg) => ({
    id: msg.id,
    avatar: '',
    date: msg.createdAt,
    isRead: msg.isRead === 1,
    message: msg.title,
    title: msg.type === 1 ? '通知' : '公告',
    sourceType: msg.sourceType,
    sourceId: msg.sourceId,
  })),
);

const showDot = computed(() => messageStore.unreadCount > 0);

// Start polling on mount
onMounted(() => {
  messageStore.startPolling();
});

const menus = computed(() => [
  {
    handler: () => {
      router.push({ name: 'Profile' });
    },
    icon: 'lucide:user',
    text: $t('page.auth.profile'),
  },
]);

const avatar = computed(() => {
  return userStore.userInfo?.avatar || preferences.app.defaultAvatar;
});

async function handleLogout() {
  messageStore.stopPolling();
  await authStore.logout(false);
}

async function handleNoticeClear() {
  Modal.confirm({
    title: '提示',
    content: '确认清空所有消息通知？',
    onOk: async () => {
      await messageStore.clearAll();
    },
  });
}

async function handleRead(item: NotificationItem) {
  if (item.id) {
    await messageStore.markRead(item.id as number);
  }
}

async function handleRemove(item: NotificationItem) {
  if (item.id) {
    await messageStore.removeMessage(item.id as number);
  }
}

async function handleMakeAll() {
  await messageStore.markAllRead();
}

function handleViewAll() {
  router.push('/system/message');
}

function handleNotificationClick(item: NotificationItem) {
  const msg = messageStore.messages.find((m) => m.id === item.id);
  if (msg?.sourceType === 'notice' && msg?.sourceId) {
    previewModalApi.setData({ id: msg.sourceId });
    previewModalApi.open();
  }
}

async function refreshPluginAwareAccess(options?: {
  skipRouteNavigation?: boolean;
}) {
  await refreshAccessibleState(router, {
    showLoadingToast: false,
    skipRouteNavigation: options?.skipRouteNavigation,
  });
}

function buildCurrentRouteTab() {
  const currentRoute = router.currentRoute.value;
  const currentRouteKey = getTabKey(currentRoute);
  const currentTab =
    tabbarStore.getTabs.find((tab) => {
      return getTabKey(tab as never) === currentRouteKey;
    }) ?? null;
  const matchedMeta =
    currentRoute.matched[currentRoute.matched.length - 1]?.meta ?? null;
  return {
    ...(currentTab ?? currentRoute),
    ...currentRoute,
    meta: matchedMeta
      ? {
          ...(currentTab?.meta ?? {}),
          ...currentRoute.meta,
          ...matchedMeta,
        }
      : {
          ...(currentTab?.meta ?? {}),
          ...currentRoute.meta,
        },
  };
}

function replacePluginAssetVersion(
  value: unknown,
  pluginId: string,
  version: string,
) {
  if (typeof value !== 'string' || !value || !pluginId || !version) {
    return value;
  }

  const match = value.match(/(\/plugin-assets\/([^/]+)\/)([^/]+)(\/.*)/);
  if (!match?.[1] || !match[2] || !match[4] || match[2] !== pluginId) {
    return value;
  }
  return `${match[1]}${version}${match[4]}`;
}

function buildPendingPluginRefreshTab(
  pending: null | PendingPluginPageRefresh,
) {
  if (!pending) {
    return null;
  }

  const currentTab = buildCurrentRouteTab();
  const nextMeta = { ...currentTab.meta };
  let changed = false;

  const nextIframeSrc = replacePluginAssetVersion(
    nextMeta.iframeSrc,
    pending.pluginId,
    pending.version,
  );
  if (nextIframeSrc !== nextMeta.iframeSrc) {
    nextMeta.iframeSrc = nextIframeSrc as string;
    changed = true;
  } else {
    const visibleIframe = findPluginIframeElement(pending.pluginId);
    const visibleIframeSource =
      visibleIframe?.getAttribute('src') || visibleIframe?.src;
    const nextVisibleIframeSource = replacePluginAssetVersion(
      visibleIframeSource,
      pending.pluginId,
      pending.version,
    );
    if (
      typeof nextVisibleIframeSource === 'string' &&
      nextVisibleIframeSource !== visibleIframeSource
    ) {
      nextMeta.iframeSrc = nextVisibleIframeSource;
      changed = true;
    }
  }

  const nextLink = replacePluginAssetVersion(
    nextMeta.link,
    pending.pluginId,
    pending.version,
  );
  if (nextLink !== nextMeta.link) {
    nextMeta.link = nextLink as string;
    changed = true;
  }

  const nextQuery = {
    ...((nextMeta.query ?? {}) as Record<string, unknown>),
  };
  const nextEmbeddedSource = replacePluginAssetVersion(
    nextQuery.embeddedSrc,
    pending.pluginId,
    pending.version,
  );
  if (nextEmbeddedSource !== nextQuery.embeddedSrc) {
    nextQuery.embeddedSrc = nextEmbeddedSource;
    nextMeta.query = nextQuery;
    changed = true;
  }

  if (!changed) {
    return null;
  }
  return {
    ...currentTab,
    meta: nextMeta,
  };
}

function syncCurrentRouteTabState(tab = buildCurrentRouteTab()) {
  // Persist the latest regenerated route meta into the tabbar store before the
  // view is remounted. IFrame tabs render from tab definitions, not directly
  // from the current route snapshot.
  tabbarStore.addTab(tab);
}

async function remountCurrentPluginRoute(tab?: ReturnType<typeof buildCurrentRouteTab>) {
  syncCurrentRouteTabState(tab);
  await nextTick();
  await tabbarStore.refresh(router);
}

async function forceReplaceCurrentRoute(path: string) {
  const currentRoute = router.currentRoute.value;
  await router.replace({
    force: true,
    hash: currentRoute.hash,
    path,
    query: currentRoute.query,
  });
}

function resolvePluginIdFromAssetURL(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '';
  }
  return value.match(/\/plugin-assets\/([^/]+)\/[^/]+\//)?.[1] ?? '';
}

function findPluginIframeElement(pluginId = '') {
  if (typeof document === 'undefined') {
    return null;
  }

  const selectors = pluginId
    ? [
        `iframe[src*="/plugin-assets/${pluginId}/"]`,
        `iframe[src*="${pluginId}/"]`,
      ]
    : ['iframe[src*="/plugin-assets/"]', 'iframe'];

  for (const selector of selectors) {
    const iframe = document.querySelector(selector);
    if (iframe instanceof HTMLIFrameElement) {
      return iframe;
    }
  }
  return null;
}

function updateVisiblePluginIframeSource(
  pending: null | PendingPluginPageRefresh,
) {
  if (!pending) {
    return false;
  }

  const iframe = findPluginIframeElement(pending.pluginId);
  if (!iframe) {
    return false;
  }

  const currentSource = iframe.getAttribute('src') || iframe.src;
  const nextSource = replacePluginAssetVersion(
    currentSource,
    pending.pluginId,
    pending.version,
  );
  if (typeof nextSource !== 'string' || nextSource === currentSource) {
    return false;
  }

  // Update the live iframe eagerly so the current tab switches generations even
  // when the router keeps the same path and only the hosted asset URL changes.
  iframe.src = nextSource;
  return true;
}

function findPluginRefreshPath(pluginId: string, version = '') {
  if (!pluginId) {
    return '';
  }

  const normalizedVersionToken = version.replaceAll('.', '-');
  const routes = router.getRoutes();
  let fallbackPath = '';
  for (const route of routes) {
    if (route.redirect) {
      continue;
    }
    const normalizedPath = route.path.replace(/^\//, '');
    if (
      normalizedPath.startsWith(`${pluginId}-`) ||
      normalizedPath.startsWith(`plugin-assets/${pluginId}/`) ||
      normalizedPath.startsWith(`plugins/${pluginId}/`)
    ) {
      if (!fallbackPath) {
        fallbackPath = route.path;
      }
      if (
        !version ||
        normalizedPath.includes(version) ||
        normalizedPath.includes(normalizedVersionToken)
      ) {
        return route.path;
      }
    }

    const authority = route.meta?.authority as string | string[] | undefined;
    if (
      typeof authority === 'string' &&
      (authority === pluginId || authority.startsWith(`${pluginId}:`))
    ) {
      if (!fallbackPath) {
        fallbackPath = route.path;
      }
    }

    const dynamicSources = [
      route.meta?.iframeSrc,
      route.meta?.link,
      (route.meta?.query as Record<string, unknown> | undefined)?.embeddedSrc,
    ];
    if (dynamicSources.some((item) => resolvePluginIdFromAssetURL(item) === pluginId)) {
      if (!fallbackPath) {
        fallbackPath = route.path;
      }
      if (
        !version ||
        dynamicSources.some(
          (item) => typeof item === 'string' && item.includes(`/${version}/`),
        )
      ) {
        return route.path;
      }
    }
  }
  return fallbackPath;
}

async function handlePluginPageRefreshNow() {
  const pending = getPendingPluginPageRefresh(router.currentRoute.value);
  const pendingRefreshTab = buildPendingPluginRefreshTab(pending);
  updateVisiblePluginIframeSource(pending);
  clearPendingPluginPageRefresh();
  notification.close(pluginPageRefreshNotificationKey);

  await refreshPluginAwareAccess({ skipRouteNavigation: true });
  if (pendingRefreshTab?.meta?.iframeSrc) {
    const refreshPath = findPluginRefreshPath(
      pending?.pluginId ?? '',
      pending?.version ?? '',
    );
    if (refreshPath) {
      await forceReplaceCurrentRoute(refreshPath);
    }
    await remountCurrentPluginRoute(pendingRefreshTab);
    await syncPluginPageGenerationBaseline();
    return;
  }

  const refreshPath = findPluginRefreshPath(
    pending?.pluginId ?? '',
    pending?.version ?? '',
  );
  if (refreshPath) {
    await forceReplaceCurrentRoute(refreshPath);
    await remountCurrentPluginRoute(pendingRefreshTab ?? undefined);
    await syncPluginPageGenerationBaseline();
    return;
  }

  if (pendingRefreshTab) {
    await remountCurrentPluginRoute(pendingRefreshTab);
    await syncPluginPageGenerationBaseline();
    return;
  }

  const fallbackPath =
    userStore.userInfo?.homePath || preferences.app.defaultHomePath || '/';
  await router.replace(fallbackPath);
}

function showPluginPageRefreshNotice(version: string) {
  notification.warning({
    key: pluginPageRefreshNotificationKey,
    message: '插件已更新',
    description: `当前插件页面已有新代际生效（${version || 'latest'}），请刷新当前页面以切换到最新版本。`,
    duration: 0,
    btn: () =>
      h(
        Button,
        {
          size: 'small',
          type: 'primary',
          onClick: () => {
            void handlePluginPageRefreshNow();
          },
        },
        { default: () => '刷新当前页面' },
      ),
  });
}

async function syncPluginPageGenerationBaseline() {
  const currentRoute = router.currentRoute.value;
  if (!getPendingPluginPageRefresh(currentRoute)) {
    clearPendingPluginPageRefresh();
    notification.close(pluginPageRefreshNotificationKey);
  }
  if (!resolvePluginPageId(currentRoute)) {
    return;
  }

  const pluginStateMap = await getPluginStateMap();
  rememberPluginPageGeneration(currentRoute, pluginStateMap);
}

function handlePluginRegistryMaybeChanged() {
  if (
    typeof document !== 'undefined' &&
    document.visibilityState &&
    document.visibilityState !== 'visible'
  ) {
    return;
  }
  void notifyPluginRegistryChangedIfNeeded();
}

// Fetch messages when notification panel is likely to open
// The Notification component triggers @read when opened
// We fetch on mount to have data ready
onMounted(() => {
  messageStore.fetchMessages();
  void getPluginStateMap();
  void syncPluginPageGenerationBaseline();
  disposePluginRegistryListener = onPluginRegistryChanged(async () => {
    const pluginStateMap = await getPluginStateMap();
    const pending = detectPendingPluginPageRefresh(
      router.currentRoute.value,
      pluginStateMap,
    );
    if (pending) {
      showPluginPageRefreshNotice(pending.version);
    }
    await refreshPluginAwareAccess();
  });
  window.addEventListener('focus', handlePluginRegistryMaybeChanged);
  document.addEventListener(
    'visibilitychange',
    handlePluginRegistryMaybeChanged,
  );
});

onBeforeUnmount(() => {
  disposePluginRegistryListener?.();
  disposePluginRegistryListener = null;
  window.removeEventListener('focus', handlePluginRegistryMaybeChanged);
  document.removeEventListener(
    'visibilitychange',
    handlePluginRegistryMaybeChanged,
  );
});

watch(
  () => ({
    enable: preferences.app.watermark,
    content: preferences.app.watermarkContent,
  }),
  async ({ enable, content }) => {
    if (enable) {
      await updateWatermark({
        content:
          content ||
          `${userStore.userInfo?.username} - ${userStore.userInfo?.realName}`,
      });
    } else {
      destroyWatermark();
    }
  },
  {
    immediate: true,
  },
);

watch(
  () => router.currentRoute.value.fullPath,
  () => {
    void syncPluginPageGenerationBaseline();
  },
  { immediate: true },
);
</script>

<template>
  <BasicLayout @clear-preferences-and-logout="handleLogout">
    <template #header-right-45>
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.layoutHeaderActionsBefore"
        class="mr-2"
      />
    </template>
    <template #header-right-145>
      <PluginSlotOutlet
        :slot-key="pluginSlotKeys.layoutHeaderActionsAfter"
        class="mr-2"
      />
    </template>
    <template #user-dropdown>
      <div class="flex items-center">
        <PluginSlotOutlet
          :slot-key="pluginSlotKeys.layoutUserDropdownAfter"
          class="mr-2"
        />
        <UserDropdown
          :avatar
          :menus
          :text="userStore.userInfo?.realName"
          :description="userStore.userInfo?.email || ''"
          :tag-text="userStore.userInfo?.username"
          @logout="handleLogout"
        />
      </div>
    </template>
    <template #notification>
      <Notification
        :dot="showDot"
        :notifications="notifications"
        @clear="handleNoticeClear"
        @click="handleNotificationClick"
        @read="handleRead"
        @remove="handleRemove"
        @make-all="handleMakeAll"
        @view-all="handleViewAll"
      />
    </template>
    <template #extra>
      <AuthenticationLoginExpiredModal
        v-model:open="accessStore.loginExpired"
        :avatar
      >
        <LoginForm />
      </AuthenticationLoginExpiredModal>
    </template>
    <template #lock-screen>
      <LockScreen :avatar @to-login="handleLogout" />
    </template>
  </BasicLayout>
  <PreviewModal />
</template>
