## MODIFIED Requirements

### Requirement: 动态插件复用公共 bridge 组件降低编写复杂度

系统 SHALL 将动态插件 bridge envelope、二进制 codec、guest 侧处理器适配、错误响应辅助与 typed guest controller 适配抽象到 `apps/lina-core/pkg` 公共组件中，避免插件作者在每个动态插件中重复编写底层 ABI、编解码样板与 envelope 到 API DTO 的手工转换逻辑。

#### Scenario: 动态插件控制器直接复用 API 请求与响应 DTO

- **WHEN** 开发者为一个已经在 `backend/api/.../v1` 中声明 DTO 的动态插件路由实现 guest controller
- **THEN** 控制器可以使用 `func(ctx context.Context, req *v1.XxxReq) (res *v1.XxxRes, err error)` 形式声明方法
- **AND** guest 路由分发器根据请求 DTO 类型名匹配运行时 `RequestType`
- **AND** 宿主构建出的动态路由合同继续复用同一份 API DTO 元数据

#### Scenario: typed guest controller 仍可访问桥接上下文并写入自定义响应

- **WHEN** typed guest controller 需要读取 `pluginId`、`requestId`、身份快照、路由元数据，或返回下载流 / 附加响应头 / 自定义状态码
- **THEN** `pkg/pluginbridge` 必须提供从 `context.Context` 读取 bridge envelope 的辅助方法
- **AND** 该组件必须提供写入响应头、原始响应体或自定义状态码的辅助方法
- **AND** 插件作者无需回退到直接声明 `func(*BridgeRequestEnvelopeV1) (*BridgeResponseEnvelopeV1, error)` 才能完成这些场景
