import type { APIRequestContext } from '@playwright/test';

import { execFileSync } from 'node:child_process';
import { existsSync } from 'node:fs';
import path from 'node:path';

import { test, expect } from '../../../fixtures/auth';
import { MenuPage } from '../../../pages/MenuPage';
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from '../../../support/api/job';

const pluginID = 'plugin-demo-dynamic';
const repoRoot = path.resolve(process.cwd(), '../..');
const runtimeArtifactPath = path.join(
  repoRoot,
  'temp',
  'output',
  `${pluginID}.wasm`,
);

type MenuNode = {
  children?: MenuNode[];
  id: number;
  name: string;
  path?: string;
  perms?: string;
  type: string;
};

type FlatMenuNode = {
  ancestors: MenuNode[];
  node: MenuNode;
};

let adminApi: APIRequestContext;
let originalInstalled = 0;
let originalEnabled = 0;

function ensureRuntimePluginArtifact() {
  if (existsSync(runtimeArtifactPath)) {
    return;
  }
  execFileSync('make', ['wasm', `p=${pluginID}`, 'out=../../temp/output'], {
    cwd: path.join(repoRoot, 'apps', 'lina-plugins'),
    stdio: 'inherit',
  });
}

function flattenMenus(nodes: MenuNode[], ancestors: MenuNode[] = []): FlatMenuNode[] {
  return nodes.flatMap((node) => [
    { ancestors, node },
    ...flattenMenus(node.children ?? [], [...ancestors, node]),
  ]);
}

async function ensurePluginInstalledAndEnabled() {
  await syncPlugins(adminApi);
  let plugin = await getPlugin(adminApi, pluginID);
  originalInstalled = plugin.installed;
  originalEnabled = plugin.enabled;

  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID);
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
}

async function restorePluginState() {
  let plugin = await getPlugin(adminApi, pluginID);

  if (originalInstalled !== 1) {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
      plugin = await getPlugin(adminApi, pluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, pluginID);
    }
    return;
  }

  if (originalEnabled !== 1 && plugin.enabled === 1) {
    await disablePlugin(adminApi, pluginID);
  }
}

test.describe('TC-140 Dynamic plugin permission menu tree regression', () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await ensurePluginInstalledAndEnabled();
  });

  test.afterAll(async () => {
    try {
      await restorePluginState();
    } finally {
      await adminApi.dispose();
    }
  });

  test('TC-140a: Dynamic route permission buttons are children of the plugin menu', async () => {
    const menuData = await expectSuccess<{ list: MenuNode[] }>(
      await adminApi.get('menu'),
    );
    const flatMenus = flattenMenus(menuData.list);
    const pluginMenu = flatMenus.find(
      ({ node }) =>
        node.perms === `${pluginID}:view` ||
        (node.path ?? '').includes(`/plugin-assets/${pluginID}/`),
    );
    expect(pluginMenu, 'missing dynamic plugin main menu').toBeTruthy();

    const dynamicRouteButtons = flatMenus.filter(({ node }) => {
      return (
        node.type === 'B' &&
        (node.perms ?? '').startsWith(`${pluginID}:`) &&
        node.perms !== `${pluginID}:view`
      );
    });
    expect(dynamicRouteButtons.length).toBeGreaterThan(0);

    for (const item of dynamicRouteButtons) {
      expect(
        item.ancestors.map((ancestor) => ancestor.id),
        `${item.node.name} should be nested below the dynamic plugin menu`,
      ).toContain(pluginMenu!.node.id);
    }
  });

  test('TC-140b: Menu tree expandable names show pointer cursor and toggle on click', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);

    await mainLayout.switchLanguage('简体中文');
    await menuPage.goto();
    await menuPage.collapseAll();

    const accessRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: '权限管理' })
      .first();
    await expect(accessRow).toBeVisible();

    const accessNameCell = accessRow
      .locator('.system-menu-name-column .vxe-cell')
      .first();
    await accessNameCell.hover();
    await expect
      .poll(async () =>
        accessNameCell.evaluate((node) => getComputedStyle(node).cursor),
      )
      .toBe('pointer');

    const childUserRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: '用户管理' })
      .first();
    await expect(childUserRow).toBeHidden();

    await accessNameCell.click();
    await expect(childUserRow).toBeVisible();

    await accessNameCell.click();
    await expect(childUserRow).toBeHidden();
  });

  test('TC-140c: Dynamic plugin button names are readable in English menu management', async ({
    adminPage,
    mainLayout,
  }) => {
    const menuPage = new MenuPage(adminPage);

    await mainLayout.switchLanguage('English');
    await menuPage.goto();
    await menuPage.expandAll();

    const pluginRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: 'Dynamic Plugin Demo' })
      .first();
    await expect(pluginRow).toBeVisible();

    const recordCreateRow = adminPage
      .locator('.vxe-body--row:visible', { hasText: 'Record Create' })
      .first();
    if (!(await recordCreateRow.isVisible({ timeout: 1000 }).catch(() => false))) {
      await pluginRow.locator('.system-menu-name-column .vxe-cell').first().click();
    }

    await expect(recordCreateRow).toBeVisible();
    await expect(
      adminPage.locator('.vxe-body--row:visible', {
        hasText: /Dynamic Route Permission:plugin-demo-dynamic/u,
      }),
    ).toHaveCount(0);
  });
});
