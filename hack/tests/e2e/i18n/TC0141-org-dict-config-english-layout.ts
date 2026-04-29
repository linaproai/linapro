import type { Locator } from '@playwright/test';

import { test, expect } from '../../fixtures/auth';
import { ensureSourcePluginEnabled } from '../../fixtures/plugin';
import { ConfigPage } from '../../pages/ConfigPage';
import { DictPage } from '../../pages/DictPage';
import {
  createAdminApiContext,
  expectSuccess,
} from '../../support/api/job';
import {
  waitForDialogReady,
  waitForDropdown,
  waitForRouteReady,
  waitForTableReady,
} from '../../support/ui';

type DeptTreeNode = {
  children?: DeptTreeNode[];
  id: number;
  label: string;
};

type ConfigListResult = {
  list: Array<{ key: string; name: string; value: string }>;
  total: number;
};

const chineseTextPattern = /[\u3400-\u9fff]/u;

function flattenDeptTree(nodes: DeptTreeNode[]): DeptTreeNode[] {
  return nodes.flatMap((node) => [node, ...flattenDeptTree(node.children ?? [])]);
}

async function readLineMetrics(locator: Locator) {
  return locator.evaluate((node) => {
    const element = node as HTMLElement;
    const style = getComputedStyle(element);
    const fontSize = Number.parseFloat(style.fontSize || '14') || 14;
    const rawLineHeight = Number.parseFloat(style.lineHeight || '');
    const lineHeight =
      Number.isFinite(rawLineHeight) && rawLineHeight > 0
        ? rawLineHeight
        : fontSize * 1.2;

    return {
      height: element.getBoundingClientRect().height,
      lineHeight,
    };
  });
}

async function expectSingleLine(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const metrics = await readLineMetrics(locator);
  expect(metrics.height, `${label} wraps unexpectedly`).toBeLessThanOrEqual(
    metrics.lineHeight * 1.6,
  );
}

async function expectRadioGroupSingleRow(locator: Locator, label: string) {
  await expect(locator, `${label} should be visible`).toBeVisible();
  const buttons = await locator
    .locator('.ant-radio-button-wrapper')
    .evaluateAll((nodes) =>
      nodes.map((node) => {
        const element = node as HTMLElement;
        return {
          text: element.textContent?.trim() ?? '',
          top: element.getBoundingClientRect().top,
        };
      }),
    );

  expect(buttons.length, `${label} should render status options`).toBeGreaterThanOrEqual(2);
  for (const button of buttons) {
    expect(
      Math.abs(button.top - buttons[0]!.top),
      `${label} option "${button.text}" wraps unexpectedly`,
    ).toBeLessThanOrEqual(2);
  }
}

async function assertConfigLocalized(
  key: string,
  expectedName: string,
  expectedValue?: string,
) {
  const api = await createAdminApiContext();
  try {
    const result = await expectSuccess<ConfigListResult>(
      await api.get(
        `config?pageNum=1&pageSize=20&key=${encodeURIComponent(key)}`,
        { headers: { 'Accept-Language': 'en-US' } },
      ),
    );
    const item = result.list.find((entry) => entry.key === key);
    expect(item?.name).toBe(expectedName);
    expect(item?.name).not.toMatch(chineseTextPattern);
    if (expectedValue !== undefined) {
      expect(item?.value).toBe(expectedValue);
      expect(item?.value).not.toMatch(chineseTextPattern);
    }
  } finally {
    await api.dispose();
  }
}

