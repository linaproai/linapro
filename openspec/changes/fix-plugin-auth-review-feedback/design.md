# 设计说明

## 上下文

`feat/plugin-auth` 合并后的审查暴露了三类系统性问题：

- 匿名 `/auth/providers` 路径在公开 DTO、service 投影和登录入口聚合中泄露了后端 SSO 跳转配置，并随 provider 数量线性放大 `sys_config` 读取频次。
- 宿主 `pluginsettings.upsertValue` 既写时间字段又走 count-then-insert/update 分支，存在并发唯一键竞争和软删除残留后无法恢复的隐患。
- OIDC 插件层和宿主层各自维护了一个 `enabled` 概念，导致登录入口显示、OAuth 发起和 OAuth 回调实际上有两个权威来源。

本变更把上述问题作为一个反馈闭环修复，目标是同时收紧公开数据面、保证 KV 存储语义、并把 provider 启停统一到宿主单一权威来源；同时整理 OIDC 插件层错误码到 `bizerr` 体系、把 OAuth handoff 错误本地化收口到运行时 i18n。

## 决策点

### D1 — 公开 provider 投影必须是按钮元数据

`/auth/providers` 是匿名接口，承担登录入口渲染语义，使用群体远大于受控管理工作台。本变更将 `auth.ListProviders` 收敛为按钮元数据投影，去掉 redirect 默认值、redirect rules JSON 和后端 SSO 投递语义，确保未登录可见的字段不携带任何路由内部规则。SSO 投递规则仍保留在认证过的 `/plugin/<id>/settings` GET/PUT 和 provider 回调内部，不暴露给匿名调用方。

### D2 — 公开 provider 列表由宿主 `IsProviderEnabled` 单源决定

仅安装但未启用的 source 插件不得出现在登录入口。`auth.ListProviders` 的 `enabled` 判定改为调用 `pluginSvc.IsProviderEnabled`，与 OIDC 插件 OAuth 启动/回调内部的 `PluginState.IsProviderEnabled` 共享同一个 platform plugin enabled snapshot，确保「登录按钮是否出现」「登录链路是否可走」「回调是否处理」始终一致。

插件私有 `Enabled` 字段同时被取消：Google/Discord 的 settings DTO、settings service 和设置页都不再有 `enabled` 开关。OIDC 插件管理员仍可以在插件被宿主禁用时打开 `/plugin/<id>/settings` 预填 client id/secret/redirect rules，但 OAuth 调用路径在宿主禁用状态下会 fail-closed。

### D3 — 高频登录页路径必须无 N+1 数据库访问

源码插件 `LoginEntry(ctx)` 实现改为静态投影，构造完登录入口元数据时不再读 `sys_config`。这条决策牺牲了登录按钮即时反映管理员动态变更登录文案的灵活度，换取登录页路径 O(1) 数据库访问以及零 SSO 配置泄露。后续若需要支持动态登录文案，应通过宿主缓存层做 snapshot，而不是回到 per-request 插件设置读取。

### D4 — `pluginsettings.upsertValue` 必须是数据库原子 upsert

宿主 KV 存储改用 PostgreSQL 原生 `INSERT ... ON CONFLICT (tenant_id, key) DO UPDATE SET ...`，避免应用层 count-then-insert/update 在并发写入时的唯一键竞争。`OnDuplicate` 列表只更新 `name`、`value` 和 `deleted_at`，其中 `deleted_at` 重置回 NULL 是关键：

- PostgreSQL 唯一索引 `(tenant_id, key)` 不带 `WHERE deleted_at IS NULL`，因此任何 soft-deleted 残留都会让后续 upsert 进入 DO UPDATE 分支。
- 如果 `OnDuplicate` 不显式重写 `deleted_at`，则 DO UPDATE 会更新 `name`/`value` 但保留旧的 `deleted_at`，结果是 GoFrame DAO 自动过滤后这一行对 GetString/List 不可见。
- 显式把 `deleted_at` 放进 `OnDuplicate` 列表后，DO UPDATE 同时执行 `deleted_at = EXCLUDED.deleted_at = NULL`，把行恢复到可见状态。

