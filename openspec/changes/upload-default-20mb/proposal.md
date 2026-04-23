## Why

当前文件上传大小上限的默认值在宿主初始化 SQL、配置模板和后端静态回退值之间并不一致，现状同时存在 10MB 和 16MB 两套基线。这会导致新环境初始化、未覆盖配置运行以及上传错误提示在不同路径下出现分裂行为。需要把默认上传上限统一提升到 20MB，让 fresh init 与宿主默认运行行为保持一致。

## What Changes

- 将内建运行参数 `sys.upload.maxSize` 的宿主默认值统一调整为 20MB，使配置管理初始化数据与实际默认行为一致。
- 对齐宿主静态上传配置回退值、配置模板默认值以及请求体大小保护链路，消除 10MB/16MB 的默认值分裂。
- 更新受影响的上传大小校验与友好错误提示测试，确保默认 20MB 在文件上传和 transport 限流路径上表现一致。

## Capabilities

### New Capabilities

### Modified Capabilities
- `config-management`: 为内建运行参数 `sys.upload.maxSize` 增加统一的 20MB 默认值约束，并要求宿主默认回退行为与该默认值保持一致。

## Impact

- `apps/lina-core/manifest/sql/007-config-management.sql` 需要调整内建配置种子数据。
- `apps/lina-core/manifest/config/config.template.yaml` 与宿主上传配置回退逻辑需要同步到 20MB。
- `apps/lina-core/internal/service/config/`、`apps/lina-core/internal/service/file/`、`apps/lina-core/internal/service/middleware/` 的上传限制与测试断言会受影响。
- 如存在嵌入或打包的 manifest 产物，需要同步再生成，避免构建结果继续携带旧默认值。
