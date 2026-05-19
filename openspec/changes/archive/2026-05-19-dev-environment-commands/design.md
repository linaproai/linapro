## Context

根 `Makefile` 当前只负责把目标转发给 `hack/tools/linactl`。`dev.setup` 已经以 Go 命令文件 `command_dev.setup.go` 实现，内部调用 `frontend.EnsureDeps` 并通过 `pnpm exec playwright install --with-deps chromium` 安装浏览器。新增能力应继续遵循跨平台开发工具规范，将长期逻辑放在 Go 工具中，而不是在 Makefile 或 shell 管道中扩展。

当前仓库有活跃 OpenSpec 变更和未提交的 `linactl` 重构，因此本次设计限定为命令重命名、环境检查和轻量测试，不扩大到重构既有工具目录。

## Goals / Non-Goals

**Goals:**

- 提供 `env.check` 命令，以稳定表格展示 Go、Node.js、pnpm、Vite、Playwright、PostgreSQL 的本机状态。
- 将 `dev.setup` 的功能迁移到 `env.setup`，并从公开命令注册和 Make 目标中移除旧入口。
- 更新 Playwright 缺失提示，让开发者使用新的 `make env.setup` 修复。
- 通过 `go test` 和 `test.scripts` 覆盖命令注册、命令文件命名治理与输出行为。

**Non-Goals:**

- 不自动安装 Go、Node.js、pnpm、PostgreSQL 或系统包管理器依赖。
- 不连接数据库或验证 PostgreSQL 用户、库名、权限、Schema 初始化状态。
- 不修改运行时 i18n 资源、前端页面、后端业务 API、权限或缓存逻辑。

## Decisions

1. `env.check` 只做工具级 smoke 检测，不做服务级验证。
   - 方案：通过 `exec.LookPath` 和版本命令读取当前版本；找不到工具时显示 `missing` 和修复备注。
   - 理由：该命令需要轻量、无副作用，并可在 Windows、Linux、macOS 上运行。
   - 备选：执行完整 `make dev` 或数据库连接验证；被排除，因为会引入长驻进程、数据库凭据和外部服务依赖。

2. 版本要求在 `linactl` 中集中定义。
   - 方案：在 `env.check` 命令文件中定义每个工具的最低或项目要求版本，例如 Go 1.25+、PostgreSQL 14+。
   - 理由：检查逻辑与展示逻辑在同一命令职责内，后续调整要求版本不需要改 Makefile。
   - 备选：从多处配置文件推导版本；被排除，因为当前配置并没有覆盖所有工具，且会让检查行为难以预测。

3. `env.setup` 复用原 `dev.setup` 实现路径。
   - 方案：将命令文件改为 `command_env.setup.go`，函数改为 `runEnvSetup`，继续调用 `frontend.EnsureDeps` 与 Playwright 安装命令。
   - 理由：行为迁移应保持兼容，避免重写已验证的安装编排。
   - 备选：保留 `dev.setup` 作为别名；被排除，因为用户明确要求去掉 `make dev.setup` 指令。

4. 表格输出由 Go 标准库格式化。
   - 方案：使用 `text/tabwriter` 输出名称、当前版本、要求版本、是否满足、备注。
   - 理由：跨平台、无 shell 依赖，适合自动化和终端阅读。

## Risks / Trade-offs

- [Risk] 不同工具版本输出格式存在差异，可能解析不到精确版本。 → 使用正则提取第一个语义化版本；无法解析时标记为未满足并保留原始输出摘要。
- [Risk] Vite 通常是项目本地依赖，未安装前无法直接执行。 → 优先检查 `apps/lina-vben/node_modules/.bin/vite`，缺失时提示运行 `make env.setup`。
- [Risk] Playwright 浏览器缓存路径在不同平台不同。 → 复用既有 Playwright 检查逻辑并将错误提示迁移到新命令。
- [Risk] 移除旧命令会打断仍使用 `make dev.setup` 的本地习惯。 → 在帮助、测试和错误提示中统一暴露 `env.setup`，让失败原因明确。

## Migration Plan

1. 新增 `env.check` 与 `env.setup` 命令注册和 Make 目标。
2. 移除 `dev.setup` 命令注册和 Make 目标，并删除或重命名旧命令文件。
3. 更新 Playwright 缺失提示和测试断言。
4. 运行 `openspec validate`、`go test ./hack/tools/linactl -count=1` 和 `go run ./hack/tools/linactl test.scripts`。

## Open Questions

- 暂无。
