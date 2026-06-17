# 插件数据服务规范

## Purpose
待定 - 由归档变更 dynamic-plugin-host-service-extension 创建。归档后更新目的。
## Requirements
### Requirement: 动态插件通过宿主数据服务访问宿主确认授权的数据表

系统 SHALL 为动态插件提供基于`table`的宿主数据服务，使插件通过宿主确认授权的数据表进行查询和变更，而不是直接获取宿主数据库连接。

#### Scenario: 插件查询授权数据表

- **WHEN** 插件调用数据服务查询一个已授权的`table`
- **THEN** 宿主校验该表属于当前 release 的最终授权快照
- **AND** 宿主仅允许访问该表允许的方法集合
- **AND** 宿主以统一结构返回查询结果

#### Scenario: 插件执行授权数据表变更

- **WHEN** 插件调用数据服务对一个已授权的`table`执行新增、更新或删除
- **THEN** 宿主仅在该表允许的操作范围内执行
- **AND** 宿主统一返回受影响行数、主键或业务结果摘要

#### Scenario: 数据表声明具有稳定治理边界

- **WHEN** 开发者为动态插件声明数据服务
- **THEN** 清单仅声明`methods`和`resources.tables`
- **AND** 插件运行时只通过表名访问该能力
- **AND** 插件不得获取底层数据库连接或 SQL 片段

### Requirement: 宿主数据服务复用当前用户权限和数据范围

系统 SHALL 在请求型执行上下文中将当前用户身份、角色权限和数据范围应用到插件数据服务调用。

#### Scenario: 登录用户调用插件数据服务

- **WHEN** 已登录用户触发一个动态插件路由并在其中调用数据服务
- **THEN** 宿主将当前用户标识、角色权限和数据范围应用到该数据服务请求
- **AND** 插件不能绕过宿主现有数据权限边界

#### Scenario: 缺少用户上下文的敏感数据调用

- **WHEN** 需要用户上下文的数据服务方法在 Hook、定时任务或匿名请求中被调用
- **THEN** 宿主拒绝执行该调用或仅允许访问显式声明的系统级资源
- **AND** 响应中返回明确的拒绝原因

### Requirement: 原始 SQL 不作为动态插件公开宿主能力

系统 SHALL 不再把原始 SQL 作为动态插件公开宿主能力，数据访问必须通过结构化数据服务和宿主确认授权的数据表完成。

#### Scenario: 插件不能直接申请 raw SQL 能力

- **WHEN** 开发者试图在插件清单或构建产物中声明 raw SQL 类宿主能力
- **THEN** 构建器或宿主装载流程直接拒绝该声明
- **AND** 宿主不向 guest 暴露通用 SQL 执行接口

#### Scenario: 所有数据访问走结构化数据服务

- **WHEN** 开发者为插件声明数据访问能力
- **THEN** 宿主文档、样例和构建校验都要求其声明基于`resources.tables`的数据服务授权
- **AND** 插件只能通过`list`、`get`、`create`、`update`、`delete`和`transaction`等结构化方法访问数据

### Requirement: 宿主数据服务通过 DAO 与 ORM 契约执行数据库操作

系统 SHALL 在宿主内部通过受控的 DAO 对象与 GoFrame `gdb` ORM 组件执行动态插件的数据服务请求，而不是将 raw SQL 或通用 SQL 执行接口暴露给 guest。

#### Scenario: 宿主执行授权数据查询

- **WHEN** 插件调用一个已授权的`table`执行`list`或`get`
- **THEN** 宿主先将该请求解析为对应的表级治理策略与 DAO 操作计划
- **AND** 宿主仅通过受控 DAO / `gdb.Model` 组装查询条件、字段投影、排序和分页
- **AND** guest 不能直接提交 SQL 片段或 JOIN 语句

#### Scenario: 宿主执行授权数据变更

