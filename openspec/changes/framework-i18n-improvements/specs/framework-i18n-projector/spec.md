## ADDED Requirements

### Requirement: 宿主必须提供统一的本地化投影组件
宿主系统 SHALL 在 `internal/service/i18n` 包内提供 `LocaleProjector` 组件,统一封装业务实体的翻译键推导、跳过策略与源文案选择。所有需要为后端拥有数据做语言投影的业务模块 MUST 通过 `LocaleProjector` 提供的语义化方法完成投影,不得在 `*_i18n.go` 适配文件内重复实现"何时翻译 / 何时跳过 / 何种 fallback"决策。`LocaleProjector` MUST 至少覆盖菜单、字典类型、字典数据、配置、内置定时任务、定时任务分组、内置角色、插件元数据,并在新增受治理实体时通过扩展该组件接入。

#### Scenario: 业务模块通过 LocaleProjector 完成投影
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` / `role` / `plugin` 等模块需要为查询结果做本地化投影
- **THEN** 模块持有 `LocaleProjector` 字段并调用语义化方法(例如 `ProjectMenu`、`ProjectDictType`)
- **AND** 模块自身代码不再判断 "是否默认语言"、"是否内置记录"、"使用哪个 Translate* 方法"

#### Scenario: 默认语言下的跳过策略由 Projector 集中决策
- **WHEN** 当前请求语言等于运行时默认语言且业务实体的可编辑字段需要保持数据库原值
- **THEN** `LocaleProjector` 在内部统一作出"跳过翻译"的决策
- **AND** 业务模块不再通过 `i18nSvc.ResolveLocale(ctx, "") == DefaultLocale` 之类的写法重复判断

### Requirement: LocaleProjector 必须为受保护内置记录提供按稳定锚点的本地化投影
`LocaleProjector` SHALL 区分"可编辑业务记录"与"框架受保护内置记录":对可编辑业务记录默认保持数据库原值,对受保护内置记录(如内置 `admin` 角色、内置默认任务分组、内置定时任务、宿主默认菜单)按稳定业务锚点提供只读列表展示位的本地化投影。Projector MUST 在源码注释中明确每类受保护记录的判定规则与翻译键推导规则。

#### Scenario: 内置 admin 角色在英文环境只读列表中显示本地化名称
- **WHEN** 管理员在 `en-US` 环境下查询角色列表
- **THEN** Projector 仅对内置 `admin` 角色按 `role.admin.name` 翻译键提供本地化展示
- **AND** 其他可编辑角色的 `name` 字段保持数据库原值

#### Scenario: 用户创建的自定义任务保持原值
- **WHEN** 管理员在任意非默认语言下查询调度任务列表
- **THEN** Projector 不会对 `is_builtin = 0` 的用户自建任务做翻译投影
- **AND** 该任务的 `name` / `description` 显示数据库原始值
