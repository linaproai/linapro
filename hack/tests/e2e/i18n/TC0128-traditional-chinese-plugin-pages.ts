import type { APIRequestContext } from "@playwright/test";

import { execFileSync } from "node:child_process";
import { rmSync } from "node:fs";
import path from "node:path";

import { test, expect } from "../../fixtures/auth";
import { ensureSourcePluginEnabled } from "../../fixtures/plugin";
import { PluginPage } from "../../pages/PluginPage";
import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "../../support/api/job";
import { waitForRouteReady } from "../../support/ui";

const dynamicPluginID = "plugin-demo-dynamic";
const repoRoot = path.resolve(process.cwd(), "../..");
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  dynamicPluginID,
  "runtime",
  `${dynamicPluginID}.wasm`,
);

const rawI18nKeyPattern =
  /\b(?:authentication|common|config|demos|dict|error|job|menu|notify|page|pages|plugin|preferences|profile|ui|validation)\.[A-Za-z][A-Za-z0-9_.:-]+\b/g;

const sourcePluginIDs = [
  "org-center",
  "content-notice",
  "monitor-loginlog",
  "monitor-online",
  "monitor-operlog",
  "monitor-server",
  "plugin-demo-source",
] as const;

const sourcePluginAuditCases = [
  {
    forbiddenTexts: [
      "Dept Name",
      "Department Management",
      "部门名称",
      "部门管理",
    ],
    path: "/system/dept",
    visibleTexts: ["部門管理", "部門名稱"],
  },
  {
    forbiddenTexts: ["Position List", "Post Name", "岗位列表", "岗位名称"],
    path: "/system/post",
    visibleTexts: ["崗位管理", "崗位列表"],
  },
  {
    forbiddenTexts: ["Notice Title", "Notice List", "通知标题", "通知列表"],
    path: "/system/notice",
    visibleTexts: ["通知公告", "公告標題"],
  },
  {
    forbiddenTexts: ["User Account", "Login Time", "用户账号", "登录时间"],
    path: "/monitor/loginlog",
    visibleTexts: ["登錄日誌", "用戶賬號"],
  },
  {
    forbiddenTexts: ["Login Account", "Force Logout", "登录账号", "强退"],
    path: "/monitor/online",
    visibleTexts: ["在線用戶", "用戶賬號"],
  },
  {
    forbiddenTexts: ["Module Name", "Operation Summary", "模块名称"],
    path: "/monitor/operlog",
    visibleTexts: ["操作日誌", "模塊名稱"],
  },
  {
    forbiddenTexts: [
      "Service Uptime",
      "Basic Info",
      "服务运行时长",
      "基础信息",
    ],
    path: "/monitor/server",
    visibleTexts: ["服務監控", "服務運行時長"],
  },
  {
    forbiddenTexts: ["Source Plugin Demo", "Demo Records"],
    path: "/plugin-demo-source-sidebar-entry",
    visibleTexts: ["源碼插件示例", "示例記錄"],
  },
] as const;

let adminApi: APIRequestContext;
let originalDynamicInstalled = 0;
let originalDynamicEnabled = 0;

function assertNoRawI18nKeys(bodyText: string, pathLabel: string) {
  const rawKeys = [...new Set(bodyText.match(rawI18nKeyPattern) || [])];
  expect(rawKeys, `${pathLabel} still shows raw i18n keys`).toEqual([]);
}

function ensureRuntimePluginArtifact() {
  execFileSync(
    "make",
    ["wasm", `p=${dynamicPluginID}`, "out=../../temp/output"],
    {
      cwd: path.join(repoRoot, "apps", "lina-plugins"),
      stdio: "inherit",
    },
  );
  rmSync(legacyRuntimeArtifactPath, { force: true });
}

async function ensureDynamicPluginInstalledAndEnabled() {
  // Re-install the current artifact even when the plugin is already installed.
  // Same-version installs refresh the active release assets, keeping this test
  // tied to the just-built mount.js instead of a stale database release.
  await installPlugin(adminApi, dynamicPluginID);
  const refreshedPlugin = await getPlugin(adminApi, dynamicPluginID);
  if (refreshedPlugin.enabled !== 1) {
    await enablePlugin(adminApi, dynamicPluginID);
  }
}

