import { test, expect } from "../../fixtures/auth";
import { execFileSync } from "node:child_process";
import { rmSync } from "node:fs";
import path from "node:path";

import {
  createAdminApiContext,
  disablePlugin,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from "../../support/api/job";

const dynamicPluginID = "plugin-demo-dynamic";
const sourcePluginID = "plugin-demo-source";
const repoRoot = path.resolve(process.cwd(), "../..");
const legacyRuntimeArtifactPath = path.join(
  repoRoot,
  "apps",
  "lina-plugins",
  dynamicPluginID,
  "runtime",
  `${dynamicPluginID}.wasm`,
);

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

async function ensureDynamicPluginAPIReady() {
  ensureRuntimePluginArtifact();
  const adminApi = await createAdminApiContext();
  try {
    await syncPlugins(adminApi);
    const restoreSourcePlugin = await ensurePluginEnabledForAPIDoc(
      adminApi,
      sourcePluginID,
    );
    const restoreDynamicPlugin = await ensurePluginEnabledForAPIDoc(
      adminApi,
      dynamicPluginID,
    );
    return async () => {
      await restoreDynamicPlugin();
      await restoreSourcePlugin();
      await adminApi.dispose();
    };
  } catch (error) {
    await adminApi.dispose();
    throw error;
  }
}

async function ensurePluginEnabledForAPIDoc(
  adminApi: Awaited<ReturnType<typeof createAdminApiContext>>,
  pluginID: string,
) {
  let plugin = await getPlugin(adminApi, pluginID);
  const originalInstalled = plugin.installed;
  const originalEnabled = plugin.enabled;
  if (plugin.installed !== 1) {
    await installPlugin(adminApi, pluginID);
    plugin = await getPlugin(adminApi, pluginID);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(adminApi, pluginID);
  }
  return async () => {
    let current = await getPlugin(adminApi, pluginID);
    if (originalInstalled !== 1) {
      if (current.enabled === 1) {
        await disablePlugin(adminApi, pluginID);
        current = await getPlugin(adminApi, pluginID);
      }
      if (current.installed === 1) {
        await uninstallPlugin(adminApi, pluginID);
      }
      return;
    }
    if (originalEnabled !== 1 && current.enabled === 1) {
      await disablePlugin(adminApi, pluginID);
    }
  };
}

test.describe("TC0129 繁体中文接口文档", () => {
  test("TC-129a: api.json 按 zh-TW 返回接口分组、摘要和参数说明", async ({
    adminPage,
  }) => {
    const restoreDynamicPlugin = await ensureDynamicPluginAPIReady();
    try {
      const response = await adminPage.request.get("/api.json?lang=zh-TW", {
        headers: { "Accept-Language": "zh-TW" },
      });
      expect(response.ok()).toBeTruthy();

      const apiDocument = await response.json();
      const documentText = JSON.stringify(apiDocument);
      expect(documentText).toContain('"用戶管理"');
      expect(documentText).toContain('"獲取用戶列表"');
      expect(documentText).toContain('"按用戶名篩選（模糊匹配）"');
      expect(documentText).toContain("目標語言編碼");
      expect(documentText).toContain('"操作日誌"');
      expect(documentText).toContain('"獲取操作日誌列表"');
      expect(documentText).toContain('"源碼插件示例"');
      expect(documentText).toContain('"查詢源碼插件示例記錄列表"');
      expect(documentText).toContain('"動態插件示例"');
      expect(documentText).toContain('"查詢動態插件示例記錄列表"');

      const userListOperation = apiDocument.paths?.["/api/v1/user"]?.get;
      expect(userListOperation?.tags).toContain("用戶管理");
      expect(userListOperation?.summary).toBe("獲取用戶列表");
      expect(userListOperation?.tags).not.toContain("User Management");
      expect(userListOperation?.summary).not.toBe("Get user list");
      expect(userListOperation?.summary).not.toBe("获取用户列表");

      const pageNumParameter = userListOperation?.parameters?.find(
        (parameter: { name?: string }) => parameter.name === "pageNum",
      );
      expect(pageNumParameter?.description).toBe("頁碼");
      expect(pageNumParameter?.description).not.toBe("Page number");
      expect(pageNumParameter?.description).not.toBe("页码");

      const usernameParameter = userListOperation?.parameters?.find(
        (parameter: { name?: string }) => parameter.name === "username",
      );
      expect(usernameParameter?.description).toBe("按用戶名篩選（模糊匹配）");
      expect(usernameParameter?.description).not.toBe("Username");

      const operlogListOperation = apiDocument.paths?.["/api/v1/operlog"]?.get;
      expect(operlogListOperation?.tags).toContain("操作日誌");
      expect(operlogListOperation?.summary).toBe("獲取操作日誌列表");
      expect(operlogListOperation?.summary).not.toBe("Get operation log list");
      expect(operlogListOperation?.summary).not.toBe("获取操作日志列表");

      const sourcePluginListOperation =
        apiDocument.paths?.["/api/v1/plugins/plugin-demo-source/records"]?.get;
      expect(sourcePluginListOperation?.tags).toContain("源碼插件示例");
      expect(sourcePluginListOperation?.summary).toBe(
        "查詢源碼插件示例記錄列表",
      );
      expect(sourcePluginListOperation?.summary).not.toBe(
        "Query source plugin demo record list",
      );
      expect(sourcePluginListOperation?.summary).not.toBe(
        "查询源码插件示例记录列表",
      );

      const dynamicPluginListOperation =
        apiDocument.paths?.[
          "/api/v1/extensions/plugin-demo-dynamic/demo-records"
        ]?.get;
      expect(dynamicPluginListOperation?.tags).toContain("動態插件示例");
      expect(dynamicPluginListOperation?.summary).toBe(
        "查詢動態插件示例記錄列表",
      );
      expect(dynamicPluginListOperation?.summary).not.toBe(
        "Query dynamic plugin demo record list",
      );
      expect(dynamicPluginListOperation?.summary).not.toBe(
        "查询动态插件示例记录列表",
      );
    } finally {
      await restoreDynamicPlugin();
    }
  });
});
