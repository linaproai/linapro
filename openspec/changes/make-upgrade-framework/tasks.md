## 1. 命令实现

- [x] 1.1 在 `command.go` 注册 `upgrade` 命令（参数 `v`、`remote`、`force`），handler 使用 `runFrameworkUpgrade` 避免与 `db.upgrade` 的 `runUpgrade` 冲突
- [x] 1.2 新增 `command_upgrade.go`：实现安全门禁、fetch、稳定 tag 解析、版本/分支 ref 解析与 `git merge`
- [x] 1.3 在 `hack/makefiles/release.mk`（或合适 makefile）新增 `make upgrade` 薄包装并透传 `v`/`remote`/`force`

## 2. 测试

- [x] 2.1 为稳定版本规范化、最新稳定 tag 选择、`v` 参数分类编写单元测试
- [x] 2.2 为脏工作区、detached HEAD、默认最新 tag、指定版本与 `v=main` 路径编写可在临时 Git 仓库运行的测试
- [x] 2.3 运行 `go test`（linactl 包）与 `linactl test.scripts` 或等价治理校验，确认 `command_upgrade.go` 命名通过

## 3. 文档

- [x] 3.1 更新 `hack/tools/linactl/README.md` 与 `README.zh-CN.md`，说明 `upgrade` 用法与参数
- [x] 3.2 记录影响分析：i18n / 缓存 / 数据权限无影响；开发工具跨平台与测试策略已覆盖

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate make-upgrade-framework --strict`
- [x] 4.2 完成后调用 `lina-review` 做规范与实现审查

## Feedback

- [x] **FB-1**: 升级源固定为官方仓库 `https://github.com/linaproai/linapro.git`（托管 remote `linapro`），移除 `remote=`，并补充忽略 origin/fork 的测试与文档
- [x] **FB-2**: `upgrade` 合并官方框架时保留本地 `apps/lina-plugins`，不自动更新插件；文档与测试覆盖
