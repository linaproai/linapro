import { test, expect } from "../../fixtures/auth";

test.describe("TC0129 繁体中文接口文档", () => {
  test("TC-129a: api.json 按 zh-TW 返回接口分组、摘要和参数说明", async ({
    adminPage,
  }) => {
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
  });
});
