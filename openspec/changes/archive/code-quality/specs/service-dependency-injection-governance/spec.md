## ADDED Requirements

### Requirement: 后端组件必须通过显式依赖注入管理运行期依赖
系统 SHALL 要求宿主和源码插件的生产后端组件通过构造函数参数逐项显式接收运行期依赖。Controller、Middleware、Service、插件宿主服务适配器和 WASM host service MUST 不在业务构造函数、请求处理、插件回调或 host service 调用路径中隐式创建关键服务依赖，MUST NOT 通过聚合依赖结构体整体传递多个接口型运行期依赖。

#### Scenario: 服务构造函数逐项接收接口依赖
- **WHEN** 宿主服务需要访问配置、插件、权限、租户、会话、缓存协调或 i18n 等运行期依赖
- **THEN** 构造函数在签名中逐项接收这些接口型依赖
- **AND** 构造函数不得在依赖缺失时静默调用其他关键服务的 `New()` 补齐依赖

#### Scenario: 禁止聚合结构体隐藏接口依赖
- **WHEN** 后端组件需要接收多个接口对象、服务对象或宿主能力适配器
- **THEN** 这些接口型依赖必须拆分为独立构造函数参数
- **AND** 不得通过 `Dependencies`、`Deps`、`Options` 或等价聚合结构体整体传递
- **AND** 依赖新增、删除或替换必须能通过 Go 编译错误暴露所有未同步调用点

### Requirement: 缓存敏感组件必须共享运行期实例或共享后端
系统 SHALL 对所有持有缓存、派生状态、失效观察状态、session/token 状态、插件运行时状态、运行时配置快照、权限快照或跨实例协调依赖的组件强制共享同一运行期实例或同一共享后端。

### Requirement: 源码插件必须通过宿主发布依赖获取宿主能力
系统 SHALL 通过源码插件 registrar 或等价宿主发布上下文向源码插件提供稳定的宿主服务目录。源码插件 Controller 和 Service MUST 通过该目录接收宿主能力适配器。

### Requirement: 初始化与注册 API 必须返回错误给调用方决策
系统 SHALL 要求宿主和源码插件的运行时初始化、源码插件注册、registrar、回调注册、路由注册、Cron 注册和中间件注册 API 在依赖缺失或校验失败时返回 `error`。这些 API MUST NOT 在内部直接 `panic` 处理可预期错误。

### Requirement: 依赖注入规则必须纳入项目规范和 lina-review 审查
系统 SHALL 将显式依赖注入、隐式构造禁止、初始化/注册错误返回和缓存敏感共享实例要求写入项目规范与 lina-review 审查标准。
