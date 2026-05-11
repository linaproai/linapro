import type { PluginDynamicState } from '#/api/system/plugin/model';
import type { PluginCapabilityKey } from '#/plugins/plugin-capabilities';
import type { PluginSlotKey } from '#/plugins/plugin-slots';
import type { Component } from 'vue';
import type { VirtualPluginSlotModuleEntry } from 'virtual:lina-plugin-slots';

import { pluginDynamicList } from '#/api/system/plugin';
import { isPluginCapabilityKey } from '#/plugins/plugin-capabilities';
import { getPluginPages } from '#/plugins/page-registry';
import { isPluginSlotKey } from '#/plugins/plugin-slots';
import { pluginSlotModules } from 'virtual:lina-plugin-slots';

type PluginRegistryListener = () => void | Promise<void>;

type PluginRegistryGlobal = typeof globalThis & {
  __linaPluginRegistryCheckPromise?: null | Promise<boolean>;
  __linaPluginRegistryListeners?: Set<PluginRegistryListener>;
  __linaPluginStateSignature?: null | string;
  __linaPluginStatePromise?: null | Promise<Map<string, PluginDynamicState>>;
};

export interface PluginSlotMeta {
  capabilities?: PluginCapabilityKey[];
  order?: number;
  pluginId?: string;
  slotKey?: PluginSlotKey;
}

export interface RegisteredPluginSlotModule {
  capabilities: PluginCapabilityKey[];
  component: Component;
  filePath: string;
  key: string;
  order: number;
  pluginId: string;
  slotKey: PluginSlotKey;
}

export interface PluginCapabilityState {
  enabled: boolean;
  pluginIds: string[];
}

export type PluginCapabilityStateMap = Map<
  PluginCapabilityKey,
  PluginCapabilityState
>;

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
      `[plugin-slot] skip unpublished capability "${value}" from ${filePath}`,
    );
  }
  return [...capabilities].sort();
}

const registeredPluginSlotModules = pluginSlotModules
  .map((item: VirtualPluginSlotModuleEntry) => {
    const match = item.filePath.match(
      /\/lina-plugins\/([^/]+)\/frontend\/slots\/(.+)\.vue$/,
    );
    if (!match?.[1] || !match[2] || !item.module.default) {
      return null;
    }

    const pluginId = item.module.pluginSlotMeta?.pluginId || match[1];
    const relativePath = match[2];
    const segments = relativePath.split('/');
    const slotKey =
      item.module.pluginSlotMeta?.slotKey ||
      segments.slice(0, Math.max(segments.length - 1, 0)).join('/');
    const slotName = segments.at(-1) || relativePath;

    if (!slotKey || !isPluginSlotKey(slotKey)) {
      console.warn(
        `[plugin-slot] skip unpublished slot "${slotKey}" from ${item.filePath}`,
      );
      return null;
    }

    return {
      capabilities: normalizeCapabilities(
        item.module.pluginSlotMeta?.capabilities,
        item.filePath,
      ),
      component: item.module.default as Component,
      filePath: item.filePath,
      key: `${pluginId}:${slotKey}:${slotName}`,
      order: item.module.pluginSlotMeta?.order ?? 0,
      pluginId,
      slotKey,
    } satisfies RegisteredPluginSlotModule;
  })
  .filter((item): item is RegisteredPluginSlotModule => item !== null)
  .sort((a, b) => {
    if (a.order !== b.order) {
      return a.order - b.order;
    }
    return a.key.localeCompare(b.key);
  });

function normalizePluginKeys(item: PluginDynamicState): string[] {
  const keys = [item.id];
  if (item.statusKey?.startsWith('sys_plugin.status:')) {
    keys.push(item.statusKey.substring('sys_plugin.status:'.length));
  }
  return keys.filter((key): key is string => !!key);
}

function getPluginRegistryGlobal() {
  return globalThis as PluginRegistryGlobal;
}

function getPluginRegistryListeners() {
  const registryGlobal = getPluginRegistryGlobal();
  registryGlobal.__linaPluginRegistryListeners ??= new Set();
  return registryGlobal.__linaPluginRegistryListeners;
}

function getPluginStatePromise() {
  return getPluginRegistryGlobal().__linaPluginStatePromise ?? null;
}

function getPluginStateSignature() {
  return getPluginRegistryGlobal().__linaPluginStateSignature ?? null;
}

function setPluginStatePromise(
  promise: null | Promise<Map<string, PluginDynamicState>>,
) {
  getPluginRegistryGlobal().__linaPluginStatePromise = promise;
}

function setPluginStateSignature(signature: null | string) {
  getPluginRegistryGlobal().__linaPluginStateSignature = signature;
}

function buildPluginStateMap(items: PluginDynamicState[]) {
  const map = new Map<string, PluginDynamicState>();
  for (const item of items) {
    for (const key of normalizePluginKeys(item)) {
      map.set(key, item);
    }
  }
  return map;
}

function isEnabled(value: unknown) {
  return value === 1 || value === '1' || value === true;
}

function isPluginEnabled(
  pluginId: string,
  pluginStateMap: Map<string, PluginDynamicState>,
) {
  const pluginState = pluginStateMap.get(pluginId);
  return (
    isEnabled(pluginState?.installed) && isEnabled(pluginState?.enabled)
  );
}

