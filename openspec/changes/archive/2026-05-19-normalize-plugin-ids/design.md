## Context

插件 ID 当前只通过 `catalog.ManifestIDPattern` 校验为 `kebab-case`，并在 manifest、源码插件注册、运行时 artifact、插件依赖、菜单 key、权限、cron handlerRef、动态资源 URL、扩展 API 路径、i18n namespace、apidoc namespace、配置和测试中被当作同一个稳定身份使用。官方插件还通过 `menu_metadata.go`、`orgcap.ProviderPluginID`、`tenantcap.ProviderPluginID` 等常量参与宿主菜单挂载和可选能力 provider 检测。

这次变更不做旧 ID 兼容。落地后，旧 ID 不再作为可安装、可启用、可依赖、可访问或可自动启用的有效插件身份；本地开发和测试环境需要重新初始化或重新同步插件治理数据。

## Goals / Non-Goals

**Goals:**

- 建立插件 ID 基础安全契约，并通过代码校验防止不安全 ID 进入 URL、文件、数据库键、菜单、权限、i18n 和 apidoc 命名空间。
- 将 `<author>-<domain>-<capability>` 保留为官方插件命名建议和仓库治理约定，避免把生态分类策略硬编码进宿主运行时。
- 将 10 个官方插件 ID 破坏式重命名为新的官方 ID。
- 同步所有由插件 ID 派生的运行时身份，包括菜单、权限、路由、资源、cron、i18n、apidoc、配置、文档和测试。
- 明确插件自有存储命名、DAO 生成和动态插件 artifact 文件名在新 ID 下的边界。
- 为后端 Go 变更、前端可观察变化、E2E、i18n、缓存一致性和数据权限影响留下可验证任务。

**Non-Goals:**

- 不提供旧插件 ID alias、重定向、自动迁移或兼容查询。
- 不保留旧 ID 的自动启用配置、旧动态 artifact、旧菜单 key、旧 i18n key 或旧 cron handlerRef。
- 不改变插件安装、启用、禁用、卸载、运行时升级和 host service 授权的业务语义。
- 不新增对外 REST API；现有插件管理和扩展 API 路径只因 `{pluginId}` 值变化而变化。
- 不实现第三方 domain 注册流程；第三方插件可自行选择符合基础安全边界的 ID。

## Decisions

### 决策 1：插件 ID 运行时只强制基础安全边界

插件 ID 继续保持整体 `kebab-case`，并限制在 64 字符内。这个约束来自实际运行时边界：插件 ID 会参与 URL path、动态资产路径、文件名、数据库键、菜单 key、权限字符串、i18n namespace 和 apidoc namespace。宿主运行时不再强制解析为三段结构，也不再根据某个 domain 或 capability 片段决定是否接受插件。

替代方案是严格要求所有插件 ID 解析为：

```text
author = 第 1 段
domain = 第 2 段
capability = 第 3 段及之后的全部 kebab-case 文本
```

该方案能强化分类，但会把命名建议变成运行时阻断，增加插件作者命名成本，并要求宿主长期维护生态 taxonomy，因此不采用。

### 决策 2：结构化命名是官方建议而不是宿主硬编码策略

官方插件推荐继续使用：

```text
<author>-<domain>-<capability>
```

其中 LinaPro 官方插件使用 `linapro` author 片段，官方基础能力可以使用 `core` capability。第三方插件可以参考该结构，但宿主不维护 domain 白名单、不保留 capability 拒绝表，也不把旧官方 ID 作为通用拒绝表。官方仓库内插件是否符合建议结构由文档、评审和治理扫描保证。

### 决策 3：`core` 是官方基础能力命名建议

`linapro-tenant-core` 和 `linapro-org-core` 表示 LinaPro 官方在 `tenant` 与 `org` 领域提供的基础能力实现。第三方插件不得使用 `*-*-core`，官方非基础能力插件也不得滥用 `core`。

这里的“不得”是官方生态和文档建议，不由宿主运行时硬编码拒绝。替代方案是使用 `linapro-org-structure`。该名称能表达组织结构，但弱化了该插件作为 `orgcap` provider 和组织基础能力实现的定位，因此不采用。

### 决策 4：官方插件采用破坏式 ID 映射

官方插件映射固定为：

```text
content-notice        -> linapro-content-notice
monitor-loginlog      -> linapro-monitor-loginlog
monitor-operlog       -> linapro-monitor-operlog
monitor-online        -> linapro-monitor-online
monitor-server        -> linapro-monitor-server
multi-tenant          -> linapro-tenant-core
org-center            -> linapro-org-core
plugin-demo-dynamic   -> linapro-demo-dynamic
plugin-demo-source    -> linapro-demo-source
demo-control          -> linapro-ops-demo-guard
```

`linapro-ops-demo-guard` 将原 `demo-control` 定位为演示环境的全局只读请求守卫，而不是普通 demo 示例能力。

### 决策 5：插件自有存储命名跟随新插件身份重新生成

由于本变更明确不考虑历史兼容性，官方插件自有业务表、索引、mock 数据、uninstall SQL、DAO/DO/Entity 生成配置和 Go 访问代码应按新插件 ID 的 snake_case 范围重新命名。例如：

