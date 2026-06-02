## ADDED Requirements

### Requirement: OIDC 插件 provider 启停必须由宿主单一权威来源决定

系统 SHALL 让 OIDC 源码插件登录按钮的可见性、OAuth 启动接口的可调用性、OAuth 回调接口的可达性这三处行为统一由宿主 `PluginState.IsProviderEnabled` 决定。OIDC 插件 MUST NOT 在自身 `Settings` 模型、settings DTO、settings 持久化、设置页表单或任何插件私有 KV 中维护一个独立的 `Enabled` 字段，因为这会与宿主 platform plugin enabled snapshot 形成两个权威来源，使匿名 `/auth/providers` 响应与 OAuth 实际可达性出现不一致。

#### Scenario: 插件被宿主禁用时登录入口与 OAuth 同时不可用

- **WHEN** `linapro-oidc-google` 插件在 platform plugin enabled snapshot 中为禁用
- **THEN** `GET /auth/providers` 响应里不出现 Google provider
- **AND** `GET /api/v1/auth/google` 返回 fail-closed 错误，不发起 Google 授权重定向
- **AND** `GET /api/v1/auth/google/callback` 拒绝处理外部 provider 回调

#### Scenario: 受保护管理接口在禁用状态下仍可预配置

- **WHEN** OIDC 插件被宿主禁用
- **AND** 已认证管理员请求 `GET /plugin/linapro-oidc-google/settings`
- **THEN** 该接口在权限校验通过后正常返回当前持久化的 client id、redirect URI 与 SSO 投递规则
- **AND** 该接口不会触发 OAuth 流程或对匿名用户暴露登录入口

### Requirement: OIDC 插件 OAuth 启动与回调必须先校验宿主启用状态

系统 SHALL 在 OIDC 插件 OAuth 控制器 `StartLogin` 与 `HandleCallback` 路径上首先调用 `isProviderEnabled(ctx)` 进行 fail-closed 判断，并仅在判断通过后再读取插件 settings。系统 MUST NOT 在 fail-closed 判断之前发起 `sys_config` 读取、外部 token 交换或 userinfo 拉取，避免在 provider 被禁用时仍触发数据库或外部 HTTP 访问。

#### Scenario: 禁用状态下 StartLogin 短路失败

- **WHEN** `linapro-oidc-google` 在 platform plugin enabled snapshot 中为禁用
- **AND** 客户端调用 `GET /api/v1/auth/google`
- **THEN** 控制器在 `isProviderEnabled(ctx)` 处直接进入失败分支
- **AND** 控制器不读 `sys_config`、不构建 Google 授权 URL、不发起 302 重定向

#### Scenario: 禁用状态下 HandleCallback 短路失败

- **WHEN** `linapro-oidc-google` 在 platform plugin enabled snapshot 中为禁用
- **AND** 客户端访问 `GET /api/v1/auth/google/callback?code=...&state=...`
- **THEN** 控制器在 `isProviderEnabled(ctx)` 处直接进入失败分支
- **AND** 控制器不读 `sys_config`、不调用 Google token 端点、不调用 Google userinfo 端点
- **AND** 控制器走 redirectWithCode 路径返回 `provider_disabled`

### Requirement: OIDC 插件自身错误必须收敛到本地 `bizerr.Code`

系统 SHALL 把 OIDC 源码插件 OAuth 控制器自身产生的错误统一定义在各插件 `backend/internal/controller/oauth/oauth_code.go` 中的 `bizerr.Code` 项，涵盖 provider 禁用、settings 不可用、authorize URL 构建失败、回调缺参数、state 校验失败、code exchange 失败、userinfo 失败、邮箱未验证、空登录结果。系统 MUST NOT 在控制器内继续以裸字符串、`gerror.New` 或机器错误码直接构造 4xx 响应或 handoff query code，必须经由本地 `bizerr.Code` 投影。

#### Scenario: 4xx 响应使用本地 bizerr 投影

- **WHEN** OIDC 插件 OAuth 控制器需要返回 4xx 响应
- **THEN** 响应 body 包含 `errorCode`、`messageKey`、`message`、`fallback`、`messageParams` 与 `providerId`
- **AND** `errorCode` 等于对应 `bizerr.Code.RuntimeCode()`
- **AND** `messageKey` 等于对应 `bizerr.Code.MessageKey()`

#### Scenario: 前端 handoff 通过 runtime code 渲染本地化文案

- **WHEN** OIDC 插件 OAuth 控制器调用 `redirectWithCode` 进入前端 OAuth handoff
- **THEN** handoff query string 中的 `error` 等于 `bizerr.Code.RuntimeCode()`
- **AND** 前端 `oauth-handoff.vue` 通过 `oauthErrorMessageKeyByCode` 把该 runtime code 映射到 `authentication.oauthHandoff.errors.*` 本地化文案
- **AND** 控制器不在 query string 中继续使用裸字符串错误码

### Requirement: 外部上游错误码不得在 OIDC 插件层重写

系统 SHALL 在 OIDC 插件 OAuth 控制器中保留两类上游稳定错误契约：外部 OAuth provider 通过回调 query 回传的 `error` 参数；宿主 `LoginByExternal` 返回的 `AUTH_*` runtime code。这两类错误 MUST NOT 在 OIDC 插件层被重写或重新映射成本地 `bizerr.Code`，应当原样经由 `redirectWithError` 透传到前端 OAuth handoff，由前端使用统一映射表完成本地化。

#### Scenario: 外部 provider 回传 error 透传

- **WHEN** 外部 provider 在回调 URL 中携带 `error=access_denied`
- **THEN** OIDC 插件控制器把 `access_denied` 原样作为 handoff query `error` 透传
- **AND** 控制器不把 `access_denied` 重新映射到任何本地 `bizerr.Code`

#### Scenario: 宿主 LoginByExternal 错误透传

- **WHEN** 宿主 `LoginByExternal` 返回 runtime code `AUTH_EXTERNAL_USER_NOT_PROVISIONED`
- **THEN** OIDC 插件控制器把 `AUTH_EXTERNAL_USER_NOT_PROVISIONED` 原样作为 handoff query `error` 透传
- **AND** 控制器不把 `AUTH_EXTERNAL_USER_NOT_PROVISIONED` 重新映射到任何本地 `bizerr.Code`

### Requirement: OAuth handoff 错误本地化键必须使用 lowerCamelCase

系统 SHALL 在前端 `authentication.oauthHandoff.errors.*` 命名空间下使用 lowerCamelCase 作为运行时 i18n 键，并 SHALL 在 `oauth-handoff.vue` 内维护一份 raw code 到 lowerCamelCase 的显式映射表。系统 MUST NOT 让前端直接把 raw 后端 / 外部 provider 错误码字符串作为 i18n 键拼接，避免不同命名风格（lower_snake_case、SCREAMING_SNAKE_CASE）污染前端 locale 资源结构。

#### Scenario: handoff 渲染 OIDC 本地错误

- **WHEN** 前端收到 handoff query `error=invalid_state`
- **THEN** `oauth-handoff.vue` 通过映射表得到 `invalidState`
- **AND** 渲染键 `authentication.oauthHandoff.errors.invalidState`

#### Scenario: handoff 渲染宿主 AUTH_* 错误

- **WHEN** 前端收到 handoff query `error=AUTH_EXTERNAL_USER_NOT_PROVISIONED`
- **THEN** `oauth-handoff.vue` 通过映射表得到 `externalUserNotProvisioned`
- **AND** 渲染键 `authentication.oauthHandoff.errors.externalUserNotProvisioned`
