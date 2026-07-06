## Why

启用 `linapro-tenant-core` 后，租户管理员在租户上下文访问用户管理、角色管理等宿主页面时，前端仍可能触发平台租户列表或登录租户候选接口。后端返回 403 符合平台控制面权限边界，但前端不应把这些受限接口作为普通页面加载路径的一部分。

## What Changes

- 为前端租户选项加载增加权限门禁：缺少 `system:tenant:list` 时不请求平台租户列表，缺少 `system:tenant:auth:login-tenants` 时不请求登录租户候选。
- 租户上下文优先使用当前租户，不因租户候选未加载而回退平台租户查询。
- 头部租户切换器复用相同门禁，避免布局启动时全局误触发受限接口。
- 补充租户管理员访问用户管理和角色管理页面的 E2E 回归，覆盖 GitHub issue 76。

## Impact

- 影响宿主前端 `apps/lina-vben/apps/web-antd` 的租户 store、用户管理租户选项加载和多租户插件头部插槽。
- 影响 `apps/lina-plugins/linapro-tenant-core` 插件前端插槽与插件自有 E2E 测试。
- 不修改后端权限模型、HTTP API 契约、数据库结构或插件清单。
