import type { APIRequestContext } from "@playwright/test";

import { Buffer } from "node:buffer";

import { test, expect } from "../../../fixtures/auth";
import {
  createAdminApiContext,
  createApiContext,
  expectSuccess,
} from "../../../support/api/job";

type CreatedID = {
  id: number;
};

type UploadedFile = {
  id: number;
  url: string;
};

let adminApi: APIRequestContext;
let limitedApi: APIRequestContext;
let limitedRoleID = 0;
let limitedUserID = 0;
let uploadedFileID = 0;
const limitedPassword = "test123456";
const missingUploadPath = "uploads/e2e-missing-file.txt";

test.describe("TC-151 Uploads route requires auth and permission", () => {
  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
    const suffix = Date.now();
    const role = await expectSuccess<CreatedID>(
      await adminApi.post("role", {
        data: {
          dataScope: 1,
          key: `e2e_up_${suffix}`,
          name: `E2E Up ${suffix}`,
          sort: 999,
          status: 1,
        },
      }),
    );
    limitedRoleID = role.id;

    const user = await expectSuccess<CreatedID>(
      await adminApi.post("user", {
        data: {
          nickname: `E2E Upload ${suffix}`,
          password: limitedPassword,
          roleIds: [limitedRoleID],
          status: 1,
          username: `e2e_upload_limited_${suffix}`,
        },
      }),
    );
    limitedUserID = user.id;
    limitedApi = await createApiContext(
      `e2e_upload_limited_${suffix}`,
      limitedPassword,
    );
  });

  test.afterAll(async () => {
    await limitedApi?.dispose();
    if (uploadedFileID > 0) {
      await adminApi.delete(`file?ids=${uploadedFileID}`).catch(() => {});
    }
    if (limitedUserID > 0) {
      await adminApi.delete(`user?ids=${limitedUserID}`).catch(() => {});
    }
    if (limitedRoleID > 0) {
      await adminApi.delete(`role?ids=${limitedRoleID}`).catch(() => {});
    }
    await adminApi?.dispose();
  });

  test("TC-151a: anonymous upload file access is rejected before filesystem lookup", async ({
    request,
  }) => {
    const response = await request.get(`/api/v1/${missingUploadPath}`);

    expect(response.status()).toBe(401);
  });

  test("TC-151b: signed-in user without file download permission is forbidden", async () => {
    const response = await limitedApi.get(missingUploadPath);

    expect(response.status()).toBe(403);
  });

  test("TC-151c: admin with file download permission reaches the protected file handler", async () => {
    const content = `storage-backed-upload-access-${Date.now()}`;
    const upload = await expectSuccess<UploadedFile>(
      await adminApi.post("file/upload", {
        multipart: {
          file: {
            name: "e2e-upload-access.txt",
            mimeType: "text/plain",
            buffer: Buffer.from(content),
          },
          scene: "other",
        },
      }),
    );
    uploadedFileID = upload.id;

    const accessPath = new URL(upload.url, "http://lina.local").pathname.replace(
      /^\/api\/v1\//,
      "",
    );
    const response = await adminApi.get(accessPath);

    expect(response.status()).toBe(200);
    expect(response.headers()["content-type"]).toMatch(/application\/octet-stream/);
    expect(await response.text()).toBe(content);
  });
});
