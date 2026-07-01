## 1. 宿主对象存储组件

- [x] 1.1 新增 `apps/lina-core/internal/service/storage` 组件，定义中立 `Service`、输入输出 DTO、对象元数据和构造函数。
- [x] 1.2 实现本地存储后端，覆盖 namespace/key 规范化、目录穿越拒绝、对象写入、读取、删除、`Stat` 和有界 `List`。
- [x] 1.3 为本地存储后端补充单元测试，覆盖合法 key、非法路径、写读删、元数据和列表上限。

## 2. 文件中心迁移

- [x] 2.1 将 `file.Service` 构造函数和 `serviceImpl` 的存储依赖改为 `storage.Service`，移除 `file.Storage` 和 `file.LocalStorage`。
- [x] 2.2 保持文件上传路径、hash 复用、历史路径读取、删除清理和公开 URL 生成语义不变。
- [x] 2.3 更新文件服务相关测试和测试替身，验证新上传路径、历史路径访问、下载读取、删除清理和文件能力适配仍通过。

## 3. 插件存储迁移

- [x] 3.1 将插件内置本地 `storagecap.Provider` 改为委托宿主 `storage.Service`，保留 `storagecap.Provider` 公开契约。
- [x] 3.2 更新插件 host services 和启动装配，确保文件中心与插件内置 provider 复用启动期同一个宿主对象存储实例。
- [x] 3.3 更新插件存储、WASM host service、分片上传和卸载清理相关测试，验证 logical path、插件/租户隔离、列表上限和 provider 状态不变。

## 4. 验证与审查

- [x] 4.1 运行 `openspec validate unify-host-object-storage --strict`。
- [x] 4.2 运行静态检索，确认 `file.Storage`、`file.NewLocalStorage` 和 `file.Service/serviceImpl` 对 `storagecap.Service` 的直接依赖不存在。
- [x] 4.3 运行覆盖变更包的 Go 测试和启动装配编译门禁。
- [x] 4.4 记录影响分析：`i18n` 无运行时文案影响；缓存一致性无影响；数据权限边界保持文件中心和插件领域原有校验；开发工具跨平台无脚本影响；测试策略为后端单元/集成编译验证，无新增 E2E。
- [x] 4.5 完成 `lina-review` 审查并处理阻塞问题。

## Verification Record

- `openspec validate unify-host-object-storage --strict`：通过。
- 静态检索：
  - `rg -n "type Storage interface|file\\.NewLocalStorage|NewLocalStorage\\(|\\bStorage\\b" apps/lina-core/internal/service/file apps/lina-core/internal/cmd/internal/httpstartup -g'*.go'`：仅剩 `file_capability.go` 注释中的插件 `Storage()` 领域说明，无旧 `file.Storage` 或 `file.NewLocalStorage`。
  - `rg -n "storagecap\\.Service|storagecap" apps/lina-core/internal/service/file/file.go apps/lina-core/internal/service/file/file_upload.go apps/lina-core/internal/service/file/file_open.go apps/lina-core/internal/service/file/file_query.go -g'*.go'`：无输出，确认 `file.Service/serviceImpl` 不依赖插件 `storagecap`。
- Go 测试：
  - `go test ./internal/service/storage ./internal/service/file ./internal/service/plugin/internal/capabilityhost ./internal/service/plugin/internal/wasm ./internal/cmd/internal/httpstartup -count=1`：通过。
  - `go test ./internal/service/storage -count=1`：通过。
  - `go test ./internal/service/file -count=1`：通过。
  - `go test ./internal/service/plugin/internal/capabilityhost -count=1`：通过。
  - `go test ./internal/service/plugin/internal/wasm -count=1`：通过。
  - `go test ./internal/cmd/internal/httpstartup -count=1`：通过。
  - `go test ./internal/cmd -count=1`：通过。
  - `go test ./internal/service/plugin/internal/runtime -run 'Test(NewWiresStorageCleanupServices|ScanPluginManifestsDiscoversRuntimePluginFromStorage|RuntimeWiring|RunReconcilerTickSafelyRecoversPanic)' -count=1`：通过。
- 格式检查：
  - `git diff --check -- apps/lina-core/internal/service/storage/storage.go apps/lina-core/internal/service/file apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_storage_local_provider.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_storage_adapter_test.go apps/lina-core/internal/service/plugin/internal/capabilityhost/capabilityhost_adapters_test.go apps/lina-core/internal/service/plugin/plugin_host_services.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_runtime.go apps/lina-core/internal/cmd/internal/httpstartup/httpstartup_routes_test.go openspec/changes/unify-host-object-storage`：通过。
- 扩展回归：
  - `go test ./internal/service/plugin ./internal/service/plugin/... -count=1`：除 `internal/service/plugin/internal/runtime` 中依赖动态样例 WASM 构建的既有失败外，其余相关包通过；失败原因是 `pkg/plugin/pluginbridge/recordstore` 的 WASI 构建字段 `Tx.invoker` 缺失，与本次存储重构无关。

## Impact Record

- `i18n` 影响：无运行时用户可见文案、API 文档源文本、插件清单或语言包资源变更。
- 缓存一致性影响：无缓存权威数据源、失效机制、跨实例同步或共享修订号变化。
- 数据权限影响：文件中心列表、详情、下载和删除继续使用原有数据权限校验；插件 `Storage()` 继续保持插件 logical path 授权、插件/租户隔离和不可见对象不泄露语义。
- 开发工具跨平台影响：不修改脚本、Makefile、CI、`linactl` 或平台相关执行入口。
- 测试策略：纯后端内部重构，无前端 UI 或用户可观察页面行为变化；使用单元测试、插件 host service/WASM 测试、启动装配编译门禁和静态检索验证，无新增 E2E。
- DI 来源检查：新增 `storage.Service` 由 HTTP 启动装配创建一次，并同时注入 `file.New` 与插件内置本地 provider；未新增运行期服务定位器、聚合依赖结构体或请求路径临时 `New()`。

## Lina Review Record

- 审查范围：`unify-host-object-storage` 相关后端服务、插件本地 provider、启动装配测试、`storage` 新组件和本变更 OpenSpec 文档；工作区中其他活跃变更、暂存内容和 `apps/lina-plugins` 子仓库状态不属于本次审查范围。
- 已读取规则：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/architecture.md`、`.agents/rules/backend-go.md`、`.agents/rules/plugin.md`、`.agents/rules/data-permission.md`、`.agents/rules/testing.md`、`.agents/rules/documentation.md`、`.agents/rules/i18n.md`。
- 结论：未发现阻塞问题。`file.Service` 依赖宿主内部 `storage.Service`，插件内置本地 provider 也复用同一启动期对象存储实例；文件中心和插件 `Storage()` 的领域契约、数据权限和 logical path 边界保持独立。

## Feedback

- [x] **FB-1**: 修正 `apps/lina-core/internal/service/file` 中缺少对应源码文件的单元测试文件命名。

### FB-1 Verification Record

- 根因：`file_access_test.go` 和 `file_runtime_params_test.go` 的测试文件基名没有同名源码文件，不符合后端单元测试命名规则。
- 修复：`file_access_test.go` 重命名为 `file_open_test.go`；`file_runtime_params_test.go` 重命名为 `file_upload_test.go`。
- 静态命名检查：`apps/lina-core/internal/service/file/*_test.go` 均存在同名 `.go` 源码文件，通过。
- `go test ./internal/service/file -count=1`：通过。
- 影响分析：仅调整测试文件命名；无运行时行为、`i18n`、缓存一致性、数据权限、开发工具跨平台、API、SQL 或 E2E 影响。
