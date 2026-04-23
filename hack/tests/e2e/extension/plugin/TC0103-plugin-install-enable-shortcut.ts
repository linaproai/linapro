import { execFileSync } from 'node:child_process';
import { mkdirSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import path from 'node:path';

import type {
  APIRequestContext,
  APIResponse,
  Page,
  Route,
} from '@playwright/test';

import { test, expect } from '../../../fixtures/auth';
import {
  createAdminApiContext,
  findPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
  updatePluginStatus,
} from '../../../fixtures/plugin';
import { LoginPage } from '../../../pages/LoginPage';
import { PluginPage } from '../../../pages/PluginPage';

const mysqlBin = process.env.E2E_MYSQL_BIN ?? 'mysql';
const mysqlUser = process.env.E2E_DB_USER ?? 'root';
const mysqlPassword = process.env.E2E_DB_PASSWORD ?? '12345678';
const mysqlDatabase = process.env.E2E_DB_NAME ?? 'lina';

const dynamicPluginID = 'plugin-install-enable-shortcut-e2e';
const dynamicPluginVersion = 'v0.1.0';
const dynamicPluginName = 'Plugin Install Enable Shortcut E2E';
const dynamicStoragePath = 'plugin-shortcut/files';
const sourcePluginID = 'content-notice';
const sourcePluginMenuTitle = '通知公告';

const installOnlyRoleName = '插件安装权限角色';
const installOnlyRoleKey = 'plugin_install_only_role';
const installOnlyUsername = 'plugin_install_only_user';
const installOnlyPassword = 'runtime123';
const installOnlyNickname = '插件安装权限用户';

function unwrapApiData(payload: any) {
  if (payload && typeof payload === 'object' && 'data' in payload) {
    return payload.data;
  }
  return payload;
}

function assertOk(response: APIResponse, message: string) {
  expect(response.ok(), `${message}, status=${response.status()}`).toBeTruthy();
}

async function expectApiSuccess<T = any>(
  response: APIResponse,
  message: string,
): Promise<T> {
  assertOk(response, message);

  const payload = (await response.json()) as {
    code?: number;
    data?: T;
    message?: string;
  };
  expect(
    payload?.code,
    `${message}, business code=${payload?.code}, business message=${payload?.message ?? ''}`,
  ).toBe(0);
  return (payload?.data ?? null) as T;
}

function repoRoot() {
  return path.resolve(process.cwd(), '../..');
}

function tempDir() {
  return path.join(repoRoot(), 'temp');
}

function artifactPath() {
  return path.join(tempDir(), `${dynamicPluginID}.wasm`);
}

function runtimeStorageArtifactPath() {
  return path.join(tempDir(), 'output', `${dynamicPluginID}.wasm`);
}

function querySQLRows(sql: string) {
  return execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      '-N',
      '-B',
      '-e',
      sql,
    ],
    { encoding: 'utf8' },
  )
    .trim()
    .split('\n')
    .filter(Boolean);
}

function execSQL(statements: string[]) {
  execFileSync(
    mysqlBin,
    [
      `-u${mysqlUser}`,
      `-p${mysqlPassword}`,
      mysqlDatabase,
      '-e',
      statements.join(' '),
    ],
    { stdio: 'ignore' },
  );
}

function cleanupDynamicPluginRows() {
  const escapedID = dynamicPluginID.replaceAll("'", "''");
  execSQL([
    `DELETE FROM sys_plugin_node_state WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_resource_ref WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_migration WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin_release WHERE plugin_id = '${escapedID}';`,
    `DELETE FROM sys_plugin WHERE plugin_id = '${escapedID}';`,
  ]);
}

function cleanupUserAndRoleRows() {
  const escapedUsername = installOnlyUsername.replaceAll("'", "''");
  const escapedRoleKey = installOnlyRoleKey.replaceAll("'", "''");
  execSQL([
    `DELETE FROM sys_user_role WHERE user_id IN (SELECT user_ids.id FROM (SELECT id FROM sys_user WHERE username = '${escapedUsername}') AS user_ids);`,
    `DELETE FROM sys_user WHERE username = '${escapedUsername}';`,
    `DELETE FROM sys_role_menu WHERE role_id IN (SELECT role_ids.id FROM (SELECT id FROM sys_role WHERE \`key\` = '${escapedRoleKey}') AS role_ids);`,
    `DELETE FROM sys_role WHERE \`key\` = '${escapedRoleKey}';`,
  ]);
}

