import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import type { APIRequestContext, Page } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  refreshPluginProjection,
  syncPlugins,
  updatePluginStatus,
} from '../../../fixtures/plugin';
import { PluginPage } from '../../../pages/PluginPage';

const pluginID = 'plugin-demo-source';
const originalMenuName = '源码插件示例';
const upgradedMenuName = '源码插件示例升级版';
const originalMenuKey = 'plugin:plugin-demo-source:sidebar-entry';
const upgradedMenuKey = 'plugin:plugin-demo-source:sidebar-entry-upgraded';
const makeBin = process.env.E2E_MAKE_BIN ?? 'make';
const mysqlBin = process.env.E2E_MYSQL_BIN ?? 'mysql';
const mysqlUser = process.env.E2E_DB_USER ?? 'root';
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? '12345678';
const mysqlDatabase = process.env.E2E_DB_NAME ?? 'lina';
const repoRoot = path.resolve(process.cwd(), '../..');
const pluginManifestPath = path.resolve(
  repoRoot,
  'apps/lina-plugins/plugin-demo-source/plugin.yaml',
);

type OriginalPluginState = {
  enabled: number;
  installed: number;
};

function extractPluginVersion(content: string) {
  const match = content.match(/^version:\s*(v\d+\.\d+\.\d+)\s*$/m);
  if (!match) {
    throw new Error('未能从 plugin-demo-source/plugin.yaml 解析版本号');
  }
  return match[1];
}

function buildHigherVersion(version: string) {
  const match = version.match(/^v(\d+)\.(\d+)\.(\d+)$/);
  if (!match) {
    throw new Error(`不支持的源码插件版本格式: ${version}`);
  }

  const major = Number(match[1]);
  const minor = Number(match[2]);
  return `v${major}.${minor + 1}.0`;
}

function buildUpgradedManifestContent(originalContent: string) {
  const originalVersion = extractPluginVersion(originalContent);
  const upgradedVersion = buildHigherVersion(originalVersion);

  let upgradedContent = originalContent.replace(
    /^version:\s*.+$/m,
    `version: ${upgradedVersion}`,
  );
  upgradedContent = upgradedContent.replaceAll(originalMenuKey, upgradedMenuKey);
  upgradedContent = upgradedContent.replace(
    /(- key: plugin:plugin-demo-source:sidebar-entry-upgraded[\s\S]*?\n\s+name:\s+)[^\n]+/,
    `$1${upgradedMenuName}`,
  );

  return {
    originalVersion,
    upgradedContent,
    upgradedVersion,
  };
}

function resetSourcePluginGovernance(pluginId: string) {
  const escapedPluginID = pluginId.replaceAll("'", "''");
  const menuKeyPattern = `plugin:${escapedPluginID}:%`;

  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      '-e',
      [
        `DELETE FROM sys_role_menu WHERE menu_id IN (SELECT id FROM (SELECT id FROM sys_menu WHERE menu_key LIKE '${menuKeyPattern}') AS plugin_menu_ids);`,
        `DELETE FROM sys_menu WHERE menu_key LIKE '${menuKeyPattern}';`,
        `DELETE FROM sys_plugin_state WHERE plugin_id = '${escapedPluginID}';`,
        `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedPluginID}';`,
        `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedPluginID}';`,
        `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedPluginID}';`,
        `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedPluginID}';`,
        `DELETE FROM sys_plugin WHERE plugin_id = '${escapedPluginID}';`,
      ].join(' '),
    ],
    {
      stdio: 'ignore',
    },
  );
}

async function restoreOriginalPluginState(
  adminApi: APIRequestContext,
  adminPage: Page,
  originalState: OriginalPluginState,
  originalManifestContent: string,
) {
  writeFileSync(pluginManifestPath, originalManifestContent, 'utf8');
  resetSourcePluginGovernance(pluginID);

  await syncPlugins(adminApi);
  if (originalState.installed === 1) {
    await installPlugin(adminApi, pluginID);
    if (originalState.enabled === 1) {
      await updatePluginStatus(adminApi, pluginID, true);
    }
  }

  await refreshPluginProjection(adminPage);
}

