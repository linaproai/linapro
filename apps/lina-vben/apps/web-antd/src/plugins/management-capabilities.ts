import { pluginCapabilityKeys } from '#/plugins/plugin-capabilities';
import {
  getPluginCapabilityProviderState,
  getPluginCapabilityStateMap,
} from '#/plugins/slot-registry';
import { useTenantStore } from '#/store/tenant';

export interface ManagementCapabilityState {
  organizationEnabled: boolean;
  tenantEnabled: boolean;
}

/**
 * Resolves host management capabilities without binding pages to official plugin IDs.
 */
export async function resolveManagementCapabilityState(
  force = false,
): Promise<ManagementCapabilityState> {
  const tenantStore = useTenantStore();
  const capabilityMap = await getPluginCapabilityStateMap(force);
  const tenantProviderState = await getPluginCapabilityProviderState(
    pluginCapabilityKeys.tenantManagement,
    force,
  );
  const tenantCapabilityEnabled =
    capabilityMap.get(pluginCapabilityKeys.tenantManagement)?.enabled === true;

  return {
    organizationEnabled: capabilityMap.get(
      pluginCapabilityKeys.organizationManagement,
    )?.enabled === true,
    tenantEnabled: tenantProviderState.observed
      ? tenantProviderState.enabled
      : tenantCapabilityEnabled || tenantStore.enabled,
  };
}

/**
 * Resolves whether tenant management should be visible in host shell state.
 */
export async function resolveTenantManagementEnabled(force = false) {
  const state = await resolveManagementCapabilityState(force);
  return state.tenantEnabled;
}
