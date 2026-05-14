## Context

hotgo 的 `cmd/watermark` 以独立 GoFrame HTTP 服务运行，核心流程是：接收截图 base64 -> 入内存队列 -> 根据租户查询策略 -> 解析策略中的 `watermark` 配置 -> 调用水印绘制库 -> 可选回调 URL 和本地保存。LinaPro 当前已经有 media 源码插件并把旧 `hg_*` 表重构为 `media_*` 表，因此 water 插件只需要读取 media 策略配置，不应该再引入旧表或独立服务进程。

## Goals

- 提供一个 `water` 源码插件，随宿主启动注册接口和菜单。
- 接口保持适合外部截图网关调用的形态，同时也提供管理页可直接测试。
- 水印策略来源统一读取 `media_*` 表，解析优先级沿用 media 插件：租户设备 > 设备 > 租户 > 全局。
- 后端必须返回 `bizerr` 业务错误，日志使用 `logger` 并传递 `ctx`。
- 纯 Go 绘制实现必须能在开发和 CI 环境稳定编译运行，不依赖本机 FFmpeg/x264/cgo 静态库。

## Non-Goals

- 不新增水印策略 CRUD，策略内容继续由 media 策略管理维护。
- 不迁移 hotgo 的 Docker/Kubernetes 部署、压测脚本、旧 gateway 调用封装和 `hg_*` 表。
- 不实现分布式任务队列；本次以进程内异步队列提供与 hotgo 等价的单实例能力，跨实例可靠任务编排留给后续独立需求。

## Architecture

### 插件结构

`apps/lina-plugins/water/` 按源码插件结构组织：

- `plugin.yaml` 声明菜单 `/water`、按钮权限和平台全局安装模式。
- `plugin_embed.go` 嵌入清单、前端和 manifest 资源。
- `backend/plugin.go` 注册 HTTP 路由。外部截图处理接口放在已登录权限链路下，管理页也使用统一鉴权/权限中间件。
- `backend/api/water/v1/*` 定义 GoFrame DTO。
- `backend/internal/service/water/*` 实现任务队列、策略解析、图片处理、回调和状态存储。
- `frontend/pages/*` 提供中文管理/测试页面。

### 接口设计

- `POST /water/snaps/{deviceType}/{deviceId}`：提交水印截图任务。请求体包含 `tenant`、`image`、设备/通道信息和可选 `callbackUrl`。响应立即返回 `taskId` 和 `success=true`。
- `GET /water/tasks/{taskId}`：查询任务状态、耗时、错误和输出图片摘要。
- `POST /water/preview`：同步预览接口，管理页用于快速验证当前策略与图片处理效果。

所有读取使用 GET，创建/执行动作使用 POST，符合 REST 语义。外部回调属于任务执行副作用，不作为查询接口。

### 策略解析

water 插件使用自己的只读 DAO 读取 media 表：

- `media_strategy`
- `media_strategy_device`
- `media_strategy_tenant`
- `media_strategy_device_tenant`

解析顺序与 media 插件保持一致：`tenantId + deviceId` 命中租户设备策略优先，其次设备策略、租户策略、全局策略。只读取 `enable=1` 的策略。策略 YAML 中的 `watermark` 节点支持两种形态：

```yaml
watermark:
  enabled: true
  text: 园区安防
  fontSize: 64
  color: "#ffffff"
  posX: 0
  posY: 0
  align: bottomRight
  opacity: 0.15
  base64: ""
```

也兼容 hotgo 旧策略直接把字段放在根级的写法。若没有匹配策略、策略未启用水印或未配置 `watermark`，任务进入 `skipped` 状态并返回原图。

### 图片处理

为了避免 cgo/FFmpeg 依赖，插件使用标准库 `image/*` 和 `golang.org/x/image` 做纯 Go 绘制：

- 支持 PNG/JPEG 输入和 `data:image/...;base64,` 前缀。
- 输出统一为 PNG data URL。
- 文字水印支持多行、颜色、透明度、九宫格对齐和绝对坐标。
- 图片水印支持 base64 输入，按原图尺寸铺底叠加透明度。

默认字体使用 `golang.org/x/image/font/basicfont`，保证构建稳定；如策略配置 font 文件路径且后续引入字体解析，可作为增强项，不阻塞本次迁移。

### 任务与回调

服务启动后懒初始化进程内队列，消费者数量从配置读取，默认 1，最小 1，最大做保护。提交接口只做参数校验和入队，消费者执行水印处理。任务状态保存在进程内带容量上限的状态仓库，用于管理页查询最近任务。

回调使用 `http.Client` 带 `time.Duration` 超时；非 2xx 响应记录为任务错误。所有用户可见错误使用 `bizerr`，内部日志使用 `logger`。

### 前端

页面挂载在内容管理下的 `/water`，中文展示：

- 截图任务提交表单：租户、设备类型、设备 ID、通道、回调 URL、图片 base64。
- 同步预览：展示输入图片和输出图片。
- 任务查询：按任务 ID 查询状态。
- 策略说明区：展示匹配来源和水印配置概要。

页面不使用运行时 i18n；该判断写入任务和审查结论。

### 数据权限与租户

water 插件读取的是 media 全平台共享配置，不按 host 租户隔离，也不新增 `host_tenant_id`。接口入参中的 `tenant` 是媒体业务租户 ID，仅用于 media 策略解析，不等同于宿主租户上下文。由于该插件主要处理外部截图数据和全局共享 media 配置，本次不接入角色数据权限过滤；访问仍受菜单/按钮权限控制，OpenSpec 记录该业务例外。

### 缓存一致性

本次不新增跨实例缓存。策略每次任务处理从数据库读取，权威数据源为 `media_*` 表，因此不存在策略缓存陈旧窗口。任务状态是本实例内存态，仅用于最近任务观测；多实例部署下调用方应查询提交任务所在实例或后续引入共享任务存储。

## Risks

- 纯 Go 绘制与 hotgo FFmpeg 绘制效果不会像素级一致，尤其是字体渲染和图片缩放细节。
- 进程内队列不是分布式可靠队列，实例重启会丢失未完成任务。
- 当前不迁移 cgo 库，因此性能目标以插件内测试与常规截图处理为准，不承诺 hotgo 独立水印集群的压测指标。
