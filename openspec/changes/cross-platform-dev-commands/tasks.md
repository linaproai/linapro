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
