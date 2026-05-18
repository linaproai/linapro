/**
 * 字典类型枚举常量
 * 用于统一管理系统中使用的字典类型标识
 */
export const DictEnum = {
  /** 系统正常/停用状态 */
  SYS_NORMAL_DISABLE: 'sys_normal_disable',
  /** 显示/隐藏状态 */
  SYS_SHOW_HIDE: 'sys_show_hide',
  /** 菜单类型 */
  SYS_MENU_TYPE: 'sys_menu_type',
  /** 用户性别 */
  SYS_USER_SEX: 'sys_user_sex',
} as const;

export type DictEnumKey = keyof typeof DictEnum;
