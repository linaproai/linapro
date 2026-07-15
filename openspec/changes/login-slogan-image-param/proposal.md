# 登录页 Slogan 插画系统参数

## 背景

登录页在左/右布局时会展示侧栏 slogan 插画。当前仅使用内置 `SloganIcon` SVG，或通过组件 prop `sloganImage` 覆盖；管理员无法在系统参数中配置自定义插画地址。

## 目标

新增受保护公共前端系统参数 `sys.auth.sloganImage`，允许管理员配置登录页 slogan 插画图片地址，并经 `/api/v1/config/public/frontend` 下发给登录页。默认使用 Vben 内置插画地址 `/slogan.svg`；空值表示不展示插画。

## 范围

- 宿主 `sys_config` 内置参数与 SQL 种子
- 公共前端配置 API 字段 `auth.sloganImage`
- 登录布局消费该配置；空值回退内置 SVG
- 相关 i18n 元数据、单元测试与 E2E

## 非目标

- 不改变左/中/右布局语义；居中布局仍不展示侧栏插画（既有行为）
- 不引入多语言 slogan 图片资源管理
- 不强制要求上传组件改造；参数值为 URL/路径文本
