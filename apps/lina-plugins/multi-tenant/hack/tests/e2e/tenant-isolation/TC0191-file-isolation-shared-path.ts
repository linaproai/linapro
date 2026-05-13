import { test, expect } from '../../support/multi-tenant';
import { scenarioTC0191 } from '../../support/multi-tenant-scenarios';

test.describe('TC-191 文件跨租户隔离', () => {
  test.use({ multiTenantMode: 'multi-tenant-enabled' });

  test('TC-191a: file storage paths include tenant buckets and platform shared paths', async ({ multiTenantMode }) => {
    expect(multiTenantMode).toBe('multi-tenant-enabled');
    await scenarioTC0191();
  });
});
