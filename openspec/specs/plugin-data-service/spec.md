# plugin-data-service Specification

## Purpose
TBD - created by archiving change dynamic-plugin-host-service-extension. Update Purpose after archive.
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

系统 SHALL 为动态插件提供`pkg/plugindb`受限 ORM 风格 SDK，作为`data service`的推荐 guest 侧访问入口；该 SDK 必须继续建立在结构化 hostService 协议之上，而不是直接向插件暴露完整`gdb.DB`、`gdb.Model`或宿主 DAO。

#### Scenario: 插件通过 plugindb 发起单表查询

- **WHEN** 插件作者使用`plugindb.Open().Table(...).WhereEq(...).Page(...).All()`等链式 API 访问数据
- **THEN** guest SDK 将该请求转换为结构化、可验证的数据访问请求
- **AND** 宿主继续按当前 release 授权快照、字段白名单和数据范围执行治理
- **AND** 插件不得借此获得 raw SQL、JOIN 或任意表达式拼接能力

#### Scenario: plugindb 作为推荐路径而兼容层暂时保留

- **WHEN** 宿主开始引入`plugindb` guest SDK
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

