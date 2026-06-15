## ADDED Requirements

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
