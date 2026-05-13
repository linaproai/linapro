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

#### Scenario: 控制器构造函数接收服务依赖
- **WHEN** 宿主或源码插件控制器依赖一个或多个服务组件
- **THEN** 控制器构造函数通过参数接收这些服务实例
- **AND** 控制器构造函数不得自行创建缓存敏感或运行期状态敏感服务实例

#### Scenario: 请求路径不得临时创建关键服务
- **WHEN** HTTP handler、中间件、插件回调或 WASM host service 正在处理一次运行期调用
- **THEN** 该路径复用构造时注入的依赖
- **AND** 该路径不得临时调用关键服务 `New()` 创建新的服务图

### Requirement: 不得通过通用容器或全局 service locator 规避显式依赖
系统 SHALL 在不引入通用 DI 容器、全局 service locator 或新增宿主私有组装层的前提下完成依赖管理。启动期已有编排、路由绑定和插件 registrar SHALL 作为显式构造边界。

#### Scenario: 启动编排持有共享实例
- **WHEN** HTTP runtime 构造宿主长生命周期服务
- **THEN** 这些服务由现有启动编排结构持有并向路由绑定、插件注册和 host service 配置传递
- **AND** 业务组件不得通过全局 registry 在运行期查询依赖

#### Scenario: 禁止新增通用 DI 容器
- **WHEN** 开发者为后端依赖管理设计方案
- **THEN** 方案不得引入第三方或自研通用 DI 容器
- **AND** 依赖关系必须保持 Go 类型签名可见

### Requirement: 缓存敏感组件必须共享运行期实例或共享后端
系统 SHALL 对所有持有缓存、派生状态、失效观察状态、session/token 状态、插件运行时状态、运行时配置快照、权限快照或跨实例协调依赖的组件强制共享同一运行期实例或同一共享后端。

#### Scenario: 中间件复用认证和权限服务实例
- **WHEN** 宿主认证、租户和权限中间件被构造
- **THEN** 中间件接收启动期构造的 `auth`、`role`、`tenant`、`config`、`i18n`、`bizctx` 和 `plugin` 依赖
- **AND** 中间件不得自行创建另一套认证、权限、租户或插件服务图

#### Scenario: 插件管理和插件运行时复用同一插件服务
- **WHEN** 插件管理控制器、插件 HTTP route dispatcher、插件 runtime cache、source route registrar 或动态插件 host service 需要插件治理状态
- **THEN** 它们复用启动期同一个插件服务实例或该实例发布的窄接口
- **AND** 不得创建会持有独立 enabled snapshot、route binding、frontend bundle、runtime i18n 或 revision observer 的插件服务实例

#### Scenario: 缓存协调后端在集群模式下保持一致
- **WHEN** `cluster.enabled=true` 且组件需要 cachecoord、kvcache、lock、session hot state 或 token state
- **THEN** 该组件使用启动期注入的 coordination-backed 服务或同一共享 coordination 后端
- **AND** 不得退回到仅当前节点可见的本地默认实例

### Requirement: 源码插件必须通过宿主发布依赖获取宿主能力
系统 SHALL 通过源码插件 registrar 或等价宿主发布上下文向源码插件提供稳定的宿主服务目录。源码插件 Controller 和 Service MUST 通过该目录接收宿主能力适配器，不得在插件生产路径中自行构造宿主内部服务图。

#### Scenario: 源码插件注册 HTTP 路由
- **WHEN** 源码插件在 `http.route.register` 回调中构造控制器和服务
- **THEN** 插件从 registrar 暴露的宿主服务目录获取 `bizctx`、`config`、`i18n`、`notify`、`auth`、`session`、`pluginstate` 等宿主能力
- **AND** 插件业务服务通过显式依赖接收这些能力

#### Scenario: 插件宿主服务适配器由宿主构造
- **WHEN** 源码插件需要使用 `pkg/pluginservice/*` 发布的宿主能力
- **THEN** 适配器实例由宿主运行期构造并通过 registrar 传递
- **AND** 插件生产路径不得调用无参 adapter 构造函数创建孤立宿主服务图

### Requirement: 依赖注入规则必须纳入项目规范和 lina-review 审查
系统 SHALL 将显式依赖注入、隐式构造禁止和缓存敏感共享实例要求写入项目规范与 `lina-review` 审查标准。审查 MUST 覆盖宿主、源码插件、插件 host service、WASM host service 和测试验证。

#### Scenario: 审查后端实现变更
- **WHEN** `lina-review` 审查任何后端 Go 变更
- **THEN** 审查检查新增或修改的组件是否通过显式依赖注入管理运行期依赖
- **AND** 审查标记生产路径中的关键服务隐式构造

#### Scenario: 审查聚合接口依赖结构体
- **WHEN** `lina-review` 审查构造函数或依赖注入设计
- **THEN** 审查标记通过聚合结构体整体传递多个接口型运行期依赖的实现
- **AND** 审查要求将接口型依赖拆分为独立构造函数参数

#### Scenario: 审查缓存敏感组件
- **WHEN** `lina-review` 审查涉及认证、权限、session、插件、配置、i18n、cachecoord、kvcache、lock、notify 或 host service 的变更
- **THEN** 审查要求说明共享实例或共享后端如何保证状态一致
- **AND** 若变更无缓存影响，审查结论必须明确说明

#### Scenario: 静态扫描阻止回归
- **WHEN** 变更完成验证
- **THEN** 项目执行静态扫描或等价治理验证，识别非测试、非启动构造边界中对关键服务 `New()` 的调用
- **AND** 任何新增违规调用必须修复或记录明确豁免理由