function getRegisteredCapabilityModules() {
  return [
    ...getPluginPages().map((item) => ({
      capabilities: item.capabilities,
      pluginId: item.pluginId,
    })),
    ...registeredPluginSlotModules.map((item) => ({
      capabilities: item.capabilities,
      pluginId: item.pluginId,
    })),
  ];
}

function buildPluginStateSignature(items: PluginDynamicState[]) {
  return items
    .map(
      (item) =>
        `${item.id}:${item.installed}:${item.enabled}:${item.version}:${item.generation}:${item.statusKey}`,
    )
    .sort()
    .join('|');
}

function setPluginStateSnapshot(items: PluginDynamicState[]) {
  const pluginStateMap = buildPluginStateMap(items);
  setPluginStateSignature(buildPluginStateSignature(items));
  setPluginStatePromise(Promise.resolve(pluginStateMap));
  return pluginStateMap;
}

async function loadPluginStateMap(force = false) {
  let pluginStatePromise = getPluginStatePromise();
  if (!pluginStatePromise || force) {
    pluginStatePromise = pluginDynamicList()
      .then((items) => {
        return setPluginStateSnapshot(items);
      })
      .catch((error) => {
        console.error('[plugin-slot] failed to load plugin state map', error);
        return new Map<string, PluginDynamicState>();
      });
    setPluginStatePromise(pluginStatePromise);
  }
  return pluginStatePromise;
}

/**
 * Returns plugin slot definitions for a given slot key.
 */
export function getPluginSlots(
  slotKey: PluginSlotKey,
): RegisteredPluginSlotModule[] {
  return registeredPluginSlotModules.filter((item) => item.slotKey === slotKey);
}

/**
 * Queries current plugin dynamic states from host backend.
 */
export async function getPluginStateMap(force = false) {
  return await loadPluginStateMap(force);
}

/**
 * Returns enabled frontend extension capabilities exposed by active plugins.
 */
export async function getPluginCapabilityStateMap(force = false) {
  const pluginStateMap = await getPluginStateMap(force);
  const capabilityMap: PluginCapabilityStateMap = new Map();
  for (const item of getRegisteredCapabilityModules()) {
    if (
      item.capabilities.length === 0 ||
      !isPluginEnabled(item.pluginId, pluginStateMap)
    ) {
      continue;
    }
    for (const capability of item.capabilities) {
      const current = capabilityMap.get(capability) ?? {
        enabled: false,
        pluginIds: [],
      };
      current.enabled = true;
      if (!current.pluginIds.includes(item.pluginId)) {
        current.pluginIds.push(item.pluginId);
      }
      capabilityMap.set(capability, current);
    }
  }
  return capabilityMap;
}

/**
 * Reports whether an extension capability is currently provided by any active plugin.
 */
export async function hasPluginCapability(
  capability: PluginCapabilityKey,
  force = false,
) {
  return (
    (await getPluginCapabilityStateMap(force)).get(capability)?.enabled === true
  );
}

/**
 * Notifies plugin-aware UI that plugin registry state changed.
 */
export async function notifyPluginRegistryChanged() {
  setPluginStatePromise(null);
  setPluginStateSignature(null);
  await Promise.allSettled(
    Array.from(getPluginRegistryListeners(), (listener) =>
      Promise.resolve(listener()),
    ),
  );
}

/**
 * Queries latest plugin dynamic state and only notifies listeners when it actually changed.
 */
export async function notifyPluginRegistryChangedIfNeeded() {
  const registryGlobal = getPluginRegistryGlobal();
  if (registryGlobal.__linaPluginRegistryCheckPromise) {
    return await registryGlobal.__linaPluginRegistryCheckPromise;
  }

  registryGlobal.__linaPluginRegistryCheckPromise = (async () => {
    try {
      // Reuse any in-flight baseline load so the first focus restore does not
      // misclassify "no snapshot yet" as "plugin registry changed".
      if (!getPluginStateSignature()) {
        await getPluginStatePromise();
      }

      const previousSignature = getPluginStateSignature();
      const items = await pluginDynamicList();
      const nextSignature = buildPluginStateSignature(items);

      if (!previousSignature) {
        setPluginStateSnapshot(items);
        return false;
      }

      if (nextSignature === previousSignature) {
        return false;
      }

      setPluginStateSnapshot(items);
      await Promise.allSettled(
        Array.from(getPluginRegistryListeners(), (listener) =>
          Promise.resolve(listener()),
        ),
      );
      return true;
    } catch (error) {
      console.error(
        '[plugin-slot] failed to check plugin registry changes',
        error,
      );
      return false;
    } finally {
      registryGlobal.__linaPluginRegistryCheckPromise = null;
    }
  })();

  return await registryGlobal.__linaPluginRegistryCheckPromise;
}

/**
 * Subscribes to plugin registry changes.
 */
export function onPluginRegistryChanged(listener: () => void | Promise<void>) {
  const pluginRegistryListeners = getPluginRegistryListeners();
  pluginRegistryListeners.add(listener);
  return () => {
    pluginRegistryListeners.delete(listener);
  };
}
