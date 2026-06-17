## Requirements

### Requirement: Datahost 表契约缓存必须绑定插件迁移状态

系统 SHALL 允许 datahost 按插件、表名和插件迁移状态缓存授权表契约。缓存权威源为当前数据库 schema、插件授权快照和插件迁移账本。插件 install、upgrade、rollback 或 uninstall SQL 成功提交后，系统 MUST 按插件失效相关表契约缓存。缓存命中不得跳过 data service 授权、字段白名单、租户过滤、数据权限、分页上限、软删除或审计治理。
