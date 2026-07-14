import type { TenantAwareLoginResult } from '#/api/tenant/model';

import { requestClient } from '#/api/request';

export namespace AuthApi {
  /** 登录接口参数 */
  export interface LoginParams {
    clientType?: 'cli' | 'desktop' | 'mobile' | 'web';
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

  /** 公开注册参数 */
  export interface RegisterParams {
    email: string;
    nickname?: string;
    password: string;
    username: string;
  }

  /** 公开注册返回值 */
  export interface RegisterResult {
    userId: number;
  }

  /** 忘记密码参数 */
  export interface ForgetPasswordParams {
    email: string;
  }

  /** 忘记密码返回值 */
  export interface ForgetPasswordResult {
    accepted: boolean;
  }

  /** 确认重置密码参数 */
  export interface ResetPasswordParams {
    password: string;
    token: string;
  }

  /** 确认重置密码返回值 */
  export interface ResetPasswordResult {
    reset: boolean;
  }
}

/**
 * 登录
 */
export async function loginApi(data: AuthApi.LoginParams) {
  return requestClient.post<AuthApi.LoginResult>('/auth/login', {
    ...data,
    clientType: 'web',
  });
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

/**
 * 公开注册账号
 */
export async function registerApi(data: AuthApi.RegisterParams) {
  return requestClient.post<AuthApi.RegisterResult>('/auth/register', data);
}

/**
 * 请求密码重置邮件
 */
export async function forgetPasswordApi(data: AuthApi.ForgetPasswordParams) {
  return requestClient.post<AuthApi.ForgetPasswordResult>(
    '/auth/forget-password',
    data,
  );
}

/**
 * 使用重置令牌设置新密码
 */
export async function resetPasswordApi(data: AuthApi.ResetPasswordParams) {
  return requestClient.post<AuthApi.ResetPasswordResult>(
    '/auth/reset-password',
    data,
  );
}
