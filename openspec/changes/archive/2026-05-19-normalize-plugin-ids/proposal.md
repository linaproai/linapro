## Why

当前官方插件 ID 缺少统一前缀和能力分层，导致官方插件与未来第三方插件容易出现命名边界不清、生态分类困难的问题。随着源码插件、动态插件、插件治理、运行时 i18n、菜单权限、cron 和宿主能力 provider 逐步稳定，插件 ID 已经成为跨运行时边界的关键身份，需要在官方插件生态扩展前完成规范化。

## What Changes

- 插件 ID 推荐使用 `<author>-<domain>-<capability>`，其中 `capability` 可由一个或多个 kebab-case 单词组成；该规则作为命名建议和官方插件治理约定，不作为宿主运行时强制校验。
- 宿主运行时仅强制插件 ID 的基础安全边界：非空、64 字符长度上限和 lowercase kebab-case，避免插件 ID 破坏 URL、文件名、数据库键、菜单、权限、i18n 或 apidoc 命名空间。
- **BREAKING**：现有官方插件 ID 统一重命名，不保留旧 ID 兼容或别名：
  - `content-notice` -> `linapro-content-notice`
  - `monitor-loginlog` -> `linapro-monitor-loginlog`
  - `monitor-operlog` -> `linapro-monitor-operlog`
  - `monitor-online` -> `linapro-monitor-online`
  - `monitor-server` -> `linapro-monitor-server`
  - `multi-tenant` -> `linapro-tenant-core`
  - `org-center` -> `linapro-org-core`
  - `plugin-demo-dynamic` -> `linapro-demo-dynamic`
  - `plugin-demo-source` -> `linapro-demo-source`
  - `demo-control` -> `linapro-ops-demo-guard`
- 插件 manifest、源码插件注册、动态插件 artifact、依赖声明、菜单 key、i18n namespace、apidoc namespace、cron handlerRef、动态资源路径、扩展 API 路径、配置、测试和文档中的官方插件 ID 同步改名。
- 新增或扩展治理扫描/测试，确保官方插件目录名、manifest ID、源码注册 ID、菜单 key 与 i18n namespace 一致。

## Capabilities

### New Capabilities

- `plugin-id-governance`: 定义插件 ID 的基础安全边界、官方插件 ID 建议结构、官方插件 ID 映射、运行时身份一致性和治理验证要求。

### Modified Capabilities

- `plugin-manifest-lifecycle`: 将插件 manifest ID、依赖声明、源码注册 ID 与动态 artifact manifest 统一接入基础安全校验，并保留注册 ID 与 manifest ID 一致性校验。
- `demo-control-guard`: 将官方演示只读保护插件的规范标识从 `demo-control` 更新为 `linapro-ops-demo-guard`，保持启用态守卫语义不变。

## Impact

- 影响官方源码插件目录、`plugin.yaml`、Go module 名称、源码注册常量、`apps/lina-plugins/lina-plugins.go`、插件聚合 `go.mod`/`go.sum`、GoFrame 生成配置和相关 import 路径。
- 影响宿主插件 manifest 校验、插件依赖校验、官方菜单稳定挂载元数据、`orgcap.ProviderPluginID`、`tenantcap.ProviderPluginID`、启动一致性检查和自动启用配置。
- 影响运行时标识派生：`sys_plugin.plugin_id`、`sys_plugin_release.plugin_id`、`sys_plugin_migration.plugin_id`、`sys_plugin_resource_ref.plugin_id`、`sys_plugin_node_state.plugin_id`、`sys_plugin_state.plugin_id`、菜单 `plugin:<id>:...`、权限字符串、cron handlerRef、动态资源 URL `/plugin-assets/<id>/<version>/...`、扩展 API `/api/v1/extensions/<id>/...`。
- 影响 i18n 与 apidoc：运行时资源键 `plugin.<plugin-id>.*`、菜单 i18n key、job i18n key、apidoc `plugins.<plugin_id_snake_case>.*` 和 README 示例必须同步。
- 影响配置和测试：`plugin.autoEnable`、开发/镜像配置、Playwright 用例、插件页面对象、后端单元测试 fixture、动态 Wasm 构建测试和 OpenSpec 规范示例需要同步。
- 不考虑历史兼容性：不实现旧 ID alias，不迁移旧运行时数据；需要通过初始化/重建测试数据或重新同步插件治理数据获得新状态。
