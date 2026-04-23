import { test, expect } from '../fixtures/auth';
import {
  waitForBusyIndicatorsToClear,
  waitForDialogReady,
  waitForTableReady,
} from '../support/ui';

async function waitForCounterStable(readCounter: () => number, stableMs: number) {
  let lastValue = readCounter();
  let lastChangedAt = Date.now();

  await expect
    .poll(
      () => {
        const currentValue = readCounter();
        if (currentValue !== lastValue) {
          lastValue = currentValue;
          lastChangedAt = Date.now();
        }
        return Date.now() - lastChangedAt;
      },
      {
        timeout: stableMs + 2000,
        intervals: [100, 100, 200, 200, 500],
      },
    )
    .toBeGreaterThanOrEqual(stableMs);
}

test.describe('Debug Export', () => {
  test('debug export flow with network logging', async ({ adminPage }) => {
    await adminPage.goto('/monitor/operlog');
    await waitForTableReady(adminPage);

    // Find the export button
    const exportBtn = adminPage.getByRole('button', { name: /导\s*出/ });
    await expect(exportBtn).toBeVisible({ timeout: 10000 });

    // Set up network listener before clicking
    const exportRequests: any[] = [];
    const exportResponseLogs: Promise<void>[] = [];
    adminPage.on('request', (request) => {
      if (request.url().includes('export')) {
        console.log('REQUEST:', request.method(), request.url());
        exportRequests.push({ method: request.method(), url: request.url() });
      }
    });

    adminPage.on('response', (response) => {
      if (response.url().includes('export')) {
        exportResponseLogs.push((async () => {
          console.log('RESPONSE:', response.status(), response.url());
          console.log('CONTENT-TYPE:', response.headers()['content-type']);
          try {
            const body = await response.body();
            console.log('BODY LENGTH:', body.length);
            console.log('BODY FIRST 100 BYTES:', body.slice(0, 100));
          } catch (e) {
            console.log('ERROR reading body:', e);
          }
        })());
      }
    });

    // Click export button
    await exportBtn.click();

    // Check if modal appeared
    const modalContent = await waitForDialogReady(adminPage.locator('.ant-modal-wrap'))
      .then((dialog) => dialog.locator('.ant-modal-content').first())
      .catch(() => null);
    const modalVisible = modalContent !== null;
    console.log('MODAL VISIBLE:', modalVisible);

    if (modalContent) {
      // Find confirm button in modal
      const confirmBtn = modalContent.getByRole('button', { name: /确\s*认/ });
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

        const response = await Promise.all([
          responsePromise,
          confirmBtn.first().click(),
        ]).then(([capturedResponse]) => capturedResponse);
        if (response) {
          console.log('GOT RESPONSE:', response.status(), response.url());
          console.log('CONTENT-TYPE:', response.headers()['content-type']);
        }
        await waitForBusyIndicatorsToClear(adminPage);
      }
    }

    await waitForCounterStable(() => exportRequests.length, 500);
    await Promise.allSettled(exportResponseLogs);
    console.log('EXPORT REQUESTS:', JSON.stringify(exportRequests, null, 2));
  });
});