`is_builtin`、`created_at`、`updated_at` 全部从 `OnDuplicate` 列表排除：`is_builtin` 是宿主治理位，不允许 KV 写入覆盖；`created_at`/`updated_at` 由 GoFrame 自动时间策略维护。

### D5 — 插件设置清空必须物理删除

`pluginsettings.SetString(ctx, pluginID, key, "")` 进入 `deleteByFullKey` 后改为 `.Unscoped().Delete()`，把行直接从 `sys_config` 物理移除。插件设置是不透明 KV 数据，没有审计或恢复用例。物理删除与 D4 配合形成防御深度：

- 新的清空操作不再产生 soft-deleted 行，源头堵住问题。
- 即便有遗留 soft-deleted 行（早期实现产生或并行写入残留），D4 的 upsert 恢复路径可以自愈，无需运维手动 SQL 清理。

### D6 — OIDC 插件本地 OAuth 错误收敛到 `bizerr.Code`

Google/Discord OAuth controller 自身产生的错误（provider 禁用、settings 不可用、authorize URL 构建失败、回调缺参数、state 校验失败、code exchange 失败、userinfo 失败、邮箱未验证、空登录结果）统一定义在各插件 `backend/internal/controller/oauth/oauth_code.go` 的 `bizerr.MustDefineWithKey` 项。

- `writeError` 4xx body 返回 `errorCode/messageKey/message/fallback/messageParams` 投影。
- `redirectWithCode` 使用 `bizerr.Code.RuntimeCode()` 生成 handoff query code，前端 `oauth-handoff.vue` 通过 raw code → lowerCamelCase 映射加载 `authentication.oauthHandoff.errors.*` 翻译。
- 外部 OAuth provider 回传的 `error` 参数和宿主 `LoginByExternal` 返回的 `AUTH_*` runtime code 是上游稳定契约，**不**在插件侧重写。

### D7 — host i18n 必须覆盖新 API 文档和新 bizerr 翻译键

新增的 `/auth/providers` 公开接口和 `error.auth.external.*` 翻译键必须在宿主 i18n 资源里有完整翻译条目：

- `manifest/i18n/zh-CN/apidoc/core-api-auth.json` 补 `ListProvidersReq`、`ListProvidersRes`、`ProviderEntity` 全字段。
- `manifest/i18n/en-US/apidoc/core-api-auth.json` 按宿主惯例保留空对象占位文件。
- `manifest/i18n/zh-CN/error.json` 与 `manifest/i18n/en-US/error.json` 补 `error.auth.external.identityInvalid` 和 `error.auth.external.userNotProvisioned` 子树。

## 风险与缓解

- **风险**：D4 把 `deleted_at` 写入 `OnDuplicate` 列表后，未来若有审计需求重新打开 sys_config 软删除，需要保留 plugin settings 的可恢复语义；当前实现是「物理 + 软删除自愈」双保险。
  **缓解**：用 D5 的物理删除作为正常路径；D4 的 `deleted_at` 重置只在异常残留下生效。后续若引入新的 KV 审计能力，应当在 audit 表里完成而不是依赖 sys_config 的 soft delete。
- **风险**：D2 把登录入口和 OAuth 回调全部绑到宿主 `IsProviderEnabled` snapshot，引入对宿主插件生命周期 freshness 的依赖。
  **缓解**：`IsProviderEnabled` 默认 fail-closed（nil PluginState、snapshot 缺失均返回 false），在 admin 重新启用后宿主 snapshot 自动刷新即可恢复，无需重启进程。
- **风险**：D3 让登录文案静态化，丧失了 admin 在线改文案的能力。
  **缓解**：当前需求里登录文案稳定不变；如未来需要动态化，应在宿主侧引入 snapshot 缓存层而不是回到 per-request 插件读。

## 验证

- 后端：宿主 + Google/Discord 插件单元测试。
- 集成：`pluginsettings_integration_test.go` 中三个 DB 回归用例覆盖「清空 → 重写 → 可见」「软删除残留 → 自愈」「清空 → 物理删除」。
- 治理：JSON 校验、静态检索、`linapro-local-build` 容器健康检查。
- OpenSpec：补 `specs/` 增量规范声明 capability 行为变化，若环境提供 `openspec` CLI 则运行 `openspec validate fix-plugin-auth-review-feedback --strict`。
