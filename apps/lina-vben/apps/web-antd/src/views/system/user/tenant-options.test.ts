import { beforeEach, describe, expect, it, vi } from 'vitest';

import { loadUserTenantOptions } from './tenant-options';

const { authLoginTenants, platformTenantList } = vi.hoisted(() => ({
  authLoginTenants: vi.fn(),
  platformTenantList: vi.fn(),
}));

vi.mock('#/api/tenant', () => ({
  authLoginTenants,
}));

vi.mock('#/api/platform/tenant', () => ({
  platformTenantList,
}));

describe('loadUserTenantOptions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('uses only the current tenant in tenant context without loading restricted tenant APIs', async () => {
    await expect(
      loadUserTenantOptions({
        accessCodes: ['system:user:query'],
        currentTenant: { code: 'alpha', id: 101, name: 'Alpha Tenant' },
        isPlatform: false,
        tenants: [
          { code: 'alpha', id: 101, name: 'Alpha Tenant' },
          { code: 'beta', id: 102, name: 'Beta Tenant' },
        ],
        userId: 7,
      }),
    ).resolves.toEqual([{ label: 'Alpha Tenant', value: 101 }]);

    expect(authLoginTenants).not.toHaveBeenCalled();
    expect(platformTenantList).not.toHaveBeenCalled();
  });

  it('does not request platform tenants when the current permissions lack the platform list permission', async () => {
    await expect(
      loadUserTenantOptions({
        accessCodes: ['system:user:query'],
        currentTenant: null,
        isPlatform: true,
        tenants: [],
        userId: 7,
      }),
    ).resolves.toEqual([]);

    expect(authLoginTenants).not.toHaveBeenCalled();
    expect(platformTenantList).not.toHaveBeenCalled();
  });

  it('loads login tenant candidates only when the login tenant permission is present', async () => {
    authLoginTenants.mockResolvedValue([
      { code: 'alpha', id: 101, name: 'Alpha Tenant', status: 'active' },
    ]);

    await expect(
      loadUserTenantOptions({
        accessCodes: ['system:tenant:auth:login-tenants'],
        currentTenant: null,
        isPlatform: true,
        tenants: [],
        userId: 7,
      }),
    ).resolves.toEqual([{ label: 'Alpha Tenant', value: 101 }]);

    expect(authLoginTenants).toHaveBeenCalledWith(7);
    expect(platformTenantList).not.toHaveBeenCalled();
  });

  it('allows wildcard permissions to load the platform tenant fallback', async () => {
    authLoginTenants.mockResolvedValue([]);
    platformTenantList.mockResolvedValue({
      items: [
        { code: 'alpha', id: 101, name: 'Alpha Tenant', status: 'active' },
      ],
    });

    await expect(
      loadUserTenantOptions({
        accessCodes: ['*'],
        currentTenant: null,
        isPlatform: true,
        tenants: [],
        userId: 7,
      }),
    ).resolves.toEqual([{ label: 'Alpha Tenant', value: 101 }]);

    expect(authLoginTenants).toHaveBeenCalledWith(7);
    expect(platformTenantList).toHaveBeenCalledWith({
      pageNum: 1,
      pageSize: 100,
      status: 'active',
    });
  });
});
