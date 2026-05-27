## Why

当前插件公共包已经收敛到`apps/lina-core/pkg/plugin`命名空间，但`capability`、`pluginhost`、`pluginbridge`和宿主内部`internal/service/plugin`之间仍存在职责交叉：普通插件消费面可以看到`*gdb.Model`、`*ghttp.Request`和宿主写入/数据范围注入方法，`pluginhost`仍以`HostServices`别名承载能力目录语义，`pluginbridge`根包仍容易被当作业务能力入口。

本项目没有兼容负担，本变更直接删除历史组件、历史方法和过宽契约，将插件贡献、插件消费宿主能力、动态插件传输协议和宿主内部运行时治理彻底拆开。

本轮需求澄清确认：本迭代按破坏性公共契约收敛执行，不保留旧 facade、旧方法、旧业务 client 或兼容转发层；如果实现阶段发现某个插件确实需要读能力，应新增 DTO 化、批量化、只读能力，而不是恢复数据库模型、HTTP 请求对象、写入接口或宿主内部治理接口。

## What Changes

- **BREAKING**：收窄`pkg/plugin/capability`普通消费面，只允许暴露状态、DTO、批量投影、只读查询和可降级能力；删除普通消费接口中的`*gdb.Model`、`*ghttp.Request`、宿主写入、数据范围注入、启动一致性和自动开通等宿主内部方法。
- **BREAKING**：将`orgcap`和`tenantcap`拆成普通消费接口、provider-facing接口、宿主内部`Scope`/`Membership`/`Provisioning`/`StartupConsistency`等窄接口；`capability.Directory`只返回普通插件消费接口。
- **BREAKING**：删除`contract.ProviderEnv.Services any`，改为每个 capability 定义自己的强类型 provider 构造环境，只注入 provider adapter 真正需要的宿主能力。
- **BREAKING**：删除`pluginhost.HostServices`、`ScopedHostServicesFactory`和`HostServicesForPlugin`等历史别名或重复 scoped 入口；源码插件 registrar 和 callback payload 通过`Services()`暴露`pluginhost.Services`源码插件运行期服务目录。
- **BREAKING**：`pluginbridge`根包不再发布业务能力 client 或宿主能力语义；动态插件业务能力 client 统一从`pkg/plugin/capability/guest`导入，`pluginbridge`只保留ABI、envelope、codec、host call、artifact section、host service wire校验和动态路由分发。
- 本迭代不改变用户可见行为，除非删除旧契约时发现既有投影字段必须同步清理；默认不新增 HTTP API、前端页面、数据库 schema、运行时文案或插件清单字段。
- 新增治理扫描或编译级约束，确保插件代码不能导入宿主内部实现、不能通过普通消费目录访问内部数据权限注入能力，宿主运行时也不能反向依赖插件公共包内部实现，并确保旧入口名、旧方法名、`contract.ProviderEnv.Services`和`pluginbridge`业务 client 没有生产残留。

## Capabilities

### New Capabilities

- 无。本变更修改既有插件宿主、能力目录、bridge和宿主边界要求，不新增独立产品能力。

### Modified Capabilities

- `core-host-boundary-governance`：收紧`apps/lina-core/internal/service/plugin`与`apps/lina-core/pkg/plugin`之间的职责边界和导入方向。
- `plugin-host-service-extension`：收紧源码插件和动态插件消费宿主能力的公共目录，删除历史`HostServices`语义别名和过宽 provider 构造环境。
- `pluginbridge-subcomponent-architecture`：将`pluginbridge`根包从兼容业务能力 facade 收敛为动态插件协议和传输边界。
- `plugin-data-service`：明确普通插件能力目录不得暴露宿主内部数据库模型注入能力，数据访问继续通过受治理 data capability 或宿主内部数据权限接口完成。

## Impact

- 影响后端公共契约：`apps/lina-core/pkg/plugin/capability/**`、`apps/lina-core/pkg/plugin/pluginhost/**`、`apps/lina-core/pkg/plugin/pluginbridge/**`。
- 影响宿主内部运行时：`apps/lina-core/internal/service/plugin/**`、`apps/lina-core/internal/service/pluginhostservices/**`、`apps/lina-core/internal/cmd/**`以及注入`orgcap`、`tenantcap`能力的宿主 service。
- 影响官方源码插件：至少包括`apps/lina-plugins/linapro-org-core/**`、`apps/lina-plugins/linapro-tenant-core/**`以及依赖源码插件 registrar 或 capability provider 的插件。
- 影响动态插件 guest SDK、WASM 构建器和 smoke 测试：`pkg/plugin/capability/guest`、`pkg/plugin/pluginbridge`、`hack/tools/linactl/internal/wasmbuilder/**`和动态插件样例。
- 用户可见行为影响：默认无影响；本迭代仅调整 Go 公共契约、宿主装配、SDK、官方插件和治理扫描，不主动改变页面交互、接口路径、响应字段、SQL 或运行时文案。
- 数据权限影响：收窄公共接口，避免普通插件接触宿主数据范围注入和写入接口；宿主内部仍需显式注入数据权限接缝并保持数据库侧过滤。
- 缓存一致性影响：provider 可用性仍以插件 enabled snapshot 和 runtime revision 为权威；本变更不新增缓存源，但必须验证不引入孤立服务实例或第二套 provider active 状态。
- `i18n`影响：预计不新增用户可见文案或翻译资源；若实现阶段修改错误码、API 文档、插件清单或前端文案，必须另行维护对应资源。
