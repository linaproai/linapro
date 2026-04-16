## ADDED Requirements

### Requirement: 插件目录与清单契约统一

系统 SHALL 为所有插件提供统一的目录结构与清单契约。源码插件 MUST 放置在 `apps/lina-plugins/<plugin-id>/` 目录下；当前 `dynamic` 动态 `wasm` 插件 MUST 能从 `plugin.dynamic.storagePath` 中被发现，并解析出与源码插件等价的 manifest 信息。

#### Scenario: 发现源码插件目录

- **WHEN** 宿主扫描 `apps/lina-plugins/` 下的插件目录
- **THEN** 仅将包含合法清单文件的目录识别为插件
- **AND** 每个插件的 `plugin-id` 在宿主范围内唯一
- **AND** 清单仅需包含插件基础信息与一级插件类型

#### Scenario: `plugin.yaml` 保持精简且可声明插件菜单

- **WHEN** 宿主解析 `plugin.yaml`
- **THEN** 清单只强制要求 `id`、`name`、`version`、`type` 等基础字段
- **AND** 宿主不再要求 `schemaVersion`、`compatibility`、`entry` 等扩展元数据
- **AND** 插件若需要向宿主注册菜单或按钮权限，必须在清单 `menus` 元数据中声明
- **AND** 前端页面、`Slot` 与 SQL 文件位置仍优先按照目录与代码约定推导，而不是在清单中重复配置

#### Scenario: 清单一级类型只保留源码与动态两类

- **WHEN** 宿主解析 `plugin.yaml` 中的 `type`
- **THEN** `type` 仅允许 `source` 或 `dynamic`
- **AND** 当前仅 `wasm` 作为动态插件的产物语义，不再作为一级插件类型
- **AND** 对历史上的 `wasm` 一级类型值，宿主在治理视角下统一按 `dynamic` 处理

#### Scenario: 安装动态插件产物

- **WHEN** 管理员上传一个 `wasm` 文件安装动态插件
- **THEN** 宿主能够解析出与源码模式一致的插件标识、名称、版本与一级插件类型
- **AND** 对缺少这些基础字段的动态插件拒绝安装
- **AND** 宿主将上传产物写入 `plugin.dynamic.storagePath/<plugin-id>.wasm`

#### Scenario: 动态插件产物使用独立存储目录

- **WHEN** 宿主发现、上传或同步一个 `dynamic` `wasm` 动态插件产物
- **THEN** 运行时产物 MUST 使用 `plugin.dynamic.storagePath/<plugin-id>.wasm` 作为宿主侧规范落盘路径
- **AND** 宿主不得再依赖 `apps/lina-plugins/<plugin-id>/plugin.yaml` 作为 runtime 发现入口
- **AND** 运行时样例插件的可读源码目录 SHOULD 与源码插件一样继续收敛在 `backend/`、`frontend/` 与 `manifest/` 下维护

#### Scenario: 当前生效 release 从稳定归档重载

- **WHEN** 一个动态插件已经存在当前生效的 release，且宿主需要再次加载其 active manifest
- **THEN** 宿主从 `plugin.dynamic.storagePath/releases/<plugin-id>/<version>/<plugin-id>.wasm` 这类稳定归档路径重载该 release
- **AND** 宿主不会因为 staging 目录里出现更新的 `wasm` 文件就立即替换当前服务中的 active release
- **AND** active manifest 在重载后仍然包含该 release 内嵌声明的 Hook、通用资源契约与菜单元数据

### Requirement: 插件生命周期状态机可治理

系统 SHALL 为插件提供可审计的生命周期状态机，并按源码插件与动态插件区分生命周期语义。

#### Scenario: 源码插件随宿主编译集成

- **WHEN** 宿主编译源码插件所在的源码树并生成 Lina 二进制
- **THEN** 源码插件的后端 Go 代码与宿主源码一起完成编译
- **AND** 源码插件在插件注册表中视为已集成，不需要额外安装步骤
- **AND** 管理员只需要管理源码插件的启用与禁用状态

#### Scenario: 源码插件首次同步后默认启用

- **WHEN** 宿主首次发现一个源码插件并将其写入插件注册表
- **THEN** 该源码插件默认处于“已集成且已启用”状态
- **AND** 宿主后续同步不会覆盖管理员对该源码插件做出的显式禁用操作

#### Scenario: 安装动态插件

- **WHEN** 管理员安装一个合法的 `wasm` 动态插件
- **THEN** 宿主创建插件安装记录与当前版本记录
- **AND** 宿主按清单依次处理迁移、资源注册、权限接入与前后端装载准备
- **AND** 插件在显式启用前不会对普通用户可见

#### Scenario: 禁用插件

- **WHEN** 管理员将已启用插件切换为禁用状态
- **THEN** 宿主停止该插件的 Hook、Slot、页面与菜单暴露
- **AND** 宿主保留插件业务数据、角色授权关系与安装记录
- **AND** 插件重新启用后可以恢复既有治理关系

#### Scenario: 卸载动态插件

