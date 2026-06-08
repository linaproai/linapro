import type { RouteLocationNormalizedGeneric } from 'vue-router';

import type { MenuRecordRaw, TabDefinition } from '@vben/types';

import type { IContextMenuItem } from '@vben-core/tabs-ui';

import { computed, ref, unref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import { useContentMaximize, useTabs } from '@vben/hooks';
import {
  ArrowLeftToLine,
  ArrowRightLeft,
  ArrowRightToLine,
  ExternalLink,
  FoldHorizontal,
  Fullscreen,
  Minimize2,
  Pin,
  PinOff,
  RotateCw,
  X,
} from '@vben/icons';
import { $t, $te, useI18n } from '@vben/locales';
import { getTabKey, useAccessStore, useTabbarStore } from '@vben/stores';
import { filterTree } from '@vben/utils';

interface MenuSnapshot {
  paths: Set<string>;
}

export function useTabbar() {
  const router = useRouter();
  const route = useRoute();
  const accessStore = useAccessStore();
  const tabbarStore = useTabbarStore();
  const { contentIsMaximize, toggleMaximize } = useContentMaximize();
  const {
    closeAllTabs,
    closeCurrentTab,
    closeLeftTabs,
    closeOtherTabs,
    closeRightTabs,
    closeTabByKey,
    getTabDisableState,
    openTabInNewWindow,
    refreshTab,
    toggleTabPin,
  } = useTabs();

  /**
   * 当前路径对应的tab的key
   */
  const currentActive = computed(() => {
    return getTabKey(route);
  });

  const { locale } = useI18n();
  const currentTabs = ref<RouteLocationNormalizedGeneric[]>();
  let previousMenuSnapshot: MenuSnapshot = {
    paths: new Set<string>(),
  };
  watch(
    [
      () => tabbarStore.getTabs,
      () => tabbarStore.updateTime,
      () => locale.value,
    ],
    ([tabs]) => {
      currentTabs.value = tabs.map((item) => wrapperTabLocale(item));
    },
  );

  /**
   * 初始化固定标签页
   */
  const initAffixTabs = () => {
    const affixTabs = filterTree(router.getRoutes(), (route) => {
      return !!route.meta?.affixTab;
    });
    tabbarStore.setAffixTabs(affixTabs);
  };

  function normalizeMenuPath(path: unknown) {
    if (typeof path !== 'string') {
      return '';
    }

    const normalized = path.split(/[?#]/u)[0]?.replace(/\/+$/u, '');
    return normalized || '/';
  }

  function collectMenuSnapshot(
    menus: MenuRecordRaw[],
    snapshot: MenuSnapshot = {
      paths: new Set<string>(),
    },
  ) {
    for (const menu of menus) {
      const path = normalizeMenuPath(menu.path);
      if (path) {
        snapshot.paths.add(path);
      }

      if (menu.children?.length) {
        collectMenuSnapshot(menu.children, snapshot);
      }
    }

    return snapshot;
  }

  function getRemovedSnapshot(currentSnapshot: MenuSnapshot): MenuSnapshot {
    return {
      paths: new Set(
        [...previousMenuSnapshot.paths].filter(
          (path) => !currentSnapshot.paths.has(path),
        ),
      ),
    };
  }

  function getTabCandidatePaths(tab: TabDefinition) {
    return [
      tab.path,
      tab.fullPath,
      tab.meta?.activePath,
      tab.meta?.link,
      ...(tab.matched?.map((item) => item.path) ?? []),
    ]
      .map((path) => normalizeMenuPath(path))
      .filter((path) => path !== '');
  }

  function isRemovedMenuTab(tab: TabDefinition, removedSnapshot: MenuSnapshot) {
    return getTabCandidatePaths(tab).some((path) =>
      removedSnapshot.paths.has(path),
    );
  }

  async function closeRemovedMenuTabs(menus: MenuRecordRaw[]) {
    const currentSnapshot = collectMenuSnapshot(menus);
    const removedSnapshot = getRemovedSnapshot(currentSnapshot);
    previousMenuSnapshot = currentSnapshot;

    if (removedSnapshot.paths.size === 0) {
      return;
    }

    const staleKeys = tabbarStore.getTabs
      .filter((tab) => !tab.meta?.affixTab)
      .filter((tab) => isRemovedMenuTab(tab, removedSnapshot))
      .map((tab) => tab.key)
      .filter((key): key is string => typeof key === 'string' && key !== '');

    if (staleKeys.length > 0) {
      await tabbarStore._bulkCloseByKeys(staleKeys);
    }
  }

  // 点击tab,跳转路由
  const handleClick = (key: string) => {
    const { fullPath, path } = tabbarStore.getTabByKey(key);
    router.push(fullPath || path);
  };

  // 关闭tab
  const handleClose = async (key: string) => {
    await closeTabByKey(key);
  };

  function wrapperTabLocale(tab: RouteLocationNormalizedGeneric) {
    const title = resolveTabTitle(tab);
    const newTabTitle = tab?.meta?.newTabTitle
      ? computed(() => translateIfExists(tab?.meta?.newTabTitle))
      : undefined;
    return {
      ...tab,
      meta: {
        ...tab?.meta,
        ...(newTabTitle ? { newTabTitle } : {}),
        title,
      },
    };
  }

  function translateIfExists(title: unknown) {
    const rawTitle = String(unref(title) ?? '');
    const titleKey = rawTitle.trim();
    if (!titleKey) {
      return '';
    }
    return $te(titleKey) ? $t(titleKey) : rawTitle;
  }

  function resolveTabTitle(tab: RouteLocationNormalizedGeneric) {
    const meta = tab?.meta;

    const i18nKey = String(meta?.i18nKey || '').trim();
    if (i18nKey && $te(i18nKey)) {
      return $t(i18nKey);
    }

    const title = translateIfExists(meta?.title);
    if (title) {
      return title;
    }

    return String(tab?.name || '');
  }

  watch(
    () => accessStore.accessMenus,
    async (menus) => {
      initAffixTabs();
      await closeRemovedMenuTabs(menus);
    },
    { immediate: true },
  );

  watch(
    () => route.fullPath,
    () => {
      const meta = route.matched?.[route.matched.length - 1]?.meta;
      tabbarStore.addTab({
        ...route,
        meta: meta || route.meta,
      });
    },
    { immediate: true },
  );

  const createContextMenus = (tab: TabDefinition) => {
    const {
      disabledCloseAll,
      disabledCloseCurrent,
      disabledCloseLeft,
      disabledCloseOther,
      disabledCloseRight,
      disabledRefresh,
    } = getTabDisableState(tab);

    const affixTab = tab?.meta?.affixTab ?? false;

    const menus: IContextMenuItem[] = [
      {
        disabled: disabledCloseCurrent,
        handler: async () => {
          await closeCurrentTab(tab);
        },
        icon: X,
        key: 'close',
        text: $t('preferences.tabbar.contextMenu.close'),
      },
      {
        handler: async () => {
          await toggleTabPin(tab);
        },
        icon: affixTab ? PinOff : Pin,
        key: 'affix',
        text: affixTab
          ? $t('preferences.tabbar.contextMenu.unpin')
          : $t('preferences.tabbar.contextMenu.pin'),
      },
      {
        handler: async () => {
          if (!contentIsMaximize.value) {
            await router.push(tab.fullPath);
          }
          toggleMaximize();
        },
        icon: contentIsMaximize.value ? Minimize2 : Fullscreen,
        key: contentIsMaximize.value ? 'restore-maximize' : 'maximize',
        text: contentIsMaximize.value
          ? $t('preferences.tabbar.contextMenu.restoreMaximize')
          : $t('preferences.tabbar.contextMenu.maximize'),
      },
      {
        disabled: disabledRefresh,
        handler: () => refreshTab(),
        icon: RotateCw,
        key: 'reload',
        text: $t('preferences.tabbar.contextMenu.reload'),
      },
      {
        handler: async () => {
          await openTabInNewWindow(tab);
        },
        icon: ExternalLink,
        key: 'open-in-new-window',
        separator: true,
        text: $t('preferences.tabbar.contextMenu.openInNewWindow'),
      },

      {
        disabled: disabledCloseLeft,
        handler: async () => {
          await closeLeftTabs(tab);
        },
        icon: ArrowLeftToLine,
        key: 'close-left',
        text: $t('preferences.tabbar.contextMenu.closeLeft'),
      },
      {
        disabled: disabledCloseRight,
        handler: async () => {
          await closeRightTabs(tab);
        },
        icon: ArrowRightToLine,
        key: 'close-right',
        separator: true,
        text: $t('preferences.tabbar.contextMenu.closeRight'),
      },
      {
        disabled: disabledCloseOther,
        handler: async () => {
          await closeOtherTabs(tab);
        },
        icon: FoldHorizontal,
        key: 'close-other',
        text: $t('preferences.tabbar.contextMenu.closeOther'),
      },
      {
        disabled: disabledCloseAll,
        handler: closeAllTabs,
        icon: ArrowRightLeft,
        key: 'close-all',
        text: $t('preferences.tabbar.contextMenu.closeAll'),
      },
    ];

    return menus.filter((item) => tabbarStore.getMenuList.includes(item.key));
  };

  return {
    createContextMenus,
    currentActive,
    currentTabs,
    handleClick,
    handleClose,
  };
}
