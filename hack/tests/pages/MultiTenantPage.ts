import type { Page } from "@playwright/test";

import { expect } from "@playwright/test";

import { MainLayout } from "./MainLayout";
import { waitForRouteReady } from "../support/ui";

type WorkbenchMode = "disabled" | "platform" | "tenant";

const tenants = [
  { id: 101, code: "alpha", name: "Alpha BU" },
  { id: 102, code: "beta", name: "Beta BU" },
];

const platformTenantRows = [
  {
    id: 101,
    code: "alpha",
    name: "Alpha BU",
    remark: "Alpha tenant remark",
    status: "active",
    createdAt: "2026-05-09 10:00:00",
  },
  {
    id: 102,
    code: "beta",
    name: "Beta BU",
    remark: "Beta tenant remark",
    status: "suspended",
    createdAt: "2026-05-09 11:00:00",
  },
];

const systemUsers = [
  {
    id: 1,
    username: "admin",
    nickname: "Administrator",
    email: "admin@example.com",
    phone: "13860000000",
    sex: 0,
    avatar: "",
    status: 1,
    remark: "",
    loginDate: "",
    createdAt: "2026-05-01 08:00:00",
    updatedAt: "2026-05-01 08:00:00",
    deptId: 0,
    deptName: "",
    postIds: [],
    roleIds: [1],
    roleNames: ["管理员"],
    tenantId: 0,
    tenantIds: [],
    tenantNames: [],
  },
  {
    id: 301,
    username: "tenant-admin",
    nickname: "Tenant Admin",
    email: "tenant@example.com",
    phone: "13860010001",
    sex: 0,
    avatar: "",
    status: 1,
    remark: "",
    loginDate: "",
    createdAt: "2026-05-09 10:00:00",
    updatedAt: "2026-05-09 10:00:00",
    deptId: 0,
    deptName: "",
    postIds: [],
    roleIds: [201],
    roleNames: ["Tenant Admin"],
    tenantId: 101,
    tenantIds: [101],
    tenantNames: ["Alpha BU"],
  },
  {
    id: 302,
    username: "tenant-beta-ops",
    nickname: "Beta Ops",
    email: "beta@example.com",
    phone: "13860020001",
    sex: 0,
    avatar: "",
    status: 1,
    remark: "",
    loginDate: "",
    createdAt: "2026-05-09 11:00:00",
    updatedAt: "2026-05-09 11:00:00",
    deptId: 0,
    deptName: "",
    postIds: [],
    roleIds: [202],
    roleNames: ["Tenant Operator"],
    tenantId: 102,
    tenantIds: [102],
    tenantNames: ["Beta BU"],
  },
];

const notFoundTextPattern = /404|Not Found|页面未找到|未找到页面/;

function ok(data: unknown) {
  return {
    contentType: "application/json",
    body: JSON.stringify({ code: 0, data }),
  };
}

function routePath(url: string) {
  return new URL(url).pathname.replace(/^\/api\/v1/, "").replace(/^\/api/, "");
}

export class MultiTenantPage {
  private lastEndImpersonateAuthorization = "";
  private lastMenusAuthorization = "";
  private lastPlatformUserAuthorization = "";
  private lastPlatformUserTenantCode = "";
  private lastSwitchTenantAuthorization = "";
  private lastSwitchTenantPayload: Record<string, any> | null = null;
  private lastUserInfoAuthorization = "";
  private lastUserCreatePayload: Record<string, any> | null = null;
  private lastUserUpdatePayload: Record<string, any> | null = null;
  private lastUserTenantFilter: null | string = null;
  private workbenchMode: WorkbenchMode = "platform";

  constructor(private page: Page) {}

  private menuNode(
    path: string,
    name: string,
    order: number,
    children: any[] = [],
  ) {
    return { children, meta: { order }, name, path };
  }

  private routeNode(
    name: string,
    path: string,
    component: string,
    title: string,
    tenantAccessMode?: "platform" | "tenant",
    children: any[] = [],
    order = 1,
  ) {
    return {
      children,
      component,
      meta: {
        icon: "lucide:circle",
        order,
        tenantAccessMode,
        title,
      },
      name,
      path,
    };
  }

  private pluginRouteNode(
    name: string,
    path: string,
    component: string,
    title: string,
    children: any[] = [],
  ) {
    return this.routeNode(name, path, component, title, undefined, children);
  }

  private menuTree(mode: WorkbenchMode) {
    const base = [
      this.menuNode("/dashboard", "Dashboard", -1, [
        this.menuNode("/dashboard/analytics", "Analytics", 1),
      ]),
    ];
    if (mode === "platform") {
      return [
        ...base,
        this.menuNode("/iam", "Access", 10, [
          this.menuNode("/system/user", "UserManagement", 1),
        ]),
        this.menuNode("/platform", "Platform", 15, [
          this.menuNode("/platform/tenants", "PlatformTenantManagement", 10),
        ]),
        this.menuNode("/org", "Organization", 16, [
          this.menuNode("/system/dept", "DeptManagement", 1),
        ]),
      ];
    }
    if (mode === "tenant") {
      return [
        ...base,
        this.menuNode("/iam", "Access", 10, [
          this.menuNode("/system/user", "UserManagement", 1),
        ]),
      ];
    }
    return base;
  }

