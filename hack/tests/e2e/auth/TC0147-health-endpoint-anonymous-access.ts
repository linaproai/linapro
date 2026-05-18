import { test, expect } from '../../fixtures/auth';

type HealthPayload = {
  data?: HealthResponse;
  mode?: string;
  status?: string;
};

type HealthResponse = {
  mode?: string;
  status: string;
};

test.describe('TC-147 Health endpoint anonymous access', () => {
  test('TC-147a: anonymous health probe returns ok status and mode', async ({
    request,
  }) => {
    const response = await request.get('/api/v1/health');

    expect(response.status()).toBe(200);
    const payload = (await response.json()) as HealthPayload;
    const health = payload.data ?? (payload as HealthResponse);

    expect(health.status).toBe('ok');
    expect(['single', 'master', 'slave']).toContain(health.mode);
  });
});
