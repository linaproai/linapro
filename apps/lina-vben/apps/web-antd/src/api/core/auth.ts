import type { TenantAwareLoginResult } from '#/api/tenant/model';

import { requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    password?: string;
    username?: string;
  }

  /** 登录接口返回值 */
  export interface LoginResult extends TenantAwareLoginResult {}

  /** 刷新 token 接口参数 */
  export interface RefreshTokenParams {
    refreshToken: string;
  }

  /** 刷新 token 接口返回值 */
  export interface RefreshTokenResult {
    accessToken: string;
    refreshToken?: string;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login', data);
}

/**
 * 退出登录
 */
export async function logoutApi() {
  return requestClient.post('/auth/logout');
}

/**
 * 刷新 access token
 */
export async function refreshTokenApi(data: AuthApi.RefreshTokenParams) {
  return requestClient.post<AuthApi.RefreshTokenResult>('/auth/refresh', data);
}
