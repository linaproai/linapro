## Why

LinaPro 的宿主需要清晰区分框架核心、默认治理基座与可选业务模块。权限、组织、内容、监控、调度、插件和开发工具曾在默认管理后台中相互缠绕，导致开发者难以判断哪些能力属于`apps/lina-core`长期稳定边界，哪些能力应通过源码插件交付、启停和替换。

本分组同时沉淀了两个宿主边界样板：`demo-control`作为官方源码插件，通过宿主发布的全局 HTTP 中间件 seam 实现演示环境写保护；`pkg/pluginbridge`子组件化拆分验证了插件公共包也必须按宿主边界和职责组织，但完整插件桥接契约由插件框架分组长期承载。

## What Changes

- 确认宿主只长期保留框架核心和管理基座，包括认证、用户、角色、菜单、插件治理、调度、配置、字典、文件、统一事件、稳定顶级菜单和能力 seam。
- 将组织、内容、在线用户、服务监控、操作日志和登录日志等非核心管理模块按官方源码插件交付，并要求宿主在插件缺失时平滑降级。
- 建立稳定的一层菜单挂载点：`dashboard`、`iam`、`org`、`setting`、`content`、`monitor`、`scheduler`、`extension`、`developer`；插件只能按语义挂载到宿主稳定目录，空目录由投影层隐藏。
- 使用能力接口、事件 Hook、HTTP 注册器、全局中间件注册器、Cron 注册器、菜单过滤和权限过滤 seam 替代散落的`if pluginEnabled`分支。
- 交付`demo-control`官方源码插件，以`plugin.autoEnable`是否包含`demo-control`作为唯一演示写保护开关，启用后按 HTTP Method 拦截`/*`写请求并保留登录、登出和受控插件治理白名单。
- 记录`pkg/pluginbridge`拆分对宿主边界的影响：桥接契约、guest SDK、ABI、transport和host service wire必须按职责分层，不能把业务能力语义混入桥接层。

## Capabilities

### New Capabilities

- `core-host-boundary-governance`
- `module-decoupling`
- `demo-control-guard`
- `plugin-http-slot-extension`

### Modified Capabilities

无。

## Impact

- 影响宿主与源码插件的职责划分、菜单挂载、HTTP/Cron/事件扩展 seam、演示环境治理和动态插件桥接代码组织。
- 交叉影响用户、组织、内容、监控、认证、配置、插件生命周期和`pluginbridge-subcomponent-architecture`能力；这些能力的当前契约由对应 owner 分组或`openspec/specs`承载。
- 不把可选业务模块重新内置到`apps/lina-core`，也不引入新的商业插件市场、签名授权或计费分发能力。
