## Why

当前普通文件上传的相对存储路径包含 `t/<tenantId>/...` 前缀，其中 `t` 只是 tenant 的缩写，对调用方和运维排查不够直观。用户希望去掉这个额外目录层，同时保留按租户 ID 分区的物理组织方式。

## What Changes

- 新上传文件的相对存储路径从 `t/<tenantId>/<yyyy>/<MM>/<filename>` 调整为 `<tenantId>/<yyyy>/<MM>/<filename>`。
- 保留 `sys_file.path` 已记录的历史路径，不执行历史文件迁移。
- 下载和 URL 访问继续以数据库记录的相对路径为准，因此旧 `t/...` 文件与新路径文件可以共存访问。
- 更新文件上传路径示例、单元测试和实现任务记录。

## Capabilities

### New Capabilities

- `file-upload-storage-path`: 约束普通文件上传的租户分区相对路径、历史路径兼容和验证要求。

### Modified Capabilities

- 无。

## Impact

- 影响 `apps/lina-core/internal/service/file` 的本地存储路径生成逻辑。
- 影响文件上传相关单元测试和 API 文档示例。
- 不改变公开上传、下载、访问 API 路径，不改变数据库 schema，不迁移现有文件。
- 不新增用户可见文案，不影响前端运行时语言包或插件 manifest/i18n。
