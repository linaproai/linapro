export const pluginSlotKeys = {
  authLoginAfter: 'auth.login.after',
  crudTableAfter: 'crud.table.after',
  crudToolbarAfter: 'crud.toolbar.after',
  dashboardWorkspaceBefore: 'dashboard.workspace.before',
  dashboardWorkspaceAfter: 'dashboard.workspace.after',
  layoutHeaderActionsBefore: 'layout.header.actions.before',
  layoutHeaderActionsAfter: 'layout.header.actions.after',
  layoutUserDropdownAfter: 'layout.user-dropdown.after',
} as const;

export type PluginSlotKey =
  (typeof pluginSlotKeys)[keyof typeof pluginSlotKeys];

export interface PublishedPluginSlot {
  description: string;
  hostLocation: string;
  key: PluginSlotKey;
}

const publishedPluginSlotKeySet = new Set<PluginSlotKey>(
  Object.values(pluginSlotKeys),
);

export const publishedPluginSlots: PublishedPluginSlot[] = [
  {
    description: '登录页表单下方的公开扩展区域，适合挂载提示信息或轻量入口。',
    hostLocation: 'auth.login',
    key: pluginSlotKeys.authLoginAfter,
  },
  {
    description: '通用 CRUD 表格区域下方的扩展区域，适合挂载说明卡片或辅助面板。',
    hostLocation: 'crud.table',
    key: pluginSlotKeys.crudTableAfter,
  },
  {
    description: '通用 CRUD 工具栏右侧扩展区域，适合挂载轻量状态或快捷操作。',
    hostLocation: 'crud.toolbar',
    key: pluginSlotKeys.crudToolbarAfter,
  },
  {
    description: '工作台主内容区顶部扩展区域，适合挂载横幅、提醒或概览块。',
    hostLocation: 'dashboard.workspace',
    key: pluginSlotKeys.dashboardWorkspaceBefore,
  },
  {
    description: '工作台主内容区底部扩展区域，适合挂载插件卡片或统计块。',
    hostLocation: 'dashboard.workspace',
    key: pluginSlotKeys.dashboardWorkspaceAfter,
  },
  {
    description: '宿主头部动作区前置扩展区域，适合挂载全局状态或入口。',
    hostLocation: 'layout.header.actions',
    key: pluginSlotKeys.layoutHeaderActionsBefore,
  },
  {
    description: '宿主头部动作区后置扩展区域，适合挂载轻量提示或快捷入口。',
    hostLocation: 'layout.header.actions',
    key: pluginSlotKeys.layoutHeaderActionsAfter,
  },
  {
    description: '右上角用户菜单左侧扩展区域，适合挂载轻量入口或状态提示。',
    hostLocation: 'layout.user-dropdown',
    key: pluginSlotKeys.layoutUserDropdownAfter,
  },
];

export function isPluginSlotKey(value: string): value is PluginSlotKey {
  return publishedPluginSlotKeySet.has(value as PluginSlotKey);
}
