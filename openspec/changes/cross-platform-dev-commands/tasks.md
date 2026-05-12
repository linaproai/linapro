## 1. 工具骨架与入口

- [x] 1.1 新增 `hack/tools/linactl` Go CLI 工具目录，补齐 `go.mod`、命令分发框架、错误输出约定和退出码处理
- [x] 1.2 将 `hack/tools/linactl` 加入 `go.work`，并提供英文 `README.md` 与中文 `README.zh-CN.md`
- [x] 1.3 实现 make 风格 `key=value` 参数解析，并覆盖 `confirm`、`rebuild`、`platforms`、`verbose`、`p` 等现有参数
- [x] 1.4 新增根目录 `make.cmd`，只负责透传参数到 `go run ./hack/tools/linactl`

## 2. 低风险目标迁移

- [x] 2.1 实现 `help` 命令，输出跨平台命令列表、参数说明和 Windows/Linux/macOS 入口示例
- [x] 2.2 实现 `prepare-packed-assets` 命令，替代 `hack/scripts/prepare-packed-assets.sh` 的目录清理和 manifest 资源复制逻辑
- [x] 2.3 实现 `wasm` 命令的插件扫描、dynamic 类型识别和指定插件构建参数，替代 `apps/lina-plugins/Makefile` 中的 `find/grep/sort` 编排
- [x] 2.4 更新根目录与子模块 Makefile，将已迁移低风险目标改为调用 `linactl`

## 3. 开发服务命令迁移

- [x] 3.1 实现 `status` 命令，跨平台展示后端与前端端口、PID 文件、日志路径和服务 readiness
- [x] 3.2 实现 `dev` 命令，编排 Wasm 构建、packed assets 准备、后端构建、前端 Vite 启动、日志写入和 HTTP readiness 等待
- [x] 3.3 实现 `stop` 命令，跨平台停止由当前工具启动且可识别的后端与前端进程，并清理 stale PID 文件
- [x] 3.4 更新 `hack/makefiles/dev.mk`，将 `dev`、`stop`、`status` 改为调用 `linactl`

## 4. 构建、数据库与测试命令迁移

- [x] 4.1 实现 `build` 命令，复用现有 `image-builder` 配置解析能力或等价逻辑，编排前端构建、资源嵌入、动态插件构建和多平台后端构建
- [x] 4.2 实现 `image-build` 与 `image` 的跨平台包装，继续调用现有 `hack/tools/image-builder` 与 Docker CLI
- [x] 4.3 实现 `init` 和 `mock` 命令，保留确认参数、防误操作提示和 PostgreSQL 连接失败诊断
- [x] 4.4 实现 `test`、`test-go`、`check-runtime-i18n` 和 `check-runtime-i18n-messages` 的跨平台包装
- [x] 4.5 实现 `cli.install`、`ctrl`、`dao`、`enums`、`service`、`pb`、`pbentity` 等 GoFrame CLI 目标的跨平台包装

## 5. 脚本治理与兼容层收敛

- [x] 5.1 评估并删除或降级不再作为主路径使用的 `.sh` 脚本，保留必要历史入口时标注兼容用途
- [x] 5.2 将根目录 `Makefile` 与 `hack/makefiles/*.mk` 中已迁移目标收敛为薄包装，避免复杂 shell 逻辑重复维护
- [x] 5.3 将 `apps/lina-core/Makefile` 和 `apps/lina-plugins/Makefile` 中已迁移目标收敛为薄包装
- [x] 5.4 明确 `make.cmd`、Makefile 与 `linactl` 的入口优先级和行为一致性

## 6. 文档与验证

- [x] 6.1 更新根目录 `README.md` 和 `README.zh-CN.md`，同步说明跨平台推荐入口、Windows `cmd.exe`、PowerShell 和 Linux/macOS 使用方式
- [x] 6.2 更新 `hack/tools/README.md` 和 `hack/tools/README.zh-CN.md`，登记 `linactl` 工具职责和维护规则
- [x] 6.3 新增 Go 单元测试覆盖参数解析、命令分发、文件复制、插件扫描、帮助输出和错误提示
- [x] 6.4 新增命令级 smoke 验证，覆盖 `make.cmd` 参数透传、Makefile 薄包装一致性和关键命令退出码
- [x] 6.5 更新 `.github/workflows/` 下与构建、测试或工具验证相关的 GitHub Actions，增加 `windows-latest` 基本命令验证
- [x] 6.6 在 Windows GitHub Actions 验证中覆盖 `go run ./hack/tools/linactl help`、`go run ./hack/tools/linactl status` 和至少一个轻量文件或插件工具命令
- [x] 6.7 在 Windows GitHub Actions 验证中覆盖 `cmd.exe` 的 `make.cmd` 或 `make` 用法，以及 PowerShell 的 `.\make.cmd` 用法
- [x] 6.8 运行 `go test` 覆盖新增工具与受影响 Go 工具模块
- [x] 6.9 运行 `openspec validate cross-platform-dev-commands --strict`
- [x] 6.10 执行 i18n 影响确认：跨平台命令工具本身不新增运行时文案；本轮反馈修复已同步多租户插件 manifest i18n 与 apidoc i18n 资源
- [x] 6.11 执行缓存一致性影响确认：本变更不新增或修改业务运行时缓存

