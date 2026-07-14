## REMOVED Requirements

### Requirement: 系统参数管理面不得展示插件命名空间配置

**REMOVED:** 该需求已撤销。系统参数 List/Export 恢复展示全部可见 `sys_config` 行（含 `plugin.*`），不再按插件命名空间过滤。

### Requirement: 插件命名空间系统参数禁止经管理面增删改

**REMOVED:** 该需求已撤销。除既有内置参数（`is_builtin` / 宿主托管键）保护外，`plugin.*` 不再单独禁止管理面增删改。

### Requirement: 系统参数管理面读取不得暴露插件命名空间

**REMOVED:** 该需求已撤销。Get / ByKey 恢复可读取 `plugin.*` 行。
