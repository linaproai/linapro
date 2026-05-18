import type { APIRequestContext } from "@playwright/test";

import { Buffer } from "node:buffer";

import { test, expect } from "../../../fixtures/auth";
import { createAdminApiContext, expectSuccess } from "../../../support/api/job";

type UploadedFile = {
  id: number;
  url: string;
};

let adminApi: APIRequestContext;
let uploadedFileID = 0;
const missingUploadPath = "uploads/e2e-missing-file.txt";

test.describe("TC-151 Uploads route is public and storage-backed", () => {
  test.beforeAll(async () => {
    adminApi = await createAdminApiContext();
  });

  test.afterAll(async () => {
    if (uploadedFileID > 0) {
      await adminApi.delete(`file?ids=${uploadedFileID}`).catch(() => {});
    }
    await adminApi?.dispose();
  });

  test("TC-151a: anonymous missing upload access reaches public file handler", async ({
    request,
  }) => {
    const response = await request.get(`/api/v1/${missingUploadPath}`);

    expect(response.status()).toBe(200);
  });

  test("TC-151b: signed-in user reaches public file handler", async () => {
    const response = await adminApi.get(missingUploadPath);

    expect(response.status()).toBe(200);
  });

  test("TC-151c: anonymous access can read an uploaded public file URL", async ({
    request,
  }) => {
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
      /^\/api\/v1/,
      "",
    );
    const response = await request.get(`/api/v1${accessPath}`);

    expect(response.status()).toBe(200);
    expect(response.headers()["content-type"]).toMatch(/application\/octet-stream/);
    expect(await response.text()).toBe(content);
  });
});
