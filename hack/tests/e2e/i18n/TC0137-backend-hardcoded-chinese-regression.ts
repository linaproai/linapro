import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  enablePlugin,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
} from "../../support/api/job";
import { waitForRouteReady } from "../../support/ui";

type DeptTreeNode = {
  children?: DeptTreeNode[];
  id: number;
  label: string;
};

const sourcePluginIDs = ["org-center", "plugin-demo-source"] as const;
const chineseSystemCopyPattern =
  /未分配部门|服务运行时长|小时|分钟|秒|这是一条来自 plugin-demo-source 接口的简要介绍/;

async function ensureSourcePluginsEnabled(
  api: APIRequestContext,
  pluginIDs: readonly string[],
) {
  await syncPlugins(api);
  for (const pluginID of pluginIDs) {
    let plugin = await getPlugin(api, pluginID);
    if (plugin.installed !== 1) {
      await installPlugin(api, pluginID);
      plugin = await getPlugin(api, pluginID);
    }
    if (plugin.enabled !== 1) {
      await enablePlugin(api, pluginID);
    }
  }
}

function flattenDeptTree(nodes: DeptTreeNode[]): DeptTreeNode[] {
  return nodes.flatMap((node) => [node, ...flattenDeptTree(node.children ?? [])]);
}

test.describe("TC-137 Backend hardcoded Chinese regression", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-137a: English backend projections do not return legacy Chinese system copy", async () => {
    const deptTree = await expectSuccess<{ list: DeptTreeNode[] }>(
      await adminApi.get("user/dept-tree", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    const unassignedNode = flattenDeptTree(deptTree.list).find((node) => node.id === 0);
    expect(unassignedNode?.label).toContain("Unassigned");
    expect(unassignedNode?.label).not.toContain("未分配部门");

    const systemInfo = await expectSuccess<{
      runDuration: string;
      runDurationSeconds: number;
    }>(
      await adminApi.get("system/info", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    expect(systemInfo.runDurationSeconds).toBeGreaterThanOrEqual(0);
    expect(systemInfo.runDuration).toMatch(/second|minute|hour|day/);
    expect(systemInfo.runDuration).not.toMatch(/小时|分钟|秒/);

    const summary = await expectSuccess<{ message: string }>(
      await adminApi.get("plugins/plugin-demo-source/summary", {
        headers: { "Accept-Language": "en-US" },
      }),
    );
    expect(summary.message).toContain("plugin-demo-source API");
    expect(summary.message).not.toMatch(chineseSystemCopyPattern);
  });

  test("TC-137b: English source-plugin page does not show legacy backend Chinese copy", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("English");
    await adminPage.goto("/plugin-demo-source-sidebar-entry?lang=en-US", {
      waitUntil: "domcontentloaded",
    });
    await waitForRouteReady(adminPage, 15_000);

    const bodyText = await adminPage.locator("body").innerText();
    expect(bodyText).toContain("Source Plugin Demo");
    expect(bodyText).toContain("Demo Records");
    expect(bodyText).not.toMatch(chineseSystemCopyPattern);
  });
});
