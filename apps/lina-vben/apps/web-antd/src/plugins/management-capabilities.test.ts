import { beforeEach, describe, expect, it, vi } from 'vitest';

import { pluginCapabilityKeys } from '#/plugins/plugin-capabilities';

import type * as ManagementCapabilities from './management-capabilities';

const {
  getPluginCapabilityProviderState,
  getPluginCapabilityStateMap,
  tenantStore,
} = vi.hoisted(() => ({
  getPluginCapabilityProviderState: vi.fn(),
  getPluginCapabilityStateMap: vi.fn(),
  tenantStore: {
    enabled: false,
  },
}));

vi.mock('#/plugins/slot-registry', () => ({
  getPluginCapabilityProviderState,
  getPluginCapabilityStateMap,
}));

vi.mock('#/store/tenant', () => ({
  useTenantStore: () => tenantStore,
}));

let managementCapabilities: typeof ManagementCapabilities;

describe('management capability resolver', () => {
  beforeEach(async () => {
    managementCapabilities = await import('./management-capabilities');
    tenantStore.enabled = false;
    getPluginCapabilityStateMap.mockResolvedValue(new Map());
    getPluginCapabilityProviderState.mockResolvedValue({
      enabled: false,
      observed: false,
      observedPluginIds: [],
      pluginIds: [],
    });
  });

  it('uses observed tenant provider runtime state instead of persisted tenant shell state', async () => {
    tenantStore.enabled = true;
    getPluginCapabilityProviderState.mockResolvedValue({
      enabled: false,
      observed: true,
      observedPluginIds: ['custom-tenant-provider'],
      pluginIds: ['custom-tenant-provider'],
    });

    await expect(
      managementCapabilities.resolveTenantManagementEnabled(),
    ).resolves.toBe(false);
    expect(getPluginCapabilityProviderState).toHaveBeenCalledWith(
      pluginCapabilityKeys.tenantManagement,
      false,
    );
  });

  it('falls back to tenant shell state when no capability provider is observed', async () => {
    tenantStore.enabled = true;

    await expect(
      managementCapabilities.resolveTenantManagementEnabled(),
    ).resolves.toBe(true);
  });

  it('returns organization capability from the capability map', async () => {
    getPluginCapabilityStateMap.mockResolvedValue(
      new Map([
        [
          pluginCapabilityKeys.organizationManagement,
          { enabled: true, pluginIds: ['custom-org-provider'] },
        ],
      ]),
    );

    await expect(
      managementCapabilities.resolveManagementCapabilityState(),
    ).resolves.toMatchObject({
      organizationEnabled: true,
      tenantEnabled: false,
    });
  });
});
