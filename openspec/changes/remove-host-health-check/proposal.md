## 背景

当前主框架内建了匿名`GET /api/v1/health`接口和`health.timeout`配置，接口内部执行数据库探测并返回宿主单机、主节点或从节点模式。这个能力把通用框架宿主与具体业务健康语义绑定在一起，但实际交付中健康检查通常需要覆盖业务依赖、外部服务、队列、模型服务或插件状态，业务应用仍会自定义健康检查接口。

为了降低 LinaPro 作为`面向可持续交付的 AI 原生全栈框架`的宿主复杂度，主框架不再提供内建业务健康检查 API 和配置入口。宿主保留系统信息中的集群与 coordination 诊断，供已认证的运维页面和测试验证使用；业务健康探测由具体业务或交付层自行定义。

## 目标

- 移除主框架内建匿名`GET /api/v1/health`接口、DTO、控制器、路由装配和 API 文档翻译资源。
- 移除`health.timeout`配置段、`config.Service.GetHealth`强类型读取契约和相关缓存、单元测试。
- 删除依赖宿主健康 API 的 E2E 与部署演示文档，调整集群 smoke 和多进程测试改用已认证系统信息诊断。
- 保留 coordination provider 和系统信息中的内部健康诊断，不删除 Redis/cluster 等基础设施自诊断能力。

## 非目标

- 不新增替代健康检查 HTTP API。
- 不改变`GET /api/v1/system/info`的权限边界或响应契约。
- 不删除 Docker Compose 中 PostgreSQL 自身的`healthcheck`，因为它是数据库容器启动依赖，不属于 LinaPro 宿主业务健康接口。
- 不为历史`/api/v1/health`路径提供兼容或重定向。

## 影响分析

- 架构：删除一个过度通用化的宿主业务健康能力，避免`lina-core`固化业务交付语义。
- API 契约：移除匿名`GET /api/v1/health`接口及`health`API 包。
- 后端 Go：修改 HTTP 路由装配、配置服务接口和测试替身。
- i18n：删除宿主`zh-CN`健康 API 文档翻译资源；无运行时 UI 文案新增。
- 测试：删除健康端点 E2E，集群验证改用登录后`GET /api/v1/system/info`。
- 开发工具跨平台：修改既有 Bash smoke 脚本，不新增平台脚本；脚本仍是原有 Redis 集群 smoke 的平台专属测试入口。
- 数据权限：不新增数据读取或写入；系统信息诊断继续走既有认证和权限校验。
- 缓存一致性：无缓存权威数据源、revision、失效或分布式同步路径变更。
