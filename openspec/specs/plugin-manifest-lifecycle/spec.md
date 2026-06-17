# 插件清单生命周期

## Purpose

定义源码插件和动态插件的插件清单发现、生命周期资源同步、只读治理查询、SQL 资源分类和语言资源扩展行为。
## Requirements
### Requirement:插件清单和生命周期必须支持新增语言时的零代码扩展

插件清单和生命周期 SHALL 在宿主新增内置语言时自动覆盖新的语言运行时 UI 翻译资源和 apidoc 翻译资源，无需修改宿主代码或各插件源码。源码插件 SHALL 在自己的 `manifest/i18n/<locale>/*.json` 和 `manifest/i18n/<locale>/apidoc/**/*.json` 中追加该语言的资源；动态插件 SHALL 在打包时将该语言的资源写入发布自定义段；宿主在加载、启用、禁用、升级和卸载流程中自动发现、加载和清理这些资源。

#### Scenario:启用新语言后插件资源自动集成
- **当** 宿主启用额外的内置语言
- **且** 源码插件提供 `manifest/i18n/<locale>/*.json`
- **则** 启用插件将该语言的资源添加到运行时翻译聚合
- **且** 禁用或卸载插件从聚合中移除这些资源
- **且** 该流程不需要宿主代码修改和不相关插件代码修改

#### Scenario:动态插件通过发布携带新语言资源
- **当** 动态插件在新版本中添加 `manifest/i18n/<locale>/*.json` 并重新打包
- **且** 宿主将插件升级到该发布
- **则** 新语言资源生效，旧发布资源不再使用
- **且** 缓存失效范围限定在受影响的插件扇区

### Requirement:插件列表查询无副作用

系统 SHALL 将插件列表查询视为无副作用的读操作。列表查询可读取发现的源码清单、动态插件注册表数据、发布快照和治理投影，但不得创建、更新或删除插件治理表数据。插件扫描和治理同步必须仅由显式同步操作或宿主启动同步操作触发。宿主启动同步也 SHALL 是差异驱动的：当插件 registry、release snapshot、菜单、权限和资源引用投影均无差异时，不得开启事务、不得写入数据库、不得执行写后回读。系统 SHALL 允许插件列表查询复用已预热的完整插件管理读模型，但该读模型必须在插件清单、动态产物、插件治理状态、发布快照、资源引用、菜单权限投影、插件声明或租户供应策略发生变化后显式失效。

