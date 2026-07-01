# 将窄接口移动到消费者侧

## Why

`apps/lina-core/internal/service/config/config.go`等生产者组件中存在大量按功能切片拆出来的内嵌接口。部分接口只是为了组成同包`Service`而存在，部分窄接口虽然确实被插件 host config、plugin config、插件能力宿主等消费者使用，但定义在生产者包中会让生产者承担消费者依赖视角，增加阅读和演进成本。

这类设计会造成两个问题：生产者组件对外暴露面看起来比实际稳定契约更复杂；消费者依赖的最小方法集合不在消费方附近，后续新增、删除或替换依赖时难以直接判断 owner、调用场景和降级语义。

## What Changes

- 将生产者包中仅用于自组合的分类接口合并回默认`Service`接口，减少导出接口数量。
- 对确有复杂度收益的跨包窄依赖接口，将定义移动到消费者包，例如插件 host config、plugin config、插件能力宿主和启动装配侧；若目标组件已有`Service`已经能清晰满足消费方需要，则直接复用该`Service`。
- 保留稳定产品契约、运行期共享状态契约和仍有业务必要的公开插件能力契约，例如`session.Store`、`jobmgmt.Scheduler`、`datascope.AccessProvider`和`menu.MenuFilter`。
- 反馈修正中同步收敛`i18n.Service`，删除无业务入口的 i18n 管理诊断 API 和源码插件消息搜索方法；不改变数据库结构、前端页面或运行时翻译包响应结构。

## Capabilities

### Modified

- `service-dependency-injection-governance`：补充消费者侧窄接口 owner 规则，要求有明确复杂度收益的消费方专用接口靠近消费者定义，同时要求可直接复用既有`Service`的消费方避免重复定义本地窄接口。

## Impact

- 影响范围：`apps/lina-core/internal/service`下配置、角色、字典、任务管理、中间件、插件、认证、i18n相关内部服务接口定义和对应启动装配调用点。
- 不影响：数据库迁移、前端页面、运行时用户可见文案和运行时翻译包响应结构。反馈修正会删除无业务入口的 i18n 管理诊断 HTTP API，并收敛`pkg/plugin`中的源码插件 i18n capability 方法。
- 验证方式：`openspec validate move-narrow-interfaces-to-consumers --strict`、静态检索确认被移除的生产者侧接口不再被引用、覆盖变更包的`go test`编译门禁。