test.describe('TC-106 源码插件升级治理', () => {
  test('TC-106a~c: 源码发现更高版本后保持旧生效版本，执行 make upgrade 后才正式切换', async ({
    adminContext,
  }) => {
    test.setTimeout(180000);

    const adminPage = await adminContext.newPage();
    const pluginPage = new PluginPage(adminPage);
    const adminApi = await createAdminApiContext();
    const originalManifestContent = readFileSync(pluginManifestPath, 'utf8');
    const { originalVersion, upgradedContent, upgradedVersion } =
      buildUpgradedManifestContent(originalManifestContent);

    let originalState: OriginalPluginState = {
      enabled: 0,
      installed: 0,
    };

    try {
      await syncPlugins(adminApi);
      const originalPlugin = await findPlugin(adminApi, pluginID);
      originalState = {
        enabled: originalPlugin?.enabled ?? 0,
        installed: originalPlugin?.installed ?? 0,
      };

      await restoreOriginalPluginState(
        adminApi,
        adminPage,
        {
          enabled: 1,
          installed: 1,
        },
        originalManifestContent,
      );
      await pluginPage.expectSidebarMenuVisible(originalMenuName);

      writeFileSync(pluginManifestPath, upgradedContent, 'utf8');

      await syncPlugins(adminApi);
      await refreshPluginProjection(adminPage);

      const pendingPlugin = await findPlugin(adminApi, pluginID);
      expect(pendingPlugin, `未找到插件: ${pluginID}`).toBeTruthy();
      expect(pendingPlugin?.version).toBe(originalVersion);

      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginID);
      await pluginPage.openPluginDetail(pluginID);
      await expect(pluginPage.pluginDetailModal()).toContainText(
        originalVersion,
      );
      await expect(pluginPage.pluginDetailModal()).not.toContainText(
        upgradedVersion,
      );
      await pluginPage.pluginDetailDialog()
        .locator('.ant-modal-close')
        .click();
      await expect(pluginPage.pluginDetailDialog()).toHaveCount(0);

      await pluginPage.expectSidebarMenuVisible(originalMenuName);
      await pluginPage.expectSidebarMenuHidden(upgradedMenuName);

      const upgradeOutput = execFileSync(
        makeBin,
        [
          'upgrade',
          'confirm=upgrade',
          'scope=source-plugin',
          `plugin=${pluginID}`,
        ],
        {
          cwd: repoRoot,
          encoding: 'utf8',
          env: process.env,
        },
      );
      expect(upgradeOutput).toContain(
        `- upgraded: ${pluginID} ${originalVersion} -> ${upgradedVersion}`,
      );
      expect(upgradeOutput).toContain('Source plugin upgrade completed. executed=1');

      await syncPlugins(adminApi);
      await refreshPluginProjection(adminPage);

      const upgradedPlugin = await findPlugin(adminApi, pluginID);
      expect(upgradedPlugin, `未找到插件: ${pluginID}`).toBeTruthy();
      expect(upgradedPlugin?.version).toBe(upgradedVersion);

      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginID);
      await pluginPage.openPluginDetail(pluginID);
      await expect(pluginPage.pluginDetailModal()).toContainText(upgradedVersion);
      await expect(pluginPage.pluginDetailModal()).not.toContainText(
        originalVersion,
      );
      await pluginPage.pluginDetailDialog()
        .locator('.ant-modal-close')
        .click();
      await expect(pluginPage.pluginDetailDialog()).toHaveCount(0);

      await pluginPage.expectSidebarMenuVisible(upgradedMenuName);
      await pluginPage.expectSidebarMenuHidden(originalMenuName);
    } finally {
      try {
        await restoreOriginalPluginState(
          adminApi,
          adminPage,
          originalState,
          originalManifestContent,
        );
      } finally {
        await adminApi.dispose();
        await adminPage.close();
      }
    }
  });
});
