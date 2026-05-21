## Why

LinaPro 的开发工具链、发布流程和测试运行时在多个维度上存在入口分散、重复维护和效率瓶颈。`linactl` 作为统一跨平台开发入口，仍保留了 `image-builder`、`build-wasm`、`runtime-i18n` 等独立工具模块；开发环境初始化入口语义不清，缺少轻量环境检查能力；release 发布链路缺少版本一致性门禁、共享测试验证和受控打标入口；Go 单元测试耗时偏高，真实 Wasm 执行和无测试包扫描拉长了 CI 周期；登录后首页运行期 SQL 存在大量重复查询。

## What Changes

### Build Tool Consolidation

- 将镜像构建实现迁移到 `hack/tools/linactl/internal/imagebuilder`，`linactl image` 与 `linactl image.build` 直接调用内部包。
- 将动态插件 Wasm 打包实现迁移到 `hack/tools/linactl/internal/wasmbuilder`，`linactl wasm` 直接调用内部包。
- 将运行时 i18n 治理扫描实现迁移到 `hack/tools/linactl/internal/runtimei18n`，`linactl i18n.check` 直接调用内部包。
- 移除 `hack/tools/image-builder`、`hack/tools/build-wasm` 与 `hack/tools/runtime-i18n` 独立工具模块及其 `go.work` 条目。
- 更新 CI 夹具、测试辅助、E2E 代码和中英文工具文档中的旧路径引用。

### Development Environment

- 新增 `make env.check` / `linactl env.check`，以表格展示本地 Go、Node.js、pnpm、Vite、Playwright 和 PostgreSQL 的名称、当前版本、要求版本、是否满足和备注。
- 新增 `make env.setup` / `linactl env.setup`，承接原 `make dev.setup` 的前端依赖安装与 Playwright Chromium 浏览器安装能力。
- 移除 `make dev.setup` / `linactl dev.setup` 公开入口。
- PostgreSQL 版本检测通过 Go 数据库连接直接查询 `SHOW server_version`，不依赖 `psql` 客户端工具。

### Release Governance

- 新增 `linactl release.tag.check` 作为唯一版本一致性校验入口，读取 `metadata.yaml` 的 `framework.version` 并校验 tag 名称。
- 在 `Release Test and Build` workflow 中增加最早执行的 release tag 版本一致性 job，所有测试和镜像发布 job 依赖该前置 job。
- 新增受控 `Create Release Tag` 手动 workflow，通过 GitHub App installation token 创建匹配 `framework.version` 的 tag。
- 将 release workflow 调整为复用共享测试验证套件，采用 Main CI 的简要测试范围（不含 E2E），镜像发布等待全部验证成功。
- 新增手动 nightly 镜像发布 workflow，直接调用镜像发布 workflow 跳过测试门禁。
- 新增内存态 Docker Compose 演示启动入口和部署测试 Compose 开发容器。

### Test Runtime Optimization

- 保留 Go 单元测试主路径中的 `-race` 并发安全检测，优化重点放在测试设计和重复重型 fixture 治理上。
- 将真实 bundled dynamic Wasm 样例执行收敛为少量 smoke 覆盖，普通单测优先使用 synthetic artifact、fake executor 或轻量测试替身。
- 为插件 runtime、catalog、integration 和生命周期测试引入可复用轻量 fixture。
- 调整 `linactl test.go` 的测试发现与执行策略，区分含 `_test.go` 的包和无测试包，输出可审计的耗时摘要。
- 优化在线会话校验流程，减少每个鉴权请求对 `sys_online_session` 的重复 SQL 往返。
- 优化插件 catalog 运行期读取路径，使用请求级快照减少 `sys_plugin_release` 重复查询。
- 将 E2E 全局默认单测试超时从 60 秒调整为 180 秒，拆分过长 E2E 流程，调整并行 worker 为 1。

## Capabilities

### New Capabilities

- `linactl-build-tool-consolidation`：定义 linactl 统一承载镜像构建、动态插件 Wasm 打包与运行时 i18n 治理扫描能力的开发工具边界。
- `go-unit-test-execution-efficiency`：约束 Go 单元测试的执行效率、测试层级、重型 fixture 复用、真实 Wasm smoke 边界、race 覆盖和可观测耗时报告。
- `login-home-sql-efficiency`：约束登录后首页运行期 SQL 数量、重复查询治理、会话校验往返、插件 catalog 读取复用和验证要求。

### Modified Capabilities

- `release-image-build`：镜像构建命令实现边界从独立工具调整为 linactl 内部组件；发布链路增加版本一致性门禁、共享测试验证和受控打标入口。
- `project-setup`：调整开发环境命令要求，新增环境检查与环境初始化入口，移除旧的 `dev.setup` 入口。
- `e2e-suite-execution-efficiency`：E2E 测试执行效率优化，包括超时调整、并行 worker 控制和长流程拆分。
- `e2e-suite-organization`：完整 E2E 由 nightly 覆盖宿主和官方插件自有 E2E；release 不运行完整 E2E。

## Impact

- 影响 `hack/tools/linactl` 及其内部组件、根 `go.work`、`Makefile`、`make.cmd`。
- 影响 `.github/workflows/` 中 nightly、release、main CI 和 reusable workflow。
- 影响 `hack/deploy/` Docker Compose 演示和测试入口。
- 影响 `apps/lina-core/internal/service/session` 和 `apps/lina-core/internal/service/plugin/internal/catalog` 的运行期读取路径。
- 影响 `hack/tests/` E2E 配置和测试用例。
- 不涉及后端 REST API 语义、数据库 schema 变更、前端页面结构或用户可见运行时文案。
