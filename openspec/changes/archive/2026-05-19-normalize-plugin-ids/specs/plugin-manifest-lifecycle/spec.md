## ADDED Requirements

### Requirement: 插件清单必须校验插件 ID 安全边界

插件清单生命周期 SHALL 在加载源码插件 manifest、动态插件 artifact manifest 和插件依赖声明时统一校验插件 ID 的基础安全边界。校验规则 MUST 复用插件 ID 治理能力定义的非空、kebab-case 和 64 字符长度限制。运行时校验 SHALL NOT 强制执行 `<author>-<domain>-<capability>` 建议结构、domain 白名单、官方 capability 保留或旧官方 ID 拒绝表。任一插件 ID 基础安全校验失败时，系统 MUST 拒绝该 manifest 或 artifact，并返回可诊断错误。

#### Scenario: 源码插件 manifest ID 校验失败
- **WHEN** 源码插件 `plugin.yaml` 声明 `id: Monitor_Server`
- **THEN** 宿主拒绝加载该源码插件 manifest
- **AND** 错误说明插件 ID 必须使用 kebab-case lowercase letters and digits

#### Scenario: 源码插件注册 ID 与 manifest ID 不一致
- **WHEN** 源码插件注册 ID 为 `linapro-monitor-server`
- **AND** `plugin.yaml` 声明 `id: linapro-monitor-loginlog`
- **THEN** 宿主拒绝加载该源码插件
- **AND** 错误说明源码插件注册 ID 与 manifest ID 不一致

#### Scenario: 动态插件 artifact manifest ID 校验失败
- **WHEN** 管理员上传动态插件 artifact
- **AND** artifact 内嵌 manifest 声明 `id: plugin_demo_dynamic`
- **THEN** 宿主拒绝该 artifact
- **AND** 错误说明动态插件 ID 不符合插件 ID 基础安全规则

#### Scenario: 插件依赖声明使用非三段 ID
- **WHEN** 插件 manifest 的 `dependencies.plugins[].id` 声明 `plugin-demo-source`
- **THEN** 宿主接受该依赖声明
- **AND** 宿主不得因其不满足建议结构而拒绝该 manifest

### Requirement: 插件清单资源命名空间必须使用当前插件 ID

插件清单生命周期 SHALL 要求 manifest 声明的菜单、权限、运行时 i18n、apidoc i18n、cron 和动态前端资产入口使用当前插件 ID 派生命名空间。宿主不得从官方旧 ID 自动推导或补齐这些资源。

#### Scenario: 菜单 key 使用旧插件 ID
- **WHEN** 插件 `linapro-content-notice` 的 manifest 声明菜单 key `plugin:content-notice:notice`
- **THEN** 宿主拒绝该 manifest
- **AND** 错误说明菜单 key 必须使用 `plugin:linapro-content-notice:` 前缀

#### Scenario: 动态插件菜单路径使用旧资产路径
- **WHEN** 动态插件 `linapro-demo-dynamic` 的 manifest 声明菜单 path `/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js`
- **THEN** 宿主拒绝或治理验证阻断该资源
- **AND** 错误说明动态资产路径必须使用 `/plugin-assets/linapro-demo-dynamic/`
