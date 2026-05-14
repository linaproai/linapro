## 1. 配置模型与校验

- [x] 1.1 扩展 `hack/tools/linactl` 的 `rootConfig`，新增 `plugins.sources` 配置结构，支持 `repo`、`root`、`ref` 和字符串数组 `items`
- [x] 1.2 在 `hack/config.yaml` 增加示例插件配置，覆盖官方来源和自定义来源的推荐写法
- [x] 1.3 实现配置校验：source 名称、repo、root、ref、items 必填检查，`items` 仅允许字符串数组
- [x] 1.4 实现插件 ID 全局唯一校验，重复插件必须在写入前失败并列出冲突 source
- [x] 1.5 实现路径安全校验，拒绝空 root、绝对路径、`..`、路径分隔符注入和 Windows drive path
- [x] 1.6 补充配置解析和校验单元测试，覆盖有效配置、重复插件、非法 root、非法 items 类型和缺失字段

## 2. 工作区去 submodule 化

- [x] 2.1 新增 `linactl plugins.init` 命令并在 `Makefile` / `make.cmd` 中提供 `make plugins.init` 包装入口
- [x] 2.2 实现 `apps/lina-plugins` 工作区状态检测，区分缺失、普通目录、submodule/gitlink、嵌套 Git 仓库和无效路径
- [x] 2.3 实现 `.gitmodules` section 删除逻辑：只删除 `apps/lina-plugins`，仅在没有其他 section 时删除整个 `.gitmodules`
- [x] 2.4 实现父仓库 submodule 配置清理，包括 `.git/config` 对应 section 和 `.git/modules/apps/lina-plugins` 元数据
- [x] 2.5 实现 gitlink 到普通目录的转换，保留 `apps/lina-plugins` 下已有插件文件内容
- [x] 2.6 补充单元测试或隔离仓库集成测试，覆盖已初始化 submodule、未初始化 submodule、普通目录、缺失目录和多 submodule `.gitmodules`

## 3. 插件安装与更新

- [x] 3.1 新增 `linactl plugins.install` / `make plugins.install`，按 `hack/config.yaml` 安装配置插件
- [x] 3.2 新增 `linactl plugins.update` / `make plugins.update`，按 `hack/config.yaml` 更新配置插件
- [x] 3.3 实现来源仓库临时 checkout，解析 `ref` 到 commit，并从 `<root>/<plugin-id>` 复制插件目录到 `apps/lina-plugins/<plugin-id>`
- [x] 3.4 确保安装/更新不会把来源仓库 `.git` 元数据写入插件目录
- [x] 3.5 实现目标目录保护：install 遇到已存在目录默认失败，update 遇到本地 dirty 默认失败，`force=1` 才允许覆盖
- [x] 3.6 支持命令子集筛选参数，例如 `p=<plugin-id>` 和 `source=<source-name>`，但 repo/root/ref 仍只来自 `hack/config.yaml`
- [x] 3.7 补充安装与更新测试，覆盖成功安装、缺失远端 plugin.yaml、目标已存在、dirty 阻断、force 覆盖和子集筛选

## 4. 锁定状态与状态检查

- [x] 4.1 设计并实现 `apps/lina-plugins/.linapro-plugins.lock.yaml`，记录插件 ID、source、repo、root、ref、resolved commit、manifest version 和内容摘要
- [x] 4.2 在 install/update 成功后写入或刷新锁定状态，失败时不得写入半成品锁文件
- [x] 4.3 新增 `linactl plugins.status` / `make plugins.status`，只读输出工作区类型、配置插件、本地插件、锁定状态和远端更新状态
- [x] 4.4 实现本地 dirty 检测，优先通过父仓库 `git status -- apps/lina-plugins/<plugin-id>` 判断用户可提交改动
- [x] 4.5 实现远端更新检测，远端不可达时输出 unknown，不修改本地状态
- [x] 4.6 补充状态命令测试，覆盖普通目录、submodule 提示、远端不可达、未配置本地插件、缺失本地插件和 orphaned lock entry

## 5. 文档与命令帮助

- [x] 5.1 更新根 `README.md` 和 `README.zh-CN.md`，说明官方仓库 submodule 与用户项目普通插件目录的区别
- [x] 5.2 更新 `hack/tools/linactl/README.md` 和 `README.zh-CN.md`，补充 `plugins.init`、`plugins.install`、`plugins.update`、`plugins.status` 用法
- [x] 5.3 更新命令帮助输出，确保 Windows、Linux、macOS 用户可通过相同 make/linactl 入口执行
- [x] 5.4 更新必要的开发规范说明，强调 `plugins.sources.items` 仅支持字符串数组且插件目录固定为 `apps/lina-plugins`

## 6. 验证与审查