test.describe('TC-141 Organization, dictionary, and config English layout regression', () => {
  test.beforeEach(async ({ adminPage, mainLayout }) => {
    await adminPage.setViewportSize({ width: 1440, height: 920 });
    await ensureSourcePluginEnabled(adminPage, 'org-center');
    await mainLayout.switchLanguage('English');
  });

  test('TC-141a: Unassigned department is localized in English projections', async ({
    adminPage,
  }) => {
    const api = await createAdminApiContext();
    try {
      const deptTree = await expectSuccess<{ list: DeptTreeNode[] }>(
        await api.get('user/dept-tree', {
          headers: { 'Accept-Language': 'en-US' },
        }),
      );
      const unassignedNode = flattenDeptTree(deptTree.list).find(
        (node) => node.id === 0,
      );
      expect(unassignedNode?.label).toContain('Unassigned Department');
      expect(unassignedNode?.label).not.toContain('未分配部门');
    } finally {
      await api.dispose();
    }

    await adminPage.goto('/system/user');
    await waitForTableReady(adminPage);
    await expect(
      adminPage.getByText(/Unassigned Department/i).first(),
    ).toBeVisible();
    await expect(adminPage.getByText('未分配部门').first()).toHaveCount(0);
  });

  test('TC-141b: Post status options stay on one row in English add form', async ({
    adminPage,
  }) => {
    await adminPage.goto('/system/post');
    await waitForTableReady(adminPage);
    await adminPage.getByRole('button', { name: /Add|新\s*增/i }).first().click();

    const drawer = await waitForDialogReady(adminPage.locator('[role="dialog"]'));
    const statusGroup = drawer
      .locator('.ant-radio-group')
      .filter({ hasText: /Normal|Enabled|Disabled/i })
      .first();
    await expectRadioGroupSingleRow(statusGroup, 'Post status radio group');

    await drawer.getByRole('button', { name: /Cancel|取\s*消/i }).click();
    await drawer.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
  });

  test('TC-141c: Dictionary type/data labels stay single-line and tag style options are translated', async ({
    adminPage,
  }) => {
    const dictPage = new DictPage(adminPage);
    await dictPage.goto();

    await adminPage
      .locator('#dict-type')
      .getByRole('button', { name: /Add|新\s*增/i })
      .click();
    let dialog = await waitForDialogReady(adminPage.locator('[role="dialog"]'));
    await expectSingleLine(
      dialog
        .locator('.ant-form-item-label, label', { hasText: 'Dictionary Type' })
        .first(),
      'Dictionary Type label',
    );
    await dialog.getByRole('button', { name: /Cancel|取\s*消/i }).click();
    await dialog.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});

    await dictPage.fillTypeSearchField('字典类型', 'sys_normal_disable');
    await dictPage.clickTypeSearch();
    await dictPage.clickTypeRow('sys_normal_disable');
    await adminPage
      .locator('#dict-data')
      .getByRole('button', { name: /Add|新\s*增/i })
      .click();

    dialog = await waitForDialogReady(adminPage.locator('[role="dialog"]'));
    await expectSingleLine(
      dialog
        .locator('.ant-form-item-label, label', { hasText: 'Tag Style' })
        .first(),
      'Tag Style label',
    );

    const tagStyleItem = dialog
      .locator('.ant-select-selector')
      .filter({ hasText: 'Please select a tag style' })
      .first();
    await tagStyleItem.click();
    const dropdown = await waitForDropdown(adminPage);
    const dropdownText = await dropdown.innerText();
    expect(dropdownText).toContain('Default');
    expect(dropdownText).toContain('Primary');
    expect(dropdownText).not.toContain('pages.system.dict.data.tagStyle');

    await adminPage.keyboard.press('Escape');
    if (await dialog.isVisible({ timeout: 1000 }).catch(() => false)) {
      await dialog
        .getByRole('button', { name: /Cancel|取\s*消/i })
        .click({ force: true })
        .catch(() => {});
      await dialog.waitFor({ state: 'hidden', timeout: 5000 }).catch(() => {});
    }
  });

  test('TC-141d: Protected login config names and built-in values are localized in English', async ({
    adminPage,
  }) => {
    await assertConfigLocalized('sys.login.blackIPList', 'Login - IP Blacklist');
    await assertConfigLocalized(
      'sys.auth.loginSubtitle',
      'Login - Form Subtitle',
      'Enter your account credentials to start managing your projects',
    );
    await assertConfigLocalized(
      'sys.auth.pageDesc',
      'Login - Page Description',
      'Built for evolving business needs, with an out-of-the-box admin entry point and a flexible pluggable extension model',
    );
    await assertConfigLocalized(
      'sys.auth.pageTitle',
      'Login - Page Title',
      'An AI-native full-stack framework engineered for sustainable delivery',
    );

    const configPage = new ConfigPage(adminPage);
    await configPage.goto();
    await configPage.fillSearchField('参数键名', 'sys.auth.pageTitle');
    await configPage.clickSearch();

    const row = configPage.findRowByExactKey('sys.auth.pageTitle');
    await expect(row).toContainText('Login - Page Title');
    await expect(row).toContainText('An AI-native full-stack framework engineered for sustainable delivery');
    await expect(row).not.toContainText('登录展示-页面标题');
    await expect(row).not.toContainText('面向可持续交付的 AI 原生全栈框架');

    await waitForRouteReady(adminPage);
  });
});
