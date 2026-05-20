## 1. 清单模型与版本约束

- [x] 1.1 扩展插件 manifest 类型，新增 `dependencies.framework.version` 与 `dependencies.plugins[]` 结构，并同步动态插件 artifact manifest 序列化/反序列化
- [x] 1.2 实现依赖字段规范化与校验，覆盖默认值、插件 ID、版本范围、自依赖、重复依赖、未知安装策略和无效结构
- [x] 1.3 确定 LinaPro 框架版本权威来源，并实现框架版本读取与语义化版本范围匹配
- [x] 1.4 更新源码插件和动态插件示例 `plugin.yaml` 注释，至少为一个官方插件样例添加依赖声明
- [x] 1.5 补充 manifest 单元测试，覆盖无依赖清单、有效依赖清单、无效依赖清单和动态插件产物依赖解析

## 2. 依赖解析器

- [x] 2.1 新增内部依赖解析组件，输入目标插件、manifest 集合、registry/release 状态和框架版本，输出检查结论、阻断原因、自动安装计划、依赖链和反向依赖
- [x] 2.2 实现依赖图构建、确定性拓扑排序、自依赖检测和循环依赖检测
- [x] 2.3 实现硬依赖、软依赖、`install:auto` 与 `install:manual` 的状态判定
- [x] 2.4 实现反向硬依赖查询，优先读取已安装发布快照，无法确认安全时保守阻断
- [x] 2.5 为依赖解析器添加单元测试，覆盖框架版本不满足、缺失依赖、版本不满足、软依赖、自动安装计划、手动依赖阻断、循环依赖和反向依赖保护

## 3. 生命周期接入

- [x] 3.1 将依赖检查接入显式插件安装路径，在任何安装副作用前生成并执行自动安装计划
- [x] 3.2 实现自动依赖安装结果记录与结构化业务错误，错误包含目标插件、依赖插件、版本范围、循环链、已安装依赖和失败原因
- [x] 3.3 将反向依赖保护接入插件卸载路径，存在下游硬依赖时阻断卸载
- [x] 3.4 将依赖解析接入 `BootstrapAutoEnable`，确保启动自动启用按拓扑顺序安装自动依赖且不隐式启用依赖插件
- [x] 3.5 将依赖校验接入源码插件升级与动态插件安装/刷新路径，确保新版本依赖不满足时不切换有效发布
- [x] 3.6 确认依赖生命周期变化在单机模式刷新本地投影，在集群模式复用插件 runtime revision/event、enabled snapshot、frontend bundle 和 i18n bundle 失效机制

## 4. API、前端与 i18n

- [x] 4.1 设计并实现插件依赖检查 API 或扩展现有插件详情/安装接口，返回框架版本检查、依赖状态、自动安装计划、手动处理项、软依赖提示和卸载阻断项
- [x] 4.2 更新插件安装响应和错误响应 DTO，使用 `bizerr` 定义稳定错误码和 messageKey，避免自由文本作为调用端合同
- [x] 4.3 更新插件管理页面，展示依赖摘要、安装确认中的自动安装计划、阻断原因和卸载下游依赖提示
- [x] 4.4 同步维护前端运行时语言包、宿主 `manifest/i18n`、插件相关 apidoc i18n JSON 和必要的动态插件清单文案
- [x] 4.5 确认前端只展示服务端依赖检查结果，不在前端重复实现依赖图或版本约束决策

## 5. 测试与回归验证

- [x] 5.1 运行插件 catalog、依赖解析器、生命周期、启动自动启用、源码插件升级和动态插件运行时相关 Go 单元测试
- [x] 5.2 增加或更新插件管理 E2E 测试，覆盖安装前依赖计划展示、依赖阻断展示和卸载反向依赖阻断
- [x] 5.3 在集群/Redis 可用环境下运行插件 runtime revision、enabled snapshot、frontend/i18n bundle 失效相关回归测试
- [x] 5.4 运行插件 full 模式验证，覆盖官方源码插件 manifest 解析、后端单测和插件自有 E2E
- [x] 5.5 运行 `openspec validate plugin-dependency-management --strict`
- [x] 5.6 运行 `git diff --check -- openspec/changes/plugin-dependency-management apps/lina-core apps/lina-vben apps/lina-plugins`
- [x] 5.7 记录 i18n、缓存一致性、数据权限、RESTful API、开发工具脚本和跨平台影响评估
- [x] 5.8 完成实现后调用 `lina-review`，重点审查依赖图语义、自动安装副作用边界、缓存一致性、i18n 覆盖、数据权限无泄露和测试覆盖

## Feedback

