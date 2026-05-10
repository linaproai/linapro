import type { APIResponse, Response } from "@playwright/test";

import { test, expect } from "../../fixtures/auth";
import { waitForRouteReady } from "../../support/ui";

const runtimeMessagesPath = "/api/v1/i18n/runtime/messages";
const runtimePersistentCacheKey = "linapro:i18n:runtime:en-US";
const runtimeETagPattern = /^"[a-z]{2}-[A-Z]{2}-\d+-[a-f0-9]{32}"$/;

async function expectRuntimeRevalidationResponse(
  response: APIResponse | Response,
  cachedEtag: string,
) {
  expect([200, 304]).toContain(response.status());

  if (response.status() === 304) {
    expect(response.headers()["etag"]).toBe(cachedEtag);

    const body = await response.body().catch((error: Error) => {
      expect(error.message).toContain("Response body is unavailable");
      return Buffer.from([]);
    });
    expect(body.length, "304 response must not carry a body").toBe(0);
    return;
  }

  const refreshedEtag = response.headers()["etag"];
  expect(refreshedEtag).toMatch(runtimeETagPattern);
  expect(refreshedEtag).not.toBe(cachedEtag);

  const payload = await response.json();
  const messages = payload?.data?.messages ?? payload?.messages;
  expect(messages, "expected refreshed messages payload on 200").toBeTruthy();
}

test.describe("TC0124 运行时翻译包 ETag 协商", () => {
  test("TC-124a: 首次请求返回 ETag 与 Cache-Control 头", async ({
    adminPage,
  }) => {
    const response = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { "Accept-Language": "en-US" },
      },
    );
    expect(response.ok()).toBeTruthy();
    expect(response.status()).toBe(200);

    const etag = response.headers()["etag"];
    expect(etag, "expected ETag header on first response").toBeTruthy();
    expect(etag).toMatch(runtimeETagPattern);

    const cacheControl = response.headers()["cache-control"];
    expect(cacheControl).toBe("private, must-revalidate");

    const payload = await response.json();
    const messages = payload?.data?.messages ?? payload?.messages;
    expect(
      messages,
      "expected messages payload to be present on 200",
    ).toBeTruthy();
  });

  test("TC-124b: 携带匹配 If-None-Match 时返回 304 或合法刷新响应", async ({
    adminPage,
  }) => {
    const firstResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { "Accept-Language": "en-US" },
      },
    );
    expect(firstResponse.ok()).toBeTruthy();
    const etag = firstResponse.headers()["etag"];
    expect(etag).toBeTruthy();

    const secondResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: {
          "Accept-Language": "en-US",
          "If-None-Match": etag!,
        },
      },
    );
    await expectRuntimeRevalidationResponse(secondResponse, etag!);
  });

  test("TC-124c: 不同语言独立 ETag,跨语言 If-None-Match 不命中", async ({
    adminPage,
  }) => {
    const enResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { "Accept-Language": "en-US" },
      },
    );
    const enEtag = enResponse.headers()["etag"];
    expect(enEtag).toMatch(runtimeETagPattern);

    const zhResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=zh-CN`,
      {
        headers: { "Accept-Language": "zh-CN" },
      },
    );
    const zhEtag = zhResponse.headers()["etag"];
    expect(zhEtag).toMatch(runtimeETagPattern);
    expect(zhEtag).not.toBe(enEtag);

    // Sending the en-US ETag while requesting zh-CN must not produce a 304.
    const crossResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=zh-CN`,
      {
        headers: {
          "Accept-Language": "zh-CN",
          "If-None-Match": enEtag!,
        },
      },
    );
    expect(crossResponse.status()).toBe(200);
  });

  test("TC-124d: If-None-Match 通配符 * 也命中 304", async ({ adminPage }) => {
    const firstResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: { "Accept-Language": "en-US" },
      },
    );
    const etag = firstResponse.headers()["etag"];
    expect(etag).toBeTruthy();

    const wildcardResponse = await adminPage.request.get(
      `${runtimeMessagesPath}?lang=en-US`,
      {
        headers: {
          "Accept-Language": "en-US",
          "If-None-Match": "*",
        },
      },
    );
    expect(wildcardResponse.status()).toBe(304);
    expect(wildcardResponse.headers()["etag"]).toBe(etag);
  });

  test("TC-124e: 前端持久化缓存会在页面重载时发送 If-None-Match", async ({
    adminPage,
    mainLayout,
  }) => {
    await adminPage.evaluate(
      (key) => localStorage.removeItem(key),
      runtimePersistentCacheKey,
    );

    await mainLayout.switchLanguage("English");

    await expect
      .poll(async () =>
        adminPage.evaluate(
          (key) => localStorage.getItem(key),
          runtimePersistentCacheKey,
        ),
      )
      .not.toBeNull();
    const cachedEntry = await adminPage.evaluate(
      (key) => localStorage.getItem(key),
      runtimePersistentCacheKey,
    );
    const etag = JSON.parse(cachedEntry || "{}")?.etag;
    expect(etag, "expected runtime i18n ETag in persistent cache").toMatch(
      runtimeETagPattern,
    );

    const revalidationRequest = adminPage.waitForRequest((request) => {
      if (!request.url().includes(runtimeMessagesPath)) {
        return false;
      }
      if (!request.url().includes("lang=en-US")) {
        return false;
      }
      return request.headers()["if-none-match"] === etag;
    });
    const revalidationResponse = adminPage.waitForResponse((response) => {
      const request = response.request();
      return (
        response.url().includes(runtimeMessagesPath) &&
        response.url().includes("lang=en-US") &&
        request.headers()["if-none-match"] === etag
      );
    });

    await adminPage.reload({ waitUntil: "domcontentloaded" });
    await waitForRouteReady(adminPage, 15_000);

    await revalidationRequest;
    const response = await revalidationResponse;
    await expectRuntimeRevalidationResponse(response, etag);
  });
});
