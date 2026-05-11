import type { Component } from 'vue';
import type { VirtualPluginPageModuleEntry } from 'virtual:lina-plugin-pages';

import type { PluginCapabilityKey } from './plugin-capabilities';

import { isPluginCapabilityKey } from './plugin-capabilities';
import { pluginPageModules } from 'virtual:lina-plugin-pages';

export interface PluginPageMeta {
  capabilities?: PluginCapabilityKey[];
  pluginId?: string;
  routePath?: string;
  title?: string;
}

export interface RegisteredPluginPage {
  component: Component;
  capabilities: PluginCapabilityKey[];
  filePath: string;
  key: string;
  pluginId: string;
  routePath: string;
  title: string;
}

function inferRoutePath(pluginId: string, pagePath: string) {
  return `${pluginId}-${pagePath.replaceAll('/', '-').replaceAll('_', '-')}`;
}

function normalizeRoutePath(routePath: string) {
  return routePath.replace(/^\//, '');
}

function normalizeCapabilities(
  values: PluginCapabilityKey[] | undefined,
  filePath: string,
) {
  const capabilities = new Set<PluginCapabilityKey>();
  for (const value of values ?? []) {
    if (isPluginCapabilityKey(value)) {
      capabilities.add(value);
      continue;
    }
    console.warn(
      `[plugin-page] skip unpublished capability "${value}" from ${filePath}`,
    );
  }
  return [...capabilities].sort();
}

const pluginPages = pluginPageModules
  .map((item: VirtualPluginPageModuleEntry) => {
    const match = item.filePath.match(
      /\/lina-plugins\/([^/]+)\/frontend\/pages\/(.+)\.vue$/,
    );
    if (!match?.[1] || !match[2] || !item.module.default) {
      return null;
    }

    const pluginId = item.module.pluginPageMeta?.pluginId || match[1];
    const pagePath = match[2];
    const routePath = normalizeRoutePath(
      item.module.pluginPageMeta?.routePath ||
        inferRoutePath(pluginId, pagePath),
    );

    return {
      component: item.module.default as Component,
      capabilities: normalizeCapabilities(
        item.module.pluginPageMeta?.capabilities,
        item.filePath,
      ),
      filePath: item.filePath,
      key: `${pluginId}:${pagePath}`,
      pluginId,
      routePath,
      title: item.module.pluginPageMeta?.title || routePath,
    } satisfies RegisteredPluginPage;
  })
  .filter((item): item is RegisteredPluginPage => item !== null)
  .sort((a, b) => a.routePath.localeCompare(b.routePath));

export function getPluginPageByRoute(routePath: string) {
  const normalizedRoutePath = normalizeRoutePath(routePath);
  return (
    pluginPages.find((item) => item.routePath === normalizedRoutePath) ?? null
  );
}

export function getPluginPages() {
  return pluginPages;
}
