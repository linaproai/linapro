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

系统 SHALL 将插件列表查询视为无副作用的读操作。列表查询可读取发现的源码清单、动态插件注册表数据、发布快照和治理投影，但不得创建、更新或删除插件治理表数据。插件扫描和治理同步必须仅由显式同步操作或宿主启动同步操作触发。宿主启动同步也 SHALL 是差异驱动的：当插件 registry、release snapshot、菜单、权限和资源引用投影均无差异时，不得开启事务、不得写入数据库、不得执行写后回读。

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

系统 SHALL 在执行插件 `manifest/sql/` 安装资产、`manifest/sql/uninstall/` 卸载资产、`manifest/sql/mock-data/` mock 资产中的任一 SQL 文件之前，先调用当前方言的 `TranslateDDL(ctx, sourceName, ddl)` 将单一 MySQL 方言来源的 SQL 内容转换为目标方言可执行内容。`sourceName` SHALL 使用插件标识、资产类型与 SQL 文件路径组合出的稳定诊断名。该规则同时适用于源码插件与动态插件、安装阶段与卸载阶段、运行时嵌入式 SQL 与开发时本地 SQL。插件源码侧 SHALL 仅维护单一 MySQL 方言来源的 SQL 文件，不得为不同数据库引擎维护多份 SQL 文件。

#### Scenario:源码插件安装时 SQL 走方言转译
- **当** 源码插件 `linapro-monitor-loginlog` 在 SQLite 模式下首次启用并执行 `manifest/sql/001-linapro-monitor-loginlog-schema.sql` 时
- **则** 插件安装管线先调用当前 SQLite 方言实例的 `TranslateDDL(ctx, sourceName, ddl)` 将 MySQL 方言 DDL 转译为 SQLite 兼容语句
- **且** 转译后的语句在 SQLite 数据库上成功执行
- **且** 插件源码 `manifest/sql/` 目录下保持单一 MySQL 方言 SQL 文件

#### Scenario:动态插件升级时 SQL 走方言转译
- **当** 动态插件升级到新版本且新版本携带新的 `manifest/sql/*.sql` 文件
- **且** 当前宿主以 SQLite 方言运行
- **则** 插件升级管线对每个新 SQL 文件调用当前 SQLite 方言实例的 `TranslateDDL(ctx, sourceName, ddl)`
- **且** 转译后的语句逐一执行
- **且** 任一文件转译或执行失败时升级管线返回失败状态

#### Scenario:插件卸载时 uninstall SQL 走方言转译
- **当** 源码插件或动态插件被卸载且 `manifest/sql/uninstall/` 下存在卸载 SQL 时
- **则** 卸载管线对每个 uninstall SQL 文件调用当前方言的 `TranslateDDL`
- **且** 转译后的 `DROP TABLE IF EXISTS` 等语句在当前数据库上成功执行
- **且** 卸载流程不依赖原 MySQL 方言专属语法

#### Scenario:插件 mock-data 加载时 SQL 走方言转译
- **当** 运维人员运行 `make mock confirm=mock` 且某插件提供 `manifest/sql/mock-data/*.sql` 时
- **则** mock 加载管线对每个 mock SQL 文件调用当前方言的 `TranslateDDL`
- **且** 转译后的 `INSERT IGNORE INTO` / `INSERT INTO` 等语句在当前数据库上成功执行

#### Scenario:插件 SQL 转译失败时安装管线快速失败
- **当** 插件某 SQL 文件包含当前方言转译器未覆盖的 MySQL 语法时
- **则** 插件安装 / 升级 / 卸载 / mock 加载管线立即停止后续 SQL 执行
- **且** 错误日志包含失败的插件标识、SQL 资产类型、失败文件名、行号提示与未覆盖关键字
- **且** 管线向上层返回失败状态，便于调用方明确定位待修复的 SQL 文件

### Requirement:插件 SQL 文件必须能被默认方言转译器处理

源码插件与动态插件提交到仓库或发布产物中的 SQL 文件 SHALL 限定在默认方言转译器（当前为 SQLite 方言 `TranslateDDL`）的覆盖范围内。具体而言，插件 SQL 不得使用以下 MySQL 特性：`FULLTEXT INDEX` / `SPATIAL INDEX` / `GENERATED ALWAYS AS` / 分区子句 / `ON DUPLICATE KEY UPDATE` / 未纳入转译器覆盖范围的数据库专用函数（`FIND_IN_SET` / `GROUP_CONCAT` / `IF()` 等）；这些语法本就由项目编码规范禁止，新增本规则强化了"在 SQLite 模式下也必须能跑通"的边界要求。若当前项目 SQL 已真实使用且语义可安全等价映射的函数（如 mock SQL 中的 `CONCAT(...)`），必须先纳入转译器与真实资产执行测试，再允许保留。

#### Scenario:插件 SQL 不使用 FULLTEXT 等未覆盖语法
- **当** 任一源码插件或动态插件提交的 `manifest/sql/*.sql` 文件被默认方言转译器处理时
- **则** 转译过程不返回"未覆盖语法"错误
- **且** 转译结果可在 SQLite 上成功执行

#### Scenario:违规 SQL 在审查与测试阶段被发现
- **当** 插件作者提交包含 `FULLTEXT INDEX` 等未覆盖语法的 SQL 文件时
- **则** 方言转译器单元测试或 E2E 测试在 SQLite 模式下报错
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

### Requirement: 显式插件安装生命周期必须执行依赖编排

显式插件安装请求 SHALL 在目标插件安装副作用前调用依赖解析和安装编排。系统必须先执行框架版本和插件依赖检查，再按自动依赖安装计划安装依赖插件，最后安装目标插件。

#### Scenario: 显式安装先处理依赖
- **WHEN** 管理员请求安装插件 `x`
- **AND** `x` 存在自动安装硬依赖 `a`
- **THEN** 系统在执行 `x` 的安装 SQL 或动态运行时协调前先安装 `a`
- **AND** `a` 安装成功后才继续安装 `x`

#### Scenario: 依赖检查失败时安装无副作用
- **WHEN** 插件 `x` 的依赖检查失败
- **THEN** 系统不得执行 `x` 的安装 SQL
- **AND** 系统不得同步 `x` 的菜单、权限、资源引用或安装状态

### Requirement: 插件安装接口必须返回依赖计划和结果

插件管理安装接口 SHALL 支持调用端获取依赖检查结果、自动安装计划和自动安装结果。安装失败时，错误响应必须以结构化业务错误表达依赖阻断原因，避免只依赖自由文本。

#### Scenario: 安装成功返回自动依赖结果
- **WHEN** 安装目标插件时自动安装了依赖插件
- **THEN** 安装响应包含目标插件 ID
- **AND** 安装响应或后续详情查询包含自动安装成功的依赖插件列表

#### Scenario: 依赖阻断返回结构化错误
- **WHEN** 安装目标插件因为依赖版本不满足失败
- **THEN** HTTP 响应包含稳定业务错误码
- **AND** 响应包含目标插件 ID、依赖插件 ID、当前版本和要求版本范围

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

