import { test, expect } from '../../../fixtures/auth';
import { createAdminApiContext, listPlugins } from '../../../fixtures/plugin';

test.describe('TC-234 Host-only plugin workspace', () => {
  test.skip(
    process.env.E2E_HOST_ONLY_PLUGINS !== '1',
    'Host-only plugin workspace assertion runs only in host-only validation.',
  );

  test('TC0234a: source plugin list is empty without the official plugin workspace', async () => {
    const adminApi = await createAdminApiContext();
    try {
      const plugins = await listPlugins(adminApi);
      const sourcePlugins = plugins.filter((plugin) => plugin.type === 'source');
      expect(sourcePlugins).toEqual([]);
    } finally {
      await adminApi.dispose();
    }
  });
});
