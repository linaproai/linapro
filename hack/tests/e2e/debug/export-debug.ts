import { test, expect } from '../../fixtures/auth';

test.describe('Debug Export', () => {
  test('debug export flow with network logging', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await adminPage.waitForLoadState('networkidle');
    await adminPage.waitForTimeout(500);

    // Find the export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible({ timeout: 10000 });

    // Set up network listener before clicking
    const exportRequests: any[] = [];
    adminPage.on('request', (request) => {
      if (request.url().includes('export')) {
        console.log('REQUEST:', request.method(), request.url());
        exportRequests.push({ method: request.method(), url: request.url() });
      }
    });

    adminPage.on('response', async (response) => {
      if (response.url().includes('export')) {
        console.log('RESPONSE:', response.status(), response.url());
        console.log('CONTENT-TYPE:', response.headers()['content-type']);
        try {
          const body = await response.body();
          console.log('BODY LENGTH:', body.length);
          console.log('BODY FIRST 100 BYTES:', body.slice(0, 100));
        } catch (e) {
          console.log('ERROR reading body:', e);
        }
      }
    });

    // Click export button
    await exportBtn.click();
    await adminPage.waitForTimeout(500);

    // Check if modal appeared
    const modalContent = adminPage.locator('.ant-modal-content');
    const modalVisible = await modalContent.isVisible().catch(() => false);
    console.log('MODAL VISIBLE:', modalVisible);

    if (modalVisible) {
      // Find confirm button in modal
      const confirmBtn = adminPage.locator('.ant-modal-content').getByRole('button', { name: /确\s*认/ });
      console.log('CONFIRM BTN COUNT:', await confirmBtn.count());

      if (await confirmBtn.count() > 0) {
        // Click confirm and wait for network response
        const responsePromise = adminPage.waitForResponse(
          (resp) => resp.url().includes('export'),
          { timeout: 10000 }
        ).catch((e) => {
          console.log('WAITFORRESPONSE ERROR:', e.message);
          return null;
        });

        await confirmBtn.first().click();

        const response = await responsePromise;
        if (response) {
          console.log('GOT RESPONSE:', response.status(), response.url());
          console.log('CONTENT-TYPE:', response.headers()['content-type']);
        }
      }
    }

    await adminPage.waitForTimeout(2000);
    console.log('EXPORT REQUESTS:', JSON.stringify(exportRequests, null, 2));
  });
});