## Feedback

- [x] **FB-1**: 允许 `.github/workflows/main-ci.yml` 在所有分支的 push 和 pull request 上触发
- [x] **FB-2**: 完整 E2E 暴露插件菜单国际化断言缺少“插件配置/Configure”按钮
- [x] **FB-3**: 完整 E2E 暴露角色数据权限下拉框源码多租户启用态判断回归
- [x] **FB-4**: 完整 E2E 暴露用户管理排序与自操作断言使用了不稳定表格定位
- [x] **FB-5**: 完整 E2E 暴露文件管理数据权限用例误用依赖组织插件的部门范围
- [x] **FB-6**: 完整 E2E 暴露监控插件用例依赖不稳定列表顺序、展示格式和错误数据范围
- [x] **FB-7**: 完整 E2E 暴露动态插件测试 fixture 缺少当前插件清单多租户字段
- [x] **FB-8**: 完整 E2E 暴露多租户场景断言与当前平台租户菜单契约不一致
- [x] **FB-9**: 完整 E2E 暴露数据权限回归用例误用依赖组织插件的部门范围
- [x] **FB-10**: 完整 E2E 暴露用户管理编辑用例使用批量按钮的不稳定定位
- [x] **FB-11**: 完整 E2E 暴露插件示例数据列帮助图标测试仍使用旧定位器
- [x] **FB-12**: 完整 E2E 暴露字典类型删除用例只依赖瞬时成功提示导致不稳定
- [x] **FB-13**: README Windows `make.cmd` 入口说明应优先展示可省略后缀的 `make` 用法
- [x] **FB-14**: 多租户插件应清理已废弃的租户解析配置与成员管理公开接口权限面
- [x] **FB-15**: `make dev` 和 `make status` 应恢复跨平台状态表格输出
- [x] 2026-05-12: FB-15 验证通过:`gofmt -w hack/tools/linactl/main.go hack/tools/linactl/main_test.go`;`go test ./hack/tools/linactl`;`go run ./hack/tools/linactl status`;`openspec validate cross-platform-dev-commands --strict`;`git diff --check -- hack/tools/linactl/main.go hack/tools/linactl/main_test.go openspec/changes/cross-platform-dev-commands/tasks.md openspec/changes/cross-platform-dev-commands/specs/cross-platform-dev-commands/spec.md`。确认 `status` 与 `dev` 最终状态展示恢复为跨平台 ASCII 表格,包含服务名、运行状态、URL、PID、PID 文件和日志路径;新增单元测试覆盖表格输出关键列与服务行。i18n 影响:本次仅修改开发工具终端输出和 OpenSpec 文档,不新增运行时菜单、按钮、表单、接口消息、插件 manifest 文案、运行时 i18n JSON、manifest i18n 或 apidoc i18n。缓存影响:不新增或修改业务运行时缓存、缓存键或失效路径。
- [x] 2026-05-12: FB-15 `/lina-review` 审查通过:审查范围限定为 `hack/tools/linactl/main.go`、`hack/tools/linactl/main_test.go`、`openspec/changes/cross-platform-dev-commands/specs/cross-platform-dev-commands/spec.md` 与本任务记录。`status` 输出使用 Go 标准库生成 ASCII 表格,不依赖 `lsof`、POSIX shell、Unicode 边框宽度或终端字体;新增表格打印逻辑显式处理 `fmt` 写入错误,单元测试自包含且可单独运行。该变更不涉及 REST API、SQL、前端 UI、运行时 i18n、数据权限、业务缓存或分布式缓存一致性。
- [x] **FB-16**: GitHub Actions 复用 workflow 调用传入未定义的 `run-windows-command-smoke` 输入导致 CI 无法启动
- [x] **FB-17**: `make dev` 应异步启动前后端进程并以当前进程状态表作为最终终端输出
- [x] 2026-05-12: FB-17 验证通过:`gofmt -w hack/tools/linactl/main.go hack/tools/linactl/main_test.go hack/tools/linactl/process_unix.go hack/tools/linactl/process_windows.go`;`go test ./hack/tools/linactl`;`openspec validate cross-platform-dev-commands --strict`;`git diff --check -- hack/tools/linactl/main.go hack/tools/linactl/main_test.go hack/tools/linactl/process_unix.go hack/tools/linactl/process_windows.go openspec/changes/cross-platform-dev-commands/tasks.md openspec/changes/cross-platform-dev-commands/specs/cross-platform-dev-commands/spec.md`;`make dev`;`ps -p $(cat temp/pids/backend.pid) -o pid,ppid,pgid,sess,stat,command`;`ps -p $(cat temp/pids/frontend.pid) -o pid,ppid,pgid,sess,stat,command`。确认 `make dev` 启动后端 PID 77972 与前端 PID 77973,命令结束后两者 PPID 均为 1 并继续运行;终端最后输出与 `make status` 一致的 `LinaPro Framework Status` 表格。修复 readiness 检查对后端根路径 `301 Location: ./` 的处理,不再因 Go HTTP client 跟随循环重定向导致启动超时;新增单元测试覆盖重定向 readiness、异步服务启动、PID 文件写入和最终状态表输出。i18n 影响:本次仅修改开发工具终端输出和 OpenSpec 文档,不新增运行时菜单、按钮、表单、接口消息、插件 manifest 文案、运行时 i18n JSON、manifest i18n 或 apidoc i18n。缓存影响:不新增或修改业务运行时缓存、缓存键或失效路径。
- [x] 2026-05-12: FB-17 `/lina-review` 审查通过:审查范围限定为 `hack/tools/linactl/main.go`、`hack/tools/linactl/main_test.go`、`hack/tools/linactl/process_unix.go`、`hack/tools/linactl/process_windows.go`、`openspec/changes/cross-platform-dev-commands/specs/cross-platform-dev-commands/spec.md` 与本任务记录。`runDev` 仍通过 `cmd.Start()` 后写入 PID 文件并 `Process.Release()` 释放子进程,Unix 使用 `Setsid`、Windows 使用新进程组与 detached process flag,保持服务脱离当前命令生命周期;最终输出复用 `runStatus` 表格。readiness client 禁止自动跟随重定向,与旧 `curl -f` 行为一致并避免后端根路径 301 循环导致超时。新增测试自包含,覆盖重定向 readiness 和 dev 异步启动/最终状态表输出。该变更不涉及 REST API、SQL、前端 UI、运行时 i18n、数据权限、业务缓存或分布式缓存一致性。
- [x] **FB-18**: GitHub Actions Go unit tests 暴露 SQLite 插件生命周期测试清理 `sys_plugin_state` 时缺少显式 WHERE 条件
- [x] 2026-05-12: FB-18 验证通过:`gofmt -w apps/lina-core/internal/service/plugin/internal/testutil/testutil_db.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_db_test.go apps/lina-core/internal/service/plugin/internal/lifecycle/migration_sqlite_test.go`;`go test -race -v ./apps/lina-core/internal/service/plugin/internal/testutil`;`cd apps/lina-core && go test -race -v ./internal/service/plugin/internal/lifecycle`;`cd apps/lina-core && go test -race -v ./...`;CI 同款逐模块 `go list -m -f '{{.Dir}}'` 后逐个模块执行 `go test -race -v ./...`;`openspec validate cross-platform-dev-commands --strict`;`git diff --check -- apps/lina-core/internal/service/plugin/internal/testutil/testutil_db.go apps/lina-core/internal/service/plugin/internal/testutil/testutil_db_test.go apps/lina-core/internal/service/plugin/internal/lifecycle/migration_sqlite_test.go openspec/changes/cross-platform-dev-commands/tasks.md`。确认共享插件测试清理 helper 对插件治理表统一使用显式列名 `plugin_id` 条件,避免 SQLite/GoFrame DELETE 安全检查误判为空条件;SQLite 生命周期测试补齐 `009-plugin-host-call.sql`,使 `sys_plugin_state` 表与清理范围一致;新增 SQLite 单元测试覆盖目标插件行删除且非目标插件行保留。i18n 影响:本次仅修改后端测试 helper、测试初始化和 OpenSpec 任务记录,不新增运行时菜单、按钮、表单、接口消息、插件 manifest 文案、运行时 i18n JSON、manifest i18n 或 apidoc i18n。缓存影响:不新增或修改业务运行时缓存、缓存键、缓存失效路径或分布式缓存一致性逻辑。
- [x] 2026-05-12: FB-18 `/lina-review` 审查通过:审查范围限定为 `apps/lina-core/internal/service/plugin/internal/testutil/testutil_db.go`、`apps/lina-core/internal/service/plugin/internal/testutil/testutil_db_test.go`、`apps/lina-core/internal/service/plugin/internal/lifecycle/migration_sqlite_test.go` 与本任务记录。共享测试清理 helper 使用显式 DAO 列名条件删除插件治理行,避免 SQLite/GoFrame 将 DO 结构体条件误判为空 DELETE;SQLite 生命周期测试初始化覆盖 `sys_plugin_state` 所属的 `009-plugin-host-call.sql`;新增测试自包含保存并恢复 GoFrame 全局数据库配置,验证目标插件清理与非目标插件保留。该变更不涉及 REST API、SQL 交付文件、前端 UI、运行时 i18n、数据权限、业务缓存或分布式缓存一致性。