- **WHEN** 插件调用一个已授权的`table`执行`create`、`update`、`delete`或`transaction`
- **THEN** 宿主仅允许对该表授权范围内的结构化数据进行变更
- **AND** 宿主通过受控 DAO / DO 对象完成写入
- **AND** 宿主不得向 guest 暴露通用 SQL 执行接口

### Requirement: 宿主在 DoCommit 层统一拦截数据服务执行与事务提交

系统 SHALL 在 GoFrame `gdb` 的提交链路上建立宿主侧统一拦截点，通过自定义 Driver/DB wrapper 包装 `DoCommit(ctx, gdb.DoCommitInput)`，对动态插件数据服务的最终数据库执行实施权限控制、审计记录和事务治理。

#### Scenario: 宿主拦截一次插件数据服务提交

- **WHEN** 动态插件触发一次数据服务调用并进入宿主数据库执行阶段
- **THEN** 宿主在 `DoCommit` 拦截点拿到当前`pluginId`、`table`、执行方法、事务状态和最终提交参数
- **AND** 宿主可在该阶段执行额外的权限校验、字段审计、风险拦截和日志记录
- **AND** 未通过治理校验的请求不得继续提交到底层数据库驱动

#### Scenario: 宿主治理插件事务边界

- **WHEN** 动态插件通过数据服务发起一个结构化`transaction`
- **THEN** 宿主通过统一的 `DoCommit` / 事务提交链路跟踪事务开启、提交和回滚
- **AND** 宿主可限制事务仅作用于当前授权的数据表集合
- **AND** 宿主能够记录事务级审计摘要，而不要求将原始 SQL 暴露给 guest

### Requirement: 动态插件通过受限 ORM 风格 SDK 访问 data service

系统 SHALL 为动态插件提供`pkg/plugindb`受限 ORM 风格 SDK，作为`data service`的推荐 guest 侧访问入口；该 SDK 必须继续建立在结构化 hostService 协议和`pluginservice/data`语义之上，而不是直接向插件暴露完整`gdb.DB`、`gdb.Model`、宿主 DAO、typed plan 内部结构或 host-side DB wrapper。

#### Scenario: 插件通过 plugindb 发起单表查询

- **WHEN** 插件作者使用`plugindb.Open().Table(...).WhereEq(...).Page(...).All()`等链式 API 访问数据
- **THEN** guest SDK 将该请求转换为结构化、可验证的数据访问请求
- **AND** 宿主继续按当前 release 授权快照、字段白名单和数据范围执行治理
- **AND** 插件不得借此获得 raw SQL、JOIN、任意表达式拼接能力或 host-side query plan 内部类型

#### Scenario: plugindb 作为推荐路径而兼容层暂时保留

- **WHEN** 宿主开始引入`plugindb`guest SDK
- **THEN** 开发文档、样例和 demo 应优先使用`plugindb.Open()`作为数据访问主路径
- **AND** 旧的`pluginbridge.Data()`可作为兼容层短期保留
- **AND** 宿主不得要求插件作者直接拼装底层 hostService envelope

### Requirement: data service 相关枚举语义值使用独立类型和常量统一管理

系统 SHALL 为`plugindb`和`data service`实现中的动作、过滤操作符、排序方向、事务 mutation 类型与访问模式等枚举语义值定义独立 Go 命名类型和常量，禁止在 builder、query plan、执行器和审计逻辑中直接写字符串字面量。

#### Scenario: 查询计划使用强类型枚举

- **WHEN** guest SDK 构造一个数据查询或事务请求
- **THEN** 动作、过滤操作符、排序方向和 mutation 类型必须使用独立命名类型与常量表达
- **AND** 结构化 query plan 在编码到协议层前完成合法性校验
- **AND** 非法枚举值必须在边界处被拒绝

#### Scenario: 宿主执行器禁止依赖枚举字面量

