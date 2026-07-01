## Why

`plugin.Service`在`apps/lina-core/internal/service/plugin/plugin.go`中继续承载数十个管理、启动、运行时、集成、能力和租户治理方法，调用方为了少量能力依赖完整插件门面，增加了审查成本和测试替身复杂度。

当前插件服务已经拆出 lifecycle、runtime、integration、frontend、upgrade 等内部组件，根门面可以进一步收敛为按真实消费边界组织的契约，避免宽接口继续扩散到普通控制器、定时任务和单一能力消费方；启动编排和`RuntimeDelegate`这类统一入口边界继续复用`Service`，避免单用途窄接口增加理解成本。

## What Changes

- **BREAKING**：收敛`plugin.Service`方法集合，删除仅作为语义包装或无生产入口的方法。
- **BREAKING**：新增按真实消费场景划分的插件 service 私有 facet 接口，对外只保留统一`Service`入口；消费者按边界使用统一入口或本地私有窄接口。
- **BREAKING**：将插件定时任务声明查询合并为单一查询契约，使用查询条件表达可执行、已安装和插件 ID 过滤。
- **BREAKING**：将插件状态变更入口合并为单一`UpdateStatus`契约，保留动态插件授权确认输入。
- 保持 HTTP API、数据库结构、插件 manifest、动态`hostServices`协议和运行时行为不变。

## Capabilities

### New Capabilities

无。

### Modified Capabilities

- `service-dependency-injection-governance`：补充消费者不得依赖过宽插件根服务接口的要求。
- `plugin-service-layout`：补充插件根 facade 接口分组、方法合并和删除无生产入口方法的要求。

## Impact

- 影响`apps/lina-core/internal/service/plugin`根服务契约、委托方法、生命周期 delegate 和测试替身。
- 影响插件管理控制器、HTTP 启动编排、`apidoc`、`cron`、`jobhandler`等依赖插件服务的构造签名或字段类型。
- 不修改 HTTP 路由、DTO、OpenAPI 元数据、SQL、前端页面、插件公共`pkg/plugin`能力契约或动态插件 wire method。
- 验证以 OpenSpec 严格校验、静态检索和 Go 包测试为主；本变更属于内部 Go 契约治理，无新增 E2E。
