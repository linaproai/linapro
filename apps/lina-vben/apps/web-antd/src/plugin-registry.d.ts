declare module 'virtual:lina-plugin-pages' {
  import type { Component } from 'vue';
  type PluginCapabilityKey =
    import('#/plugins/plugin-capabilities').PluginCapabilityKey;

  export interface VirtualPluginPageModule {
    default?: Component;
    pluginPageMeta?: {
      capabilities?: PluginCapabilityKey[];
      pluginId?: string;
      routePath?: string;
      title?: string;
    };
  }

  export interface VirtualPluginPageModuleEntry {
    filePath: string;
    module: VirtualPluginPageModule;
  }

  export const pluginPageModules: VirtualPluginPageModuleEntry[];
}

declare module 'virtual:lina-plugin-slots' {
  import type { Component } from 'vue';
  type PluginCapabilityKey =
    import('#/plugins/plugin-capabilities').PluginCapabilityKey;
  type PluginSlotKey = import('#/plugins/plugin-slots').PluginSlotKey;

  export interface VirtualPluginSlotModule {
    default?: Component;
    pluginSlotMeta?: {
      capabilities?: PluginCapabilityKey[];
      order?: number;
      pluginId?: string;
      slotKey?: PluginSlotKey;
    };
  }

  export interface VirtualPluginSlotModuleEntry {
    filePath: string;
    module: VirtualPluginSlotModule;
  }

  export const pluginSlotModules: VirtualPluginSlotModuleEntry[];
}

declare module 'virtual:lina-app-third-party-locales' {
  export type ThirdPartyLocaleLoader = () => Promise<{
    default?: unknown;
    [key: string]: unknown;
  }>;

  export const antdLocaleLoaders: Record<string, ThirdPartyLocaleLoader>;
  export const dayjsLocaleLoaders: Record<string, ThirdPartyLocaleLoader>;
}

declare module 'virtual:lina-vxe-locales' {
  export type ThirdPartyLocaleLoader = () => Promise<{
    default?: unknown;
    [key: string]: unknown;
  }>;

  export const vxeLocaleLoaders: Record<string, ThirdPartyLocaleLoader>;
}