  private routeTree(mode: WorkbenchMode) {
    const base = [
      this.routeNode(
        "Dashboard",
        "/dashboard",
        "",
        "page.dashboard.title",
        undefined,
        [
          this.routeNode(
            "Analytics",
            "/dashboard/analytics",
            "#/views/dashboard/analytics/index.vue",
            "page.dashboard.analytics",
          ),
        ],
      ),
    ];
    if (mode === "platform") {
      return [
        ...base,
        this.routeNode(
          "IAM",
          "/iam",
          "",
          "page.routes.system.accessManagement",
          undefined,
          [
            this.routeNode(
              "UserManagement",
              "/system/user",
              "#/views/system/user/index.vue",
              "page.routes.system.userManagement",
            ),
          ],
          10,
        ),
        this.routeNode(
          "Platform",
          "/platform",
          "",
          "menu.platform.title",
          "platform",
          [
            this.pluginRouteNode(
              "PluginMultiTenantPlatformTenants",
              "/platform/tenants",
              "#/views/system/plugin/dynamic-page",
              "plugin.multi-tenant.menu.platform.tenants.name",
            ),
          ],
          15,
        ),
        this.routeNode(
          "Org",
          "/org",
          "",
          "page.routes.system.organization",
          undefined,
          [
            this.routeNode(
              "DeptManagement",
              "/system/dept",
              "#/views/system/dept/index.vue",
              "page.routes.system.deptManagement",
            ),
          ],
          16,
        ),
      ];
    }
    if (mode === "tenant") {
      return [
        ...base,
        this.routeNode(
          "IAM",
          "/iam",
          "",
          "page.routes.system.accessManagement",
          undefined,
          [
            this.routeNode(
              "UserManagement",
              "/system/user",
              "#/views/system/user/index.vue",
              "page.routes.system.userManagement",
            ),
          ],
          10,
        ),
      ];
    }
    return base;
  }

  private flattenPaths(
    items: Array<{ children?: any[]; path?: string }>,
  ): string[] {
    return items.flatMap((item) => [
      item.path || "",
      ...this.flattenPaths(item.children ?? []),
    ]);
  }

  private findRoute(
    items: Array<{ children?: any[]; path?: string }>,
    path: string,
  ): any | null {
    for (const item of items) {
      if (item.path === path) {
        return item;
      }
      const child = this.findRoute(item.children ?? [], path);
      if (child) {
        return child;
      }
    }
    return null;
  }

  private assertPrunedWorkbenchProjection(mode: WorkbenchMode) {
    const routeTree = this.routeTree(mode);
    const routePaths = this.flattenPaths(routeTree);
    const menuPaths = this.flattenPaths(this.menuTree(mode));
    for (const path of [
      "/platform/tenant-members",
      "/tenant",
      "/tenant/members",
      "/tenant/plugins",
    ]) {
      expect(routePaths).not.toContain(path);
      expect(menuPaths).not.toContain(path);
    }

    if (mode === "platform") {
      const iamOrder = this.findRoute(routeTree, "/iam")?.meta?.order;
      const platformOrder = this.findRoute(routeTree, "/platform")?.meta?.order;
      const orgOrder = this.findRoute(routeTree, "/org")?.meta?.order;
      expect(iamOrder).toBeLessThan(platformOrder);
      expect(platformOrder).toBeLessThan(orgOrder);
    }
  }

  private userInfo(mode = this.workbenchMode) {
    return {
      userId: 1,
      username: mode === "tenant" ? "tenant-admin" : "admin",
      realName: mode === "tenant" ? "Tenant Admin" : "Administrator",
      avatar: "",
      email: "admin@example.com",
      permissions: ["*"],
      roles: ["admin"],
      homePath:
        mode === "platform"
          ? "/platform/tenants"
          : mode === "tenant"
            ? "/system/user"
            : "/dashboard/analytics",
      menus: this.menuTree(mode),
    };
  }

