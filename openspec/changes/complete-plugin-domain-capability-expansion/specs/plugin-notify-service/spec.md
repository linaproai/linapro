## ADDED Requirements

### Requirement: 通知消息读取必须返回类型化投影
系统 SHALL 将插件可见通知消息读取收敛到稳定`MessageProjection`或等价类型化 DTO。动态`messages.batch_get` MUST 返回类型化消息投影，不得使用未治理`map[string]any`作为长期协议载荷。

#### Scenario: 批量读取通知消息
- **WHEN** 插件批量读取消息 ID
- **THEN** 系统返回当前 actor 可见的类型化消息投影
- **AND** 不暴露通知内部存储结构或任意扩展字段 map

### Requirement: 通知必须支持按业务来源批量读取
系统 SHALL 提供`Notifications.BatchGetBySource`和动态`messages.by_source.batch_get`，按`SourceType + SourceIDs`集合化读取当前 actor 可见消息。

#### Scenario: 按来源读取消息
- **WHEN** 插件按一组业务来源 ID 读取通知消息
- **THEN** 系统批量返回这些来源下当前 actor 可见的消息投影
- **AND** 不得对每个来源执行一次消息列表查询作为常规实现

### Requirement: 通知必须支持可见性校验
系统 SHALL 提供`Notifications.EnsureVisible`和动态`messages.visible.ensure`，用于写入关联或执行动作前校验消息目标可见。任一不可见、不存在或未授权消息 MUST 整体拒绝。

#### Scenario: 消息可见性校验失败
- **WHEN** 插件校验的消息集合包含不可见消息
- **THEN** 系统返回结构化拒绝错误
- **AND** 不暴露该消息是不存在还是不可见
