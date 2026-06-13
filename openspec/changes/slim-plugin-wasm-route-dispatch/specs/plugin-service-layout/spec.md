## ADDED Requirements

### Requirement: 动态插件 route 分发必须按职责拆分

系统 SHALL 将动态插件 route 分发实现按入口编排、路由匹配、鉴权权限、请求 envelope 和响应写回拆分到职责明确的`route*.go`文件。`apps/lina-core/internal/service/plugin/internal/runtime/route.go`MUST 保持为入口和核心 dispatcher 编排文件，行数 MUST 不超过`400`行。拆分不得改变动态插件公开 API 路径、内部 route contract、访问级别、权限查询、数据权限边界、缓存 freshness 检查或响应 envelope 语义。

#### Scenario: route 入口文件保持瘦身

- **WHEN** 静态测试读取`apps/lina-core/internal/service/plugin/internal/runtime/route.go`
- **THEN** 文件行数不得超过`400`行
- **AND** 文件不得重新承载 JWT 解析、角色菜单查询、path pattern 编译、请求 envelope 构造和响应写回的完整实现

#### Scenario: 动态路由公开契约不变

- **WHEN** 动态插件 route contract 声明`path: /api/v1/reports/{id}`、`method: GET`和受保护访问级别
- **THEN** 宿主仍按既有`/x/{plugin-id}/api/v1/...`公开路径匹配和分发
- **AND** 鉴权、权限菜单查询和 session touch 语义与拆分前保持一致
- **AND** 请求传递给动态插件的 route snapshot、路径参数、header、cookie、query、body 和响应写回语义保持一致

#### Scenario: route 拆分不新增运行期依赖

- **WHEN** 系统构造插件 runtime service
- **THEN** route 分发仍使用既有构造函数传入的 executor、registry、session、auth、权限过滤和缓存 freshness 能力
- **AND** 拆分文件不得通过包级变量、service locator 或临时`New()`创建运行期依赖
