## 1. 规范与实现

- [x] 1.1 更新`host-runtime-operations`增量规范，移除宿主内建匿名健康探测端点要求，并明确业务健康由业务应用自行提供。
- [x] 1.2 删除健康 API DTO、控制器、公共路由装配、权限审计豁免和 apidoc 翻译资源。
- [x] 1.3 删除`health.timeout`配置段、`config.Service.GetHealth`契约、静态配置缓存和相关单元测试。
- [x] 1.4 更新部署演示配置、Compose 和 README，移除对`/api/v1/health`的依赖。
- [x] 1.5 删除健康端点 E2E，并将集群 smoke 与多进程集成测试改为通过已认证`GET /api/v1/system/info`读取 coordination 诊断。

## 2. 验证与审查

- [x] 2.1 运行`openspec validate remove-host-health-check --strict`。
- [x] 2.2 运行配置服务、HTTP 启动绑定、权限审计、集群包和宿主启动包 Go 测试门禁。
- [x] 2.3 运行 E2E TypeScript 校验或相关测试资产校验。
- [x] 2.4 运行静态检索确认非归档实现不再引用宿主健康 API、`GetHealth`、`HealthConfig`或`health.timeout`。
- [x] 2.5 执行`lina-review`审查并修复严重问题。

## 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | 主框架把数据库连通性、集群主从模式和业务可用性统一抽象成匿名健康 API，并提供`health.timeout`配置；但业务健康通常取决于业务依赖和插件状态，后续交付仍会自定义健康接口，导致宿主能力复杂且复用价值低。 |
| 处理 | 删除`apps/lina-core/api/health`、`apps/lina-core/internal/controller/health`、`config.Service.GetHealth`、`HealthConfig`、静态健康配置缓存、健康 API apidoc 翻译、公共路由绑定和权限审计豁免；交付配置和部署说明不再声明或调用`/api/v1/health`。 |
| 测试调整 | 删除`hack/tests/e2e/auth/TC008-health-endpoint-anonymous-access.ts`，将过期会话测试前移为`TC008`；Redis cluster smoke 和多进程集成测试改为登录后读取`GET /api/v1/system/info`中的 coordination 诊断。 |
| 规范 | 更新`host-runtime-operations`基线和增量规范，明确宿主不得内建业务健康检查接口；同步将系统信息、拓扑和 Redis coordination 规范中的旧“健康 API”表述收敛为系统信息诊断。 |
| DI 来源 | 删除健康控制器依赖后无新增运行期依赖；HTTP 路由装配少创建一个控制器实例；系统信息诊断继续复用启动期已有`sysInfoSvc`、`authSvc`、`middlewareSvc`和权限链路。 |
| `i18n`影响 | 删除宿主`zh-CN`健康 API apidoc 翻译资源并重新生成 packed manifest；未新增运行时 UI 文案、菜单、按钮、表单、语言包或翻译缓存逻辑。 |
| 缓存一致性影响 | 无缓存权威数据源、revision、失效触发点、跨实例同步、TTL 或故障降级变更；删除的健康 API 不参与缓存。 |
| 数据权限影响 | 无新增数据读写、列表、详情、导出、下载、聚合或执行动作；系统信息诊断继续要求认证和`about:system:list`权限，不新增匿名数据可见性。 |
| 开发工具跨平台影响 | 修改既有`hack/tests/scripts/run-redis-cluster-smoke.sh`的探测逻辑，未新增脚本或默认开发入口；该脚本仍是原有 Redis 集群 smoke 的 Bash 专属测试入口。交付 Compose 仅移除 LinaPro 容器自身健康 API 探针，保留 PostgreSQL service healthcheck。 |
| 外部规则加载 | 已读取`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/documentation.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`goframe-v2`、`lina-e2e`和`karpathy-guidelines`。确认无业务数据权限、缓存一致性和插件目录结构实现影响。 |

## 验证记录

| 命令 | 结果 |
|------|------|
| `openspec validate remove-host-health-check --strict` | 通过。 |
| `cd apps/lina-core && go test ./internal/service/config ./internal/cmd/internal/httpstartup ./internal/service/middleware ./internal/service/role ./internal/service/cluster ./internal/service/apidoc ./internal/cmd -count=1` | 通过。 |
| `cd apps/lina-core && go test ./internal/packed -count=1` | 通过。 |
| `cd hack/tests && pnpm test:validate && pnpm exec tsc --noEmit -p tsconfig.json` | 通过，校验`250`个 E2E 文件。 |
| `rg --no-ignore -n "GetHealth\|HealthConfig\|health\\.timeout\|core-api-health\|api/health\|internal/controller/health" apps/lina-core hack ...` | 通过，无匹配。 |
| `rg --no-ignore -n '/api/v1/health([^[:alnum:]]|$)' apps/lina-core hack ...` | 通过，无匹配。 |
| `rg --no-ignore -n "health:" apps/lina-core/manifest/config hack/deploy apps/lina-core/internal/packed/manifest/config ...` | 通过，无匹配。 |
| `find apps/lina-core/api/health apps/lina-core/internal/controller/health -maxdepth 2 -type f -print` | 通过，无文件。 |
| `bash -n hack/tests/scripts/run-redis-cluster-smoke.sh` | 通过。 |
| `git diff --check` | 通过。 |

## Lina 审查报告

**变更：** `remove-host-health-check`
**范围：** 反馈级 OpenSpec 新变更及本次健康检查删除相关实现。
**审查文件数：** 38 个版本控制候选文件，另检查本地默认配置`apps/lina-core/manifest/config/config.yaml`。
**范围来源：** `git status --short`、`git ls-files --others --exclude-standard`、`git -C apps/lina-plugins status --short`、本变更任务上下文和静态检索结果。
**已读取规则文件：** `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/api-contract.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/i18n.md`、`.agents/rules/dev-tooling.md`、`.agents/rules/documentation.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、`.agents/instructions/markdown-format.instructions.md`。

