## Why

hotgo 中已有截图加水印服务，但它是独立进程并依赖旧的 `hg_*` 媒体表。LinaPro 需要把这部分能力迁移为源码插件，复用当前 media 插件的 `media_*` 策略与绑定配置，并纳入统一插件治理、权限、前端页面和测试流程。

## What Changes

- 新增 `water` 源码插件，提供水印截图处理入口、结果回调、任务状态查询和前端测试页面。
- 将 hotgo 的水印策略解析迁移到 LinaPro，旧 `hg_strategy`、`hg_strategy_tenant`、`hg_strategy_device`、`hg_strategy_device_tenant` 对应读取当前 `media_strategy`、`media_strategy_tenant`、`media_strategy_device`、`media_strategy_device_tenant`。
- 保留 hotgo 的 `watermark` YAML 节点语义，支持文字水印、图片水印、透明度、对齐和坐标配置。
- 不迁移 hotgo 的独立 HTTP 服务进程、旧网关调用链、`hg_*` 表或 cgo/FFmpeg 静态库绑定；插件内采用可随 Go 插件构建交付的图片处理实现。
- 新增自动化测试，覆盖策略解析、参数校验、图片处理和前端接口联动。

## Capabilities

### New Capabilities

- `water-plugin`: 水印源码插件的安装、接口、任务处理、media 策略集成和管理页面能力。

### Modified Capabilities

- `media-plugin`: 明确 media 策略表作为水印插件读取的共享策略来源，`hg_*` 旧表名在 LinaPro 中映射为 `media_*` 表。

## Impact

- 新增 `apps/lina-plugins/water/` 源码插件目录、插件清单、后端接口、服务、前端页面和测试。
- 更新 `apps/lina-plugins/go.mod`、`apps/lina-plugins/lina-plugins.go`，让宿主编译并注册 water 源码插件。
- 读取 media 插件已安装的 `media_*` 表，不新增水印策略表；如 media 表不存在或未安装，接口必须返回稳定业务错误。
- 前端页面中文硬编码展示，符合用户对 media/water 业务模块“不使用 i18n”的要求；API 文档资源不新增运行时 i18n。
