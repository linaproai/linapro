import { test, expect } from '@host-tests/fixtures/auth';
import {
  createAdminApiContext,
  createApiContext,
  createRole,
  createUser,
  deleteRole,
  deleteUser,
  disablePlugin,
  enablePlugin,
  expectBusinessError,
  expectSuccess,
  getPlugin,
  installPlugin,
  syncPlugins,
  uninstallPlugin,
} from '@host-tests/support/api/job';

type APIRequestContext = Awaited<ReturnType<typeof createAdminApiContext>>;

type PluginState = {
  enabled: number;
  installed: number;
};

type MessageItem = {
  id: number;
  isRead: number;
  title: string;
};

type NoticeCreateResult = {
  id: number;
};

const password = "test123456";
const pluginID = "content-notice";

async function ensurePluginEnabled(api: APIRequestContext, id: string) {
  await syncPlugins(api);
  const plugin = await getPlugin(api, id);
  if (plugin.installed !== 1) {
    await installPlugin(api, id);
  }
  if (plugin.enabled !== 1) {
    await enablePlugin(api, id);
  }
  return plugin;
}

async function restorePluginState(
  api: APIRequestContext,
  id: string,
  state: PluginState,
) {
  if (state.installed !== 1) {
    await uninstallPlugin(api, id).catch(() => {});
    return;
  }
  if (state.enabled !== 1) {
    await disablePlugin(api, id).catch(() => {});
    return;
  }
  await enablePlugin(api, id).catch(() => {});
}

async function listMessages(api: APIRequestContext) {
  return expectSuccess<{ list: MessageItem[]; total: number }>(
    await api.get("user/message?pageNum=1&pageSize=100"),
  );
}

async function findMessageByTitle(api: APIRequestContext, title: string) {
  const result = await listMessages(api);
  return result.list.find((item) => item.title === title);
}

test.describe("TC-174 用户消息自隔离边界", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let otherApi: APIRequestContext;
  let originalPluginState: PluginState;
  let roleID = 0;
  let limitedUserID = 0;
  let otherUserID = 0;
  let noticeID = 0;
  let messageTitle = "";

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const plugin = await ensurePluginEnabled(adminApi, pluginID);
    originalPluginState = {
      enabled: plugin.enabled,
      installed: plugin.installed,
    };

    const suffix = Date.now().toString();
    roleID = (
      await createRole(adminApi, {
        name: `MsgScope${suffix.slice(-10)}`,
        key: `e2e_msg_all_${suffix}`,
        menuIds: [],
        dataScope: 1,
        sort: 970,
      })
    ).id;

    const limitedUsername = `e2e_msg_limited_${suffix}`;
    const otherUsername = `e2e_msg_other_${suffix}`;
    limitedUserID = (
      await createUser(adminApi, {
        username: limitedUsername,
        password,
        nickname: "E2E Message Limited",
        roleIds: [roleID],
      })
    ).id;
    otherUserID = (
      await createUser(adminApi, {
        username: otherUsername,
        password,
        nickname: "E2E Message Other",
      })
    ).id;

    limitedApi = await createApiContext(limitedUsername, password);
    otherApi = await createApiContext(otherUsername, password);
    await expectSuccess(await limitedApi.delete("user/message/clear"));
    await expectSuccess(await otherApi.delete("user/message/clear"));

    messageTitle = `E2E Message Boundary ${suffix}`;
    noticeID = (
      await expectSuccess<NoticeCreateResult>(
        await adminApi.post("notice", {
          data: {
            title: messageTitle,
            type: 1,
            content: "E2E message boundary content",
            status: 1,
          },
        }),
      )
    ).id;

    await expect
      .poll(async () => Boolean(await findMessageByTitle(limitedApi, messageTitle)))
      .toBeTruthy();
    await expect
      .poll(async () => Boolean(await findMessageByTitle(otherApi, messageTitle)))
      .toBeTruthy();
  });

  test.afterAll(async () => {
    if (noticeID > 0) {
      await adminApi.delete(`notice/${noticeID}`).catch(() => {});
    }
    await limitedApi?.delete("user/message/clear").catch(() => {});
    await otherApi?.delete("user/message/clear").catch(() => {});
    await limitedApi?.post("auth/logout").catch(() => {});
    await otherApi?.post("auth/logout").catch(() => {});
    await limitedApi?.dispose();
    await otherApi?.dispose();
    for (const userID of [limitedUserID, otherUserID]) {
      if (userID > 0) {
        await deleteUser(adminApi, userID).catch(() => {});
      }
    }
    if (roleID > 0) {
      await deleteRole(adminApi, roleID).catch(() => {});
    }
    if (originalPluginState) {
      await restorePluginState(adminApi, pluginID, originalPluginState);
    }
    await adminApi?.dispose();
  });

  test("TC-174a~c: 全部数据权限不读取、标记或删除他人消息", async () => {
    const limitedMessage = await findMessageByTitle(limitedApi, messageTitle);
    const otherMessage = await findMessageByTitle(otherApi, messageTitle);
    expect(limitedMessage).toBeTruthy();
    expect(otherMessage).toBeTruthy();
    expect(limitedMessage!.id).not.toBe(otherMessage!.id);

    await expectSuccess(await limitedApi.get(`user/message/${limitedMessage!.id}`));
    await expectBusinessError(
      await limitedApi.get(`user/message/${otherMessage!.id}`),
    );

    await expectSuccess(await limitedApi.put(`user/message/${otherMessage!.id}/read`));
    const otherAfterRead = await findMessageByTitle(otherApi, messageTitle);
    expect(otherAfterRead?.isRead).toBe(0);

    await expectSuccess(await limitedApi.delete(`user/message/${otherMessage!.id}`));
    const otherAfterDelete = await findMessageByTitle(otherApi, messageTitle);
    expect(otherAfterDelete).toBeTruthy();
  });
});