- **WHEN** 宿主解析并执行一个结构化 data 请求
- **THEN** 执行器、审计器和治理逻辑必须基于命名类型与常量进行分派和校验
- **AND** 不得在业务逻辑中直接以`\"eq\"`、`\"in\"`、`\"like\"`、`\"asc\"`等字符串字面量表达这些枚举语义

### Requirement:插件公共数据能力不得暴露宿主数据库模型

系统 SHALL 确保插件公共数据能力和普通 capability 消费接口不暴露`*gdb.Model`、原始 SQL、DAO、DO、Entity 或宿主内部查询对象。需要数据库侧数据范围注入的宿主逻辑 MUST 通过宿主内部窄接口完成，动态插件数据访问 MUST 继续通过受治理 data service 和`hostServices`授权快照完成。

#### Scenario:普通插件消费组织或租户能力

- **WHEN** 普通源码插件或动态插件读取组织、租户或数据相关能力
- **THEN** 返回值使用 DTO、值对象、批量投影或结构化 data service 响应
- **AND** 不返回`*gdb.Model`、DAO、DO、Entity 或可拼接 SQL 的对象

#### Scenario:宿主内部执行数据范围过滤

- **WHEN** 宿主 service 需要在数据库查询阶段注入组织或租户数据范围
- **THEN** 该 service 使用宿主内部`ScopeService`等窄接口
- **AND** 过滤必须在数据库查询阶段完成，不能先查询全量数据再在内存过滤

#### Scenario:动态插件访问数据

- **WHEN** 动态插件需要查询或变更宿主授权的数据表
- **THEN** 它通过`pkg/plugin/capability/guest.Directory.Data()`获取受治理的`pkg/plugin/capability/data` facade 并发起结构化请求
- **AND** 宿主按当前 release 的`hostServices`授权快照、用户上下文和数据权限执行治理
- **AND** guest 不能获取宿主数据库连接、`gdb.Model`或 raw SQL 执行入口

### Requirement:数据能力接口必须优先支持批量投影和有界装配

系统 SHALL 为组织、租户和插件数据访问提供批量投影或结构化分页契约，避免普通插件或宿主列表路径通过循环调用单条详情、逐项 provider 查询或前端瀑布式调用完成装配。

#### Scenario:用户列表装配组织和租户投影

- **WHEN** 宿主或插件需要为用户列表装配部门、岗位或租户标签
- **THEN** 它使用批量投影接口一次性传入当前页用户 ID 集合
- **AND** provider 或宿主实现通过集合化查询、投影查询、缓存或快照完成装配
- **AND** 不得对每个用户循环调用单项详情方法

#### Scenario:动态插件列表查询

- **WHEN** 动态插件通过 data service 查询列表
- **THEN** 宿主在数据库侧完成授权、过滤、排序和分页
- **AND** 只返回当前接口需要的字段或稳定投影
- **AND** 不得先加载大集合到内存后分页或过滤

### Requirement: plugindb 的 Host-Side 实现不得作为公共 API

系统 SHALL 将`plugindb`中的 typed plan、plan codec、host DB wrapper、DoCommit 拦截、执行器、审计上下文和治理校验实现视为宿主内部细节。插件可导入的公共 API MUST 限于 guest 侧受限 DSL、公开枚举、公开 DTO 和必要 facade；host-side 实现 MUST 位于`plugindb/internal/**`、`pluginservice/data`内部边界或宿主`internal/service/plugin/internal/datahost`中。

#### Scenario: 插件不能导入 Host DB Wrapper

- **WHEN** 插件代码尝试 import `pkg/plugindb/internal/host`或旧的公开 host-side DB 包
- **THEN** Go internal 边界或治理扫描阻止该依赖
- **AND** 插件只能通过`plugindb`guest DSL 发起结构化 data service 请求

#### Scenario: Host 执行器通过 Facade 解码计划

- **WHEN** 宿主 data host 需要执行动态插件数据请求
- **THEN** 宿主通过受控 facade 或内部包解码并校验 typed plan
- **AND** 解码后的内部计划不得回传给插件或成为插件接口契约