function cleanupWorkspace() {
  rmSync(artifactPath(), { force: true });
  rmSync(runtimeStorageArtifactPath(), { force: true });
}

function writeULEB128(buffer: number[], value: number) {
  let current = value >>> 0;
  while (true) {
    let byte = current & 0x7f;
    current >>>= 7;
    if (current !== 0) {
      byte |= 0x80;
    }
    buffer.push(byte);
    if (current === 0) {
      return;
    }
  }
}

function appendCustomSection(buffer: number[], name: string, payload: Buffer) {
  const section: number[] = [];
  writeULEB128(section, Buffer.byteLength(name));
  section.push(...Buffer.from(name));
  section.push(...payload);

  buffer.push(0x00);
  writeULEB128(buffer, section.length);
  buffer.push(...section);
}

function writeRuntimeArtifact() {
  mkdirSync(tempDir(), { recursive: true });

  const bytes: number[] = [0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00];
  appendCustomSection(
    bytes,
    'lina.plugin.manifest',
    Buffer.from(
      JSON.stringify({
        description: 'Plugin used to verify the install-and-enable shortcut.',
        id: dynamicPluginID,
        name: dynamicPluginName,
        type: 'dynamic',
        version: dynamicPluginVersion,
      }),
    ),
  );
  appendCustomSection(
    bytes,
    'lina.plugin.dynamic',
    Buffer.from(
      JSON.stringify({
        abiVersion: 'v1',
        frontendAssetCount: 0,
        runtimeKind: 'wasm',
        sqlAssetCount: 0,
      }),
    ),
  );
  appendCustomSection(
    bytes,
    'lina.plugin.backend.host-services',
    Buffer.from(
      JSON.stringify([
        {
          methods: ['info.now'],
          service: 'runtime',
        },
        {
          methods: ['list', 'get'],
          resources: {
            paths: [dynamicStoragePath],
          },
          service: 'storage',
        },
      ]),
    ),
  );

  writeFileSync(artifactPath(), Buffer.from(bytes));
}

async function uploadDynamicPlugin(adminApi: APIRequestContext) {
  const response = await adminApi.post('plugins/dynamic/package', {
    multipart: {
      overwriteSupport: '1',
      file: {
        buffer: readFileSync(artifactPath()),
        mimeType: 'application/wasm',
        name: path.basename(artifactPath()),
      },
    },
  });
  await expectApiSuccess(response, '管理员上传动态插件失败');
}

async function prepareDynamicPlugin(adminApi: APIRequestContext) {
  cleanupDynamicPluginRows();
  cleanupWorkspace();
  writeRuntimeArtifact();
  await uploadDynamicPlugin(adminApi);
  await expect
    .poll(async () => Boolean(await findPlugin(adminApi, dynamicPluginID)))
    .toBe(true);
}

async function resetSourcePlugin(
  adminApi: APIRequestContext,
  installed: boolean,
  enabled = false,
) {
  await syncPlugins(adminApi);
  let plugin = await findPlugin(adminApi, sourcePluginID);
  if (installed) {
    if (plugin?.installed !== 1) {
      await installPlugin(adminApi, sourcePluginID);
      plugin = await findPlugin(adminApi, sourcePluginID);
    }
    if (enabled && plugin?.enabled !== 1) {
      await updatePluginStatus(adminApi, sourcePluginID, true);
    }
    if (!enabled && plugin?.enabled === 1) {
      await updatePluginStatus(adminApi, sourcePluginID, false);
    }
    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.installed ?? 0)
      .toBe(1);
    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.enabled ?? -1)
      .toBe(enabled ? 1 : 0);
    return;
  }
  if (plugin?.installed === 1) {
    await uninstallPlugin(adminApi, sourcePluginID);
  }
  await expect
    .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.installed ?? 1)
    .toBe(0);
}

