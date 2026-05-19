## Why

当前开发环境初始化入口仍使用 `make dev.setup`，但其职责实际是准备本地环境依赖，与 `make dev` 启动开发服务的语义混在一起。仓库还缺少一个轻量、跨平台的环境检查入口，开发者无法在安装前快速确认 Go、Node.js、pnpm、Vite、Playwright 和 PostgreSQL 是否满足开发要求。

## What Changes

- 新增 `make env.check` / `linactl env.check`，以表格展示本地开发工具的名称、当前版本、要求版本、是否满足和备注。
- 新增 `make env.setup` / `linactl env.setup`，承接原 `make dev.setup` 的前端依赖安装与 Playwright Chromium 浏览器安装能力。
- 移除 `make dev.setup` / `linactl dev.setup` 公开入口，并将相关提示更新为 `make env.setup`。
- 保持根 `Makefile` 与 Windows `make.cmd` 仍作为薄包装入口，真实逻辑继续收敛在 `hack/tools/linactl`。

## Capabilities

### New Capabilities

### Modified Capabilities

- `project-setup`: 调整开发环境命令要求，新增环境检查与环境初始化入口，并移除旧的 `dev.setup` 入口。

## Impact

- 影响根目录 `Makefile` 拆分目标、`hack/tools/linactl` 命令注册和相关工具测试。
- 不改变 LinaPro 后端运行时 API、数据库 Schema、权限模型、数据权限、插件契约、前端页面、运行时 i18n 资源或业务缓存行为。
- 不新增外部依赖；继续调用已有的 `go`、`node`、`pnpm`、`vite`、`playwright`、`psql` 等专业工具。