async function restoreDynamicPluginState() {
  let plugin = await getPlugin(adminApi, dynamicPluginID);

  if (originalDynamicInstalled !== 1) {
    if (plugin.enabled === 1) {
      await disablePlugin(adminApi, dynamicPluginID);
      plugin = await getPlugin(adminApi, dynamicPluginID);
    }
    if (plugin.installed === 1) {
      await uninstallPlugin(adminApi, dynamicPluginID);
    }
    return;
  }

  if (originalDynamicEnabled !== 1 && plugin.enabled === 1) {
    await disablePlugin(adminApi, dynamicPluginID);
  }
}

test.describe("TC0128 繁体中文插件页面巡检", () => {
  test.beforeAll(async () => {
    ensureRuntimePluginArtifact();
    adminApi = await createAdminApiContext();
    await syncPlugins(adminApi);
    const plugin = await getPlugin(adminApi, dynamicPluginID);
    originalDynamicInstalled = plugin.installed;
    originalDynamicEnabled = plugin.enabled;
  });

  test.afterAll(async () => {
    try {
      await restoreDynamicPluginState();
    } finally {
      await adminApi.dispose();
    }
  });

  test("TC-128a: 源码插件页面展示繁体中文标签且不泄漏源标签", async ({
    adminPage,
    mainLayout,
  }) => {
    for (const pluginID of sourcePluginIDs) {
      await ensureSourcePluginEnabled(adminPage, pluginID);
    }

    await mainLayout.switchLanguage("繁體中文");

    for (const auditCase of sourcePluginAuditCases) {
      await test.step(auditCase.path, async () => {
        await adminPage.goto(auditCase.path, { waitUntil: "domcontentloaded" });
        await waitForRouteReady(adminPage, 15_000);

        const bodyText = await adminPage.locator("body").innerText();
        for (const text of auditCase.visibleTexts) {
          expect(bodyText).toContain(text);
        }
        for (const text of auditCase.forbiddenTexts) {
          expect(bodyText).not.toContain(text);
        }
        assertNoRawI18nKeys(bodyText, auditCase.path);
      });
    }
  });

  test("TC-128b: 动态插件嵌入页和独立页使用繁体中文运行时资源", async ({
    adminPage,
    mainLayout,
  }) => {
    await ensureDynamicPluginInstalledAndEnabled();
    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage, 15_000);

    const pluginPage = new PluginPage(adminPage);
    await mainLayout.switchLanguage("繁體中文");
    await pluginPage.clickSidebarMenuItem("動態插件示例");
    await waitForRouteReady(adminPage, 15_000);

    await expect(pluginPage.pluginDemoDynamicTitle()).toContainText(
      "動態插件示例已生效",
    );
    await expect(
      pluginPage.pluginDemoDynamicOpenStandaloneButton(),
    ).toContainText("打開獨立頁面");

    const embeddedText = await adminPage.locator("body").innerText();
    expect(embeddedText).not.toContain("Dynamic Plugin Demo Is Live");
    expect(embeddedText).not.toContain("Standalone Page");
    assertNoRawI18nKeys(embeddedText, "plugin-demo-dynamic embedded page");

    const [standalonePage] = await Promise.all([
      adminPage.context().waitForEvent("page"),
      pluginPage.pluginDemoDynamicOpenStandaloneButton().click(),
    ]);
    await standalonePage.waitForLoadState("domcontentloaded");
    await standalonePage.waitForLoadState("networkidle").catch(() => {});

    await expect.poll(async () => standalonePage.url()).toContain("lang=zh-TW");
    const standaloneText = await standalonePage.locator("body").innerText();
    expect(standaloneText).toContain("獨立頁面已成功打開");
    expect(standaloneText).toContain(
      "當前頁面由 plugin-demo-dynamic 直接以託管靜態資源形式提供",
    );
    expect(standaloneText).not.toContain("Standalone Page Opened Successfully");
    expect(standaloneText).not.toContain("Dynamic Plugin Demo");
    assertNoRawI18nKeys(standaloneText, "plugin-demo-dynamic standalone page");
    await standalonePage.close();
  });
});