#### Scenario:从管理页面查询插件列表
- **当** 管理员打开插件管理并调用 `GET /api/v1/plugins` 时
- **则** 系统返回插件列表和当前治理状态
- **且** GET 请求不写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`

#### Scenario:显式同步插件
- **当** 管理员通过 `POST /api/v1/plugins/sync` 触发插件同步时
- **则** 系统扫描源码插件和动态插件产物
- **且** 系统可从清单同步注册表、发布快照、资源索引、菜单和权限治理数据

#### Scenario:启动同步无差异时不产生数据库副作用
- **当** 宿主启动同步发现插件清单与现有治理投影完全一致
- **则** 系统不得为该插件写入 `sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu` 或 `sys_role_menu`
- **且** 系统不得开启空事务或为了刷新启动快照重复回读同一治理行

#### Scenario:列表读模型命中时保持完整治理字段
- **当** 插件管理列表读模型已预热且仍有效
- **则** `GET /api/v1/plugins` 可以直接复用该读模型返回结果
- **且** 返回字段必须保持与同步构建路径一致，包含依赖检查、宿主服务授权信息、声明路由、mock 数据标识、运行时升级状态和租户供应策略
- **且** 系统不得为了提升首屏速度删除弹窗依赖的治理字段

#### Scenario:插件治理变化后列表读模型失效
- **当** 插件同步、动态包上传、安装、卸载、启用、禁用、源码升级、动态升级或租户供应策略更新成功后
- **则** 系统必须显式失效插件管理列表读模型
- **且** 单机部署中后续列表查询必须重建本地读模型
- **且** 集群部署中其他实例必须通过共享修订号或等价机制观察到失效并刷新本地读模型

### Requirement: 插件清单处理必须与治理副作用分离

系统 SHALL 将插件清单扫描、解析、校验和访问视为清单事实源读取能力。清单处理路径不得隐藏写入`sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu`、`sys_role_menu`或其他治理表，也不得隐藏触发菜单同步、权限同步、资源引用同步、hook 分发、运行时节点状态同步或缓存失效。需要治理写入或副作用时，调用方 MUST 在显式同步、安装、卸载、启用、禁用、升级或启动同步编排入口中按顺序调用清单、存储和副作用能力。

#### Scenario: 清单扫描仅读取资源

- **WHEN** 系统扫描源码插件或动态插件 manifest 资源
- **THEN** 系统只解析、校验和返回清单声明
- **AND** 扫描过程不得写入插件治理表
- **AND** 扫描过程不得触发菜单、权限、资源引用或 hook 副作用

#### Scenario: 显式同步插件治理资源

- **WHEN** 管理员或宿主启动同步编排确认需要同步插件治理资源
- **THEN** 编排入口先通过`catalog`读取和校验 manifest
- **AND** 再通过`store`写入注册表、发布快照、授权快照或治理投影
- **AND** 最后显式调用菜单同步、资源引用同步、hook 分发或运行时状态同步能力

#### Scenario: 插件列表查询复用清单和治理投影

- **WHEN** 插件列表或详情查询需要展示清单声明和治理状态
- **THEN** 查询路径可以组合`catalog`清单投影与`store`治理投影
- **AND** 查询路径不得借由`catalog`扫描隐式修复或写入治理表

### Requirement: 插件治理写入后的副作用调用点必须可追踪

系统 SHALL 在插件显式同步、安装、卸载、启用、禁用、升级、动态包上传和租户供应策略更新等治理路径中保留可追踪的副作用调用点。每个菜单同步、资源引用同步、hook 分发、运行时节点状态同步和缓存失效操作 MUST 由当前编排入口显式触发，且错误处理语义必须与治理写入顺序一致。

#### Scenario: 插件启用触发状态同步

- **WHEN** 插件启用编排成功写入插件治理状态
- **THEN** 编排入口显式触发必要的菜单、资源引用、运行时状态或 hook 副作用
- **AND** 副作用调用不得隐藏在`catalog.SetPluginStatus`或等价清单 helper 内

#### Scenario: 插件安装同步资源引用

- **WHEN** 插件安装编排完成 manifest 校验并写入安装治理状态
- **THEN** 编排入口显式同步插件资源引用投影
- **AND** 资源引用同步失败时安装流程必须按既有错误处理语义返回失败或执行回滚

#### Scenario: 插件治理变化触发缓存失效

- **WHEN** 插件同步、动态包上传、安装、卸载、启用、禁用、升级或租户供应策略写入成功
- **THEN** 编排入口必须保留既有插件管理读模型和运行时派生缓存失效语义
- **AND** 缓存失效不得因为治理写入 owner 从`catalog`迁移到`store`而遗漏

### Requirement:插件宿主服务元数据查找必须避免模式探测错误

系统 SHALL 通过只读元数据查询读取插件列表宿主服务投影的宿主数据库元数据。该查找不得触发 `information_schema.TABLES` 的错误业务表模式探测；如果数据库不支持元数据查找或查找失败，插件列表 API SHALL 降级返回原始表名。

#### Scenario:解析动态插件权限的数据表注释
- **当** 插件列表项声明 `data.resources.tables` 时
- **则** 系统尝试读取表注释用于权限审查展示
- **且** 查找不发出 `SHOW FULL COLUMNS FROM TABLES` 错误

#### Scenario:元数据查找不可用
- **当** 当前数据库方言不支持宿主表注释查找或查找失败时
- **则** 插件列表 API 仍成功返回
- **且** hostServices 权限展示使用原始表名作为回退信息

### Requirement:插件清单 SQL 资源必须将 mock-data 分类为独立资产类型

扫描插件 `manifest/sql/` 目录时，宿主 SHALL 区分安装、卸载和 mock-data SQL 资产，不重叠。`manifest/sql/*.sql` 属于安装资产，`manifest/sql/uninstall/*.sql` 属于卸载资产，`manifest/sql/mock-data/*.sql` 属于 mock 资产。Mock SQL 文件不得出现在安装或卸载资产列表中。源码插件和动态插件必须使用相同的扫描逻辑。

#### Scenario:安装资产列表排除 mock-data 文件
- **当** 宿主解析包含 `manifest/sql/001-schema.sql` 和 `manifest/sql/mock-data/001-mock.sql` 的插件的安装 SQL 资产时
- **则** 返回的安装资产列表仅包含 `001-schema.sql`
- **且** 不包含 `mock-data/001-mock.sql` 或该路径的任何变体

#### Scenario:Mock 资产扫描仅返回 mock-data 文件
- **当** 宿主解析同一插件的 mock SQL 资产时
- **则** 返回的资产列表仅包含 `manifest/sql/mock-data/` 下的文件
- **且** 文件按文件名升序排列

### Requirement:动态插件打包必须保留 mock-data 目录约定

动态插件打包 SHALL 在产物文件系统视图中保留 `manifest/sql/mock-data/`，并使用与源码插件相同的运行时扫描方法。打包工具和产物模式不得为此引入不同的 mock-data 路径或额外的清单字段。

#### Scenario:动态插件升级保留 mock-data 可见性
- **当** 动态插件在新版本中添加或修改 `manifest/sql/mock-data/*.sql`
- **且** 宿主升级到新产物
- **则** mock SQL 扫描反映新版本内容
- **且** mock-data 目录通过产物文件系统视图保持可见

### Requirement:插件 SQL 资源执行前必须经当前方言转译

系统 SHALL 在执行插件 `manifest/sql/` 安装资产、`manifest/sql/uninstall/` 卸载资产、`manifest/sql/mock-data/` mock 资产中的任一 SQL 文件之前，先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)` 入口处理单一 PostgreSQL 14+ 方言来源的 SQL 内容。`sourceName` SHALL 使用插件标识、资产类型与 SQL 文件路径组合出的稳定诊断名。该规则同时适用于源码插件与动态插件、安装阶段与卸载阶段、运行时嵌入式 SQL 与开发时本地 SQL。插件源码侧 SHALL 仅维护单一 PostgreSQL 方言来源的 SQL 文件，不得为不同数据库引擎维护多份 SQL 文件。

#### Scenario:源码插件安装时 SQL 走方言转译
- **当** 源码插件 `linapro-monitor-loginlog` 在 PostgreSQL 模式下首次启用并执行 `manifest/sql/001-linapro-monitor-loginlog-schema.sql` 时
- **则** 插件安装管线先调用当前 PostgreSQL 方言实例的 `TranslateDDL(ctx, sourceName, ddl)`
- **且** 返回的 PostgreSQL SQL 语句逐一成功执行
- **且** 插件源码 `manifest/sql/` 目录下保持单一 PostgreSQL 方言 SQL 文件

#### Scenario:动态插件升级时 SQL 走方言转译
- **当** 动态插件升级到新版本且新版本携带新的 `manifest/sql/*.sql` 文件
- **且** 当前宿主以 PostgreSQL 方言运行
- **则** 插件升级管线对每个新 SQL 文件调用当前 PostgreSQL 方言实例的 `TranslateDDL(ctx, sourceName, ddl)`
- **且** 返回的语句逐一执行
- **且** 任一文件转译或执行失败时升级管线返回失败状态

#### Scenario:插件卸载时 uninstall SQL 走方言转译
- **当** 源码插件或动态插件被卸载且 `manifest/sql/uninstall/` 下存在卸载 SQL 时
- **则** 卸载管线对每个 uninstall SQL 文件调用当前方言的 `TranslateDDL`
- **且** 转译后的 `DROP TABLE IF EXISTS` 等语句在当前数据库上成功执行
- **且** 卸载流程不依赖原 MySQL 方言专属语法

#### Scenario:插件 mock-data 加载时 SQL 走方言转译
- **当** 运维人员运行 `make db.mock confirm=mock` 且某插件提供 `manifest/sql/mock-data/*.sql` 时
- **则** mock 加载管线对每个 mock SQL 文件调用当前方言的 `TranslateDDL`
- **且** 返回的 `INSERT INTO` 等 PostgreSQL 语句在当前数据库上成功执行

#### Scenario:插件 SQL 转译失败时安装管线快速失败
- **当** 插件某 SQL 文件包含当前方言转译器未覆盖的 MySQL 语法时
- **则** 插件安装 / 升级 / 卸载 / mock 加载管线立即停止后续 SQL 执行
- **且** 错误日志包含失败的插件标识、SQL 资产类型、失败文件名、行号提示与未覆盖关键字
- **且** 管线向上层返回失败状态，便于调用方明确定位待修复的 SQL 文件

### Requirement:插件 SQL 文件必须符合 PostgreSQL-only SQL 源约束

源码插件与动态插件提交到仓库或发布产物中的 SQL 文件 SHALL 使用 PostgreSQL 14+ 语法，并遵守 `sql-source-syntax` 中的 PostgreSQL-only 源约束。插件 SQL 不得使用 MySQL 或 SQLite 特有语法，例如 `FULLTEXT INDEX` / `SPATIAL INDEX` / `ON DUPLICATE KEY UPDATE` / `INSERT IGNORE` / `FIND_IN_SET` / `GROUP_CONCAT` / `IF()` / `COLLATE NOCASE` 等。使用 PostgreSQL 高级特性前必须新立 OpenSpec 变更评估可维护性、升级策略、DAO 兼容性和测试覆盖。

#### Scenario:插件 SQL 不使用 MySQL 或 SQLite 特有语法
- **当** 任一源码插件或动态插件提交 `manifest/sql/*.sql` 文件时
- **则** SQL 文件不包含 MySQL 或 SQLite 特有语法
- **且** 语句可在 PostgreSQL 上直接成功执行

#### Scenario:违规 SQL 在审查与测试阶段被发现
- **当** 插件作者提交包含 `FULLTEXT INDEX` 等未覆盖语法的 SQL 文件时
- **则** SQL 源治理扫描、PostgreSQL 资产 smoke 或审查在合并前报错
- **且** 错误消息明确指向违规文件与未覆盖关键字
- **且** 该 SQL 在 PR / 变更审查阶段必须被修复，方可进入主干

### Requirement: 插件清单生命周期必须识别依赖声明

插件清单发现、显式同步、发布快照和只读治理查询 SHALL 保留并暴露插件 `dependencies` 声明。源码插件和动态插件的依赖声明必须在清单校验、同步到发布快照和插件列表投影时使用同一结构化语义。

#### Scenario: 同步源码插件依赖声明
- **WHEN** 源码插件 `plugin.yaml` 包含 `dependencies`
- **AND** 管理员执行显式插件同步
- **THEN** 系统校验依赖声明
- **AND** 系统在插件发布快照中保留依赖声明
- **AND** 插件列表或详情查询可返回依赖摘要

#### Scenario: 同步动态插件依赖声明
- **WHEN** 动态插件产物 manifest 包含 `dependencies`
- **AND** 系统解析动态插件产物
- **THEN** 系统使用与源码插件相同的依赖校验规则
- **AND** 动态插件发布快照保留依赖声明

### Requirement: 插件卸载生命周期必须检查反向依赖

插件卸载请求 SHALL 在执行卸载副作用前检查已安装插件的硬依赖。如果下游已安装插件依赖目标插件，卸载生命周期必须阻断。

#### Scenario: 卸载前发现反向依赖
- **WHEN** 管理员请求卸载插件 `base`
- **AND** 已安装插件 `consumer` 硬依赖 `base`
- **THEN** 系统拒绝卸载 `base`
- **AND** 系统返回 `consumer` 作为下游依赖

#### Scenario: 反向依赖读取使用发布快照
- **WHEN** 已安装源码插件的当前工作区 manifest 不可读取
- **THEN** 系统优先使用已安装发布快照中的依赖声明执行反向依赖检查
- **AND** 无法确认依赖安全时采用保守阻断策略

### Requirement: 插件生命周期和租户供应策略治理必须要求平台上下文

插件显式同步、上传、安装、卸载、启用、禁用、升级、安装模式变更和租户供应策略写入 SHALL 被视为平台生命周期治理。调用方除具备对应权限字符串外，还 MUST 处于平台上下文；租户上下文和代管租户上下文均不得执行这些操作。

#### Scenario: 租户上下文修改新租户自动启用策略被拒绝

- **WHEN** 租户用户调用插件新租户自动启用策略更新接口
- **THEN** 系统 MUST 返回平台上下文 required 的结构化业务错误
- **AND** 不修改 `sys_plugin.auto_enable_for_new_tenants`
- **AND** 不触发插件状态缓存刷新

#### Scenario: 平台上下文修改新租户自动启用策略

- **WHEN** 平台管理员在平台上下文更新插件新租户自动启用策略
- **THEN** 系统在校验插件能力、安装模式和既有治理规则后保存策略
- **AND** 触发插件治理缓存或运行时快照的显式作用域失效
- **AND** 后续新租户供应按更新后的策略执行

#### Scenario: 代管上下文不能执行插件生命周期治理

- **WHEN** 平台管理员正在代管某租户
- **AND** 调用插件同步、上传、安装、卸载、启用、禁用或升级接口
- **THEN** 系统 MUST 拒绝该操作
- **AND** 不修改平台插件治理数据

### Requirement: 插件生命周期 SQL 与迁移账本必须事务一致

系统 SHALL 在 PostgreSQL 默认运行环境下，将插件 install、upgrade、uninstall 和 rollback 生命周期 SQL 文件执行与对应 `sys_plugin_migration` 账本记录放入同一事务边界。任一 SQL 文件转译、语句执行或账本写入失败时，系统 MUST 回滚本次生命周期 SQL 和账本写入。

#### Scenario: 生命周期 SQL 中途失败时回滚账本
- **WHEN** 插件 P 安装期间执行多个 `manifest/sql/*.sql` 文件
- **AND** 其中一个 SQL 语句执行失败
- **THEN** 系统回滚本次安装生命周期中已执行的 SQL 语句
- **AND** 系统不得写入表示该失败 SQL 文件已成功完成的 `sys_plugin_migration` 记录

#### Scenario: 迁移账本写入失败时回滚 SQL
- **WHEN** 插件 P 升级期间 SQL 文件已经执行成功
- **AND** 对应 `sys_plugin_migration` 账本写入失败
- **THEN** 系统回滚本次升级生命周期 SQL
- **AND** 插件 P 不得进入升级成功状态

#### Scenario: rollback SQL 使用相同事务语义
- **WHEN** 插件 P 的 rollback SQL 被执行
- **THEN** rollback SQL 文件执行和 rollback 方向迁移账本写入在同一事务中完成
- **AND** 任一步失败都会回滚本次 rollback SQL 与账本写入

### Requirement: 插件生命周期 rollback 失败必须进入权威诊断

系统 SHALL 将插件生命周期失败后的 rollback 失败纳入权威失败诊断。rollback SQL、菜单恢复、前端资源恢复、权限治理恢复或发布状态恢复失败时，系统 MUST 保留原始失败原因和 rollback 失败原因；不得只写 warning 日志后返回原始错误。

#### Scenario: rollback SQL 失败被返回给调用方
- **WHEN** 动态插件 P 安装失败后执行 rollback SQL
- **AND** rollback SQL 执行失败
- **THEN** 系统返回或记录同时包含安装原始失败和 rollback SQL 失败的诊断
- **AND** 插件 P 的运行时状态标记为失败或需要人工处理

#### Scenario: 治理资源恢复失败被记录
- **WHEN** 动态插件 P 升级失败后系统尝试恢复菜单、前端资源或权限治理资源
- **AND** 任一恢复动作失败
- **THEN** 系统将恢复失败写入插件发布、节点状态或 registry 的失败诊断
- **AND** 后续管理或协调流程可以读取该失败原因

### Requirement:插件 manifest 资源必须支持插件自作用域只读读取

系统 SHALL 为源码插件和动态插件提供`HostServices.Manifest()`能力，使插件代码能够只读读取当前插件`manifest/`目录下的原始资源。读取范围 MUST 绑定当前插件 ID，不得允许插件读取宿主 manifest、其他插件 manifest、任意文件系统路径或 URL。`metadata.yaml` SHALL 在插件实际提供该文件时作为可通过该能力读取的普通可选资源，但系统不得要求所有插件都提交`metadata.yaml`，也不得为 `metadata.yaml` 保留独立的 `Metadata` 服务、`metadata` host service、`metadata.get` 或等价插件可见读取入口。

#### Scenario:源码插件读取自身 metadata 普通资源

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "metadata.yaml")`
- **AND** `apps/lina-plugins/plugin-a/manifest/metadata.yaml`存在
- **THEN** 系统返回该文件内容
- **AND** 读取作用域限定为`plugin-a`的`manifest/`目录
- **AND** 该读取不经过独立的 `Metadata` 服务

#### Scenario:源码插件读取自身 config 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "config/config.example.yaml")`
- **AND** `apps/lina-plugins/plugin-a/manifest/config/config.example.yaml`存在
- **THEN** 系统返回该文件原始内容
- **AND** 不把该读取结果作为插件运行期有效配置自动生效

#### Scenario:源码插件读取自身 SQL 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "sql/001-schema.sql")`
- **AND** `apps/lina-plugins/plugin-a/manifest/sql/001-schema.sql`存在
- **THEN** 系统返回该 SQL 文件原始内容
- **AND** 不执行该 SQL

#### Scenario:源码插件读取自身 i18n 资源原文

- **WHEN** 源码插件`plugin-a`调用`HostServices.Manifest().Get(ctx, "i18n/zh-CN/plugin.json")`
- **AND** `apps/lina-plugins/plugin-a/manifest/i18n/zh-CN/plugin.json`存在
- **THEN** 系统返回该 JSON 文件原始内容
- **AND** 不把该读取作为翻译资源注册或缓存失效动作

#### Scenario:插件未提供 metadata

- **WHEN** 插件未维护`manifest/metadata.yaml`
- **THEN** 插件无需为了目录规范提交空白或占位 metadata 文件
- **AND** 插件清单不得申请读取不存在的`metadata.yaml`

#### Scenario:动态插件读取 artifact 中的 metadata 普通资源

- **WHEN** 动态插件`plugin-a`调用`manifest.get`读取`metadata.yaml`
- **AND** 当前 active release artifact 携带`manifest/metadata.yaml`
- **AND** 当前授权快照允许读取`metadata.yaml`
- **THEN** 系统从该 active release 的资源快照返回文件内容
- **AND** 该内容绑定当前 active release 的 checksum 或 generation

#### Scenario:动态插件读取已授权专用目录资源原文

- **WHEN** 动态插件`plugin-a`调用`manifest.get`读取`config/config.example.yaml`、`sql/001-schema.sql`或`i18n/zh-CN/plugin.json`
- **AND** 当前 active release artifact 携带对应资源
- **AND** 当前授权快照允许读取对应路径
- **THEN** 系统从该 active release 的资源快照返回文件原始内容
- **AND** 不触发配置生效、SQL 执行或翻译资源注册

#### Scenario:插件扫描 YAML manifest 资源到结构体

- **WHEN** 插件调用`HostServices.Manifest().Scan(ctx, "metadata.yaml", "", &metadata)`
- **AND** `metadata.yaml`是合法 YAML 文档
- **THEN** 系统将文件内容绑定到插件提供的结构体
- **AND** 结构体业务语义和验证逻辑由插件内部维护

### Requirement:Manifest 资源读取必须执行路径安全治理

系统 SHALL 将`Manifest()`的路径参数解释为相对当前插件`manifest/`根目录的 slash 路径。系统 MUST 拒绝空根读取、绝对路径、路径穿越、Windows drive path、URL、跨插件路径和未授权路径；动态插件还 MUST 按`plugin.yaml`中 `service: manifest` 的 `resources.paths` 与宿主确认后的授权快照校验。系统 MUST 允许合法的 `config/`、`sql/` 和 `i18n/` manifest 相对路径参与同一套路径安全和授权校验。

#### Scenario:合法相对路径被允许

- **WHEN** 插件读取`metadata.yaml`、`resources/policy.yaml`、`config/config.example.yaml`、`sql/001-schema.sql`或`i18n/zh-CN/plugin.json`
- **AND** 该路径位于当前插件`manifest/`目录下且满足授权策略
- **THEN** 系统允许读取该资源
- **AND** 返回内容不包含其他插件资源

#### Scenario:路径穿越被拒绝

- **WHEN** 插件读取`../other-plugin/manifest/metadata.yaml`或`../../apps/lina-core/manifest/config/config.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不访问目标文件系统路径

#### Scenario:绝对路径和 URL 被拒绝

- **WHEN** 插件读取`/etc/passwd`、`C:\\secret.yaml`或`http://example.com/config.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不发起本地文件或网络读取

#### Scenario:动态插件未授权 manifest 路径被拒绝

- **WHEN** 动态插件当前授权快照只允许读取`metadata.yaml`
- **AND** 插件调用`manifest.get`读取`config/config.example.yaml`
- **THEN** 系统拒绝该请求
- **AND** 不返回 artifact 中该资源内容

### Requirement:插件 manifest 专用生命周期资源不得被通用 Manifest 读取混用

系统 SHALL 保持插件`manifest/sql/`、`manifest/i18n/`和`manifest/config/`等专用目录的既有治理边界。`HostServices.Manifest()`用于读取插件自身打包的 manifest 原始资源，可以读取这些专用目录中的文件原文，但不得绕过 SQL 生命周期管线、i18n 资源管线或插件配置服务。插件运行期有效配置 MUST 通过`HostServices.Config()`读取。

#### Scenario:配置目录可读取原文但有效配置通过 Config 读取

- **WHEN** 插件需要读取打包的`manifest/config/config.yaml`原文
- **THEN** 插件 MAY 使用`HostServices.Manifest().Get(ctx, "config/config.yaml")`
- **AND** 该读取只返回打包原文
- **AND** 插件需要读取运行期有效配置时 MUST 使用`HostServices.Config()`

#### Scenario:SQL 和 i18n 资源继续由专用管线处理

- **WHEN** 插件安装 SQL 或 i18n 资源需要被宿主加载
- **THEN** 系统继续使用插件生命周期、数据库和 i18n 管线扫描`manifest/sql/`和`manifest/i18n/`
- **AND** `HostServices.Manifest()`不得成为执行 SQL 或加载翻译包的替代入口

#### Scenario:旧 Metadata 服务语义被移除

- **WHEN** 插件代码或动态插件清单需要读取`manifest/metadata.yaml`
- **THEN** 系统只提供`HostServices.Manifest()`或`manifest.get`作为读取入口
- **AND** 系统不得继续发布`Metadata()`、`metadata.get`或`service: metadata`读取入口

### Requirement: 显式插件安装生命周期必须执行依赖检查

显式插件安装请求 SHALL 在目标插件安装副作用前调用依赖解析和依赖检查。系统必须先执行框架版本和插件依赖检查，确认所有声明的插件依赖已经安装且版本满足后，才安装目标插件。系统 MUST NOT 根据插件清单自动安装依赖插件，也 MUST NOT 通过`required`、`install`或等价字段驱动软依赖、手动安装策略或自动安装策略。

#### Scenario: 显式安装先检查依赖

- **WHEN** 管理员请求安装插件`x`
- **AND** `x`在`dependencies.plugins`中声明依赖`a`
- **THEN** 系统在执行`x`的安装 SQL 或动态运行时协调前检查`a`是否已安装且版本满足
- **AND** `a`未安装或版本不满足时拒绝安装`x`
- **AND** 系统不得自动安装`a`

#### Scenario: 依赖检查失败时安装无副作用

- **WHEN** 插件`x`的依赖检查失败
- **THEN** 系统不得执行`x`的安装 SQL
- **AND** 系统不得同步`x`的菜单、权限、资源引用或安装状态

### Requirement: 插件安装接口必须返回依赖检查结果

插件管理安装接口 SHALL 支持调用端获取依赖检查结果。安装失败时，错误响应必须以结构化业务错误表达依赖阻断原因，避免只依赖自由文本。接口响应和详情投影 MUST NOT 返回由插件清单驱动的自动安装计划、自动安装结果、软依赖提示或手动安装策略。

#### Scenario: 依赖阻断返回结构化错误

- **WHEN** 安装目标插件因为依赖缺失或依赖版本不满足失败
- **THEN** HTTP 响应包含稳定业务错误码
- **AND** 响应包含目标插件 ID、依赖插件 ID、当前版本和要求版本范围

#### Scenario: 安装成功不返回自动依赖结果

- **WHEN** 目标插件安装成功
- **THEN** 安装响应包含目标插件 ID
- **AND** 安装响应不得包含自动安装成功的依赖插件列表

### Requirement: 插件生命周期编排下沉后必须保持治理语义

系统 SHALL 在将插件生命周期编排迁入 lifecycle 子组件后保持现有安装、卸载、启用、禁用、状态变更、源码插件生命周期、启动自动启用和租户生命周期钩子的治理语义。平台上下文守卫、依赖检查、反向依赖阻断、host service authorization、SQL migration、资源引用同步、菜单权限同步、hook 分发、runtime state 同步和缓存失效不得因迁移遗漏或改变顺序。

#### Scenario: 插件安装语义保持

- **WHEN** 管理员安装 source 或 dynamic 插件
- **THEN** 系统仍执行依赖检查、manifest 校验、host service authorization、SQL migration、registry/release 写入、资源和菜单同步、hook 分发和缓存失效
- **AND** 任一步骤失败时保持迁移前的回滚或失败返回语义

#### Scenario: 插件卸载语义保持

- **WHEN** 管理员卸载插件
- **THEN** 系统仍在副作用前检查反向依赖和 lifecycle veto
- **AND** 保持 runtime 停止、资源引用清理、菜单权限清理、uninstall SQL、storage cleanup 和缓存失效顺序
- **AND** force 与 purge storage 语义不因 lifecycle 子组件迁移改变

#### Scenario: 启动自动启用语义保持

- **WHEN** 宿主启动并处理`plugin.autoEnable`
- **THEN** 系统仍只对配置目标插件执行自动安装/启用
- **AND** 保持 startup auto-enable 的依赖检查、mock-data 策略、租户供应策略和启动统计语义

### Requirement: 插件生命周期业务控制参数必须显式传递

系统 SHALL 将生命周期普通业务控制参数作为方法参数、options 或稳定输入结构显式传递。安装 mock data、startup auto-enable 标记、依赖检查结果和类似控制语义 MUST NOT 通过普通请求 context key 隐式改变生命周期行为。仅用于一次启动编排的只读大快照 MAY 继续通过 context 传递，但不得改变生命周期业务语义。

#### Scenario: 安装 mock data

- **WHEN** 调用方请求安装插件并选择加载 mock-data
- **THEN** 该选择通过 install options 或等价显式输入传入 lifecycle
- **AND** lifecycle 不通过 context key 判断是否加载 mock-data

#### Scenario: 启动自动启用

- **WHEN** startup bootstrap 触发插件自动安装或启用
- **THEN** startup 语义通过显式 options 或内部启动入口传入 lifecycle
- **AND** 普通 HTTP 安装请求无法通过伪造 context key 获得 startup-only 行为

### Requirement: 租户生命周期钩子必须通过 lifecycle 子组件编排

系统 SHALL 由 lifecycle 子组件编排租户删除、租户插件禁用和新租户供应相关的插件 lifecycle precondition 与 notification。根门面和 tenant capability adapter MUST 只依赖窄接口，不得复制租户生命周期扫描和 veto 汇总逻辑。

#### Scenario: 租户删除前检查插件 veto

- **WHEN** 租户能力在删除租户前请求插件 precondition
- **THEN** lifecycle 子组件扫描需要参与租户删除 veto 的已启用插件
- **AND** 使用统一 veto 汇总返回结构化错误
- **AND** 根门面不复制该扫描逻辑

#### Scenario: 新租户供应插件

- **WHEN** 新租户创建后需要按平台策略供应插件
- **THEN** lifecycle 子组件读取 auto-enable for new tenants 策略并执行供应编排
- **AND** 供应完成后通过统一插件变化发布入口失效派生缓存

### Requirement: 插件清单读取必须复用有界读模型缓存

系统 SHALL 为插件清单事实源提供内部读模型缓存。源码插件 manifest、动态插件 desired manifest、release manifest 与 release YAML 快照 MUST 按各自权威数据源和不可变性分开缓存。缓存读取路径 MUST 不写入插件治理表，不触发菜单、权限、资源引用、hook 或运行时状态副作用。

#### Scenario: 动态插件稳态扫描复用文件状态守卫

- **WHEN** 系统扫描动态插件 artifact 目录且某 artifact 的路径、大小和修改时间均未变化
- **THEN** 系统复用已缓存的 manifest 解析结果
- **AND** 不重新整文件读取、哈希和解析该 artifact

#### Scenario: 单插件读取不全量扫描

- **WHEN** 调用方按`pluginID`读取单个插件清单且缓存索引已包含该插件 artifact 路径
- **THEN** 系统只读取或校验该插件对应 artifact 的缓存条目
- **AND** 不为了单插件详情查询扫描全部插件 artifact

#### Scenario: Release 快照按校验和复用

- **WHEN** 调用方读取同一`pluginID`、`releaseID`和`checksum`对应的 release manifest
- **THEN** 系统复用已有解析快照
- **AND** 不重复反序列化同一 release YAML 内容

#### Scenario: 清单读取保持无副作用

- **WHEN** 插件列表、详情、依赖检查、OpenAPI 投影或 hook 分发读取清单缓存
- **THEN** 读取过程不得写入`sys_plugin`、`sys_plugin_release`、`sys_plugin_resource_ref`、`sys_menu`或`sys_role_menu`
- **AND** 所有治理同步仍由显式同步、安装、升级、启用、禁用或启动同步编排入口触发

### Requirement: 插件清单缓存必须支持按插件显式失效

系统 SHALL 在插件同步、动态包上传、安装、卸载、启用、禁用、升级、active release 切换或源码插件同步成功后，按`pluginID`或 artifact 路径失效对应清单缓存。全局失效只允许用于全局配置或无法确定影响范围的治理事件。

#### Scenario: 单插件升级只失效目标插件清单

- **WHEN** 动态插件`P`升级到新的 active release
- **THEN** 系统失效插件`P`的 desired manifest、release manifest 和相关索引条目
- **AND** 其他插件清单缓存保持可用

#### Scenario: 卸载插件清理 release 快照

- **WHEN** 插件`P`被卸载且其 release 不再作为当前权威发布使用
- **THEN** 系统清理或淘汰插件`P`相关 release manifest 缓存
- **AND** 后续读取插件`P`必须回源确认其治理状态

