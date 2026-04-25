## Why

默认管理工作台当前仍显示 Vben 开源框架默认 Logo，与 LinaPro 的项目品牌不一致。需要将工作台 Logo 与浏览器 favicon 替换为项目提供的本地资产，避免生产交付继续暴露上游模板品牌。

## What Changes

- 将管理工作台的 Logo 资源替换为项目内本地 `linapro-mark.png`。
- 将管理工作台的 `favicon.ico` 替换为项目提供的 favicon 文件。
- 调整默认偏好配置，使工作台壳层与认证页统一使用新的本地图标 Logo，并恢复应用名文本展示。
- 调整宿主内置公共前端配置初始化数据，避免启动后将 Logo 覆盖回 Vben 默认资源。
- 新增用户默认头像系统参数，默认值为 `/avatar.webp`，并通过现有公开前端配置供工作台头像兜底使用。
- 不引入新的后端接口端点、数据表结构变更或兼容性迁移。

## Capabilities

### New Capabilities

- `workbench-brand-assets`: 约束默认管理工作台使用 LinaPro 本地品牌资产，包括工作台 Logo 与浏览器 favicon。

### Modified Capabilities

- 无。

## Impact

- 影响 `apps/lina-vben` 默认管理工作台的静态资源与默认偏好配置。
- 影响 `apps/lina-core` 公开前端配置的初始化种子数据、服务兜底值与接口示例。
- 影响登录页、工作台基础布局中通过 `preferences.logo.source` 渲染的图标 Logo 与应用名文本组合。
- 影响用户未设置头像时通过 `preferences.app.defaultAvatar` 使用的头像兜底资源。
- 影响浏览器标签页展示的 favicon。
- 不影响后端 API、数据库结构、权限模型或插件运行时。