  private async mockShellApis() {
    await this.page.route(
      /\/api(?:\/v1)?\/config\/public\/frontend$/,
      async (route) => {
        await route.fulfill(ok({}));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/i18n\/runtime\/locales(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(
          ok({
            enabled: true,
            items: [
              {
                locale: "zh-CN",
                name: "简体中文",
                nativeName: "简体中文",
                isDefault: true,
              },
              { locale: "en-US", name: "English", nativeName: "English" },
            ],
            locale: "zh-CN",
          }),
        );
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/i18n\/runtime\/messages(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(ok({ messages: {} }));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/plugins\/dynamic(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(
          ok({
            list:
              this.workbenchMode === "disabled"
                ? []
                : [
                    {
                      enabled: 1,
                      generation: 1,
                      id: "multi-tenant",
                      installed: 1,
                      statusKey: "sys_plugin.status:multi-tenant",
                      version: "v0.1.0",
                    },
                    {
                      enabled: 1,
                      generation: 1,
                      id: "org-center",
                      installed: 1,
                      statusKey: "sys_plugin.status:org-center",
                      version: "v0.1.0",
                    },
                  ],
          }),
        );
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/user\/message\/count(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(ok({ count: 0 }));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/user\/message(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(ok({ list: [], total: 0 }));
      },
    );
    await this.page.route(/\/api(?:\/v1)?\/user\/info$/, async (route) => {
      this.lastUserInfoAuthorization =
        route.request().headers().authorization || "";
      await route.fulfill(ok(this.userInfo()));
    });
    await this.page.route(
      /\/api(?:\/v1)?\/menus\/all(?:\?.*)?$/,
      async (route) => {
        this.lastMenusAuthorization =
          route.request().headers().authorization || "";
        await route.fulfill(ok({ list: this.routeTree(this.workbenchMode) }));
      },
    );
  }

  private async mockTenantApis() {
    await this.page.route(
      /\/api(?:\/v1)?\/platform\/tenants\/\d+\/impersonate$/,
      async (route) => {
        this.workbenchMode = "tenant";
        await route.fulfill(
          ok({
            token: "impersonation-token",
            tenant: platformTenantRows[0],
          }),
        );
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/platform\/tenants\/\d+\/end-impersonate$/,
      async (route) => {
        this.workbenchMode = "platform";
        this.lastEndImpersonateAuthorization =
          route.request().headers().authorization || "";
        await route.fulfill(ok({}));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/auth\/switch-tenant$/,
      async (route) => {
        this.workbenchMode = "tenant";
        this.lastSwitchTenantAuthorization =
          route.request().headers().authorization || "";
        this.lastSwitchTenantPayload = route.request().postDataJSON();
        await route.fulfill(ok({ accessToken: "beta-token" }));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/auth\/login-tenants(?:\?.*)?$/,
      async (route) => {
        const userId = new URL(route.request().url()).searchParams.get(
          "userId",
        );
        await route.fulfill(ok({ list: userId === "1" ? [tenants[0]] : [] }));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/platform\/tenants(?:\/\d+(?:\/status)?)?(?:\?.*)?$/,
      async (route) => {
        const method = route.request().method();
        const path = routePath(route.request().url());
        if (method === "GET") {
          await route.fulfill(
            ok({ list: platformTenantRows, total: platformTenantRows.length }),
          );
          return;
        }
        if (method === "PUT" && path.endsWith("/status")) {
          const tenant = platformTenantRows.find((item) =>
            path.includes(`/${item.id}/`),
          );
          const body = route.request().postDataJSON() as { status?: string };
          await route.fulfill(ok({ ...tenant, status: body.status }));
          return;
        }
        if (method === "POST") {
          const body = route.request().postDataJSON();
          await route.fulfill(
            ok({
              id: 103,
              status: "active",
              createdAt: "2026-05-09 12:00:00",
              ...body,
            }),
          );
          return;
        }
        if (method === "PUT") {
          const tenant = platformTenantRows.find((item) =>
            path.endsWith(`/${item.id}`),
          );
          await route.fulfill(
            ok({ ...tenant, ...route.request().postDataJSON() }),
          );
          return;
        }
        await route.fulfill(ok(platformTenantRows[0]));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/platform\/tenant\/resolver-config(?:\?.*)?$/,
      async (route) => {
        if (route.request().method() === "GET") {
          await route.fulfill(
            ok({
              chain: [
                "override",
                "jwt",
                "session",
                "header",
                "subdomain",
                "default",
              ],
              onAmbiguous: "prompt",
              reservedSubdomains: ["www", "api", "admin"],
              rootDomain: "example.test",
            }),
          );
          return;
        }
        await route.fulfill(ok({}));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/dict\/data\/type\/sys_normal_disable$/,
      async (route) => {
        await route.fulfill(
          ok({
            list: [
              { label: "正常", value: "1", tagStyle: "success" },
              { label: "停用", value: "0", tagStyle: "error" },
            ],
          }),
        );
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/role\/options(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(
          ok({
            list: [
              { id: 1, key: "admin", name: "管理员" },
              { id: 201, key: "tenant-admin", name: "Tenant Admin" },
            ],
          }),
        );
      },
    );
    await this.page.route(/\/api(?:\/v1)?\/user(?:\?.*)?$/, async (route) => {
      if (route.request().method() === "GET") {
        const tenantId = new URL(route.request().url()).searchParams.get(
          "tenantId",
        );
        if (!tenantId) {
          const headers = route.request().headers();
          this.lastPlatformUserAuthorization = headers.authorization || "";
          this.lastPlatformUserTenantCode = headers["x-tenant-code"] || "";
        }
        this.lastUserTenantFilter = tenantId;
        const list = tenantId
          ? systemUsers.filter((item) =>
              item.tenantIds.includes(Number(tenantId)),
            )
          : systemUsers;
        await route.fulfill(ok({ list, total: list.length }));
        return;
      }
      if (route.request().method() === "POST") {
        this.lastUserCreatePayload = route.request().postDataJSON();
        await route.fulfill(ok({ id: 303 }));
        return;
      }
      await route.fulfill(ok(systemUsers[0]));
    });
    await this.page.route(/\/api(?:\/v1)?\/user\/\d+$/, async (route) => {
      const path = routePath(route.request().url());
      const id = Number(path.split("/").pop());
      const user = systemUsers.find((item) => item.id === id) || systemUsers[0];
      if (route.request().method() === "GET") {
        await route.fulfill(ok(user));
        return;
      }
      if (route.request().method() === "PUT") {
        this.lastUserUpdatePayload = route.request().postDataJSON();
        await route.fulfill(ok({}));
        return;
      }
      await route.fulfill(ok(user));
    });
    await this.page.route(
      /\/api(?:\/v1)?\/user\/dept-tree(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(ok({ list: [] }));
      },
    );
    await this.page.route(
      /\/api(?:\/v1)?\/plugins(?:\?.*)?$/,
      async (route) => {
        await route.fulfill(
          ok({
            list: [
              {
                id: "org-center",
                pluginId: "org-center",
                installed: 0,
                enabled: 0,
                status: 0,
              },
            ],
            total: 1,
          }),
        );
      },
    );
  }

  private async installAuthState(mode: WorkbenchMode) {
    await this.page.addInitScript(
      ({ modeValue, tenantState }) => {
        localStorage.clear();
        localStorage.setItem(
          "lina-web-antd-5.6.0-dev-core-access",
          JSON.stringify({
            accessToken:
              modeValue === "tenant" ? "alpha-token" : "platform-token",
            accessCodes: ["*"],
            refreshToken: null,
            isLockScreen: false,
          }),
        );
        localStorage.setItem(
          "linapro:tenant-state",
          JSON.stringify(tenantState),
        );
      },
      {
        modeValue: mode,
        tenantState:
          mode === "platform"
            ? {
                enabled: true,
                currentTenant: null,
                tenants: [],
                impersonation: { active: false },
              }
            : mode === "tenant"
              ? {
                  enabled: true,
                  currentTenant: tenants[0],
                  tenants,
                  impersonation: { active: false },
                }
              : {
                  enabled: false,
                  currentTenant: null,
                  tenants: [],
                  impersonation: { active: false },
                },
      },
    );
  }

  private async prepareWorkbench(mode: WorkbenchMode) {
    this.workbenchMode = mode;
    await this.installAuthState(mode);
    await this.mockShellApis();
    await this.mockTenantApis();
  }

  async mockWorkbenchShell() {
    await this.prepareWorkbench("platform");
  }

  async mockLoginTenantSelection() {
    this.workbenchMode = "tenant";
    await this.mockShellApis();
    await this.page.route(
      /\/api(?:\/v1)?\/config\/public\/frontend$/,
      async (route) => {
        await route.fulfill(
          ok({
            auth: {
              loginSubtitle: "",
            },
          }),
        );
      },
    );
    await this.page.route(/\/api(?:\/v1)?\/auth\/login$/, async (route) => {
      await route.fulfill(
        ok({
          preToken: "pre-token",
          tenants,
        }),
      );
    });
    await this.page.route(
      /\/api(?:\/v1)?\/auth\/select-tenant$/,
      async (route) => {
        await route.fulfill(ok({ accessToken: "tenant-token" }));
      },
    );
  }

  async gotoPlatformTenants() {
    await this.prepareWorkbench("platform");
    this.assertPrunedWorkbenchProjection("platform");
    await this.page.goto("/platform/tenants");
    await waitForRouteReady(this.page);
  }

  async gotoSystemUsers() {
    await this.prepareWorkbench("platform");
    this.assertPrunedWorkbenchProjection("platform");
    await this.page.goto("/system/user");
    await waitForRouteReady(this.page);
  }

  async gotoTenantSystemUsers() {
    await this.prepareWorkbench("tenant");
    this.assertPrunedWorkbenchProjection("tenant");
    await this.page.goto("/system/user");
    await waitForRouteReady(this.page);
  }

  async gotoDisabledPlatformRoute() {
    await this.prepareWorkbench("disabled");
    await this.page.goto("/platform/tenants");
    await waitForRouteReady(this.page);
  }

  async expectPlatformTenantWorkbench() {
    await expect(this.page.getByTestId("platform-tenants-page")).toBeVisible();
    await expect(
      this.page.getByText(/插件页面未找到|Plugin page not found/),
    ).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-switcher")).toBeVisible();
    await this.expectHeaderTenantSwitcherLoadsOptions();
    await expect(
      this.page.getByTestId("platform-tenants-page").getByText("Alpha BU"),
    ).toBeVisible();
    await expect(
      this.page.getByTestId("platform-tenants-page").getByText("Beta BU"),
    ).toBeVisible();
    await expect(this.page.getByTestId("tenant-create")).toBeVisible();
    await expect(this.page.getByTestId("tenant-create")).toHaveText(
      /^新\s*增$/,
    );
    await expect(this.page.getByTestId("tenant-suspend-101")).toBeVisible();
    await expect(this.page.getByTestId("tenant-resume-102")).toBeVisible();
    await expect(this.page.getByTestId("tenant-archive-101")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-members-101")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-member-modal")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-impersonate-101")).toBeVisible();
    await expect(this.page.getByTestId("tenant-delete-101")).toBeVisible();
    await this.expectTenantImpersonationTooltip();
    await expect(this.page.getByText("2026-05-09 10:00:00")).toBeVisible();
    await expect(this.page.getByText("2026-05-09 11:00:00")).toBeVisible();
    await expect(this.page.getByText("套餐")).toHaveCount(0);
    await expect(this.page.getByText("归档")).toHaveCount(0);
    await expect(this.page.getByText("已暂停")).toHaveCount(0);
    await expect(this.page.getByText("暂停").first()).toBeVisible();
    await expect(
      this.page.getByRole("button", { name: /^更多$|^More$/ }),
    ).toHaveCount(0);
    await this.expectTenantCreateModalFields();
    await this.expectTenantEditModalFields();
  }

  private async expectTenantImpersonationTooltip() {
    await this.page.getByTestId("tenant-impersonate-101").hover();
    await expect(this.page.locator(".ant-tooltip:visible")).toContainText(
      "以平台身份进入该租户视角",
    );
  }

  async expectTenantSearchInlineLayout() {
    await expect
      .poll(async () => {
        const locators = [
          this.page.getByTestId("tenant-search-code"),
          this.page.getByTestId("tenant-search-name"),
          this.page.getByTestId("tenant-search-status"),
          this.page.getByRole("button", { name: /^重\s*置$|^Reset$/ }),
          this.page.getByRole("button", { name: /^搜\s*索$|^Search$/ }),
        ];
        const boxes = await Promise.all(
          locators.map(async (locator) => locator.first().boundingBox()),
        );
        if (boxes.some((box) => !box)) {
          return false;
        }
        const visibleBoxes = boxes as Array<{
          height: number;
          width: number;
          x: number;
          y: number;
        }>;
        const centers = visibleBoxes.map((box) => box.y + box.height / 2);
        const minCenter = Math.min(...centers);
        const maxCenter = Math.max(...centers);
        const ordered = visibleBoxes.every((box, index) => {
          if (index === 0) {
            return true;
          }
          return box.x >= visibleBoxes[index - 1]!.x;
        });
        return maxCenter - minCenter <= 18 && ordered;
      })
      .toBe(true);
  }

  private async expectHeaderTenantSwitcherLoadsOptions() {
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const parsed = JSON.parse(
            localStorage.getItem("linapro:tenant-state") || "{}",
          );
          return parsed.tenants?.length ?? 0;
        }),
      )
      .toBeGreaterThan(0);
    await this.page.getByTestId("tenant-switcher-select").click();
    const alphaOption = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option", {
        hasText: "Alpha BU",
      })
      .first();
    await expect(alphaOption).toBeVisible();
    await this.page.keyboard.press("Escape");
  }

  private async expectTenantCreateModalFields() {
    await this.page.getByTestId("tenant-create").click();
    await expect(this.page.getByTestId("tenant-form")).toBeVisible();
    await expect(this.page.getByTestId("tenant-code-input")).toBeEnabled();
    await expect(this.page.getByTestId("tenant-code-input")).toHaveValue("");
    await expect(this.page.getByTestId("tenant-name-input")).toHaveValue("");
    await expect(this.page.getByTestId("tenant-plan-input")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-remark-input")).toHaveValue("");
    await this.expectTenantModalInlineLayout();

    const createResponsePromise = this.page.waitForResponse((response) => {
      const request = response.request();
      return (
        request.method() === "POST" &&
        routePath(response.url()) === "/platform/tenants"
      );
    });
    await this.page.getByTestId("tenant-code-input").fill("gamma");
    await this.page.getByTestId("tenant-name-input").fill("Gamma BU");
    await this.page
      .getByTestId("tenant-remark-input")
      .fill("Gamma tenant remark");
    await this.clickTenantModalConfirm("新增租户");
    const createResponse = await createResponsePromise;
    expect(createResponse.request().postDataJSON()).toMatchObject({
      code: "gamma",
      name: "Gamma BU",
      remark: "Gamma tenant remark",
    });
    expect(createResponse.request().postDataJSON()).not.toHaveProperty("plan");
    await expect(this.page.getByTestId("tenant-form")).toHaveCount(0);
  }

  private async expectTenantEditModalFields() {
    await this.page.getByTestId("tenant-edit-101").click();
    await expect(this.page.getByTestId("tenant-form")).toBeVisible();
    await expect(this.page.getByTestId("tenant-code-input")).toBeDisabled();
    await expect(this.page.getByTestId("tenant-code-input")).toHaveValue(
      "alpha",
    );
    await expect(this.page.getByTestId("tenant-name-input")).toHaveValue(
      "Alpha BU",
    );
    await expect(this.page.getByTestId("tenant-plan-input")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-remark-input")).toHaveValue(
      "Alpha tenant remark",
    );
    await this.expectTenantModalInlineLayout();

    const updateResponsePromise = this.page.waitForResponse((response) => {
      const request = response.request();
      return (
        request.method() === "PUT" &&
        routePath(response.url()) === "/platform/tenants/101"
      );
    });
    await this.page.getByTestId("tenant-name-input").fill("Alpha Business");
    await this.page
      .getByTestId("tenant-remark-input")
      .fill("Updated alpha tenant remark");
    await this.clickTenantModalConfirm("编辑租户");
    const updateResponse = await updateResponsePromise;
    expect(updateResponse.request().postDataJSON()).toMatchObject({
      name: "Alpha Business",
      remark: "Updated alpha tenant remark",
    });
    expect(updateResponse.request().postDataJSON()).not.toHaveProperty("code");
    expect(updateResponse.request().postDataJSON()).not.toHaveProperty("plan");
    await expect(this.page.getByTestId("tenant-form")).toHaveCount(0);
  }

  private async expectTenantModalInlineLayout() {
    await this.expectTenantFieldInline("tenant-code-input", "租户编码");
    await this.expectTenantFieldInline("tenant-name-input", "租户名称");
    await this.expectTenantFieldInline("tenant-remark-input", "备注", "top");
  }

  private async expectTenantFieldInline(
    inputTestId: string,
    labelText: string,
    verticalAlignment: "center" | "top" = "center",
  ) {
    const input = this.page.getByTestId(inputTestId);
    const formItem = input.locator(
      `xpath=ancestor::*[contains(normalize-space(.), "${labelText}")][1]`,
    );
    const label = formItem
      .locator(
        `xpath=.//*[contains(normalize-space(.), "${labelText}") and not(self::input) and not(self::textarea)][1]`,
      )
      .first();
    await expect(label).toContainText(labelText);

    const [labelBox, inputBox] = await Promise.all([
      label.boundingBox(),
      input.boundingBox(),
    ]);
    if (!labelBox || !inputBox) {
      throw new Error(`Tenant modal field ${inputTestId} is not visible`);
    }

    expect(labelBox.x).toBeLessThan(inputBox.x);
    expect(labelBox.x + labelBox.width).toBeLessThanOrEqual(inputBox.x + 8);

    if (verticalAlignment === "top") {
      expect(Math.abs(labelBox.y - inputBox.y)).toBeLessThanOrEqual(20);
      return;
    }

    const labelCenterY = labelBox.y + labelBox.height / 2;
    const inputCenterY = inputBox.y + inputBox.height / 2;
    expect(Math.abs(labelCenterY - inputCenterY)).toBeLessThanOrEqual(20);
  }

  private async clickTenantModalConfirm(title: string) {
    await this.page
      .getByRole("dialog", { name: title })
      .getByRole("button", { name: /^确\s*认$|^Confirm$/ })
      .click();
  }

  private async clickUserDrawerConfirm(title: string) {
    await this.page
      .getByRole("dialog", { name: title })
      .getByRole("button", { name: /^确\s*认$|^Confirm$/ })
      .click();
  }

  async expectMultiTenantDisabledFallback() {
    await expect(this.page.getByTestId("platform-tenants-page")).toHaveCount(0);
    await expect(this.page.getByTestId("tenant-switcher")).toHaveCount(0);
    await expect(this.page).toHaveURL(/\/dashboard\/analytics/);
    await expect(
      this.page.getByTestId("dashboard-analytics-page"),
    ).toBeVisible();
  }

  async expectSystemUserTenantWorkbench() {
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(
      this.page.getByText(/插件页面未找到|Plugin page not found/),
    ).toHaveCount(0);
    await expect(this.page.getByText("admin", { exact: true })).toBeVisible();
    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-admin",
      }),
    ).toContainText("Alpha BU");
    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-beta-ops",
      }),
    ).toContainText("Beta BU");
    await expect(
      this.page.locator(".vxe-header--column", { hasText: "所属租户" }).first(),
    ).toBeVisible();
    if (this.workbenchMode === "platform") {
      await expect(this.page.getByTestId("user-tenant-filter")).toBeVisible();
      await this.expectUserTenantFilter();
      await this.expectUserTenantDrawer();
    } else {
      await expect(this.page.getByTestId("user-tenant-filter")).toHaveCount(0);
      await expect(
        this.page.getByTestId("user-drawer-tenant-select"),
      ).toHaveCount(0);
    }
    await expect(this.page.getByTestId("tenant-switcher")).toBeVisible();
  }

  private async expectUserTenantFilter() {
    await this.page.getByTestId("user-tenant-filter").click();
    const alphaOption = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option", {
        hasText: "Alpha BU",
      })
      .first();
    await expect(alphaOption).toBeVisible();
    await expect(
      this.page.locator(
        ".ant-select-dropdown:visible .ant-select-item-option",
        {
          hasText: "Beta BU",
        },
      ),
    ).toHaveCount(0);
    await alphaOption.click();
    const userListResponsePromise = this.page.waitForResponse((response) => {
      if (!/\/api(?:\/v1)?\/user(?:\?.*)?$/.test(response.url())) {
        return false;
      }
      return new URL(response.url()).searchParams.get("tenantId") === "101";
    });
    await this.page.getByRole("button", { name: /^搜\s*索$|^Search$/ }).click();
    await userListResponsePromise;
    expect(this.lastUserTenantFilter).toBe("101");
    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-admin",
      }),
    ).toContainText("Alpha BU");
    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-beta-ops",
      }),
    ).toHaveCount(0);
  }

  private async expectUserTenantDrawer() {
    await this.page.getByRole("button", { name: /^新\s*增$|^Add$/ }).click();
    await expect(
      this.page.getByRole("dialog", { name: "新增用户" }),
    ).toBeVisible();
    await this.selectUserDrawerTenants(["Alpha BU"]);

    const createResponsePromise = this.page.waitForResponse((response) => {
      const request = response.request();
      return (
        request.method() === "POST" && routePath(response.url()) === "/user"
      );
    });
    const createDialog = this.page.getByRole("dialog", { name: "新增用户" });
    await createDialog
      .getByPlaceholder(/请输入账号|请输入用户名/)
      .fill("drawer-tenant-user");
    await createDialog.getByPlaceholder("请输入密码").fill("admin123");
    await createDialog
      .getByPlaceholder("请输入昵称")
      .fill("Drawer Tenant User");
    await this.clickUserDrawerConfirm("新增用户");
    await createResponsePromise;
    expect(this.lastUserCreatePayload).toMatchObject({
      tenantIds: [101],
      username: "drawer-tenant-user",
    });
    await expect(createDialog).toHaveCount(0);

    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-admin",
      }),
    ).toBeVisible();
    const userRow = this.page
      .locator(".vxe-table--main-wrapper .vxe-body--row:visible", {
        hasText: "tenant-admin",
      })
      .first();
    const userRowID = await userRow.getAttribute("rowid");
    expect(userRowID, "未找到租户用户行 rowid: tenant-admin").toBeTruthy();
    await this.page
      .locator(
        `.vxe-table--fixed-right-wrapper .vxe-body--row[rowid=\"${userRowID}\"]`,
      )
      .getByRole("button", { name: /^编\s*辑$|^Edit$/ })
      .click();
    const editDialog = this.page.getByRole("dialog", { name: "编辑用户" });
    await expect(editDialog).toBeVisible();
    await expect(
      this.page.getByTestId("user-drawer-tenant-select"),
    ).toContainText("Alpha BU");
    await this.page.getByTestId("user-drawer-tenant-select").click();
    await expect(
      this.page.locator(
        ".ant-select-dropdown:visible .ant-select-item-option",
        {
          hasText: "Beta BU",
        },
      ),
    ).toHaveCount(0);
    await editDialog.getByPlaceholder("请输入昵称").click();
    await expect(editDialog).toBeVisible();

    const updateResponsePromise = this.page.waitForResponse((response) => {
      const request = response.request();
      return (
        request.method() === "PUT" && routePath(response.url()) === "/user/301"
      );
    });
    await this.clickUserDrawerConfirm("编辑用户");
    await updateResponsePromise;
    expect(this.lastUserUpdatePayload).toMatchObject({
      id: 301,
      tenantIds: [101],
    });
    await expect(editDialog).toHaveCount(0);
  }

  private async selectUserDrawerTenants(labels: string[], clear = false) {
    const select = this.page.getByTestId("user-drawer-tenant-select");
    await expect(select).toBeVisible();
    if (clear) {
      await select
        .locator(".ant-select-selection-item-remove")
        .first()
        .evaluate((element) => {
          (element as HTMLElement).click();
        });
    }
    for (const label of labels) {
      await select.click();
      const option = this.page
        .locator(".ant-select-dropdown:visible .ant-select-item-option", {
          hasText: label,
        })
        .first();
      await expect(option).toBeVisible();
      await option.click();
    }
  }

  async expectTenantMemberManagementUsesUserPage() {
    await this.gotoTenantSystemUsers();
    await this.expectSystemUserTenantWorkbench();
  }

  async expectRemovedManagementRoutesFallback() {
    await this.gotoPlatformTenants();
    await this.page.goto("/platform/tenant-members");
    await waitForRouteReady(this.page);
    await expect(
      this.page.getByTestId("platform-tenant-members-page"),
    ).toHaveCount(0);
    await expect(this.page.getByText(notFoundTextPattern)).toBeVisible();

    await this.page.goto("/tenant/members");
    await waitForRouteReady(this.page);
    await expect(this.page.getByTestId("tenant-members-page")).toHaveCount(0);
    await expect(this.page).toHaveURL(/\/platform\/tenants/);

    await this.gotoTenantSystemUsers();
    await this.page.goto("/tenant/members");
    await waitForRouteReady(this.page);
    await expect(this.page.getByTestId("tenant-members-page")).toHaveCount(0);
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(this.page.getByText("用户列表")).toBeVisible();

    await this.page.goto("/tenant/plugins");
    await waitForRouteReady(this.page);
    await expect(this.page.getByTestId("tenant-plugins-page")).toHaveCount(0);
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(this.page.getByText("用户列表")).toBeVisible();
  }

  async exerciseTenantSelectionLogin() {
    await this.mockLoginTenantSelection();
    await this.page.goto("/auth/login");
    await this.page
      .locator(
        '#username, [name="username"], input[placeholder*="用户名"], input[placeholder*="username"]',
      )
      .first()
      .fill("tenant-user");
    await this.page
      .locator(
        '#password, [name="password"], input[placeholder*="密码"], input[placeholder*="password"]',
      )
      .first()
      .fill("tenant-pass");
    await this.page.locator('button[aria-label="login"]').click();
    await expect(this.page.getByTestId("login-tenant-selector")).toBeVisible();
    await expect(this.page.locator('button[aria-label="login"]')).toHaveCount(
      0,
    );
    await expect(
      this.page.locator(
        '#username, [name="username"], input[placeholder*="用户名"], input[placeholder*="username"]',
      ),
    ).toHaveCount(0);
    await expect(
      this.page.locator(
        '#password, [name="password"], input[placeholder*="密码"], input[placeholder*="password"]',
      ),
    ).toHaveCount(0);
    const tenantSelect = this.page
      .getByTestId("login-tenant-form")
      .getByRole("combobox");
    await expect(tenantSelect).toBeVisible();
    await expect(
      this.page.getByText("请选择本次要进入的租户"),
    ).toBeVisible();
    await expect(
      this.page.getByText("请输入您的账户信息以开始管理您的项目"),
    ).toHaveCount(0);
    await expect(this.page.getByTestId("login-tenant-confirm")).toBeVisible();
    const selectBox = await tenantSelect.boundingBox();
    const confirmBox = await this.page
      .getByTestId("login-tenant-confirm")
      .boundingBox();
    expect(selectBox).toBeTruthy();
    expect(confirmBox).toBeTruthy();
    expect(confirmBox!.y - (selectBox!.y + selectBox!.height)).toBeGreaterThan(
      16,
    );

    await tenantSelect.click();
    await expect(
      this.page.getByRole("option", { name: "Alpha BU (alpha)" }),
    ).toBeVisible();
    const betaOption = this.page.getByRole("option", {
      name: "Beta BU (beta)",
    });
    await expect(betaOption).toBeVisible();
    await betaOption.click();

    const selectTenantResponse = this.page.waitForResponse((response) =>
      /\/api(?:\/v1)?\/auth\/select-tenant$/.test(response.url()),
    );
    await this.page.getByTestId("login-tenant-confirm").click();
    await selectTenantResponse;
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const raw = localStorage.getItem("linapro:tenant-state") || "{}";
          return JSON.parse(raw)?.currentTenant?.id ?? 0;
        }),
      )
      .toBe(102);
  }

  async exerciseTenantSwitch() {
    await this.gotoPlatformTenants();
    const tenantSwitcher = this.page.getByTestId("tenant-switcher-select");
    await tenantSwitcher.click();
    const alphaOption = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option", {
        hasText: "Alpha BU",
      })
      .first();
    await expect(alphaOption).toBeVisible();
    const impersonationResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/platform\/tenants\/101\/impersonate$/.test(item.url()),
    );
    await alphaOption.click();
    await impersonationResponsePromise;
    await waitForRouteReady(this.page);
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(this.page.getByText("用户列表")).toBeVisible();
    await expect(this.page.getByTestId("platform-tenants-page")).toHaveCount(0);
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const parsed = JSON.parse(
            localStorage.getItem("linapro:tenant-state") || "{}",
          );
          return parsed.currentTenant?.id ?? 0;
        }),
      )
      .toBe(101);
    await expect(this.page.getByTestId("impersonation-banner")).toBeVisible();
    await expect(this.page.getByTestId("impersonation-banner")).toContainText(
      "Alpha BU",
    );
    await this.expectImpersonationBannerBeforeTenantSwitcher();
    await this.expectStoredImpersonationOriginalToken();
  }

  async exerciseTenantUserSwitch() {
    await this.prepareWorkbench("tenant");
    await this.page.goto("/system/user");
    await waitForRouteReady(this.page);
    await expect(this.page).toHaveURL(/\/system\/user/);

    const tenantSwitcher = this.page.getByTestId("tenant-switcher-select");
    await tenantSwitcher.click();
    const betaOption = this.page
      .locator(".ant-select-dropdown:visible .ant-select-item-option", {
        hasText: "Beta BU",
      })
      .first();
    await expect(betaOption).toBeVisible();
    const switchResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/auth\/switch-tenant$/.test(item.url()),
    );
    this.lastMenusAuthorization = "";
    this.lastUserInfoAuthorization = "";
    await betaOption.click();
    await switchResponsePromise;
    await waitForRouteReady(this.page);

    expect(this.lastSwitchTenantAuthorization).toBe("Bearer alpha-token");
    expect(this.lastSwitchTenantPayload).toMatchObject({ tenantId: 102 });
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(this.page.getByText("用户列表")).toBeVisible();
    expect(this.lastUserInfoAuthorization).toBe("Bearer beta-token");
    expect(this.lastMenusAuthorization).toBe("Bearer beta-token");
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const accessState = JSON.parse(
            localStorage.getItem("lina-web-antd-5.6.0-dev-core-access") || "{}",
          );
          const tenantState = JSON.parse(
            localStorage.getItem("linapro:tenant-state") || "{}",
          );
          return {
            accessToken: accessState.accessToken,
            currentTenantId: tenantState.currentTenant?.id ?? 0,
          };
        }),
      )
      .toEqual({
        accessToken: "beta-token",
        currentTenantId: 102,
      });
  }

  async exerciseImpersonation() {
    await this.gotoPlatformTenants();
    await new MainLayout(this.page).switchLanguage("English");
    const impersonationResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/platform\/tenants\/101\/impersonate$/.test(item.url()),
    );
    await this.page.getByTestId("tenant-impersonate-101").click();
    await impersonationResponsePromise;
    await waitForRouteReady(this.page);
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(
      this.page.getByText(/用户列表|User List/).first(),
    ).toBeVisible();
    await expect(this.page.getByTestId("platform-tenants-page")).toHaveCount(0);
    await expect(this.page.getByTestId("impersonation-banner")).toBeVisible();
    await expect(this.page.getByTestId("impersonation-banner")).toContainText(
      "Alpha BU",
    );
    await this.expectCompactEnglishImpersonationBanner();
    await this.expectImpersonationBannerBeforeTenantSwitcher();
    await this.expectStoredImpersonationOriginalToken();
    const endImpersonateResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/platform\/tenants\/101\/end-impersonate$/.test(
        item.url(),
      ),
    );
    await this.page.getByTestId("impersonation-exit").click();
    await endImpersonateResponsePromise;
    await waitForRouteReady(this.page);
    expect(this.lastEndImpersonateAuthorization).toBe(
      "Bearer impersonation-token",
    );
    await expect(this.page.getByTestId("impersonation-banner")).toHaveCount(0);
    await expect(this.page.getByTestId("platform-tenants-page")).toBeVisible();
    await this.expectPlatformUserListAfterImpersonationExit();
  }

  async exerciseDirectImpersonationDefaultRoute() {
    await this.gotoPlatformTenants();
    const impersonationResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/platform\/tenants\/101\/impersonate$/.test(item.url()),
    );
    await this.page.getByTestId("tenant-impersonate-101").click();
    await impersonationResponsePromise;
    await waitForRouteReady(this.page);
    await expect(this.page).toHaveURL(/\/system\/user/);
    await expect(this.page.getByText("用户列表")).toBeVisible();
    await expect(this.page.getByTestId("platform-tenants-page")).toHaveCount(0);
    await expect(this.page.getByTestId("impersonation-banner")).toContainText(
      "Alpha BU",
    );

    const endImpersonateResponsePromise = this.page.waitForResponse((item) =>
      /\/api(?:\/v1)?\/platform\/tenants\/101\/end-impersonate$/.test(
        item.url(),
      ),
    );
    await this.page.getByTestId("impersonation-exit").click();
    await endImpersonateResponsePromise;
    await waitForRouteReady(this.page);
    await expect(this.page).toHaveURL(/\/platform\/tenants/);
    await expect(this.page.getByTestId("platform-tenants-page")).toBeVisible();
    await expect(this.page.getByTestId("impersonation-banner")).toHaveCount(0);
  }

  private async expectImpersonationBannerBeforeTenantSwitcher() {
    const [bannerBox, switcherBox] = await Promise.all([
      this.page.getByTestId("impersonation-banner").boundingBox(),
      this.page.getByTestId("tenant-switcher").boundingBox(),
    ]);
    if (!bannerBox || !switcherBox) {
      throw new Error("Impersonation banner or tenant switcher is not visible");
    }
    expect(bannerBox.x + bannerBox.width).toBeLessThanOrEqual(
      switcherBox.x + 2,
    );
  }

  private async expectCompactEnglishImpersonationBanner() {
    await expect(this.page.locator("html")).toHaveAttribute("lang", "en-US");
    await expect(this.page.getByTestId("impersonation-exit")).toHaveText(
      "Exit",
    );
    await expect(
      this.page.getByTestId("impersonation-banner-text"),
    ).toHaveAttribute(
      "title",
      "Acting as platform administrator for tenant Alpha BU",
    );
    const bannerBox = await this.page
      .getByTestId("impersonation-banner")
      .boundingBox();
    if (!bannerBox) {
      throw new Error("Impersonation banner is not visible");
    }
    expect(bannerBox.width).toBeLessThanOrEqual(242);
  }

  private async expectStoredImpersonationOriginalToken() {
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          return (
            localStorage.getItem(
              "linapro:tenant-impersonation-original-token",
            ) || ""
          );
        }),
      )
      .toBe("platform-token");
  }

  private async expectPlatformUserListAfterImpersonationExit() {
    this.lastPlatformUserAuthorization = "";
    this.lastPlatformUserTenantCode = "__unset__";
    await this.page.goto("/system/user");
    await waitForRouteReady(this.page);
    await expect(
      this.page.locator(".vxe-body--row:visible", {
        hasText: "tenant-beta-ops",
      }),
    ).toContainText("Beta BU");
    expect(this.lastPlatformUserAuthorization).toBe("Bearer platform-token");
    expect(this.lastPlatformUserTenantCode).toBe("");
    await expect
      .poll(async () =>
        this.page.evaluate(() => {
          const accessState = JSON.parse(
            localStorage.getItem("lina-web-antd-5.6.0-dev-core-access") || "{}",
          );
          const tenantState = JSON.parse(
            localStorage.getItem("linapro:tenant-state") || "{}",
          );
          return {
            accessToken: accessState.accessToken,
            currentTenantId: tenantState.currentTenant?.id ?? 0,
            impersonationActive: tenantState.impersonation?.active === true,
            originalToken:
              localStorage.getItem(
                "linapro:tenant-impersonation-original-token",
              ) || "",
          };
        }),
      )
      .toEqual({
        accessToken: "platform-token",
        currentTenantId: 0,
        impersonationActive: false,
        originalToken: "",
      });
  }

  async expectPlatformTenantUiDifference() {
    await this.gotoPlatformTenants();
    await expect(this.page.getByTestId("platform-tenants-page")).toBeVisible();
    await this.page.goto("/tenant/members");
    await waitForRouteReady(this.page);
    await expect(this.page.getByTestId("tenant-members-page")).toHaveCount(0);
    await expect(this.page.getByText(notFoundTextPattern)).toBeVisible();

    await this.gotoTenantSystemUsers();
    await expect(this.page.getByTestId("tenant-members-page")).toHaveCount(0);
    await this.page.goto("/platform/tenants");
    await waitForRouteReady(this.page);
    await expect(this.page.getByTestId("platform-tenants-page")).toHaveCount(0);
    await expect(this.page).toHaveURL(/\/system\/user/);
  }

  async expectScenarioEvidence(label: string) {
    await this.page.evaluate((value) => {
      document.body.dataset.multiTenantScenario = value;
    }, label);
    await expect
      .poll(async () =>
        this.page.evaluate(() => document.body.dataset.multiTenantScenario),
      )
      .toBe(label);
  }
}
