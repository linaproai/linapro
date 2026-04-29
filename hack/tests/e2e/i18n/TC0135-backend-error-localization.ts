import type { APIRequestContext, APIResponse } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import {
  createAdminApiContext,
  enablePlugin,
  getPlugin,
  installPlugin,
  syncPlugins,
} from "../../support/api/job";

type ErrorEnvelope = {
  code: number;
  errorCode?: string;
  message?: string;
  messageKey?: string;
  messageParams?: Record<string, unknown>;
};

const sourcePluginIDs = [
  "content-notice",
  "org-center",
  "monitor-loginlog",
  "monitor-operlog",
] as const;

const errorCases = [
  {
    errorCode: "CONTENT_NOTICE_NOT_FOUND",
    messageKey: "error.content.notice.not.found",
    messages: {
      "en-US": "Notice does not exist",
      "zh-CN": "通知公告不存在",
      "zh-TW": "通知公告不存在",
    },
    path: "notice/99999999",
  },
  {
    errorCode: "ORG_DEPT_NOT_FOUND",
    messageKey: "error.org.dept.not.found",
    messages: {
      "en-US": "Department does not exist",
      "zh-CN": "部门不存在",
      "zh-TW": "部門不存在",
    },
    path: "dept/99999999",
  },
  {
    errorCode: "ORG_POST_NOT_FOUND",
    messageKey: "error.org.post.not.found",
    messages: {
      "en-US": "Post does not exist",
      "zh-CN": "岗位不存在",
      "zh-TW": "崗位不存在",
    },
    path: "post/99999999",
  },
  {
    errorCode: "MONITOR_LOGINLOG_NOT_FOUND",
    messageKey: "error.monitor.loginlog.not.found",
    messages: {
      "en-US": "Login log does not exist",
      "zh-CN": "登录日志不存在",
      "zh-TW": "登錄日誌不存在",
    },
    path: "loginlog/99999999",
  },
  {
    errorCode: "MONITOR_OPERLOG_NOT_FOUND",
    messageKey: "error.monitor.operlog.not.found",
    messages: {
      "en-US": "Operation log does not exist",
      "zh-CN": "操作日志不存在",
      "zh-TW": "操作日誌不存在",
    },
    path: "operlog/99999999",
  },
] as const;

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

async function expectBackendError(
  response: APIResponse,
): Promise<ErrorEnvelope> {
  const payload = (await response.json()) as ErrorEnvelope;
  expect(payload.code).not.toBe(0);
  return payload;
}

test.describe("TC-135 Backend error localization", () => {
  let adminApi: APIRequestContext;

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    await ensureSourcePluginsEnabled(adminApi, sourcePluginIDs);
  });

  test.afterAll(async () => {
    await adminApi.dispose();
  });

  test("TC-135a: business errors keep stable codes while messages follow request locale", async () => {
    for (const errorCase of errorCases) {
      for (const locale of ["zh-CN", "en-US", "zh-TW"] as const) {
        const payload = await expectBackendError(
          await adminApi.get(errorCase.path, {
            headers: { "Accept-Language": locale },
          }),
        );

        expect(payload.errorCode).toBe(errorCase.errorCode);
        expect(payload.messageKey).toBe(errorCase.messageKey);
        expect(payload.message).toBe(errorCase.messages[locale]);
      }
    }
  });
});
