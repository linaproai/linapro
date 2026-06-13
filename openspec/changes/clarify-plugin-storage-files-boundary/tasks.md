## 1. 规范与边界

- [x] 1.1 补充`plugin-storage-service`和`plugin-host-domain-capabilities`增量规范，明确`Storage()`与`Files()`领域职责、动态`hostServices`声明边界和数据权限影响。
  - 记录：已新增`plugin-storage-service`与`plugin-host-domain-capabilities`增量规范；`Storage()`定义为插件私有对象存储领域，`Files()`定义为宿主文件中心投影与可见性校验领域；动态`storage`使用`resources.paths`，动态`files`不使用路径资源。
- [x] 1.2 运行`openspec validate clarify-plugin-storage-files-boundary --strict`，确认新增规范满足 OpenSpec 校验。
  - 验证：`openspec validate clarify-plugin-storage-files-boundary --strict`已通过。

## 2. 源码插件示例迁移

- [x] 2.1 修改`linapro-demo-source`前先检查插件根目录`AGENTS.md`，然后将 demo service 构造函数改为显式接收`storagecap.Service`，并在`backend/plugin.go`通过`registrar.Services().Storage()`注入。
  - 记录：`apps/lina-plugins/linapro-demo-source/AGENTS.md`不存在，按顶层规范执行；`demo.New`已显式接收`storagecap.Service`，路由注册校验并注入`registrar.Services().Storage()`。DI 来源为宿主启动期`capability.Services`，经`sourceServicesProvider`按插件 ID 作用域化后发布给`pluginhost.Services`，源码插件复用该共享能力目录，不创建独立服务图。
- [x] 2.2 将`linapro-demo-source`附件保存、下载、替换、删除和卸载清理逻辑迁移到`storagecap.Service`，业务表只保存插件 logical path、附件名和必要展示元数据。
  - 记录：附件保存使用`storagecap.Service.Put`写入`demo-record-files/...` logical path；下载使用`Get`返回的 reader 流式输出；替换、删除和失败回滚使用`Delete`；卸载清理通过源码插件卸载输入发布的插件作用域`Services()`获取`Storage()`，按业务表记录的`tenant_id + attachment_path`清理，并对已记录租户执行有界前缀清理。
- [x] 2.3 移除或收敛源码插件示例中直接读取`upload.path`、拼接宿主本地路径、依赖`os/gfile`管理插件附件内容的实现；如需历史本地附件兼容，在任务记录中说明兼容读取和清理边界。
  - 记录：已移除附件链路中的`upload.path`读取、宿主本地路径拼接和`os/gfile`文件内容管理；本次全新项目无历史兼容负担，不保留旧本地附件读取兼容。

## 3. 文档同步

- [x] 3.1 更新`apps/lina-core/pkg/plugin/README.md`与`README.zh-CN.md`，用场景表明确`Storage()`是插件私有对象存储，`Files()`是宿主文件中心投影与可见性校验。
- [x] 3.2 更新`linapro-demo-source`中英文 README，说明源码插件附件通过`Storage()`管理，并记录卸载清理语义。
- [x] 3.3 更新`linapro-demo-dynamic`中英文 README，补充`service: storage`与`service: files`并列示例，避免把`Storage()`和`Files()`混用。

## 4. 验证与审查

- [x] 4.1 运行覆盖源码插件后端的 Go 编译门禁和相关单元测试；若迁移触发控制器或路由绑定变化，运行对应启动绑定测试。
  - 验证：`GOWORK=off go test ./backend/... -count=1`（源码插件后端）通过；`go test ./internal/service/plugin/internal/capabilityhost -count=1`通过；`go test ./pkg/plugin/pluginhost ./internal/service/plugin/internal/lifecycle -count=1`通过；`go test ./internal/service/plugin -run 'TestPluginLifecycleOrchestrationStaticBoundaries|Test.*Lifecycle|TestRegisterUninstallHandlerPublishesPolicySnapshot' -count=1`通过；`go test ./internal/cmd -count=1`通过。
- [x] 4.2 运行`openspec validate clarify-plugin-storage-files-boundary --strict`和`git diff --check`，并检查中英文 README 镜像内容一致。
  - 验证：`openspec validate clarify-plugin-storage-files-boundary --strict`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过；中英文 README 镜像文件存在且均包含`Storage`/`storage`边界说明；静态检索确认源码 demo 附件链路不再包含`upload.path`、`buildDemoAttachmentFullPath`、`gfile`、`os.Open`或`FullPath`物理路径依赖。
