# Tasks

## Summary

- [x] 交付跨平台`linactl`主入口、Windows`make.cmd`、Makefile 薄包装、环境检查/初始化、服务启停、构建、镜像、Wasm、测试、init/mock 和 GoFrame 代码生成命令。
- [x] 整合独立工具模块到`linactl/internal/`，保留公开命令稳定，移除默认路径中对旧`image-builder`、`build-wasm`、`runtime-i18n`和本地`gf`的依赖。
- [x] 扩展`linactl ctrl`和`linactl dao`支持宿主和插件后端目标目录，收敛根、宿主和插件目录 Makefile 代码生成入口，引入共享`hack/makefiles/plugin.codegen.mk`。
- [x] 建立 Agent 多资源桥接、月度 OpenSpec 自动归档、release/nightly 发布治理、受控 tag 创建、source-plugin 升级、安装脚本、demo/test Compose 和`lina-perf-audit`手动审计技能。
- [x] 反馈闭环：升级治理、数据库启动、安装脚本、性能审计、跨平台命令、工具整合、环境命令、镜像发布、release 版本治理、monthly archive、代码生成目标目录、插件 Makefile 硬编码路径和 Redis cluster smoke 脚本命令名称共处理`FB-*`系列问题。
- [x] 验证：覆盖 Go 单元测试、命令 smoke、Windows 基本命令、YAML/shell 语法、Docker/Redis/CI 等价验证、安装脚本 smoke、性能审计 dry-run、OpenSpec 校验、diff 空白检查和`lina-review`审查。
- [x] 治理：本历史分组主要影响开发工具、CI、文档、OpenSpec 和 Agent skill；运行时代码、HTTP API、数据库 schema、权限、数据权限、插件运行时、前端 UI、运行时 i18n 和业务缓存影响由对应 owner 变更承载。本次压缩不修改运行时资源。
- [x] 迁移插件代码生成配置：将`backend/hack/config.yaml`迁移到插件根`hack/config.yaml`，解耦`workDir`和`configDir`；`linactl ctrl`和`linactl dao`只保留`dir=`参数；`plugins.check`扫描新路径并阻断旧路径。
- [x] 迁移插件自定义构建指令：从插件`Makefile`变量收敛到`hack/config.yaml`的`build.commands`；删除`apps/lina-plugins`根`go.mod`/`go.sum`/`lina-plugins.go`，由`linactl`自动生成聚合模块。
- [x] 复用 GoFrame 停机配置：移除自定义`shutdown.timeout`和`config.Service.GetShutdown`，改用`server.gracefulShutdownTimeout`；资源清理 deadline 从`Server.GetGracefulShutdownTimeout()`派生。
- [x] FB-1 至 FB-11：收敛`ctrl`/`dao`参数、移除根 Makefile 兼容、删除插件工作区级重复入口、修正`hack/config.yaml`文档定位、删除根`go.mod`/`go.sum`、修复旧版动态插件 release snapshot 兼容。
- [x] 验证：`go test ./hack/tools/linactl/... -count=1`通过；`make plugins.check`通过；代表性插件`make dao`和`make ctrl`烟测通过；`openspec validate`通过；静态检索确认旧路径无活动配置残留。
- [x] 交付`Go`静态检查门禁：仓库配置与版本锁定、`linactl lint.go`跨平台入口、宿主/插件完整模式、`staticcheck U1000`多目标死代码归并、`env.setup`预装工具、主`CI`/发布验证阻断。
- [x] FB-1~FB-10（静态检查）：收敛首批 linter 噪声、按 wasip1 矩阵归并死代码、统一`staticcheck U1000`、自动安装锁定版本、`env.setup`预装、恢复动态示例 lifecycle 导出契约、builder 配置迁入`hack/config.yaml`并删除`backend/*/*.yaml`兼容路径；验证覆盖`linactl`单测、`make lint`、wasmbuilder/runtime 相关测试与 OpenSpec 校验。
- [x] 治理：静态检查与 builder 配置变更无运行时`i18n`/缓存/数据权限/HTTP API 影响；跨平台入口仍为 Go/`linactl`；无需新增 E2E。
- [x] 删除`linactl`历史兼容面：旧变量/回退路径/`plugins=auto`/snake_case/布尔别名/环境变量覆盖/调试入口；`wasm`仅`dir=`。
- [x] FB-1~FB-6：删除未用 helper、收敛 kebab-case 与`dir=`、修复 CI 静态扫描与 runtime 测试`plugin_dir`误用；验证`linactl`全包测试、lint.go、相关 runtime 测试与 OpenSpec 校验。