### 发现的问题

- 未发现剩余阻塞问题。
- 审查过程中发现本地默认配置`apps/lina-core/manifest/config/config.yaml`仍保留`health:`段，已删除并重新运行`make pack.assets`、静态检索和门禁验证；该文件未纳入 Git 跟踪，但作为本地运行配置残留一并清理。

### 规则域结论

- OpenSpec 和文档：通过。变更文档为中文，`proposal.md`、`tasks.md`和增量规范已通过`openspec validate remove-host-health-check --strict`。
- 架构边界：通过。删除宿主内建业务健康检查契约，业务健康语义回到业务应用或插件边界，降低`lina-core`通用宿主耦合。
- API 契约：通过。已删除匿名`GET /api/v1/health`DTO、控制器和路由绑定；未新增兼容路径或重定向。
- 后端 Go：通过。删除健康控制器、配置服务健康读取契约和测试替身方法；无新增运行期依赖，路由装配少创建一个控制器实例。
- 测试和 E2E：通过。删除健康端点匿名访问 E2E 后，认证模块文件编号保持`TC001`到`TC008`连续；替代诊断验证由 Go 多进程测试和 Redis smoke 脚本覆盖。
- `i18n`：通过。删除宿主健康 API 的`zh-CN`apidoc 翻译资源并重新打包 manifest；未新增运行时 UI 文案或翻译缓存逻辑。
- 缓存一致性：无影响。删除的健康 API 不参与缓存、revision、失效、TTL 或跨实例同步。
- 数据权限：无影响。未新增数据读写、列表、详情、导出、下载、聚合或执行动作；`system/info`诊断继续走认证和权限链路。
- 开发工具跨平台：通过但保留既有边界。修改的是原有 Redis 集群 Bash smoke 脚本，未新增默认开发入口；已运行`bash -n`语法烟测。
- 插件目录：未命中。本次未修改`apps/lina-plugins/<plugin-id>`；已展开`apps/lina-plugins`子仓库状态，确认其变更与本次反馈文件无交集。

### 验证证据

- `openspec validate remove-host-health-check --strict`：通过。
- `cd apps/lina-core && go test ./internal/service/config ./internal/cmd/internal/httpstartup ./internal/service/middleware ./internal/service/role ./internal/service/cluster ./internal/service/apidoc ./internal/cmd ./internal/packed -count=1`：通过。
- `cd hack/tests && pnpm test:validate`：通过，校验`250`个 E2E 文件。
- `cd hack/tests && pnpm exec tsc --noEmit -p tsconfig.json`：通过。
- 健康 API、`GetHealth`、`HealthConfig`、`health.timeout`和`health:`静态检索：通过，非 OpenSpec 说明文档范围无残留。
- `bash -n hack/tests/scripts/run-redis-cluster-smoke.sh`：通过。
- `git diff --check`：通过。

### 摘要

- 严重：0
- 警告：0

### 建议操作

1. 本变更可以进入用户确认阶段；当前工作区存在大量其他活跃改动和`apps/lina-plugins`子仓库改动，提交或归档前应继续按变更范围隔离审查。