function lookupMenuID(menuKey: string) {
  const rows = querySQLRows(
    `SELECT id FROM sys_menu WHERE menu_key = '${menuKey.replaceAll("'", "''")}' LIMIT 1;`,
  );
  expect(rows.length, `未找到菜单: ${menuKey}`).toBe(1);
  return Number.parseInt(rows[0]!, 10);
}

async function getAdminDeptID(adminApi: APIRequestContext) {
  const response = await adminApi.get('user/1');
  const payload = await expectApiSuccess<{ deptId?: number }>(
    response,
    '查询管理员详情失败',
  );
  return payload?.deptId && payload.deptId > 0 ? payload.deptId : undefined;
}

async function createInstallOnlyRole(adminApi: APIRequestContext) {
  const menuIDs = [
    lookupMenuID('extension'),
    lookupMenuID('extension:plugin:list'),
    lookupMenuID('extension:plugin:query'),
    lookupMenuID('extension:plugin:install'),
  ];

  const response = await adminApi.post('role', {
    data: {
      dataScope: 3,
      key: installOnlyRoleKey,
      menuIds: menuIDs,
      name: installOnlyRoleName,
      remark: 'Plugin management install-only role',
      sort: 10,
      status: 1,
    },
  });
  const payload = await expectApiSuccess<{ id?: number }>(
    response,
    '创建安装权限角色失败',
  );
  expect(payload?.id, '角色创建成功后应返回角色ID').toBeTruthy();
  return payload.id as number;
}

async function createInstallOnlyUser(
  adminApi: APIRequestContext,
  deptID: number | undefined,
  roleID: number,
) {
  const data: Record<string, any> = {
    nickname: installOnlyNickname,
    password: installOnlyPassword,
    roleIds: [roleID],
    status: 1,
    username: installOnlyUsername,
  };
  if (deptID !== undefined) {
    data.deptId = deptID;
  }

  const response = await adminApi.post('user', {
    data,
  });
  const payload = await expectApiSuccess<{ id?: number }>(
    response,
    '创建安装权限用户失败',
  );
  expect(payload?.id, '用户创建成功后应返回用户ID').toBeTruthy();
  return payload.id as number;
}

async function loginAsInstallOnlyUser(page: Page) {
  const loginPage = new LoginPage(page);
  await loginPage.goto();
  await loginPage.loginAndWaitForRedirect(
    installOnlyUsername,
    installOnlyPassword,
  );
}

async function interceptEnableFailure(route: Route) {
  await route.fulfill({
    body: JSON.stringify({
      code: 1,
      message: '模拟启用失败',
    }),
    contentType: 'application/json',
    status: 500,
  });
}

