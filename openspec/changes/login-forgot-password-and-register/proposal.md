## Why

登录页当前按历史规范刻意隐藏「忘记密码」与「创建账号」入口，并把对应认证子路由回退到 `/auth/login`。用户需要参考 Vben 登录页样式，补齐这两项公开入口与页面能力，使未登录用户可以按标准中后台交互进入找回密码与注册流程。

## What Changes

- 在标准登录页按 Vben `AuthenticationLogin` 样式展示「忘记密码?」与「还没有账号? 创建账号」入口。
- 创建账号入口固定放在「其他登录方式」区域下方，对齐 Vben 登录页截图结构。
- 将 `/auth/forget-password`、`/auth/register` 从回退登录改为渲染独立认证子页（继续沿用 Vben 认证布局与表单组件）。
- 通过系统参数 `sys.auth.forgetPasswordEnabled`、`sys.auth.registerEnabled` 控制入口与子路由可达性，默认 `true`；关闭后隐藏入口并回退登录页。
- 公开自助注册：`POST /auth/register` 创建平台账号并分配内置 `user` 角色。
- 邮件密码重置：`POST /auth/forget-password` 发送一次性链接，`POST /auth/reset-password` 确认新密码。
- 手机号登录、扫码登录入口保持隐藏；对应子路由继续回退登录页。
- 更新 `login-page-presentation` 规范与登录相关 E2E，使验收与实现一致。
## Capabilities

### New Capabilities

- `auth-account-recovery-and-register`: 登录页忘记密码/创建账号入口、独立认证子页展示与表单交互能力。

### Modified Capabilities

- `login-page-presentation`: 调整“仅暴露用户名密码登录、隐藏未实现入口”的要求，允许展示并进入忘记密码与创建账号子流程；手机号/扫码登录仍不作为公开能力。

## Impact

- 前端：`apps/lina-vben/apps/web-antd` 登录页、认证路由、忘记密码/注册视图。
- 测试：`hack/tests/e2e/auth/TC006-login-page-presentation.ts`、`hack/tests/pages/LoginPage.ts`，以及必要时新增认证子页 E2E。
- 规范：`openspec/specs/login-page-presentation` 增量修改；新增 `auth-account-recovery-and-register` 能力规范。
- 后端 API：本变更不新增公开注册/重置密码 HTTP 契约；继续由管理员在用户管理中创建账号与重置密码。
