import axios from 'axios';
import MockAdapter from 'axios-mock-adapter';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  authenticateResponseInterceptor,
  errorMessageResponseInterceptor,
} from './preset-interceptors';
import { RequestClient } from './request-client';

describe('requestClient', () => {
  let mock: MockAdapter;
  let requestClient: RequestClient;

  beforeEach(() => {
    mock = new MockAdapter(axios);
    requestClient = new RequestClient();
  });

  afterEach(() => {
    mock.reset();
  });

  it('should successfully make a GET request', async () => {
    mock.onGet('test/url').reply(200, { data: 'response' });

    const response = await requestClient.get('test/url');

    expect(response.data).toEqual({ data: 'response' });
  });

  it('should successfully make a POST request', async () => {
    const postData = { key: 'value' };
    const mockData = { data: 'response' };
    mock.onPost('/test/post', postData).reply(200, mockData);
    const response = await requestClient.post('/test/post', postData);
    expect(response.data).toEqual(mockData);
  });

  it('should successfully make a PUT request', async () => {
    const putData = { key: 'updatedValue' };
    const mockData = { data: 'updated response' };
    mock.onPut('/test/put', putData).reply(200, mockData);
    const response = await requestClient.put('/test/put', putData);
    expect(response.data).toEqual(mockData);
  });

  it('should successfully make a DELETE request', async () => {
    const mockData = { data: 'delete response' };
    mock.onDelete('/test/delete').reply(200, mockData);
    const response = await requestClient.delete('/test/delete');
    expect(response.data).toEqual(mockData);
  });

  it('should handle network errors', async () => {
    mock.onGet('/test/error').networkError();
    try {
      await requestClient.get('/test/error');
      expect(true).toBe(false);
    } catch (error: any) {
      expect(error.isAxiosError).toBe(true);
      expect(error.message).toBe('Network Error');
    }
  });

  it('should handle timeout', async () => {
    mock.onGet('/test/timeout').timeout();
    try {
      await requestClient.get('/test/timeout');
      expect(true).toBe(false);
    } catch (error: any) {
      expect(error.isAxiosError).toBe(true);
      expect(error.code).toBe('ECONNABORTED');
    }
  });

  it('should skip global error messages for silent requests', async () => {
    const client = new RequestClient();
    const localMock = new MockAdapter(client.instance);
    const makeErrorMessage = vi.fn();
    client.addResponseInterceptor(
      errorMessageResponseInterceptor(makeErrorMessage),
    );

    localMock.onGet('/test/default-error').networkErrorOnce();
    await expect(client.get('/test/default-error')).rejects.toMatchObject({
      message: 'Network Error',
    });
    expect(makeErrorMessage).toHaveBeenCalledTimes(1);

    localMock.onGet('/test/silent-error').networkErrorOnce();
    await expect(
      client.get('/test/silent-error', { silentErrorMessage: true }),
    ).rejects.toMatchObject({
      message: 'Network Error',
    });
    expect(makeErrorMessage).toHaveBeenCalledTimes(1);

    localMock.reset();
  });

  it('should successfully upload a file', async () => {
    const fileData = new Blob(['file contents'], { type: 'text/plain' });

    mock.onPost('/test/upload').reply((config) => {
      return config.data instanceof FormData && config.data.has('file')
        ? [200, { data: 'file uploaded' }]
        : [400, { error: 'Bad Request' }];
    });

    const response = await requestClient.upload('/test/upload', {
      file: fileData,
    });
    expect(response.data).toEqual({ data: 'file uploaded' });
  });

  it('should successfully download a file as a blob', async () => {
    const mockFileContent = new Blob(['mock file content'], {
      type: 'text/plain',
    });

    mock.onGet('/test/download').reply(200, mockFileContent);

    const res = await requestClient.download('/test/download');

    expect(res.data).toBeInstanceOf(Blob);
  });

  it('should reject queued 401 requests when refresh fails and not replay them', async () => {
    // Two concurrent requests both receive 401. The first triggers a refresh
    // attempt and the second is queued. When the refresh fails, the queued
    // request must reject with the refresh error instead of being replayed
    // with an empty token (which would re-enter the 401 → refresh loop).
    const client = new RequestClient();
    const doRefreshToken = vi
      .fn()
      .mockRejectedValue(new Error('refresh failed'));
    const doReAuthenticate = vi.fn().mockResolvedValue(undefined);
    client.addResponseInterceptor(
      authenticateResponseInterceptor({
        client,
        doReAuthenticate,
        doRefreshToken,
        enableRefreshToken: true,
        formatToken: (token: null | string) =>
          token ? `Bearer ${token}` : null,
      }),
    );

    const localMock = new MockAdapter(client.instance);
    let aCalls = 0;
    let bCalls = 0;
    localMock.onGet('/protected/a').reply(() => {
      aCalls += 1;
      return [401, { code: 401, message: 'Unauthorized' }];
    });
    localMock.onGet('/protected/b').reply(() => {
      bCalls += 1;
      return [401, { code: 401, message: 'Unauthorized' }];
    });

    const results = await Promise.allSettled([
      client.get('/protected/a'),
      client.get('/protected/b'),
    ]);

    expect(results[0].status).toBe('rejected');
    expect(results[1].status).toBe('rejected');
    // Each protected URL must be hit exactly once. Without this fix the
    // queued request would be replayed with an empty token and trigger a
    // second 401 → refresh chain.
    expect(aCalls).toBe(1);
    expect(bCalls).toBe(1);
    expect(doRefreshToken).toHaveBeenCalledTimes(1);
    expect(doReAuthenticate).toHaveBeenCalledTimes(1);
    expect(client.isRefreshing).toBe(false);
    expect(client.refreshTokenQueue).toEqual([]);

    localMock.reset();
  });

  it('should not hang requests that 401 during the doReAuthenticate window', async () => {
    // After refresh fails, the interceptor still has `isRefreshing = true`
    // while it awaits doReAuthenticate. A 401 that arrives in that window
    // gets queued, and without a second drain it would never settle. This
    // test makes doReAuthenticate slow enough to interleave a late 401.
    const client = new RequestClient();
    const doRefreshToken = vi
      .fn()
      .mockRejectedValue(new Error('refresh failed'));
    let releaseReauth: () => void = () => undefined;
    const reauthGate = new Promise<void>((resolve) => {
      releaseReauth = resolve;
    });
    const doReAuthenticate = vi
      .fn()
      .mockImplementation(async () => reauthGate);
    client.addResponseInterceptor(
      authenticateResponseInterceptor({
        client,
        doReAuthenticate,
        doRefreshToken,
        enableRefreshToken: true,
        formatToken: (token: null | string) =>
          token ? `Bearer ${token}` : null,
      }),
    );

    const localMock = new MockAdapter(client.instance);
    localMock
      .onGet('/protected/first')
      .reply(401, { code: 401, message: 'Unauthorized' });
    localMock
      .onGet('/protected/late')
      .reply(401, { code: 401, message: 'Unauthorized' });

    const firstPending = client.get('/protected/first');
    // Yield long enough for the first response interceptor to fail refresh,
    // start awaiting doReAuthenticate, and leave isRefreshing = true.
    await new Promise((resolve) => setTimeout(resolve, 10));
    expect(client.isRefreshing).toBe(true);

    const latePending = client.get('/protected/late');
    // Yield once more so the late request actually reaches the response
    // interceptor and lands in the queue.
    await new Promise((resolve) => setTimeout(resolve, 10));
    expect(client.refreshTokenQueue.length).toBe(1);

    // Releasing the reauth gate must drain the late queue entry, not leave
    // it hanging once isRefreshing flips back to false.
    releaseReauth();

    const results = await Promise.allSettled([firstPending, latePending]);
    expect(results[0].status).toBe('rejected');
    expect(results[1].status).toBe('rejected');
    expect(doRefreshToken).toHaveBeenCalledTimes(1);
    expect(doReAuthenticate).toHaveBeenCalledTimes(1);
    expect(client.isRefreshing).toBe(false);
    expect(client.refreshTokenQueue).toEqual([]);

    localMock.reset();
  });
});
