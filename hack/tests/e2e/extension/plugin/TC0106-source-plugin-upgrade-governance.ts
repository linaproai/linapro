import { execFileSync } from 'node:child_process';
import { readFileSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import type { APIRequestContext } from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  syncPlugins,
  updatePluginStatus,
} from '../../../fixtures/plugin';
import { PluginPage } from '../../../pages/PluginPage';

const pluginID = 'plugin-demo-source';
const originalMenuName = '源码插件示例';
const upgradedMenuName = '源码插件示例升级版';
const originalMenuKey = 'plugin:plugin-demo-source:sidebar-entry';
const upgradedMenuKey = 'plugin:plugin-demo-source:sidebar-entry-upgraded';
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

type UserMenuNode = {
  children?: UserMenuNode[];
  name: string;
  type: string;
};

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: Awaited<ReturnType<APIRequestContext['get']>>, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

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

async function fetchCurrentUserMenus(
  adminApi: APIRequestContext,
): Promise<UserMenuNode[]> {
  const response = await adminApi.get('user/info');
  assertOk(response, '查询当前用户信息失败');
  const payload = unwrapApiData(await response.json());
  return payload?.menus ?? [];
}

function hasMenuName(list: UserMenuNode[], name: string): boolean {
  return list.some((item) => {
    if (item.name === name) {
      return true;
    }
    return hasMenuName(item.children ?? [], name);
  });
}

async function restoreOriginalPluginState(
  adminApi: APIRequestContext,
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
}

test.describe('TC-106 源码插件升级治理', () => {
  test('TC-106a~b: 源码发现更高版本后保持旧生效版本，未显式升级前不自动切换', async ({
    adminContext,
  }) => {
    test.setTimeout(120000);

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
        {
          enabled: 1,
          installed: 1,
        },
        originalManifestContent,
      );

      writeFileSync(pluginManifestPath, upgradedContent, 'utf8');

      await syncPlugins(adminApi);

      const pendingPlugin = await findPlugin(adminApi, pluginID);
      expect(pendingPlugin, `未找到插件: ${pluginID}`).toBeTruthy();
      expect(pendingPlugin?.version).toBe(originalVersion);

      const currentMenus = await fetchCurrentUserMenus(adminApi);
      expect(
        hasMenuName(currentMenus, originalMenuName),
        `应继续保留旧菜单名称: ${originalMenuName}`,
      ).toBeTruthy();
      expect(
        hasMenuName(currentMenus, upgradedMenuName),
        `未显式升级前不应出现新菜单名称: ${upgradedMenuName}`,
      ).toBeFalsy();

      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(pluginID);
      await expect(pluginPage.pluginRow(pluginID)).toContainText(originalVersion);
      await expect(pluginPage.pluginRow(pluginID)).not.toContainText(
        upgradedVersion,
      );
    } finally {
      try {
        await restoreOriginalPluginState(
          adminApi,
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
