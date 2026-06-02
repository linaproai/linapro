# 修复插件认证 PR 反馈

## Why

`feat/plugin-auth` 在合并审查中暴露了几类问题：匿名 `/auth/providers` 返回了后端 SSO redirect 配置，登录页 provider 聚合路径会随 provider 数量读取插件设置，`pluginsettings` 写入存在 count-then-insert/update 并发竞争和手写时间字段，OAuth handoff 与 OIDC 插件页面存在硬编码文案，且已安装但未启用的认证插件仍可能显示在匿名登录入口。

这些问题会影响公开接口最小披露、登录页高频路径性能、插件设置写入可靠性、运行时多语言一致性和插件启停语义，需要作为当前插件认证反馈闭环修复。

## What Changes

- 收窄匿名 `/auth/providers` 公开 DTO、service projection 和前端类型，只返回登录按钮所需静态元数据。
- 将 Google/Discord provider 的 public `LoginEntry` 改为静态投影，避免匿名登录页读取插件设置并消除 N+1 `sys_config` 查询路径。
- 将 `pluginsettings` 写入改为数据库原子 upsert，并让 GoFrame 继续维护 `created_at`、`updated_at`。
- 使用认证 provider 专用的 `IsProviderEnabled` 过滤匿名登录入口，避免仅安装未启用的插件显示按钮。
- 将 OAuth handoff、登录 provider 按钮和 OIDC 插件页面可见文案接入运行时 i18n，并统一 OAuth error locale key 命名。
- 取消 OIDC 插件设置页里的私有“启用登录”开关，登录入口是否出现和 OAuth 回调是否可用统一由宿主插件 provider enablement 状态控制。
- 补充后端单元测试和治理验证，覆盖公开 DTO 字段收窄、provider 过滤、upsert 冲突更新列和 panic allowlist。

## Impact

- `i18n`：有影响。宿主认证语言包新增 OAuth handoff 文案，OIDC 插件页面复用宿主工作台 `plugins.oidc.*` 文案，Google 插件页面标题改为 i18n key；Google/Discord 插件自身已启用 `i18n.enabled: true`，插件菜单/插件元数据继续由各自 `manifest/i18n` 维护。
- 缓存一致性：有影响。`/auth/providers` 和 OIDC OAuth 回调均改用 `PluginState.IsProviderEnabled` / `pluginSvc.IsProviderEnabled`，该路径读取插件 runtime enabled snapshot，并通过现有 runtime cache freshness 机制刷新；本变更不新增缓存或失效机制。
- 数据权限：匿名 `/auth/providers` 是公开资源例外，只暴露 button metadata，不返回 redirect rules、receiver URL、secret 或租户/用户数据；不接入角色数据权限。
- 接口性能：有影响并已优化。匿名 provider 聚合不再读取每个插件 settings，数据装配成本随 provider 数量仅进行 registry 遍历和 enabled snapshot 判断。
- 数据库：不新增 SQL/DAO。`pluginsettings` 写入使用现有 `sys_config` 唯一键 `(tenant_id, key)` 的原子 upsert，不手写时间字段。
- 测试策略：后端行为使用单元测试覆盖；前端文案/表单类治理使用 JSON 校验、静态检索和构建健康检查验证。
- 开发工具跨平台：无影响。本变更不修改脚本、Makefile、CI 或跨平台工具入口。
- 设置接口边界：`/plugin/<id>/settings` GET/PUT 在 OIDC 插件被宿主禁用后仍可由具备 `linapro-oidc-google:settings:*` 或 `linapro-oidc-discord:settings:*` 权限的管理员访问。该接口受 `Auth+Tenancy+Permission` 中间件保护，未挂载任何匿名登录入口，仅用于让管理员在启用插件之前预先配置 client id、redirect URI 和 SSO 投递规则。OAuth 发起 `/api/v1/auth/<provider>` 与回调 `/api/v1/auth/<provider>/callback` 仍统一由 `PluginState.IsProviderEnabled` 控制，禁用状态下设置接口可用不会触发 OAuth 流程或对外暴露登录入口。
- 默认跳转地址：OIDC 插件 SPA landing fallback 与配置默认值统一为 `/dashboard/analytics`，避免登录后跳到工作台根路径再二次跳转的体验，同时保持 SSO 投递规则 (`backendRedirects`) 行为不变。
- 错误契约：OIDC 插件自身产生的 OAuth 错误收敛到各插件 `backend/internal/controller/oauth/oauth_code.go` 的 `bizerr.Code` 定义；handoff 仍通过稳定 runtime code 查询参数传给前端，外部 provider 回传错误和宿主认证 `AUTH_*` 错误保持上游契约。
