import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0193 } from '../../support/multi-tenant-scenarios';

test.describe('TC-193 任务跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-193a: tenant jobs and platform built-in jobs persist separately', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0193();
  });
});
