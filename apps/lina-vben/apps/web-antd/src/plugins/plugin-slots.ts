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
    description:
      'Public extension area below the login form for hints or lightweight entries.',
    hostLocation: 'auth.login',
    key: pluginSlotKeys.authLoginAfter,
  },
  {
    description:
      'Extension area below CRUD tables for help cards or supporting panels.',
    hostLocation: 'crud.table',
    key: pluginSlotKeys.crudTableAfter,
  },
  {
    description:
      'Extension area on the right side of CRUD toolbars for status or quick actions.',
    hostLocation: 'crud.toolbar',
    key: pluginSlotKeys.crudToolbarAfter,
  },
  {
    description:
      'Extension area above the workspace main content for banners, alerts, or summaries.',
    hostLocation: 'dashboard.workspace',
    key: pluginSlotKeys.dashboardWorkspaceBefore,
  },
  {
    description:
      'Extension area below the workspace main content for plugin cards or metrics.',
    hostLocation: 'dashboard.workspace',
    key: pluginSlotKeys.dashboardWorkspaceAfter,
  },
  {
    description:
      'Extension area before host header actions for global status or entries.',
    hostLocation: 'layout.header.actions',
    key: pluginSlotKeys.layoutHeaderActionsBefore,
  },
  {
    description:
      'Extension area after host header actions for hints or quick entries.',
    hostLocation: 'layout.header.actions',
    key: pluginSlotKeys.layoutHeaderActionsAfter,
  },
  {
    description:
      'Extension area before the user menu for lightweight entries or status hints.',
    hostLocation: 'layout.user-dropdown',
    key: pluginSlotKeys.layoutUserDropdownAfter,
  },
];

export function isPluginSlotKey(value: string): value is PluginSlotKey {
  return publishedPluginSlotKeySet.has(value as PluginSlotKey);
}
