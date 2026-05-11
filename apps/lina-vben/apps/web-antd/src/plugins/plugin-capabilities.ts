export const pluginCapabilityKeys = {
  organizationManagement: 'organization.management',
  tenantManagement: 'tenant.management',
} as const;

export type PluginCapabilityKey =
  (typeof pluginCapabilityKeys)[keyof typeof pluginCapabilityKeys];

const pluginCapabilityKeySet = new Set<PluginCapabilityKey>(
  Object.values(pluginCapabilityKeys),
);

export function isPluginCapabilityKey(
  value: string,
): value is PluginCapabilityKey {
  return pluginCapabilityKeySet.has(value as PluginCapabilityKey);
}
