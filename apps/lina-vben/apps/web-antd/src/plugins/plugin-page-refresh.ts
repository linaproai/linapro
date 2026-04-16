import type { PluginDynamicState } from '#/api/system/plugin/model';
import type { RouteLocationNormalizedLoaded } from 'vue-router';

import { getPluginPageByRoute } from '#/plugins/page-registry';

export interface PendingPluginPageRefresh {
  generation: number;
  pluginId: string;
  version: string;
}

type PluginPageGenerationBaseline = PendingPluginPageRefresh;

const pluginAssetRoutePattern = /\/plugin-assets\/([^/]+)\/([^/]+)\//;

let pluginPageGenerationBaseline: null | PluginPageGenerationBaseline = null;
let pendingPluginPageRefresh: null | PendingPluginPageRefresh = null;

function normalizeRoutePath(route: RouteLocationNormalizedLoaded) {
  return route.path.replace(/^\//, '');
}

function extractPluginIdFromAssetURL(value: unknown) {
  if (typeof value !== 'string' || !value) {
    return '';
  }
  return value.match(pluginAssetRoutePattern)?.[1] ?? '';
}

function extractPluginIdFromVisibleIframe() {
  if (typeof document === 'undefined' || typeof window === 'undefined') {
    return '';
  }

  // IFrame routes can keep the current page content alive outside of the route
  // meta refresh window. When the router has not yet rebuilt `meta.iframeSrc`,
  // the visible iframe still exposes the active plugin asset URL.
  for (const iframe of Array.from(document.querySelectorAll('iframe'))) {
    if (!(iframe instanceof HTMLIFrameElement)) {
      continue;
    }
    const style = window.getComputedStyle(iframe);
    const rect = iframe.getBoundingClientRect();
    if (
      style.display === 'none' ||
      style.visibility === 'hidden' ||
      rect.width <= 0 ||
      rect.height <= 0
    ) {
      continue;
    }

    const pluginId = extractPluginIdFromAssetURL(
      iframe.getAttribute('src') || iframe.src,
    );
    if (pluginId) {
      return pluginId;
    }
  }
  return '';
}

function mergeRouteQuery(route: RouteLocationNormalizedLoaded) {
  return {
    ...((route.meta?.query ?? {}) as Record<string, unknown>),
    ...(route.query as Record<string, unknown>),
  };
}

export function resolvePluginPageId(route: RouteLocationNormalizedLoaded) {
  const pluginPage = getPluginPageByRoute(normalizeRoutePath(route));
  if (pluginPage?.pluginId) {
    return pluginPage.pluginId;
  }

  const mergedQuery = mergeRouteQuery(route);
  const dynamicSources = [
    mergedQuery.embeddedSrc,
    route.meta?.iframeSrc,
    route.meta?.link,
    route.fullPath,
    route.path,
  ];
  for (const source of dynamicSources) {
    const pluginId = extractPluginIdFromAssetURL(source);
    if (pluginId) {
      return pluginId;
    }
  }
  return extractPluginIdFromVisibleIframe();
}

function findPluginState(
  pluginId: string,
  pluginStateMap: Map<string, PluginDynamicState>,
) {
  if (!pluginId) {
    return null;
  }
  return pluginStateMap.get(pluginId) ?? null;
}

export function clearPendingPluginPageRefresh() {
  pendingPluginPageRefresh = null;
}

export function getPendingPluginPageRefresh(
  route: RouteLocationNormalizedLoaded,
) {
  if (!pendingPluginPageRefresh) {
    return null;
  }
  const pluginId = resolvePluginPageId(route);
  if (!pluginId || pluginId !== pendingPluginPageRefresh.pluginId) {
    return null;
  }
  return pendingPluginPageRefresh;
}

export function rememberPluginPageGeneration(
  route: RouteLocationNormalizedLoaded,
  pluginStateMap: Map<string, PluginDynamicState>,
) {
  if (pendingPluginPageRefresh) {
    return;
  }

  const pluginId = resolvePluginPageId(route);
  const pluginState = findPluginState(pluginId, pluginStateMap);
  if (!pluginState) {
    pluginPageGenerationBaseline = null;
    return;
  }

  pluginPageGenerationBaseline = {
    pluginId,
    generation: pluginState.generation,
    version: pluginState.version,
  };
}

export function detectPendingPluginPageRefresh(
  route: RouteLocationNormalizedLoaded,
  pluginStateMap: Map<string, PluginDynamicState>,
) {
  const pluginId = resolvePluginPageId(route);
  if (!pluginId) {
    pluginPageGenerationBaseline = null;
    pendingPluginPageRefresh = null;
    return null;
  }

  const pluginState = findPluginState(pluginId, pluginStateMap);
  if (!pluginState) {
    return null;
  }
  if (pluginState.installed !== 1 || pluginState.enabled !== 1) {
    pendingPluginPageRefresh = null;
    return null;
  }
  if (
    !pluginPageGenerationBaseline ||
    pluginPageGenerationBaseline.pluginId !== pluginId
  ) {
    pluginPageGenerationBaseline = {
      pluginId,
      generation: pluginState.generation,
      version: pluginState.version,
    };
    return null;
  }
  if (pluginPageGenerationBaseline.generation === pluginState.generation) {
    return null;
  }

  pendingPluginPageRefresh = {
    pluginId,
    generation: pluginState.generation,
    version: pluginState.version,
  };
  return pendingPluginPageRefresh;
}