- [x] 4.3 按`lina-review`完成审查，记录已读取规则、DI 来源、数据权限、缓存一致性、`i18n`影响、开发工具跨平台影响和测试策略。
  - 审查：已读取`AGENTS.md`、`openspec.md`、`backend-go.md`、`plugin.md`、`documentation.md`、`architecture.md`、`data-permission.md`、`api-contract.md`、`testing.md`、`i18n.md`、`cache-consistency.md`和 Markdown 格式规则；命中`openspec-apply-change`、`lina-review`和`goframe-v2`技能。活跃变更`clarify-plugin-storage-files-boundary`仍未归档；工作区存在另一活跃变更`remove-dynamic-plugin-i18n-host-service`相关无关差异，本次未回滚也未纳入本变更完成依据。
  - 范围：父仓库状态和未跟踪文件已检查，`apps/lina-plugins`子仓库状态、差异和未跟踪文件已展开；`apps/lina-plugins/linapro-demo-source/AGENTS.md`和`apps/lina-plugins/linapro-demo-dynamic/AGENTS.md`均不存在，按顶层规范和命中规则执行。
  - DI 来源：源码插件`Storage()`来自宿主启动期创建的`capability.Services`，经`sourceServicesProvider.SourceServicesForPlugin(pluginID)`按插件 ID 作用域化后传入生命周期卸载输入，再由`linapro-demo-source`显式注入`storagecap.Service`；复用宿主共享 provider 和启动期能力目录，不在业务路径临时`New()`独立服务图。
  - 数据权限：源码插件 CRUD 仍通过`tenantFilter`约束记录可见性；下载先按记录可见性读取业务记录，再通过同租户上下文访问对象；卸载清理属于宿主生命周期管理例外，按业务表投影读取`tenant_id + attachment_path`并绑定目标租户作用域删除插件私有对象。审查中发现 Storage 适配器原先可能让宿主请求租户覆盖插件显式租户，已修复为显式`bizctxcap.WithCurrentContext`优先，并新增单元测试覆盖。
  - 缓存一致性：本次未新增缓存、未修改缓存失效策略或跨实例协调；`Storage()`继续复用启动期共享 provider runtime 和本地 provider，未引入节点本地派生缓存。
  - `i18n`：`linapro-demo-source`启用插件 i18n；本次只更新附件存储不可用错误和移除废弃本地文件错误资源，中英文插件错误资源同步维护。未修改前端运行时文案、菜单、路由或 API DTO 文档元数据。
  - 接口与文档：未修改 HTTP 路由、方法、DTO 或公开 API 签名；下载实现改为内部流式读取`Storage()`返回的 reader。中英文 README 镜像均已同步补充`Storage()`与`Files()`边界。
  - 开发工具跨平台：未修改`Makefile`、脚本、CI、代码生成或工具入口，无开发工具跨平台影响。
  - 测试策略：本次是后端能力边界和文档治理变更，无前端用户可观察页面流程变化，未触发 E2E 质量审查；通过 Go 单元测试、启动绑定包测试、OpenSpec 严格校验、diff 空白检查、README 镜像检查和静态检索覆盖。剩余风险为未做真实对象存储 provider 集成测试，但 provider key、租户作用域、生命周期 DI 和源码插件编译门禁已由现有替身测试覆盖。

## Feedback

- [x] **FB-1**: 移除通用`Storage().Put`固定`32MB`大小上限，改为流式写入 provider；`linapro-demo-source`示例附件业务限制调整为`32MB`。
  - 根因：原`storageAdapter.Put`为检测内容类型会先把完整`Body`读入内存，因此用`MaxObjectBytes = 32MB`做硬保护；这把内存保护误固化为通用`Storage()`领域大小上限。
  - 修复：`storageAdapter.Put`改为只读取前`512`字节探测`Content-Type`，再用`io.MultiReader`将探测前缀拼回原流并透传给 provider；移除`storagecap.MaxObjectBytes`和`PLUGIN_STORAGE_OBJECT_TOO_LARGE`通用错误码；`PutInput.Size`继续透传 provider，但不作为通用拒绝条件。`linapro-demo-source`示例附件业务上限调整为`32MB`，业务错误码和本地化参数继续由插件自身治理。
  - 审查：按`lina-review`完成反馈级审查，已读取`AGENTS.md`、`openspec.md`、`backend-go.md`、`plugin.md`、`architecture.md`、`data-permission.md`、`testing.md`、`cache-consistency.md`、`i18n.md`、`documentation.md`、`api-contract.md`和 Markdown 格式规则，并使用`lina-feedback`、`lina-review`、`goframe-v2`技能。审查范围来自父仓库和`apps/lina-plugins`子仓库状态；`linapro-demo-source`和`linapro-demo-dynamic`插件根目录均无本地`AGENTS.md`。
  - 影响分析：本次不新增运行期依赖，不改变`storagecap.Service`公开方法签名、HTTP API、DTO、动态 host service wire 协议、SQL 或 DAO；无数据权限边界变化，插件对象仍按插件 ID 和租户作用域隔离；无缓存一致性影响；无开发工具跨平台影响；无前端页面或用户可观察 UI 流程变化，不触发 E2E 质量审查；`i18n`仅沿用已有 demo 附件大小错误资源，不新增翻译键。
  - 验证：新增`TestStorageAdapterStreamsPutWithoutFixedObjectLimit`和`TestStorageAdapterContentTypeProbePreservesBody`覆盖移除固定大小拒绝和流式探测不吞 body；`go test ./internal/service/plugin/internal/capabilityhost ./pkg/plugin/capability/storagecap -count=1`通过；`GOWORK=off go test ./backend/... -count=1`通过；`go test ./pkg/plugin/pluginhost ./internal/service/plugin/internal/lifecycle -count=1`通过；`go test ./internal/service/plugin -run 'TestPluginLifecycleOrchestrationStaticBoundaries|Test.*Lifecycle|TestRegisterUninstallHandlerPublishesPolicySnapshot' -count=1`通过；`go test ./internal/cmd -count=1`通过；`openspec validate clarify-plugin-storage-files-boundary --strict`通过；`git diff --check && git -C apps/lina-plugins diff --check`通过；静态检索确认通用`MaxObjectBytes`、`CodeStorageObjectTooLarge`、`PLUGIN_STORAGE_OBJECT_TOO_LARGE`和旧`10MB`限制无残留。
  - 剩余风险：本次使用本地 provider 和测试 provider 覆盖流式写入契约，未做真实远端对象存储 provider 的大对象集成测试；远端 provider 若需要自身大小、费用或带宽限额，应在 provider 或业务层独立治理。
