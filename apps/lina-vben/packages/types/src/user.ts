import type { MenuRecordRaw } from '@vben-core/typings';

/** 用户信息 */
interface UserInfo {
  /**
   * 用户描述
   */
  desc: string;
  /**
   * 首页地址
   */
  homePath: string;

  /**
   * accessToken
   */
  token: string;
  /**
   * 用户菜单树
   */
  menus?: MenuRecordRaw[];
  /**
   * 用户权限标识列表
   */
  permissions?: string[];
}

export type { UserInfo };
