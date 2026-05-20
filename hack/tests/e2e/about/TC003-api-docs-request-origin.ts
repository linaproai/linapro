import { test, expect } from '../../fixtures/auth';
import { config } from '../../fixtures/config';

type OpenApiDocument = {
  servers?: Array<{
    url?: string;
  }>;
};

test.describe('TC-3 API docs request origin', () => {
  test('TC-3a: OpenAPI server URL follows frontend proxy and backend direct entrypoints', async ({
    adminPage,
    request,
  }) => {
    const proxiedResponse = await adminPage.request.get('/api.json?lang=zh-CN');
    expect(proxiedResponse.ok()).toBeTruthy();
    const proxiedDocument = (await proxiedResponse.json()) as OpenApiDocument;

    expect(proxiedDocument.servers?.[0]?.url).toBe(
      config.frontendProxyBackendOrigin,
    );

    const directResponse = await request.get(
      `${config.backendBaseURL}/api.json?lang=zh-CN`,
      {
        headers: {
          Host: '127.0.0.1:18088',
        },
      },
    );
    expect(directResponse.ok()).toBeTruthy();
    const directDocument = (await directResponse.json()) as OpenApiDocument;

    expect(directDocument.servers?.[0]?.url).toBe('http://127.0.0.1:18088');
  });
});
