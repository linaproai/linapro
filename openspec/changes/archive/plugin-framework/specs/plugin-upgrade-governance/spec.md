## Requirements

### Requirement: 插件升级必须由统一升级编排组件执行

系统 SHALL 将源码插件升级和动态插件升级的 preview、execute、失败记账、release 提升、治理资源同步和缓存发布纳入同一升级编排模型。source 与 dynamic 插件可以保留不同执行策略，但共享依赖校验、反向依赖保护、失败诊断、治理守卫边界和缓存发布骨架。

### Requirement: 插件升级失败诊断必须使用单一账本约定

系统 SHALL 使用一套 `sys_plugin_migration` 升级失败诊断约定表达 source 与 dynamic 插件升级失败。失败 phase、error code、message key、fallback、目标 release 和原始错误信息 MUST 由统一升级模型归一化。

### Requirement: 插件升级治理守卫必须只在公开入口执行一次

系统 SHALL 在公开插件升级入口执行平台治理守卫，并禁止统一升级组件通过再入公开插件服务方法重复执行守卫或重复发布缓存。

### Requirement: 插件升级缓存发布必须复用插件变化发布入口

系统 SHALL 在 source 和 dynamic 插件升级成功、失败或失败诊断变化后，通过统一插件变化发布入口发布作用域化变化。发布必须包含插件 ID、插件类型和 reason，并继续复用 `plugin-runtime` revision controller。
