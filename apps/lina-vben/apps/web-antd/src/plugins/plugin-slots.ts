export const pluginSlotKeys = {
  authLoginAfter: 'auth.login.after',
  /**
   * Platform social account icons (Google / Discord / QQ / …).
   * Host renders these as a Vben-style icon row under “其他登录方式”.
   * Protocol / directory logins (generic OIDC, LDAP) stay on auth.login.after.
   */
  authLoginSocial: 'auth.login.social',
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
      'Full-width button stack below the login form for protocol or directory logins (generic OIDC, LDAP, enterprise SSO).',
    hostLocation: 'auth.login',
    key: pluginSlotKeys.authLoginAfter,
  },
  {
    description:
      'Icon-row extension under the login form for third-party platform accounts (Google, Discord, QQ, WeChat, GitHub). Host wraps with an “other login methods” divider.',
    hostLocation: 'auth.login',
    key: pluginSlotKeys.authLoginSocial,
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
