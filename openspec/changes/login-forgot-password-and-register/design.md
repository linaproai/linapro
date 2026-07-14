## Context

历史变更 `user-management` / `login-page-presentation` 将忘记密码、注册、手机登录、扫码登录等未完成入口从登录页隐藏，并将对应子路由重定向到 `/auth/login`，避免用户误判产品能力。

当前仓库已保留 Vben 同源资产：

- 登录组件 `AuthenticationLogin` 默认支持 `showForgetPassword` / `showRegister`
- 视图 `forget-password.vue`、`register.vue` 与公共 UI 组件已存在，但路由仍 redirect，登录页开关关闭
- E2E `TC006` 明确断言这些入口隐藏且子路由回退

用户反馈要求参考 Vben 登录页样式，补齐「忘记密码」与「创建账号」能力。宿主目前仅有管理员侧用户创建与重置密码 API，没有公开自助注册或邮件重置契约。

## Goals / Non-Goals

**Goals:**

- 登录页按 Vben 样式展示忘记密码、创建账号入口（保留记住账号与主登录按钮布局）。
- `/auth/forget-password`、`/auth/register`、`/auth/reset-password` 渲染独立认证页。
- 公开自助注册写入平台账号并分配内置 `user` 角色。
- 忘记密码通过邮件发送一次性重置链接；确认后更新密码并清理会话。
- 系统参数开关默认开启，服务端强制校验。

**Non-Goals:**

- 不启用手机号登录、扫码登录。
- 不实现短信验证码与图形验证码（本期用限流）。
- 不修改管理员侧用户管理重置密码契约。

## Decisions

### D1: 宿主公开 Auth API 完整交付注册与邮件重置

- **选择**：`POST /auth/register`、`POST /auth/forget-password`、`POST /auth/reset-password`。
- **原因**：用户要求完整功能而非页面占位；账号与 JWT 归属宿主域。
- **邮件**：通过 `notifycap.EmailDeliveryOrNil()` 发送；未配置邮件通道时找回密码返回“暂时不可用”。

### D2: 复用 Vben 认证 UI，不自研登录表单布局

- **选择**：继续使用 `@vben/common-ui` 的 `AuthenticationLogin` / `AuthenticationForgetPassword` / `AuthenticationRegister`。
- **原因**：用户明确要求参考 Vben 样式；现有组件已提供链接位置、标题 emoji、返回登录与表单布局。
- **备选**：自建 Ant Design 表单 — 拒绝，样式一致性成本更高。

### D3: 仅放开 forget-password / register 路由

- **选择**：`code-login`、`qrcode-login` 继续 redirect 到登录页；登录页 `showCodeLogin` / `showQrcodeLogin` 保持 `false`。
- **原因**：用户仅要求忘记密码与创建账号；避免重新暴露未交付的手机号/扫码能力。

### D4: 提交反馈使用前端 i18n + notification

- **选择**：在应用视图层提交 handler 中调用 `notification`/`message`，文案走 `authentication.*` 语言包。
- **原因**：无后端契约变更时仍需可验证的用户反馈；避免 `console.log` 伪实现。

### D5: 创建账号入口移出 AuthenticationLogin 默认插槽

- **选择**：登录页将 `showRegister=false`，在外部登录与社交登录区域之后自绘「还没有账号? 创建账号」。
- **原因**：Vben 截图中该块位于「其他登录方式」下方；组件内默认顺序在登录按钮后、第三方登录后但登录页插件插槽在组件外，必须在 `login.vue` 统一编排顺序。

### D6: 系统参数开关默认开启并进入公开前端白名单

- **选择**：新增 `sys.auth.forgetPasswordEnabled` / `sys.auth.registerEnabled`（`true`/`false`，默认 `true`），纳入 `publicFrontendSettingSpecs` 与 `/config/public/frontend` 的 `auth` 字段。
- **原因**：登录页未登录即可读取；与现有 `sys.auth.*` 登录展示参数一致，可在参数设置中管理。
## Risks / Trade-offs

- **[Risk] 用户期望提交后真正创建账号或收到邮件** → 页面副标题与成功提示明确“请联系系统管理员”；后续可另开变更补自助 API。
- **[Risk] 与历史“隐藏未实现入口”规范冲突** → 本变更同步修改 `login-page-presentation`，并更新 E2E。
- **[Trade-off] 表单字段收集后不落库** → 优先样式与导航完整，避免未治理的公开写接口。

## Migration Plan

1. 打开登录页开关并切换路由组件。
2. 完善两页提交反馈与 i18n。
3. 更新 TC006 与必要时新增子页用例。
4. 回滚：重新关闭开关并将路由改回 redirect。

## Open Questions

无。公开自助注册与邮件重置留作后续独立变更。
