## ADDED Requirements

### Requirement: 插件生命周期缓存失效必须通过单一变化发布入口

系统 SHALL 为插件同步、动态包上传、安装、卸载、启用、禁用、源码升级、动态升级、租户供应策略更新和启动自动启用等插件治理变化提供单一插件变化发布入口。该入口 MUST 复用`plugin-runtime`revision controller，统一失效 runtime 派生缓存、插件管理读模型、frontend bundle、i18n runtime bundle 和 WASM 相关派生状态，不得创建额外仅当前节点可见的缓存域或分散发布路径。

#### Scenario: 生命周期写入后发布变化

- **WHEN** 插件安装、卸载、启用、禁用或状态变更成功写入治理状态
- **THEN** lifecycle 编排调用统一插件变化发布入口
- **AND** 入口发布`plugin-runtime`revision 并记录 reason
- **AND** 插件管理读模型和 runtime 派生缓存均观察同一 revision 失效

#### Scenario: 租户供应策略变化后发布变化

- **WHEN** 平台管理员更新插件新租户供应策略
- **THEN** 系统通过统一插件变化发布入口失效受影响的插件管理和运行时派生缓存
- **AND** 不绕过`plugin-runtime`revision controller 创建独立本地失效路径

#### Scenario: 审查缓存失效入口

- **WHEN** 变更新增插件治理写路径或迁移生命周期编排
- **THEN** 静态治理或审查确认该路径最终调用统一插件变化发布入口
- **AND** 对无缓存影响路径必须记录无影响判断
