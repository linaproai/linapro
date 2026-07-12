import {
  expect,
  request as playwrightRequest,
  type APIRequestContext,
} from "@playwright/test";

import { test } from "../../../fixtures/auth";
import { config } from "../../../fixtures/config";
import { MenuPage } from "../../../pages/MenuPage";
import { waitForBusyIndicatorsToClear } from "../../../support/ui";

const apiBaseURL = config.apiBaseURL;

async function createAdminApiContext(): Promise<APIRequestContext> {
  const loginApi = await playwrightRequest.newContext({ baseURL: apiBaseURL });
  const loginResponse = await loginApi.post("auth/login", {
    data: {
      username: config.adminUser,
      password: config.adminPass,
      clientType: "web",
    },
  });
  expect(loginResponse.ok()).toBeTruthy();

  const loginResult = await loginResponse.json();
  const accessToken = loginResult.data?.accessToken;
  expect(accessToken).toBeTruthy();
  await loginApi.dispose();

  return playwrightRequest.newContext({
    baseURL: apiBaseURL,
    extraHTTPHeaders: {
      Authorization: `Bearer ${accessToken}`,
    },
  });
}

async function createMenuTree(
  api: APIRequestContext,
  suffix: string,
): Promise<{ parentId: number; childId: number; parentName: string; childName: string }> {
  const parentName = `E2E级联父-${suffix}`;
  const childName = `E2E级联子-${suffix}`;

  const parentResponse = await api.post("menu", {
    data: {
      parentId: 0,
      name: parentName,
      type: "D",
      path: `e2e-cascade-parent-${suffix}`,
      icon: `mdi:folder-e2e-${suffix}`,
      sort: 999,
      visible: 1,
      status: 1,
      isFrame: 0,
      isCache: 0,
    },
  });
  expect(parentResponse.ok()).toBeTruthy();
  const parentBody = await parentResponse.json();
  const parentId = Number(parentBody.data?.id ?? parentBody.data);
  expect(parentId).toBeGreaterThan(0);

  const childResponse = await api.post("menu", {
    data: {
      parentId,
      name: childName,
      type: "M",
      path: `e2e-cascade-child-${suffix}`,
      component: "system/menu/index",
      perms: `e2e:menu:cascade:${suffix}`,
      icon: `mdi:file-e2e-${suffix}`,
      sort: 1,
      visible: 1,
      status: 1,
      isFrame: 0,
      isCache: 0,
    },
  });
  expect(childResponse.ok()).toBeTruthy();
  const childBody = await childResponse.json();
  const childId = Number(childBody.data?.id ?? childBody.data);
  expect(childId).toBeGreaterThan(0);

  return { parentId, childId, parentName, childName };
}

async function deleteMenuCascade(api: APIRequestContext, menuId: number) {
  const response = await api.delete(`menu/${menuId}`, {
    params: { cascadeDelete: true },
  });
  // Best-effort cleanup; ignore already-deleted rows.
  if (!response.ok()) {
    return;
  }
}

async function getMenuFlags(
  api: APIRequestContext,
  menuId: number,
): Promise<{ status: number; visible: number }> {
  const response = await api.get(`menu/${menuId}`);
  expect(response.ok()).toBeTruthy();
  const body = await response.json();
  return {
    status: Number(body.data?.status),
    visible: Number(body.data?.visible),
  };
}

test.describe("TC003 菜单状态与显示开关", () => {
  test("TC003a: 列表渲染状态与显示开关", async ({
    authenticatedPage: adminPage,
  }) => {
    const menuPage = new MenuPage(adminPage);
    await menuPage.goto();

    await expect(
      adminPage.getByTestId("menu-status-switch").first(),
    ).toBeVisible({ timeout: 10000 });
    await expect(
      adminPage.getByTestId("menu-visible-switch").first(),
    ).toBeVisible({ timeout: 10000 });

    // Switch labels must be translated text, not raw i18n keys.
    await expect(
      adminPage.getByTestId("menu-status-switch").first(),
    ).toContainText(/启用|停用/);
    await expect(
      adminPage.getByTestId("menu-visible-switch").first(),
    ).toContainText(/显示|隐藏/);
  });

  test("TC003b: 停用父级目录级联停用子级，启用同步恢复子级", async ({
    authenticatedPage: adminPage,
  }) => {
    const api = await createAdminApiContext();
    const suffix = Date.now().toString().slice(-8);
    const { parentId, childId, parentName, childName } = await createMenuTree(
      api,
      `d${suffix}`,
    );
    const menuPage = new MenuPage(adminPage);

    try {
      await menuPage.goto();
      await menuPage.revealMenuRow(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);

      await expect(menuPage.statusSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "true",
        { timeout: 15000 },
      );
      await expect(menuPage.statusSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "true",
        { timeout: 15000 },
      );

      await menuPage.toggleStatus(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);

      await expect(menuPage.statusSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "false",
      );
      await expect(menuPage.statusSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "false",
      );

      const afterDisable = await getMenuFlags(api, childId);
      expect(afterDisable.status).toBe(0);

      // Re-enable parent; child is restored together.
      await menuPage.toggleStatus(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);
      await expect(menuPage.statusSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "true",
      );
      await expect(menuPage.statusSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "true",
      );

      const afterEnable = await getMenuFlags(api, childId);
      expect(afterEnable.status).toBe(1);
      const parentFlags = await getMenuFlags(api, parentId);
      expect(parentFlags.status).toBe(1);
    } finally {
      await deleteMenuCascade(api, parentId);
      await api.dispose();
      await waitForBusyIndicatorsToClear(adminPage).catch(() => {});
    }
  });

  test("TC003c: 隐藏父级目录级联隐藏子级，显示同步恢复子级", async ({
    authenticatedPage: adminPage,
  }) => {
    const api = await createAdminApiContext();
    const suffix = Date.now().toString().slice(-8);
    const { parentId, childId, parentName, childName } = await createMenuTree(
      api,
      `h${suffix}`,
    );
    const menuPage = new MenuPage(adminPage);

    try {
      await menuPage.goto();
      await menuPage.revealMenuRow(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);

      await expect(menuPage.visibleSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "true",
        { timeout: 15000 },
      );
      await expect(menuPage.visibleSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "true",
        { timeout: 15000 },
      );

      await menuPage.toggleVisible(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);

      await expect(menuPage.visibleSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "false",
      );
      await expect(menuPage.visibleSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "false",
      );

      const afterHide = await getMenuFlags(api, childId);
      expect(afterHide.visible).toBe(0);

      // Re-show parent; child is restored together.
      await menuPage.toggleVisible(parentName);
      await menuPage.expandMenuRow(parentName);
      await menuPage.revealMenuRow(childName);
      await expect(menuPage.visibleSwitch(parentName)).toHaveAttribute(
        "aria-checked",
        "true",
      );
      await expect(menuPage.visibleSwitch(childName)).toHaveAttribute(
        "aria-checked",
        "true",
      );

      const afterShow = await getMenuFlags(api, childId);
      expect(afterShow.visible).toBe(1);
      const parentFlags = await getMenuFlags(api, parentId);
      expect(parentFlags.visible).toBe(1);
    } finally {
      await deleteMenuCascade(api, parentId);
      await api.dispose();
      await waitForBusyIndicatorsToClear(adminPage).catch(() => {});
    }
  });
});
