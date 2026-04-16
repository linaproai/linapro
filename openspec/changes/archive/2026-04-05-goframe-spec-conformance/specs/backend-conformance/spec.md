## ADDED Requirements

### Requirement: 控制器与服务层实现约束
后端生产代码 SHALL 遵循仓库定义的 GoFrame v2 分层约束，控制器依赖通过构造函数注入，服务层组件按约定目录与命名组织。

#### Scenario: 控制器依赖初始化
- **WHEN** 控制器依赖一个或多个服务组件
- **THEN** 这些依赖在对应的 `_new.go` 构造函数中完成初始化
- **AND** 接口方法内部不再临时调用 `service.New()` 创建依赖

#### Scenario: 服务组件拆分
- **WHEN** 某个服务组件存在多个职责子模块
- **THEN** 代码按组件前缀和子模块后缀拆分到独立文件
- **AND** 不使用与组件名无关的裸文件名承载子模块逻辑

### Requirement: ORM 与软删除一致性
后端生产代码 SHALL 使用 GoFrame 推荐的 ORM 方式访问数据库，并遵循自动软删除与时间维护约定。

#### Scenario: 查询软删除表
- **WHEN** 代码查询包含 `deleted_at` 字段的表
- **THEN** 查询逻辑依赖 GoFrame 自动软删除过滤
- **AND** 生产代码不手写 `WhereNull(deleted_at)` 或等价 SQL 条件

#### Scenario: 更新和写入数据
- **WHEN** 代码执行数据库写入、更新或关联关系维护
- **THEN** 生产代码使用 DO 对象传递 `Data`
- **AND** 不手工维护 `created_at`、`updated_at`、`deleted_at` 这些由框架自动维护的字段

### Requirement: 公开符号文档完整
后端公开方法、结构体和关键公开字段 SHALL 具有符合 Go 文档习惯的注释，便于生成文档和长期维护。

#### Scenario: 新增或整改公开符号
- **WHEN** 代码中存在导出方法、导出结构体或关键导出字段
- **THEN** 其声明前包含紧邻且语义明确的注释
- **AND** 注释可被 Go doc 正常识别，而不是仅保留分隔说明或脱离声明的备注
