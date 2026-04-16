declare module 'virtual:lina-plugin-pages' {
  import type { Component } from 'vue';

  export interface VirtualPluginPageModule {
    default?: Component;
    pluginPageMeta?: {
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
  type PluginSlotKey = import('#/plugins/plugin-slots').PluginSlotKey;

  export interface VirtualPluginSlotModule {
    default?: Component;
    pluginSlotMeta?: {
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