- [x] **FB-1**: 新增跨平台 `make tidy` 指令，扫描仓库 Go 模块并逐个执行 `go mod tidy`
- [x] **FB-2**: 合并插件生命周期 facade 中重复的 `Install` 与 `InstallWithResult` 安装入口
- [x] **FB-3**: 合并插件生命周期 facade 中重复的 `Uninstall` 与 `UninstallWithOptions` 卸载入口
- [x] **FB-4**: 移除 `make pb`、`make pbentity`、`make enums`、`make deploy`、`make service` 指令入口
- [x] **FB-5**: 清理根目录 `make help` 中残留的 `pb`、`pbentity`、`enums`、`deploy`、`service` 命令展示
- [x] **FB-6**: 根目录 `make help` 隐藏仅适用于内部目录的 `dao`、`cli`、`cli.install`、`ctrl` 命令

## 验证与影响记录

- 5.1 已运行 `go test ./internal/service/plugin/internal/catalog ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/dependency ./internal/service/plugin/internal/sourceupgrade ./internal/service/plugin ./internal/controller/plugin -count=1`，并补充运行依赖生命周期、控制器投影、runtime revision、共享 enabled snapshot、pluginruntimecache、i18n runtime bundle 相关回归测试。
- 5.2 新增 `TC0235-plugin-dependency-management.ts`，覆盖安装前自动依赖计划展示、依赖阻断展示并禁用安装提交、卸载反向依赖阻断并禁用卸载提交；已用 `E2E_BROWSER_CHANNEL=chrome` 运行通过。
- 5.3 已临时启动本机无持久化 Redis 并运行 `LINA_TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service/coordination -run 'TestRedisProviderIntegration|TestRedisProviderFailureIntegration' -count=1 -v`；同时运行 `LINA_TEST_REDIS_ADDR=127.0.0.1:6379 go test ./internal/service/plugin/internal/runtime ./internal/service/plugin/internal/integration ./internal/service/pluginruntimecache ./internal/service/i18n ./internal/service/plugin -run 'Test.*Revision|Test.*Shared|Test.*Cache|Test.*DynamicPlugin|TestBootstrapAutoEnableWaitsUntilCurrentNodeBecomesPrimary|Test.*Cluster|Test.*Topology' -count=1`，完成后已关闭临时 Redis。
- 5.4 已运行官方插件后端 `GOWORK=$(pwd)/temp/go.work.plugins go test ./apps/lina-plugins/... -count=1`，并运行插件 full E2E 入口 `E2E_BROWSER_CHANNEL=chrome pnpm test:module -- plugins -- --project chromium`，结果 165 passed。
- i18n 影响：已同步前端运行时语言包、宿主 `manifest/i18n` 错误翻译、`zh-CN` apidoc i18n，并重新执行 `go run ../../hack/tools/linactl prepare-packed-assets` 生成嵌入资源；`en-US` apidoc 仍保持英文源文本策略。
- 缓存一致性影响：只读依赖检查不触发缓存失效；安装、卸载、升级路径复用现有插件 runtime revision/event、enabled snapshot、frontend bundle、runtime i18n bundle 的插件作用域失效机制。单机模式走本地投影刷新；集群模式继续依赖既有 cluster/topology 与 coordination 抽象。
- 数据权限影响：依赖检查只读取插件 catalog、registry、release snapshot 和清单元数据，不读取或暴露租户业务数据；未新增业务数据操作面。
- RESTful API 影响：新增只读接口为 `GET /plugins/{id}/dependencies`；安装仍为 `POST /plugins/{id}/install`，卸载仍为 `DELETE /plugins/{id}`。
- 开发工具与跨平台影响：未新增 Shell/PowerShell 脚本；资源打包使用既有 Go 工具 `linactl prepare-packed-assets`；新增 E2E 和 TypeScript 代码不引入平台专属命令。
- 5.8 已完成 `lina-review`：发现并修复依赖检查结果中安装方向 `blockers` 与卸载方向反向阻断混用的问题，新增 `reverseBlockers` API/前端字段，保留 `blockers` 仅用于安装/升级方向阻断；补充服务层未知 release snapshot 反向阻断测试、控制器投影测试和 TC0235 E2E mock 覆盖。复查未发现新的阻断问题。
- FB-2 已将 `LifecycleManagementService.Install` 调整为唯一安装入口并返回 `DependencyCheckResult`，删除 facade 层 `InstallWithResult`；控制器继续从 `Install` 返回值投影安装响应，其他调用点丢弃结果仅保留错误处理。验证已运行 `go test ./internal/service/plugin ./internal/controller/plugin -count=1`、`go test ./internal/cmd -count=1`、`openspec validate plugin-dependency-management --strict`、`git diff --check -- apps/lina-core/internal/controller/plugin apps/lina-core/internal/service/plugin openspec/changes/plugin-dependency-management`，并通过 `rg -n "InstallWithResult"` 确认生产代码无遗留旧入口。i18n、缓存一致性、数据权限、RESTful API 和开发工具脚本均无新增影响；`lina-review` 复查未发现阻断问题。
- FB-3 已将 root `LifecycleManagementService.Uninstall` 调整为唯一卸载入口并显式接收 `UninstallOptions`，删除 root facade 层 `UninstallWithOptions`；控制器继续按请求传递 `PurgeStorageData` 与 `Force`，默认调用点显式传入 `UninstallOptions{PurgeStorageData: true}` 以保持原卸载策略。底层 dynamic runtime service 的 `UninstallWithOptions` 属于内部边界，本次未合并。验证已运行 `go test ./internal/service/plugin ./internal/controller/plugin -count=1`、`go test ./internal/cmd -count=1`、`openspec validate plugin-dependency-management --strict`、`git diff --check -- apps/lina-core/internal/controller/plugin apps/lina-core/internal/service/plugin openspec/changes/plugin-dependency-management`，root facade 复查未发现旧入口遗留。i18n、缓存一致性、数据权限、RESTful API 和开发工具脚本均无新增影响；`lina-review` 复查未发现阻断问题。
- FB-4 已移除 `apps/lina-core/hack/hack.mk` 中的 `pb`、`pbentity`、`enums`、`deploy`、`service` make target，并清理 `apps/lina-core/Makefile` 中仅供 `deploy` 使用的 `NAMESPACE` 与 `DEPLOY_NAME` 变量。验证已运行 `make -C apps/lina-core -n pb`、`make -C apps/lina-core -n pbentity`、`make -C apps/lina-core -n enums`、`make -C apps/lina-core -n deploy`、`make -C apps/lina-core -n service`，均返回 `No rule to make target`；已运行 `go test ./hack/tools/linactl -count=1`、`openspec validate plugin-dependency-management --strict`、`git diff --check -- apps/lina-core/Makefile apps/lina-core/hack/hack.mk openspec/changes/plugin-dependency-management/tasks.md`，并通过静态扫描确认仓库中无这些 make target 或文档命令引用残留。i18n、缓存一致性、数据权限和 RESTful API 无影响；开发工具影响为移除过期 Makefile 包装入口，未新增平台专属脚本或命令依赖；`lina-review` 复查未发现阻断问题。
- FB-5 已从 `hack/tools/linactl` 命令注册表移除 `pb`、`pbentity`、`enums`、`deploy`、`service`，并删除不再可达的 `runDeploy` 实现，确保根目录 `make help` 不再展示这些命令。验证已运行 `make help`，输出中不再包含上述命令；已运行 `go test ./hack/tools/linactl -count=1`、`openspec validate plugin-dependency-management --strict`、`git diff --check -- apps/lina-core/Makefile apps/lina-core/hack/hack.mk hack/tools/linactl/command.go hack/tools/linactl/command_ops.go openspec/changes/plugin-dependency-management/tasks.md`，并通过静态扫描确认 `linactl` 注册文本无遗留。i18n、缓存一致性、数据权限和 RESTful API 无影响；开发工具影响为同步移除跨平台 CLI 帮助入口和不可达 deploy 逻辑，未新增平台专属脚本或命令依赖；`lina-review` 复查未发现阻断问题。
- FB-6 已为 `linactl` 命令增加内部命令标记，根目录 `make help` 默认隐藏 `dao`、`cli`、`cli.install`、`ctrl`，并保留 `linactl help --all` 展示完整命令表；`apps/lina-core` 内部目录的 `make dao`、`make ctrl`、`make cli`、`make cli.install` 仍可用。验证已运行 `make help | rg "^  (dao|cli|cli\\.install|ctrl)\\s"`，无输出；运行 `cd hack/tools/linactl && go run . help --all | rg "^  (dao|cli|cli\\.install|ctrl)\\s"` 可看到四个内部命令；运行 `make -C apps/lina-core -n dao`、`make -C apps/lina-core -n ctrl`、`make -C apps/lina-core -n cli`、`make -C apps/lina-core -n cli.install` 确认内部 target 保留；已运行 `go test ./hack/tools/linactl -count=1`、`openspec validate plugin-dependency-management --strict`、`git diff --check -- hack/tools/linactl/app.go hack/tools/linactl/command.go hack/tools/linactl/constants_types.go hack/tools/linactl/main_test.go openspec/changes/plugin-dependency-management/tasks.md`。i18n、缓存一致性、数据权限和 RESTful API 无影响；开发工具影响为区分根目录帮助视图与内部工具命令视图，未新增平台专属脚本或命令依赖；`lina-review` 复查未发现阻断问题。
