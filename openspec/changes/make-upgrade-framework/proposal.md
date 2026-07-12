## Why

二次开发或 fork 后的业务仓库需要持续合并上游框架更新。当前缺少统一的跨平台入口，开发者只能手动查找稳定 tag、`git fetch` 再 `git merge`，步骤易错且在 Windows 上更难复用。需要提供 `make upgrade` / `linactl upgrade`，默认合并最新稳定版本，并支持指定版本或 `main` 分支。

## What Changes

- 新增跨平台命令 `linactl upgrade`，由根 `Makefile` 以 `make upgrade` 薄包装暴露。
- 默认行为：从配置的 Git remote 拉取标签，解析最新稳定版本（`vMAJOR.MINOR.PATCH`，不含预发布后缀），合并到当前本地分支。
- 支持 `v=<version>` 指定稳定版本号（如 `v0.5.0`）合并到当前分支。
- 支持 `v=main`（或其它分支名）将 remote 上对应分支合并到当前本地分支。
- 升级源硬编码为官方仓库 `https://github.com/linaproai/linapro.git`（托管 remote `linapro`），不使用本地 `origin`/fork，不接受 `remote=`。
- 合并时保留本地 `apps/lina-plugins`，不自动更新插件；插件更新仍走 `make plugins.update`。
- 默认拒绝在脏工作区或 detached HEAD 上执行；`force=1` 可跳过脏工作区检查。
- 同步更新 `linactl` 中英文 README 中的命令说明。

## Capabilities

### New Capabilities

- `framework-upgrade`: 通过 `make upgrade` / `linactl upgrade` 将上游框架稳定版本或指定 ref 合并到当前本地分支。

### Modified Capabilities

- （无）现有 baseline 规范不涉及框架源码升级入口。

## Impact

- `hack/tools/linactl/`：新增 `upgrade` 命令实现、注册与测试。
- `hack/makefiles/`：新增 `upgrade` Make 目标包装。
- `hack/tools/linactl/README.md` 与 `README.zh-CN.md`：补充用法说明。
- 依赖本机可用的 `git`；不引入新的第三方 Go 依赖。
- 不影响运行时 HTTP API、数据权限、缓存、插件宿主契约或前端页面。
