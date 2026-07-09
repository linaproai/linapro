## Requirements

### Requirement: Role 模块必须发布动态路由访问投影契约

系统 SHALL 由 `role` 模块发布面向动态插件运行时的访问投影契约。该契约 MUST 基于 token access snapshot 和 `permission-access` 修订号返回当前 token、用户和租户下的权限字符串、角色名、数据范围、unsupported 数据范围标记和超管标记。插件运行时 MUST 不直接读取角色治理表来构建同类投影。

### Requirement: 访问投影必须携带租户维度且不泄漏内部模型

系统 SHALL 在动态路由访问投影的缓存键、输入校验和输出 DTO 中保留租户维度。该契约返回的投影 MUST 是 `role` 模块自有 DTO 或值对象，不得暴露 DAO、DO、Entity 或可修改共享快照。
