# plugin-manifest-lifecycle Specification

## Purpose
TBD - created by archiving change plugin-framework. Update Purpose after archive.
## Requirements
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

#### Scenario: 动态插件通过嵌入资源声明生成清单与 SQL 快照

- **WHEN** 动态插件作者使用 `go:embed` 声明 `plugin.yaml`、`manifest/sql` 与 `manifest/sql/uninstall`
- **THEN** 构建器必须从该嵌入文件系统中读取这些资源
- **AND** 运行时产物中嵌入的 manifest 与 SQL 快照必须继续作为宿主安装、上传和生命周期治理的真相源
- **AND** 宿主不得改为通过 guest 运行时方法动态获取这些治理资源

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

### Requirement: 动态插件清单可声明结构化宿主服务策略

系统 SHALL 允许动态插件在`plugin.yaml`中仅声明结构化`hostServices`策略，用于描述需要的宿主 service、method、资源申请和治理参数；宿主内部 capability 分类必须根据这些声明自动推导，而不是要求作者重复维护顶层`capabilities`字段。其中`storage`服务当前通过`resources.paths`声明逻辑路径申请，`data`服务当前通过`resources.tables`声明数据表申请。

#### Scenario: 插件声明宿主服务策略

- **WHEN** 开发者编写动态插件清单
- **THEN** 清单可以声明`hostServices`元数据
- **AND** 每个声明至少包含 service、method 集合以及资源申请或策略参数
- **AND** 清单不再需要单独声明顶层`capabilities`
- **AND** 构建器对未知 service、未知 method 和非法策略直接报错

#### Scenario: 宿主读取宿主服务策略快照

- **WHEN** 宿主查看一个动态插件的 manifest 快照或 release 快照
- **THEN** 宿主可以恢复该插件声明的宿主服务策略
- **AND** 管理员可以据此审查插件计划访问的宿主能力范围

#### Scenario: 插件声明资源申请而非宿主底层连接

- **WHEN** 开发者在清单中声明宿主服务依赖
- **THEN** 对`storage`服务，插件只声明稳定的逻辑路径或路径前缀`resources.paths`
- **AND** 对`network`服务，插件只声明 URL 模式列表
- **AND** 对`data`服务，插件在`resources`节点下声明需要访问的表名列表`tables`
- **AND** 对`cache`、`lock`和`notify`等低优先级服务，当前仍可继续使用逻辑`resourceRef`规划，其中分别表示缓存命名空间、逻辑锁名和通知通道标识
- **AND** 插件清单不得固化数据库连接、宿主文件绝对路径、缓存地址或密钥明文
- **AND** 真实资源绑定由宿主安装流程或管理员配置完成

### Requirement: 宿主服务资源申请纳入插件治理资源索引

系统 SHALL 将动态插件声明的宿主服务资源申请统一纳入`sys_plugin_resource_ref`治理资源索引；该表用于承载 release 级别的插件治理资源投影，而不只是镜像某个名为`resourceRef`的作者侧字段。对`storage`记录逻辑路径申请，对`network`记录 URL 模式申请，对`data`记录表名申请，对`cache`、`lock`、`notify`等低优先级服务继续记录逻辑资源引用。

#### Scenario: 安装或升级动态插件同步治理资源索引

- **WHEN** 宿主安装或升级一个声明了宿主服务资源的动态插件
- **THEN** 宿主将这些资源申请同步为插件资源归属记录
- **AND** 资源类型能够区分`host-storage`、`host-upstream`、`host-data-table`、`host-cache`、`host-lock`和`host-notify-channel`
- **AND** 这些记录可以参与审计、卸载和回滚治理

#### Scenario: 卸载或回滚动态插件更新治理资源索引

- **WHEN** 宿主卸载一个动态插件或将其回滚到旧 release
- **THEN** 宿主同步更新对应的宿主服务资源申请记录
- **AND** 当前 release 不再使用的逻辑路径、URL 模式、低优先级服务逻辑`resourceRef`或数据表声明不得继续保留为生效态

#### Scenario: 激活 release 时恢复逻辑引用绑定

- **WHEN** 宿主激活一个动态插件 release
- **THEN** 宿主根据 release 快照恢复资源申请的最终授权状态
- **AND** 运行时后续只按该快照解释宿主服务调用

### Requirement: 资源型宿主服务申请在安装或启用时需要宿主确认授权

系统 SHALL 在动态插件安装或启用阶段展示所有资源型宿主服务权限申请，并由宿主管理员确认最终授权结果。

#### Scenario: 安装时展示宿主服务权限申请

- **WHEN** 宿主准备安装一个声明了资源型 hostServices 的动态插件
- **THEN** 宿主展示插件申请的 service、method、资源标识（如`path`、URL 模式、`resourceRef`或`table`）及其治理参数摘要
- **AND** 当申请项属于`data` service 且宿主可解析表级说明时，宿主同时展示表名对应的人类可读说明，避免管理员只能依赖裸表名判断用途
- **AND** 管理员可以基于该清单审查插件计划访问的宿主资源范围

#### Scenario: 启用时确认或收窄宿主服务授权

- **WHEN** 宿主准备启用一个声明了资源型 hostServices 的动态插件 release
- **THEN** 宿主允许管理员批准、收窄或拒绝这些资源申请
- **AND** 宿主将最终确认结果持久化为当前 release 的授权快照
- **AND** 运行时后续只按这份最终快照解释宿主服务调用