test.describe('TC-103 插件安装弹窗快捷启用', () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    cleanupDynamicPluginRows();
    cleanupUserAndRoleRows();
    cleanupWorkspace();
    await resetSourcePlugin(adminApi, true, true);
    await adminApi.dispose();
  });

  test.afterEach(async () => {
    cleanupDynamicPluginRows();
    cleanupUserAndRoleRows();
    cleanupWorkspace();
    await resetSourcePlugin(adminApi, true, true);
  });

  test('TC0103a: 动态插件可在授权审查弹窗中直接安装并启用', async ({
    adminPage,
  }) => {
    await prepareDynamicPlugin(adminApi);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(dynamicPluginID);
    await pluginPage.openInstallAuthorization(dynamicPluginID);
    await expect(pluginPage.hostServiceAuthInstallAndEnableButton()).toBeVisible();

    await pluginPage.confirmInstallAndEnable();

    await expect
      .poll(async () => (await findPlugin(adminApi, dynamicPluginID))?.installed ?? 0)
      .toBe(1);
    await expect
      .poll(async () => (await findPlugin(adminApi, dynamicPluginID))?.enabled ?? 0)
      .toBe(1);
    await expect(
      pluginPage.messageNotice('插件已安装并启用'),
    ).toBeVisible();

    const plugin = await findPlugin(adminApi, dynamicPluginID);
    expect(plugin?.enabled).toBe(1);
    await pluginPage.searchByPluginId(dynamicPluginID);
    await expect(pluginPage.pluginEnabledSwitch(dynamicPluginID)).toHaveAttribute(
      'aria-checked',
      'true',
    );
  });

  test('TC0103b: 安装并启用的第二步失败时保留已安装未启用状态', async ({
    adminPage,
  }) => {
    await prepareDynamicPlugin(adminApi);

    const enablePath = `**/api/v1/plugins/${dynamicPluginID}/enable`;
    await adminPage.route(enablePath, interceptEnableFailure);
    try {
      const pluginPage = new PluginPage(adminPage);
      await pluginPage.gotoManage();
      await pluginPage.searchByPluginId(dynamicPluginID);
      await pluginPage.installAndEnablePlugin(dynamicPluginID);

      await expect(
        pluginPage.messageNotice('插件已安装，但启用失败，请稍后重试。'),
      ).toBeVisible();
      await expect(pluginPage.messageNotice('模拟启用失败')).toBeVisible();
      await expect
        .poll(async () => (await findPlugin(adminApi, dynamicPluginID))?.installed ?? 0)
        .toBe(1);
      await expect
        .poll(async () => (await findPlugin(adminApi, dynamicPluginID))?.enabled ?? 1)
        .toBe(0);

      await pluginPage.searchByPluginId(dynamicPluginID);
      await expect(
        pluginPage.pluginEnabledSwitch(dynamicPluginID),
      ).toHaveAttribute('aria-checked', 'false');
      await pluginPage.expectUninstallActionVisible(dynamicPluginID);
    } finally {
      await adminPage.unroute(enablePath, interceptEnableFailure);
    }
  });

  test('TC0103c: 源码插件可在安装弹窗中直接安装并启用', async ({
    adminPage,
  }) => {
    await resetSourcePlugin(adminApi, false);

    const pluginPage = new PluginPage(adminPage);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(sourcePluginID);
    await pluginPage.openInstallAuthorization(sourcePluginID);
    await expect(pluginPage.hostServiceAuthInstallAndEnableButton()).toBeVisible();

    await pluginPage.confirmInstallAndEnable();

    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.installed ?? 0)
      .toBe(1);
    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.enabled ?? 0)
      .toBe(1);
    await expect(
      pluginPage.messageNotice('插件已安装并启用'),
    ).toBeVisible();

    await pluginPage.searchByPluginId(sourcePluginID);
    await expect(pluginPage.pluginEnabledSwitch(sourcePluginID)).toHaveAttribute(
      'aria-checked',
      'true',
    );
    await adminPage.goto('/system/notice');
    await expect(adminPage.getByText(sourcePluginMenuTitle).first()).toBeVisible();
  });

  test('TC0103d: 仅具备安装权限的用户看不到安装并启用按钮', async ({
    page,
  }) => {
    await resetSourcePlugin(adminApi, false);
    cleanupUserAndRoleRows();

    const deptID = await getAdminDeptID(adminApi);
    const roleID = await createInstallOnlyRole(adminApi);
    await createInstallOnlyUser(adminApi, deptID, roleID);

    await loginAsInstallOnlyUser(page);

    const pluginPage = new PluginPage(page);
    await pluginPage.gotoManage();
    await pluginPage.searchByPluginId(sourcePluginID);
    await pluginPage.openInstallAuthorization(sourcePluginID);
    await expect(
      pluginPage.hostServiceAuthInstallAndEnableButton(),
    ).toHaveCount(0);
    await expect(pluginPage.hostServiceAuthConfirmButton()).toBeVisible();

    await pluginPage.confirmHostServiceAuthorization();

    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.installed ?? 0)
      .toBe(1);
    await expect
      .poll(async () => (await findPlugin(adminApi, sourcePluginID))?.enabled ?? 1)
      .toBe(0);

    await pluginPage.searchByPluginId(sourcePluginID);
    await expect(pluginPage.pluginEnabledSwitch(sourcePluginID)).toHaveClass(
      /ant-switch-disabled/,
    );
  });
});