```text
plugin_linapro_tenant_core_tenant      -> plugin_linapro_tenant_core_tenant
plugin_linapro_org_core_dept          -> plugin_linapro_org_core_dept
plugin_linapro_demo_source_record       -> plugin_linapro_demo_source_record
plugin_linapro_demo_dynamic_record      -> plugin_linapro_demo_dynamic_record
```

单表插件使用 `plugin_<plugin_id_snake_case>`，多表插件在此基础上追加业务后缀。这样与既有“插件自有业务表使用 plugin_id snake_case 范围命名”的规范保持一致。

替代方案是只改运行时插件 ID、保留旧表名。该方案改动更小，但会留下新插件 ID 与插件存储命名不一致的问题，不利于长期治理，因此不采用。

### 决策 6：运行时校验与官方仓库治理分层执行

运行时必须在 manifest 加载时校验 ID 基础安全边界，发现空 ID、超长 ID 或不安全字符时 fail-fast。治理扫描用于覆盖运行时校验难以覆盖的官方仓库一致性：

- 插件目录名等于 manifest ID。
- 源码插件 `pluginhost.NewSourcePlugin(id)` 等注册 ID 等于 manifest ID。
- 菜单 key 使用 `plugin:<plugin-id>:` 前缀。
- 运行时 i18n key 使用 `plugin.<plugin-id>.` 前缀。
- apidoc i18n key 使用 `plugins.<plugin_id_snake_case>.` 前缀。
- 动态 artifact 文件名、hosted asset 路径和扩展 API 测试使用新 ID。
- 官方插件旧 ID 不再出现在运行时代码、配置、测试和新规范正向路径中。

### 决策 7：缓存一致性沿用插件 ID 作用域失效

本变更不新增缓存类型，但会改变插件 ID 作用域。安装、启用、禁用、同步、运行时升级和 i18n 刷新仍必须使用显式插件 ID scope 失效；集群模式下继续依赖既有共享修订号、事件广播或分布式缓存策略。禁止为了改名在普通业务路径中清空所有插件、所有语言或所有扇区缓存。

## Risks / Trade-offs

- [风险] 旧 ID 残留在配置或测试中导致启动自动启用失败或插件不可见。缓解：增加旧 ID 残留扫描，覆盖配置、Go、TS、Vue、JSON、YAML、README 和 OpenSpec 活跃文档。
- [风险] Go module、import path、目录名和 generated DAO 不一致导致编译失败。缓解：按插件逐个重命名 module/import/replace，重新运行对应 `gf gen dao` 或等价生成流程，并执行插件聚合编译测试。
- [风险] 业务表改名遗漏 mock/uninstall/DAO 访问点。缓解：以 SQL 表名清单驱动更新，补充插件包 Go 测试和 SQL/DAO 静态扫描。
- [风险] 动态插件 artifact、资源 URL 和 E2E 测试使用旧 ID。缓解：重新构建动态插件 artifact，更新 `/plugin-assets/<id>/...` 与 `/api/v1/extensions/<id>/...` 断言。
- [风险] i18n key 改名遗漏导致页面显示原始 key。缓解：运行 runtime i18n 扫描、JSON 校验和现有 i18n E2E。
- [风险] 旧运行时数据库中仍有旧插件治理行。缓解：本变更不做兼容迁移，执行前要求重建或清理本地数据库；依赖反向检查忽略未发现且无可信 release snapshot 的无关陈旧行，避免它们阻断新插件生命周期。

## Migration Plan

1. 新增插件 ID 基础安全校验组件，接入 manifest、依赖和动态 artifact 校验。
2. 重命名官方插件目录、manifest ID、Go module/import、源码注册常量、GoFrame 生成配置和聚合 workspace。
3. 更新官方插件自有 SQL 表、索引、mock 数据、uninstall SQL、DAO/DO/Entity 和服务访问代码。
4. 更新宿主官方插件常量、provider 插件 ID、稳定菜单父级映射、启动一致性、autoEnable 配置和示例配置。
5. 更新菜单 key、权限、cron handlerRef、动态资源路径、扩展 API 路径、i18n/apidoc key、文档和 E2E。
6. 增加治理扫描和单元测试，确认官方旧 ID 在运行时文件正向路径中无残留，新 ID 全部满足基础安全边界。
7. 执行后端 Go 编译门禁、插件聚合测试、前端 typecheck、E2E 校验、相关 Playwright 用例和 OpenSpec 严格校验。

Rollback 策略：本变更为破坏式改名，不提供运行时回滚脚本。若实现中发现风险过大，应在合并前整体回退本变更分支；已初始化到新 ID 的本地环境需要重新初始化才能回到旧状态。

## Open Questions

- `linapro-monitor-loginlog` 与 `linapro-monitor-operlog` 本轮按用户确认保留压缩 capability：`loginlog`、`operlog`。未来是否再改为 `login-log`、`operation-log` 应另行提出破坏式变更。
- 是否需要为第三方插件作者提供单独的 `linactl plugin.id.check` 命令。本轮先通过 Go 测试/治理扫描覆盖仓库内插件。
