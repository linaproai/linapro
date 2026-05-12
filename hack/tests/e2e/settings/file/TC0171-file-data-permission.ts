import fs from "node:fs";
import path from "node:path";

import type { APIRequestContext } from "@playwright/test";

import { test, expect } from "../../../fixtures/auth";
import {
  createAdminApiContext,
  createApiContext,
  createRole,
  createUser,
  deleteRole,
  deleteUser,
  expectBusinessError,
  expectSuccess,
  getMenuIdsByPermsWithAncestors,
} from "../../../support/api/job";

type UploadResult = {
  id: number;
  original: string;
};

type FileListItem = {
  id: number;
  original: string;
};

const password = "test123456";

async function uploadTextFile(
  api: APIRequestContext,
  filePath: string,
) {
  return expectSuccess<UploadResult>(
    await api.post("file/upload", {
      multipart: {
        scene: "other",
        file: fs.createReadStream(filePath),
      },
    }),
  );
}

async function listFiles(api: APIRequestContext, original: string) {
  return expectSuccess<{ list: FileListItem[]; total: number }>(
    await api.get(
      `file?pageNum=1&pageSize=100&original=${encodeURIComponent(original)}`,
    ),
  );
}

test.describe("TC-171 文件管理数据权限", () => {
  let adminApi: APIRequestContext;
  let limitedApi: APIRequestContext;
  let roleID = 0;
  let userID = 0;
  let visibleFileID = 0;
  let hiddenFileID = 0;
  let visiblePath = "";
  let hiddenPath = "";
  let visibleName = "";
  let hiddenName = "";

  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const suffix = Date.now().toString();
    visibleName = `e2e_visible_${suffix}.txt`;
    hiddenName = `e2e_hidden_${suffix}.txt`;
    visiblePath = path.join("/tmp", visibleName);
    hiddenPath = path.join("/tmp", hiddenName);
    fs.writeFileSync(visiblePath, "visible file");
    fs.writeFileSync(hiddenPath, "hidden file");

    const menuIds = await getMenuIdsByPermsWithAncestors(adminApi, [
      "system:file:query",
      "system:file:upload",
      "system:file:download",
      "system:file:remove",
    ]);
    roleID = (
      await createRole(adminApi, {
        name: `FileScope${suffix.slice(-10)}`,
        key: `e2e_file_self_${suffix}`,
        menuIds,
        dataScope: 4,
        sort: 970,
      })
    ).id;
    const username = `e2e_file_scope_${suffix}`;
    userID = (
      await createUser(adminApi, {
        username,
        password,
        nickname: "E2E File Scope",
        roleIds: [roleID],
      })
    ).id;
    limitedApi = await createApiContext(username, password);

    hiddenFileID = (await uploadTextFile(adminApi, hiddenPath)).id;
    visibleFileID = (await uploadTextFile(limitedApi, visiblePath)).id;
  });

  test.afterAll(async () => {
    if (visibleFileID > 0) {
      await adminApi.delete(`file/${visibleFileID}`).catch(() => {});
    }
    if (hiddenFileID > 0) {
      await adminApi.delete(`file/${hiddenFileID}`).catch(() => {});
    }
    await limitedApi?.post("auth/logout").catch(() => {});
    await limitedApi?.dispose();
    if (userID > 0) {
      await deleteUser(adminApi, userID).catch(() => {});
    }
    if (roleID > 0) {
      await deleteRole(adminApi, roleID).catch(() => {});
    }
    await adminApi?.dispose();
    for (const filePath of [visiblePath, hiddenPath]) {
      if (filePath && fs.existsSync(filePath)) {
        fs.unlinkSync(filePath);
      }
    }
  });

  test("TC-171a~c: 文件列表、下载和删除按上传人范围过滤", async () => {
    const visibleFiles = await listFiles(limitedApi, "e2e_");
    const visibleOriginals = visibleFiles.list.map((item) => item.original);
    expect(visibleOriginals).toContain(visibleName);
    expect(visibleOriginals).not.toContain(hiddenName);

    await expectSuccess(await limitedApi.get(`file/detail/${visibleFileID}`));
    await expectBusinessError(
      await limitedApi.get(`file/detail/${hiddenFileID}`),
    );
    await expectBusinessError(
      await limitedApi.get(`file/download/${hiddenFileID}`),
    );
    await expectBusinessError(await limitedApi.delete(`file/${hiddenFileID}`));

    const hiddenStillExists = await listFiles(adminApi, hiddenName);
    expect(hiddenStillExists.list.map((item) => item.id)).toContain(
      hiddenFileID,
    );
  });
});
