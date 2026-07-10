## ADDED Requirements

### Requirement: 升级预览与执行由 lifecycle 编排

系统 SHALL 通过 lifecycle 编排入口提供源码插件升级与动态插件 runtime 升级预览/执行能力。升级实现可保留独立内部包，但 MUST 由 lifecycle 构造、持有并向根 facade 暴露。

#### Scenario: 管理端请求 runtime 升级预览

- **WHEN** 操作者请求某个待升级插件的 runtime 升级预览
- **THEN** 根 facade 调用 lifecycle 拥有的升级能力
- **AND** 预览结果仍包含版本对比、依赖检查、SQL 摘要与 hostServices 差异
- **AND** 该路径不要求根 facade 直接依赖 upgrade 包构造函数

#### Scenario: 管理端确认执行 runtime 升级

- **WHEN** 操作者确认执行 runtime 升级
- **THEN** lifecycle 拥有的升级能力完成锁、状态迁移、SQL/回调与 release 切换编排
- **AND** 动态副作用仍通过 runtime reconcile / lifecycle callback 执行