- [x] 6.1 运行 `cd hack/tools/linactl && go test ./... -count=1`
- [x] 6.2 运行 `go run ./hack/tools/linactl test.scripts`
- [x] 6.3 运行 `openspec validate plugin-workspace-management --strict`
- [x] 6.4 运行 `git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`
- [x] 6.5 记录 i18n 影响结论：本变更仅涉及开发工具命令和文档，不新增前端运行时、接口文档或插件 manifest i18n；若实现中新增运行时文案则同步维护翻译
- [x] 6.6 记录缓存一致性结论：本变更不新增运行时缓存、缓存键或跨实例失效逻辑
- [x] 6.7 记录数据权限结论：本变更不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界
- [x] 6.8 记录 RESTful API 结论：本变更不新增后端 REST API
- [x] 6.9 完成实现后调用 `lina-review`，重点审查去 submodule 化安全性、路径安全、dirty 保护、跨平台实现和测试覆盖

## Verification Notes

- i18n: 本变更仅新增开发工具命令输出、README 文档和 `hack/config.yaml` 示例，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- Cache: 本变更不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑；插件目录变化仍由既有构建和插件同步流程处理。
- Data permission: 本变更不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- RESTful API: 本变更不新增后端 REST API。
- Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`。
- Review: 已完成 `lina-review` 审查；修复了新增 helper 未使用、新增 Go 测试函数缺少注释、更新命令未检查已提交锁摘要漂移、以及 submodule section 删除可能误删后续普通 Git config section 的问题。确认 `plugins.init/install/update/status` 均由 `linactl` Go 工具承载，`Makefile` 仅作为包装入口；`plugins.status` 在当前真实仓库中识别 `apps/lina-plugins` 为 submodule 并只读输出提示，未修改真实插件工作区。

## Feedback

- [x] **FB-1**: 清理已删除升级技能在项目文档、运行时提示和 i18n 资源中的残留描述

## Feedback Verification Notes

- FB-1 i18n: 已同步更新源码插件待升级错误提示的 `en-US` 与 `zh-CN` 运行时错误翻译；未新增前端运行时文案、菜单、路由、按钮、接口文档或插件 manifest 文案。
- FB-1 cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-1 data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-1 RESTful API: 不新增后端 REST API。
- FB-1 dev tools: 仅清理已删除命令路径的说明文字，不新增或修改开发工具/脚本入口。
- FB-1 tests: 已通过 `go test ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin -run 'TestBuildSourcePluginUpgradePendingErrorIncludesBulkCommand|TestValidateSourcePluginUpgradeReadinessFailsForPendingUpgrade' -count=1`、`openspec validate plugin-workspace-management --strict`、残留静态扫描和 `git diff --check`。
- [x] **FB-2**: 支持 `plugins.sources.<name>.items` 使用字符串 `"*"` 展开安装来源 root 下全部插件，禁止与显式插件 ID 混用
- [x] **FB-3**: 将测试工具入口从 `make test-*` / `linactl test-*` 统一改为 `make test.*` / `linactl test.*`

### Feedback Verification Notes

- FB-2 i18n: 仅调整开发工具配置语义、命令输出和 README 文档，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-2 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-2 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-2 RESTful API: 不新增后端 REST API。
- FB-2 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`go run ./hack/tools/linactl test.scripts`、`openspec validate plugin-workspace-management --strict`、`git diff --check -- hack/config.yaml Makefile make.cmd hack/tools/linactl README.md README.zh-CN.md openspec/changes/plugin-workspace-management`。
- FB-3 i18n: 仅调整开发工具命令名称、CI 调用和治理文档引用，不新增前端运行时文案、接口文档、插件 manifest i18n 或 apidoc i18n。
- FB-3 Cache: 不新增运行时缓存、缓存键、缓存失效路径、订阅或跨实例一致性逻辑。
- FB-3 Data permission: 不新增或修改 HTTP/API 数据操作接口，不涉及角色数据权限边界。
- FB-3 RESTful API: 不新增后端 REST API。
- FB-3 Dev tools: `make test.go`、`make test.host`、`make test.plugins`、`make test.scripts` 已替代旧的 `make test-go`、`make test-host`、`make test-plugins`、`make test-scripts`；`linactl` 同步提供 `test.go`、`test.host`、`test.plugins`、`test.scripts`，并用单元测试断言旧 hyphen 命令不再注册。
- FB-3 Tests: 已通过 `cd hack/tools/linactl && go test ./... -count=1`、`make -n test.go test.host test.plugins test.scripts`、`make -n test.go plugins=0`、`make -n test-go` 确认旧目标不存在、`go run ./hack/tools/linactl test.scripts`、`go run ./hack/tools/linactl help`、`go run ./hack/tools/linactl help test.go` / `test.host` / `test.plugins` / `test.scripts`、`go run github.com/rhysd/actionlint/cmd/actionlint@latest .github/workflows/reusable-backend-unit-tests.yml`、`openspec validate plugin-workspace-management --strict`、旧命令名静态扫描和 `git diff --check`。
- FB-3 Review: 已完成 `lina-review` 审查；确认 `hack/makefiles/test.mk` 只保留点号命名测试目标，`linactl` 命令注册表只暴露 `test.go`、`test.host`、`test.plugins`、`test.scripts`，旧 hyphen 命令仅作为单元测试负向断言和本任务说明出现。变更不新增 REST API、业务数据操作、运行时缓存或 i18n 资源。
