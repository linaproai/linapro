## 1. 规则与基线

- [x] 1.1 读取并记录本变更命中的规则文件和技能入口。
- [x] 1.2 复查`apps/lina-core/internal/service/plugin`根包文件数量、低行数文件和大测试文件分布。

## 2. 生产文件组织

- [x] 2.1 合并根包同职责低行数 host service 配置文件。
- [x] 2.2 合并`runtimecache`中同职责低行数 revision 观察文件。
- [x] 2.3 合并 tenant governance provisioning policy 到 startup consistency 治理文件，保持根插件`Service`公开契约、缓存刷新、runtime upgrade 和 lifecycle 行为不变。

## 3. 单元测试组织

- [x] 3.1 将`plugin_list_test.go`中的 data table comment 与 runtime upgrade preview 测试拆到关联测试文件，并将 host service、tenant provisioning 测试归位到合并后的生产职责文件。
- [x] 3.2 将列表查找和字符串包含 helper 收敛到根包测试支撑文件。
- [x] 3.3 确认测试仍然自包含、顺序无关，清理逻辑由当前测试显式注册。

## 4. 验证与审查

- [x] 4.1 运行 Go 格式化、相关 Go 测试、OpenSpec 严格校验和 diff 检查。
- [x] 4.2 记录`i18n`、缓存一致性、数据权限、开发工具跨平台和测试策略影响。
- [x] 4.3 调用`lina-review`完成代码和规范审查。

## Feedback

- [x] **FB-1**: 修复根插件服务测试文件缺少对应生产源码文件的问题。

## Execution Notes

- 规则与技能：已使用`lina-feedback`和`goframe-v2`；已读取`openspec`、`architecture`、`backend-go`、`plugin`、`testing`、`cache-consistency`、`data-permission`、`i18n`和`documentation`规则。
- 生产文件：`ConfigureWasmHostServices`并入`plugin_host_services.go`；tenant provisioning policy 并入`plugin_startup_consistency.go`；`ObservedRevision`方法并入`runtimecache.go`。
- DI 来源检查：本轮不新增运行期依赖、不修改构造函数签名、不改变启动装配传递路径；`ConfigureWasmHostServices`仅移动文件位置，所有依赖仍由宿主启动期显式传入并复用既有共享实例或共享后端。
- 测试文件：新增`plugin_data_table_comment_test.go`、`plugin_runtime_upgrade_preview_test.go`和`plugin_host_services_test.go`；删除已失去生产文件对应关系的`plugin_wasm_host_services_test.go`、`plugin_tenant_provisioning_policy_test.go`和`plugin_test_helpers_test.go`。
- `FB-1`根包命名闭环：新增`plugin_topology.go`、`plugin_capability_revision.go`、`plugin_platform_guard.go`、`plugin_startup_snapshot.go`和`plugin_management_list_cache.go`；将`plugin_dependency_lifecycle_test.go`归位为`plugin_dependency_test.go`；将启动快照、依赖缓存和管理列表缓存测试按对应生产文件拆分；根包`*_test.go`到生产`*.go`的机械扫描已清零。
- `FB-1`递归子包检查：`internal/catalog`、`internal/frontend`、`internal/integration`、`internal/lifecycle`、`internal/runtime`和`internal/wasm`仍存在子包测试命名债务，本轮未扩展修改，避免与当前工作区中这些子包已有并行改动混杂；已通过`go test ./internal/service/plugin/internal/... -count=1`确认编译和行为测试未被本轮根包整理破坏。
- `i18n`影响：无运行时文案、API 文档源文本、插件清单或语言包变更。
- 缓存一致性影响：仅移动 runtime cache revision 观察方法和既有 cache invalidation 调用位置，不改变权威数据源、失效触发点、跨实例同步机制或最大陈旧时间。
- 数据权限影响：不新增或修改数据读取、写入、导出、下载或插件数据访问路径。
- 开发工具跨平台影响：不修改脚本、CI、生成入口或`linactl`。
- 测试策略：纯内部治理重构，无用户可观察行为变化，不新增 E2E；使用 Go 单元测试、OpenSpec 严格校验和 diff 检查闭环。
- 验证结果：`openspec validate refactor-plugin-service-layout --strict`通过；`cd apps/lina-core && go test ./internal/service/plugin -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/runtimecache -count=1`通过；`cd apps/lina-core && go test ./internal/service/plugin/internal/... -count=1`通过；`git diff --check -- apps/lina-core/internal/service/plugin openspec/changes/refactor-plugin-service-layout/tasks.md`通过。
- 审查结果：`lina-review`已按反馈级范围完成，未发现阻塞问题；E2E 质量审查未触发，原因是本轮仅为后端内部文件组织和单元测试组织调整，无用户可观察行为或 E2E 资产变更。
