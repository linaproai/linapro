import type { Locator } from '@host-tests/support/playwright';

import { test, expect } from '@host-tests/fixtures/auth';
import { waitForRouteReady } from '@host-tests/support/ui';

const monitorPayload = {
  code: 0,
  data: {
    dbInfo: {
      idle: 1,
      inUse: 1,
      maxOpenConns: 100,
      openConns: 2,
      version: '8.0.35',
    },
    nodes: [
      {
        collectAt: '2026-04-29 10:00:00',
        cpu: {
          cores: 8,
          modelName: 'Apple M-series',
          usagePercent: 28.4,
        },
        disks: [
          {
            free: 549_755_813_888,
            fsType: 'APFS',
            path: '/',
            total: 1_099_511_627_776,
            usagePercent: 50,
            used: 549_755_813_888,
          },
        ],
        goInfo: {
          gcPauseNs: 120_000,
          gfVersion: 'v2.10.0',
          goroutines: 88,
          processCpu: 2.2,
          processMemory: 1.8,
          serviceUptime: '1h 5m',
          version: 'go1.24.0',
        },
        memory: {
          available: 8_589_934_592,
          total: 17_179_869_184,
          usagePercent: 50,
          used: 8_589_934_592,
        },
        network: {
          bytesRecv: 2_147_483_648,
          bytesSent: 1_073_741_824,
          recvRate: 4096,
          sendRate: 2048,
        },
        nodeIp: '127.0.0.1',
        nodeName: 'local-dev-node',
        server: {
          arch: 'arm64',
          bootTime: '2026-04-29 08:00:00',
          hostname: 'local-dev-node',
          os: 'darwin',
          startTime: '2026-04-29 09:00:00',
          uptime: 7200,
        },
      },
    ],
  },
  message: 'OK',
};

async function expectSingleLine(locator: Locator, label: string) {
  await locator.scrollIntoViewIfNeeded();
  await expect(locator, `${label} should be visible`).toBeVisible();
  const metrics = await locator.evaluate((node) => {
    const element = node as HTMLElement;
    const style = getComputedStyle(element);

    return {
      text: element.textContent?.trim() ?? '',
      whiteSpace: style.whiteSpace,
    };
  });
  expect(metrics.text, `${label} should not be empty`).not.toBe('');
  expect(metrics.text, `${label} should not contain line breaks`).not.toContain(
    '\n',
  );
  expect(metrics.whiteSpace, `${label} should prevent wrapping`).toBe('nowrap');
}

test.describe('TC-142 Server monitor English disk table regression', () => {
  test('TC-142a: Disk table key columns keep English headers and values on one line', async ({
    adminPage,
    mainLayout,
  }) => {
    await adminPage.setViewportSize({ width: 1366, height: 900 });
    await adminPage.route('**/api/v1/monitor/server**', async (route) => {
      await route.fulfill({
        body: JSON.stringify(monitorPayload),
        contentType: 'application/json',
        status: 200,
      });
    });

    await mainLayout.switchLanguage('English');
    await adminPage.goto('/monitor/server');
    await waitForRouteReady(adminPage);

    const diskTable = adminPage.locator('.server-monitor-disk-table').first();
    await diskTable.scrollIntoViewIfNeeded();
    await expect(diskTable).toBeVisible();

    for (const header of ['File System', 'Total', 'Used', 'Available']) {
      await expectSingleLine(
        diskTable
          .locator('thead th:visible', { hasText: new RegExp(`^${header}$`) })
          .first(),
        `${header} header`,
      );
    }

    const firstRowCells = diskTable
      .locator('tbody tr.ant-table-row:visible')
      .first()
      .locator('td:visible');
    await expectSingleLine(firstRowCells.nth(1), 'File System value');
    await expectSingleLine(firstRowCells.nth(2), 'Total value');
    await expectSingleLine(firstRowCells.nth(3), 'Used value');
    await expectSingleLine(firstRowCells.nth(4), 'Available value');

    await diskTable.screenshot({
      path: 'test-results/TC0142-server-monitor-disk-table-english.png',
    });
  });
});