- **WHEN** 管理员卸载一个动态插件
- **THEN** 宿主移除该插件在宿主侧注册的菜单、资源引用、运行时产物与挂载信息
- **AND** 宿主默认不删除插件自己的业务数据表或业务数据
- **AND** 宿主保留卸载审计信息

### Requirement: 插件菜单通过清单元数据治理

系统 SHALL 使用 `plugin.yaml` 或动态产物嵌入 manifest 中的 `menus` 元数据管理插件菜单和按钮权限，而不是要求插件通过 SQL 直接操作 `sys_menu` 与 `sys_role_menu`。

#### Scenario: 源码插件同步菜单

- **WHEN** 宿主同步一个源码插件清单
- **THEN** 宿主依据该插件 `menus` 元数据幂等写入或更新对应的 `sys_menu`
- **AND** 宿主依据 `parent_key` 解析真实 `parent_id`
- **AND** 宿主为这些菜单补齐默认管理员角色授权，而不要求插件 SQL 再手工写入 `sys_role_menu`

#### Scenario: 安装动态插件注册菜单

- **WHEN** 管理员安装一个动态插件
- **THEN** 宿主在执行插件安装 SQL 后，继续依据 manifest `menus` 元数据幂等写入或更新对应的 `sys_menu`
- **AND** 插件安装 SQL 可以继续负责业务表与业务种子数据，但不再承担插件菜单注册职责

#### Scenario: 卸载动态插件删除菜单

- **WHEN** 管理员卸载一个动态插件
- **THEN** 宿主在插件卸载 SQL 成功执行后，依据 manifest `menus` 元数据删除对应的 `sys_role_menu` 关联与 `sys_menu`
- **AND** 删除范围仅限该插件 manifest 中声明的菜单键，不依赖插件 SQL 手工维护删除语句
- **AND** 若插件未声明任何菜单，宿主跳过菜单删除步骤

#### Scenario: 升级插件

- **WHEN** 管理员为已安装插件安装更高版本的 release
- **THEN** 宿主为插件创建新的 release 记录与代际信息
- **AND** 旧 release 在新 release 生效前保持可回退
- **AND** 升级失败时宿主能够回滚到上一个可用 release

#### Scenario: 升级失败 release 保持隔离

- **WHEN** 一个动态插件在升级、迁移或前端资源切换过程中失败并触发回滚
- **THEN** 宿主将失败 release 标记为 `failed`
- **AND** 宿主恢复注册表指针到上一个稳定 release
- **AND** 失败 release 的公共前端资源不会继续通过 `/plugin-assets/<plugin-id>/<version>/...` 对外提供

#### Scenario: 源码插件不暴露安装卸载操作

- **WHEN** 管理员查看源码插件的插件管理操作项
- **THEN** 宿主不会为源码插件展示安装或卸载操作
- **AND** 源码插件仅暴露同步发现、启用和禁用等适用操作

### Requirement: 插件资源归属与迁移记录可追踪

系统 SHALL 记录插件对宿主资源与迁移的占用关系，以支持卸载、重装、升级、审计与故障恢复。

#### Scenario: 插件注册宿主资源

- **WHEN** 插件在安装期间创建或声明菜单、权限、配置、字典、静态资源或其他宿主治理资源
- **THEN** 宿主记录该资源与插件、release 的归属关系
- **AND** 这些引用关系可以被查询、审计和用于卸载清理

#### Scenario: 执行插件迁移

- **WHEN** 插件安装或升级需要执行 SQL 或其他迁移步骤
- **THEN** 宿主记录每个迁移项的执行顺序、版本、校验摘要、执行结果与时间
- **AND** 同一个 release 的同一个迁移项不会被重复执行

#### Scenario: 插件版本 SQL 命名与目录约束

- **WHEN** 插件在 `manifest/sql/` 目录下提供安装阶段 SQL
- **THEN** 安装 SQL 文件 MUST 使用与宿主一致的命名格式 `{序号}-{当前迭代名称}.sql`
- **AND** 这些安装 SQL 文件 MUST 放在插件的 `manifest/sql/` 根目录下，供宿主按顺序扫描执行
- **AND** 插件卸载 SQL MUST 独立放在 `manifest/sql/uninstall/` 目录下
- **AND** 宿主初始化顺序执行流程 MUST 只扫描 `manifest/sql/` 根目录，不得误执行 `manifest/sql/uninstall/` 下的卸载 SQL

#### Scenario: 插件菜单治理不依赖整型菜单 ID

- **WHEN** 宿主根据插件 manifest `menus` 元数据同步宿主菜单与按钮权限
- **THEN** 菜单记录 MUST 使用 `menu_key` 作为菜单稳定标识
- **AND** 父子关系 MUST 通过 `parent_key` 解析真实 `parent_id`，而不是写死固定整型 `parent_id`
- **AND** 插件安装、升级与卸载流程 MUST 不依赖固定整型 `id`

#### Scenario: 安装过程部分失败

- **WHEN** 插件在迁移、资源注册或产物准备过程中任一步骤失败
- **THEN** 宿主将插件状态标记为失败或待人工介入
- **AND** 宿主回滚尚未生效的宿主治理资源
- **AND** 宿主保留失败上下文供后续诊断