### Requirement: data service 只能访问当前插件自有表

系统 SHALL 将动态插件`data`服务限定为当前插件自有表的结构化访问能力。可授权表名 MUST 属于当前插件命名空间或宿主明确标记为该插件自有资源的表。宿主核心`sys_*`表和官方能力插件表 MUST NOT 通过`data`服务授权给普通插件。

#### Scenario: 动态插件访问自有表

- **WHEN** 动态插件声明`data`服务访问当前插件自有表
- **THEN** 宿主校验表名归属当前插件命名空间
- **AND** 授权快照仅包含该插件自有表和允许的方法集合

#### Scenario: 动态插件声明宿主核心表

- **WHEN** 动态插件声明`data`服务访问`sys_user`、`sys_role`、`sys_dict_data`、`sys_online_session`、`sys_plugin`或其他宿主核心表
- **THEN** 构建、安装或启用校验失败
- **AND** 运行时授权快照不得包含该表

#### Scenario: 动态插件声明官方能力插件表

- **WHEN** 动态插件声明`data`服务访问组织、租户、`AI`或其他官方能力插件内部表
- **THEN** 宿主拒绝该`data`授权
- **AND** 插件必须改为调用对应领域能力方法

### Requirement: data service 治理必须保留插件自有表数据边界

系统 SHALL 在动态插件访问自有表时继续执行当前用户、租户、数据权限、字段投影、分页、排序、软删除和审计治理。`data`服务收窄为插件自有表后，MUST NOT 放宽现有结构化请求、事务、字段白名单、分页上限和原始 SQL 禁止规则。

#### Scenario: 插件自有表列表查询

- **WHEN** 动态插件通过`data`服务查询当前插件自有表列表
- **THEN** 宿主在数据库侧完成授权、租户过滤、数据权限过滤、排序和分页
- **AND** 只返回当前请求允许的字段投影

#### Scenario: 插件自有表事务变更

- **WHEN** 动态插件通过`data`服务对当前插件自有表执行结构化事务
- **THEN** 宿主限制事务仅作用于当前插件已授权自有表集合
- **AND** 继续记录插件`ID`、表名、方法、事务结果和审计摘要

### Requirement: Datahost 表契约缓存必须绑定插件迁移状态

系统 SHALL 允许 datahost 按插件、表名和插件迁移状态缓存授权表契约。缓存权威源为当前数据库 schema、插件授权快照和插件迁移账本。插件 install、upgrade、rollback 或 uninstall SQL 成功提交后，系统 MUST 按插件失效相关表契约缓存。缓存命中不得跳过 data service 授权、字段白名单、租户过滤、数据权限、分页上限、软删除或审计治理。

#### Scenario: 表契约缓存命中

- **WHEN** 动态插件重复访问同一已授权自有表
- **AND** 该插件迁移状态和授权快照未变化
- **THEN** datahost 可以复用已构建的表契约
- **AND** 不必每次清理并重新读取数据库字段元数据

#### Scenario: 插件升级 DDL 后失效表契约

- **WHEN** 插件`P`升级 SQL 成功提交并改变自有表结构
- **THEN** 系统失效插件`P`相关 datahost 表契约缓存
- **AND** 下一次 data service 调用基于 live schema 重建契约

#### Scenario: 缓存不可用时回源读取

- **WHEN** 表契约缓存未命中、已失效或后端不可用
- **THEN** datahost 回源读取当前数据库 schema 并执行既有校验
- **AND** 不因缓存失败放宽结构化 data service 治理边界

#### Scenario: 授权快照变化后拒绝旧字段

- **WHEN** 插件`P`的 data service 授权快照收缩可访问字段或方法
- **THEN** 后续 data service 调用必须按新授权快照校验
- **AND** 不得因为旧表契约缓存继续允许已收回的字段或方法

