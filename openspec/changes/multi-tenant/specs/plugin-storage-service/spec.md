## ADDED Requirements

### Requirement: 插件文件存储路径租户前缀
插件通过 host storage service 上传文件 SHALL 自动在路径前缀注入 `tenant=<id>`,具体路径 `/storage/t/<tenant_id>/plugin-<plugin-id>/...`。

#### Scenario: 默认租户隔离
- **WHEN** 插件 P 在租户 A 上下文中上传文件
- **THEN** 文件路径 `/storage/t/A/plugin-P/yyyy/mm/dd/{file_id}`
- **AND** sys_file 表记录 `tenant_id=A`

### Requirement: 跨租户访问需显式平台接口
插件读取文件 SHALL 校验 `bizctx.TenantId` 与 `sys_file.tenant_id` 匹配;不匹配返回 403。平台管理员管理平台模式只能通过显式 `/platform/*` 只读接口或专用平台 host service 跨租户访问;impersonation 模式不绕过目标租户过滤。

#### Scenario: 跨租户读取被拒
- **WHEN** 插件在租户 B 上下文中尝试读取租户 A 的文件 fileX
- **THEN** 返回 403 `bizerr.CodePluginFileTenantForbidden`

#### Scenario: 平台显式只读访问
- **WHEN** 平台管理员通过平台 host service 读取租户 B 的文件用于排障
- **THEN** service 校验平台权限后返回文件内容
- **AND** 记录跨租户只读审计

### Requirement: 平台共享文件路径
插件需要存储跨租户共享文件(如平台 logo)SHALL 通过 `storage.SaveAsPlatform(...)` 显式写入 `/storage/t/0/...`,需处于平台上下文、有效数据权限为全部数据权限并具备对应 `system:*` 存储写入功能权限。

#### Scenario: 平台共享上传
- **WHEN** 平台管理员上传通用模板
- **THEN** 路径 `/storage/t/0/plugin-<id>/...`
- **AND** 所有租户可读
