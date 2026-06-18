## 1. 规范与实现

- [x] 1.1 更新 `host-runtime-operations` 增量规范，要求宿主拥有的运行时资源清理复用 GoFrame `server.gracefulShutdownTimeout`，不再使用顶层 `shutdown.timeout`。
- [x] 1.2 移除自定义 `shutdown` 配置读取契约、缓存和测试覆盖。
- [x] 1.3 更新 HTTP 运行时关闭编排，使资源清理 deadline 从当前 GoFrame Server 的 `GetGracefulShutdownTimeout()` 派生。
- [x] 1.4 更新配置样例，将停机超时配置移入 `server.gracefulShutdownTimeout`，移除顶层 `shutdown` 配置段。

## 2. 验证与审查

- [x] 2.1 运行 `openspec validate reuse-goframe-shutdown-config --strict`。
- [x] 2.2 运行配置服务和 HTTP 启动包测试：`cd apps/lina-core && go test ./internal/service/config ./internal/cmd/internal/httpstartup -count=1`。
- [x] 2.3 运行宿主启动绑定编译门禁：`cd apps/lina-core && go test ./internal/cmd -count=1`。
- [x] 2.4 运行静态检查：`rg -n "GetShutdown|ShutdownConfig|shutdown\\.timeout|^shutdown:" apps/lina-core/internal apps/lina-core/manifest/config hack/deploy` 和 `git diff --check`。
- [x] 2.5 执行 `lina-review` 审查并修复严重问题。

## 实施记录

| 项目 | 记录 |
|------|------|
| 根因 | 主框架自定义 `shutdown.timeout` 被读取后手动写入 GoFrame Server 优雅停止配置，同时宿主资源清理又读取同一自定义配置，形成重复配置入口和重复解析逻辑。 |
| 处理 | 删除 `config.Service.GetShutdown`、`ShutdownConfig` 和 `config_shutdown.go`；`configureHTTPServer` 不再手动设置 GoFrame 优雅停止超时；`shutdownHTTPRuntime` 改为从当前 `*ghttp.Server.GetGracefulShutdownTimeout()` 派生宿主资源清理 deadline。 |
| 配置 | 交付配置和部署样例使用 GoFrame 原生 `server.gracefulShutdownTimeout: 30`；顶层 `shutdown` 配置段已移除；本地忽略的 `manifest/config/config.yaml` 和 `internal/packed/manifest/config/config.template.yaml` 也同步移除旧入口。 |
| DI 来源 | 无新增运行期依赖；`shutdownHTTPRuntime` 从既有 HTTP 启动编排中接收已创建的 GoFrame Server 实例，owner 为 `httpstartup.Run`，传递路径为 `Run -> shutdownHTTPRuntime`。 |
| `i18n` 影响 | 无运行时用户可见文案、菜单、路由、按钮、API 文档源文本、插件清单、语言包或翻译资源影响。 |
| 缓存一致性影响 | 无缓存、快照、订阅、修订号或分布式失效路径变更。 |
| 数据权限影响 | 无数据读写、租户、组织、可见性或权限过滤逻辑变更。 |
| 开发工具跨平台影响 | 未新增或修改脚本、Makefile、CI、`linactl` 或平台相关执行入口；仅同步 YAML 配置模板内容。 |
| 测试策略 | 纯后端启动/配置契约变更，无前端 UI 和用户端工作流变化，不触发 E2E；使用单元测试、宿主启动包编译门禁、OpenSpec 校验和静态扫描覆盖。 |

## 验证记录

| 命令 | 结果 |
|------|------|
| `openspec validate reuse-goframe-shutdown-config --strict` | 通过。 |
| `cd apps/lina-core && go test ./internal/service/config ./internal/cmd/internal/httpstartup -count=1` | 通过。 |
| `cd apps/lina-core && go test ./internal/cmd -count=1` | 通过。 |
| `rg -n "GetShutdown\|ShutdownConfig\|shutdown\\.timeout\|^shutdown:" apps/lina-core/internal apps/lina-core/manifest/config hack/deploy` | 通过，无匹配。 |
| `git diff --check` | 通过。 |

## Lina 审查报告

| 项目 | 结论 |
|------|------|
| 变更 | `reuse-goframe-shutdown-config` |
| 范围 | 反馈级审查：本变更新建 OpenSpec 文件、HTTP 启动/停机编排、配置服务契约、相关测试、交付配置和部署样例。工作区存在其他未提交改动，均与本反馈无关，未纳入本次审查。 |
| 已读取规则 | `AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/backend-go.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/dev-tooling.md`。 |
| 发现的问题 | 未发现阻塞问题；警告 0 个。 |
| 后端 Go | 通过。未新增服务构造函数、数据库访问、接口、缓存敏感服务或业务错误；Go 编译门禁已覆盖变更包和宿主启动绑定包。 |
| OpenSpec/文档 | 通过。新建变更使用中文，增量规范为最小 `MODIFIED` requirement，`openspec validate` 已通过。 |
| 测试 | 通过。新增同包单元测试验证宿主资源清理复用 GoFrame Server 配置；无用户可观察 UI 或端到端流程变化，未触发 E2E 质量审查。 |
| 架构 | 通过。移除重复配置入口，未新增抽象层或污染 `lina-core` 核心领域契约。 |
| 开发工具 | 通过。仅同步配置模板内容，无脚本、工具或平台执行入口变化。 |
