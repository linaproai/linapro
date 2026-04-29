import { test, expect } from "../../fixtures/auth";
import { waitForRouteReady } from "../../support/ui";

const rawI18nKeyPattern =
  /\b(?:authentication|common|config|demos|dict|error|job|menu|notify|page|pages|plugin|preferences|profile|ui|validation)\.[A-Za-z][A-Za-z0-9_.:-]+\b/g;

const hostPageAuditCases = [
  {
    forbiddenTexts: ["Analytics", "Workspace", "工作台", "仪表盘"],
    path: "/dashboard/analytics",
    visibleTexts: ["工作臺", "訪問量"],
  },
  {
    forbiddenTexts: ["Workspace", "Quick Navigation", "工作台", "快捷导航"],
    path: "/dashboard/workspace",
    visibleTexts: ["工作臺"],
  },
  {
    forbiddenTexts: ["Username", "User Management", "用户名", "用户管理"],
    path: "/system/user",
    visibleTexts: ["用戶管理", "用戶列表"],
  },
  {
    forbiddenTexts: ["Role Management", "Role Name", "角色名称"],
    path: "/system/role",
    visibleTexts: ["角色管理", "角色列表"],
  },
  {
    forbiddenTexts: ["Menu Management", "Menu Name", "菜单管理", "菜单名称"],
    path: "/system/menu",
    visibleTexts: ["菜單管理", "菜單列表"],
  },
  {
    forbiddenTexts: ["Dictionary Management", "Dictionary Type", "字典类型"],
    path: "/system/dict",
    visibleTexts: ["字典管理", "字典類型列表"],
  },
  {
    forbiddenTexts: [
      "Config Management",
      "Config Name",
      "参数管理",
      "参数名称",
    ],
    path: "/system/config",
    visibleTexts: ["參數設置列表", "參數名稱"],
  },
  {
    forbiddenTexts: ["File Management", "File Type", "文件类型"],
    path: "/system/file",
    visibleTexts: ["文件管理", "文件列表"],
  },
  {
    forbiddenTexts: ["Scheduled Jobs", "Job Name", "任务管理", "任务名称"],
    path: "/system/job",
    visibleTexts: ["任務管理", "定時任務列表"],
  },
  {
    forbiddenTexts: ["Job Groups", "Group Code", "任务组", "分组编码"],
    path: "/system/job-group",
    visibleTexts: ["任務分組列表", "分組編碼"],
  },
  {
    forbiddenTexts: ["Job Logs", "Job Snapshot", "任务日志", "任务快照"],
    path: "/system/job-log",
    visibleTexts: ["執行日誌列表", "錯誤摘要"],
  },
  {
    forbiddenTexts: ["Plugin Management", "Synchronize Plugins"],
    path: "/system/plugin",
    visibleTexts: ["插件管理", "插件列表"],
  },
  {
    forbiddenTexts: ["System Info", "Backend", "系统信息", "后端"],
    path: "/about/system-info",
    visibleTexts: ["系統信息", "後端"],
  },
  {
    forbiddenTexts: ["API Docs", "API Documentation", "接口文档"],
    path: "/about/api-docs",
    visibleTexts: ["接口文檔"],
  },
  {
    forbiddenTexts: ["Profile", "Personal Center", "个人中心"],
    path: "/profile",
    visibleTexts: ["個人中心"],
  },
] as const;

function assertNoRawI18nKeys(bodyText: string, path: string) {
  const rawKeys = [
    ...new Set(
      [...bodyText.matchAll(rawI18nKeyPattern)]
        .filter(
          (match) =>
            bodyText.slice(Math.max(0, (match.index ?? 0) - 4), match.index) !==
            "sys.",
        )
        .map((match) => match[0]),
    ),
  ];
  expect(rawKeys, `${path} still shows raw i18n keys`).toEqual([]);
}

test.describe("TC0127 繁体中文宿主页面内容巡检", () => {
  test("TC-127a: 默认宿主菜单页面展示繁体中文标签且不泄漏中英文源标签", async ({
    adminPage,
    mainLayout,
  }) => {
    await mainLayout.switchLanguage("繁體中文");

    for (const auditCase of hostPageAuditCases) {
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

  test("TC-127b: 登录页展示繁体中文公共文案且不泄漏原始翻译键", async ({
    loginPage,
  }) => {
    await loginPage.goto();
    await loginPage.switchLanguage("繁體中文");

    await expect(loginPage.loginSubtitle).toContainText(
      "請輸入您的帳戶信息以開始管理您的項目",
    );

    const bodyText = await loginPage.getBodyText();
    expect(bodyText).toContain("歡迎回來");
    expect(bodyText).toContain("登錄");
    await expect(loginPage.usernameInput).toHaveAttribute(
      "placeholder",
      /請輸入用戶名/,
    );
    expect(bodyText).not.toContain(
      "AI-driven full-stack development framework",
    );
    expect(bodyText).not.toContain("请输入您的帐户信息以开始管理您的项目");
    assertNoRawI18nKeys(bodyText, "/auth/login");
  });
});
