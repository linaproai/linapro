## ADDED Requirements

### Requirement: Data 服务必须支持插件自有表批量读取
系统 SHALL 提供动态`data.batch_get`，按主键集合批量读取当前插件自有表中已授权、当前租户和数据权限范围内的记录。实现 MUST 复用现有 data host 表授权、字段投影、租户过滤、软删除和 record store plan 校验。

#### Scenario: 批量读取插件自有表
- **WHEN** 动态插件对已授权插件自有表调用`data.batch_get`
- **THEN** 系统返回可见记录和不透明缺失集合
- **AND** 不得逐主键调用`data.get`作为常规实现

#### Scenario: 批量读取包含不可见记录
- **WHEN** 请求主键集合中包含不可见或不存在记录
- **THEN** 系统不返回这些记录
- **AND** 不区分记录不存在、租户外、数据权限外或未授权

### Requirement: Data exists 和 count 不得进入本批实现
系统 SHALL 将`data.exists`和`data.count`保留为后续独立设计候选。任何实现这些方法的尝试 MUST 先冻结过滤 DSL、字段白名单、租户和数据权限注入、软删除、聚合不泄露策略和测试矩阵。

#### Scenario: 插件声明 data count
- **WHEN** 动态插件声明或调用`data.count`
- **THEN** 本批系统不得发布该方法
- **AND** 调用方必须等待后续独立 OpenSpec 设计
