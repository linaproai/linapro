## ADDED Requirements

### Requirement: 业务模块必须拥有自身本地化投影规则
宿主系统 SHALL 让菜单、字典、配置、定时任务、角色、插件运行时等业务模块在各自模块边界内维护本地化投影规则。`internal/service/i18n` SHALL 只提供语言解析、翻译查找、资源加载、缓存、缺失检查等基础能力,不得引用业务实体、业务保护规则或业务翻译键推导逻辑。业务模块可以通过窄接口依赖 `ResolveLocale`、`Translate`、`TranslateSourceText` 等底层能力,但不得要求 i18n 基础服务反向认识业务模块。

#### Scenario: 业务模块在自身边界内完成投影
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` / `role` / `plugin runtime` 等模块需要为查询结果做本地化展示
- **THEN** 模块在自己的 `*_i18n.go` 或等价模块内推导翻译键、判断是否跳过默认语言、判断是否属于受保护内置记录
- **AND** `internal/service/i18n` 不 import 这些业务模块或业务实体

#### Scenario: i18n 基础服务只提供底层能力
- **WHEN** 业务模块需要翻译一个展示字段
- **THEN** 业务模块调用 `Translate` / `TranslateSourceText` 等底层方法并传入自身拥有的 key 与 fallback
- **AND** i18n 服务不暴露 `ProjectMenu`、`ProjectDictType`、`ProjectBuiltinJob` 等按业务实体命名的方法

### Requirement: 受保护内置记录的判定规则必须留在所属业务模块
业务模块 SHALL 在自身服务中维护受保护内置记录的判定规则与翻译键约定。例如:内置 `admin` 角色的展示投影由 `role` 模块维护,默认任务分组和内置任务的展示投影由 `jobmgmt` 模块维护,插件元数据展示投影由 plugin runtime 模块维护。通用 i18n 基础服务不得包含这些业务判定常量。

#### Scenario: 内置 admin 角色在英文环境只读列表中显示本地化名称
- **WHEN** 管理员在 `en-US` 环境下查询角色列表
- **THEN** `role` 模块仅对内置 `admin` 角色按 `role.builtin.admin.name` 翻译键提供本地化展示
- **AND** 其他可编辑角色的 `name` 字段保持数据库原值

#### Scenario: 用户创建的自定义任务保持原值
- **WHEN** 管理员在任意非默认语言下查询调度任务列表
- **THEN** `jobmgmt` 模块不会对 `is_builtin = 0` 的用户自建任务做翻译投影
- **AND** 该任务的 `name` / `description` 显示数据库原始值
