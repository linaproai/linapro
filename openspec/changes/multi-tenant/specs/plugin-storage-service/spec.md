## ADDED Requirements

### Requirement: 插件文件存储路径租户前缀
插件通过 host storage service 上传文件 SHALL 自动在路径前缀注入 `tenant=<id>`,具体路径 `/storage/t/<tenant_id>/plugin-<plugin-id>/...`。

#### Scenario: 默认租户隔离
- **WHEN** 插件 P 在租户 A 上下文中上传文件
- **THEN** 文件路径 `/storage/t/A/plugin-P/yyyy/mm/dd/{file_id}`
- **AND** sys_file 表记录 `tenant_id=A`

### Requirement: 跨租户访问需平台 bypass
插件读取文件 SHALL 校验 `bizctx.TenantId` 与 `sys_file.tenant_id` 匹配;不匹配返回 403,平台管理员 bypass。

#### Scenario: 跨租户读取被拒
- **WHEN** 插件在租户 B 上下文中尝试读取租户 A 的文件 fileX
- **THEN** 返回 403 `bizerr.CodePluginFileTenantForbidden`

### Requirement: 平台共享文件路径
插件需要存储跨租户共享文件(如平台 logo)SHALL 通过 `storage.SaveAsPlatform(...)` 显式写入 `/storage/t/0/...`,需具备 `platform:storage:write` 权限。

#### Scenario: 平台共享上传
- **WHEN** 平台管理员上传通用模板
- **THEN** 路径 `/storage/t/0/plugin-<id>/...`
- **AND** 所有租户可读
