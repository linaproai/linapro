## ADDED Requirements

### Requirement: 插件 ID 必须满足运行时安全边界

系统 SHALL 只在运行时强制校验插件 ID 的基础安全边界：插件 ID 不能为空、总长度 MUST 不超过运行时 `plugin_id` 字段允许的 64 字符，并且 MUST 使用小写字母、数字和 hyphen 组成的 kebab-case 文本，以保证其可安全用于 URL path、动态资产路径、文件名、数据库键、菜单 key、权限字符串、i18n namespace 和 apidoc namespace。宿主运行时 SHALL NOT 强制校验 `<author>-<domain>-<capability>` 结构、domain 白名单、官方 capability 保留或旧官方 ID 拒绝表。

#### Scenario: 接受官方建议的结构化插件 ID
- **WHEN** 插件 manifest 声明 `id: linapro-content-notice`
- **THEN** 系统接受该插件 ID
- **AND** 该 ID 可继续作为菜单、权限、资源、i18n 和 apidoc 的派生命名空间

#### Scenario: 接受非三段结构的扩展插件 ID
- **WHEN** 插件 manifest 声明 `id: demo-control`
- **THEN** 系统接受该插件 ID
- **AND** 系统不得因其不满足 `<author>-<domain>-<capability>` 建议结构而拒绝 manifest

#### Scenario: 接受自定义 domain 片段
- **WHEN** 插件 manifest 声明 `id: acme-random-report`
- **THEN** 系统接受该插件 ID
- **AND** 系统不得因 `random` 不在宿主内置列表中而拒绝 manifest

#### Scenario: 拒绝不安全字符
- **WHEN** 插件 manifest 声明 `id: Acme_Report`
- **THEN** 系统拒绝该 manifest
- **AND** 错误说明插件 ID 必须使用 kebab-case lowercase letters and digits

#### Scenario: 拒绝超长插件 ID
- **WHEN** 插件 manifest 声明超过 64 字符的 ID
- **THEN** 系统拒绝该 manifest
- **AND** 错误说明插件 ID length must not exceed 64 characters

### Requirement: 官方插件 ID 必须使用规范化映射

系统 SHALL 将 LinaPro 官方随仓库发布的插件 ID 规范化为以下映射，并不得继续在官方插件的运行时配置、manifest、源码注册、菜单、权限、cron、i18n、apidoc、测试或文档正向路径中使用旧官方 ID。该映射只约束 LinaPro 官方插件资产，不作为宿主对第三方插件 ID 的运行时拒绝表。

| 旧 ID | 新 ID |
| --- | --- |
| `content-notice` | `linapro-content-notice` |
| `monitor-loginlog` | `linapro-monitor-loginlog` |
| `monitor-operlog` | `linapro-monitor-operlog` |
| `monitor-online` | `linapro-monitor-online` |
| `monitor-server` | `linapro-monitor-server` |
| `multi-tenant` | `linapro-tenant-core` |
| `org-center` | `linapro-org-core` |
| `plugin-demo-dynamic` | `linapro-demo-dynamic` |
| `plugin-demo-source` | `linapro-demo-source` |
| `demo-control` | `linapro-ops-demo-guard` |

#### Scenario: 官方插件清单使用新 ID
- **WHEN** 宿主扫描 `apps/lina-plugins/linapro-org-core/plugin.yaml`
- **THEN** manifest ID 为 `linapro-org-core`
- **AND** 宿主不得在官方插件清单中发现 `org-center`

#### Scenario: 官方自动启用配置使用新 ID
- **WHEN** 宿主读取仓库默认 `plugin.autoEnable`
- **THEN** 官方插件项使用规范化新 ID
- **AND** 默认配置中不得继续使用 `multi-tenant`、`org-center` 或其他旧官方 ID

### Requirement: 插件运行时身份必须使用当前插件 ID

系统 SHALL 在运行时身份边界使用当前插件 ID，不得为官方旧 ID 提供 alias、重定向或兼容查询。该边界包括插件管理 API、扩展 API、动态前端资产 URL、菜单 key、权限字符串、cron handlerRef、插件状态表、发布表、迁移表、资源引用表、节点状态表、插件 KV 状态表和 host service 授权记录。

#### Scenario: 新扩展 API 路径使用当前 ID
- **WHEN** 动态插件 `linapro-demo-dynamic` 暴露扩展 API
- **THEN** 宿主公开路径使用 `/api/v1/extensions/linapro-demo-dynamic/...`
- **AND** 宿主不得通过 `/api/v1/extensions/plugin-demo-dynamic/...` 暴露同一官方插件

#### Scenario: 新动态资产路径使用当前 ID
- **WHEN** 动态插件 `linapro-demo-dynamic` 提供前端资产
- **THEN** 宿主资产路径使用 `/plugin-assets/linapro-demo-dynamic/<version>/...`
- **AND** 宿主不得通过 `/plugin-assets/plugin-demo-dynamic/<version>/...` 暴露同一官方插件资产

#### Scenario: 新 cron handlerRef 使用当前 ID
- **WHEN** 插件 `linapro-monitor-server` 注册内置定时任务
- **THEN** handlerRef 使用 `plugin:linapro-monitor-server/cron:<name>`
- **AND** 系统不得继续生成 `plugin:monitor-server/cron:<name>`

### Requirement: 仓库治理扫描必须验证官方插件 ID 一致性

系统 SHALL 提供自动化验证，确保官方插件目录名、manifest ID、源码插件注册 ID、动态 artifact manifest、依赖声明、菜单 key、运行时 i18n key、apidoc i18n key、配置和测试 fixture 使用同一个当前插件 ID。验证失败时变更不得通过。

#### Scenario: 目录名与 manifest ID 不一致
- **WHEN** 插件目录为 `apps/lina-plugins/linapro-content-notice`
- **AND** 该目录的 `plugin.yaml` 声明 `id: content-notice`
- **THEN** 治理验证失败
- **AND** 错误指出目录名与 manifest ID 不一致

#### Scenario: i18n namespace 使用旧 ID
- **WHEN** 插件 `linapro-content-notice` 的运行时语言包包含 `plugin.content-notice.name`
- **THEN** 治理验证失败
- **AND** 错误指出运行时 i18n key 必须使用 `plugin.linapro-content-notice.` 前缀

#### Scenario: apidoc namespace 使用当前 ID
- **WHEN** 插件 `linapro-demo-dynamic` 的 apidoc 语言包包含 `plugins.linapro_demo_dynamic`
- **THEN** 治理验证通过
