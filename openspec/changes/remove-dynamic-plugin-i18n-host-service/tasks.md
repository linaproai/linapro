## 1. Specification

- [x] 1.1 更新`plugin-host-service-extension`增量规范，明确动态插件不开放`i18n`host service，源码插件`I18n()`能力保留。

## 2. Implementation

- [x] 2.1 从动态插件 host service catalog、协议别名、guest SDK 目录、`WASM`registry 和授权判断中移除`i18n`。
- [x] 2.2 更新或补充单元测试，覆盖`service: i18n`声明被拒绝、catalog/dispatcher 不再要求`i18n`。

## 3. Documentation

- [x] 3.1 更新`apps/lina-core/pkg/plugin`中英文`README`，同步动态插件多语言边界。
- [x] 3.2 更新官网领域能力与`i18n`专题文档，去掉动态插件自行读取或管理`manifest/i18n`的说明。

## 4. Verification

- [x] 4.1 运行`openspec validate remove-dynamic-plugin-i18n-host-service --strict`。
- [x] 4.2 运行相关 Go 测试、静态检索和格式检查。
- [x] 4.3 完成`lina-review`审查，并记录影响分析：`i18n`有边界调整；缓存、数据权限、开发工具跨平台无实现影响；不新增运行期依赖。

## Impact Record

- `i18n`影响：有。动态插件不再发布`service: i18n`、`host:i18n`、`locale.get`、`messages.translate`或`messages.keys.find`；动态插件继续通过`manifest/i18n`交付资源，由宿主统一发现、合并、缓存和前端分发；源码插件`capability.Services.I18n()`保留。
- 缓存一致性影响：无缓存实现变更。未修改翻译缓存权威数据源、失效触发点、跨实例同步机制或缓存后端，仅移除动态 guest 访问入口。
- 数据权限影响：无数据读取、写入、下载、聚合或执行类接口变更；未新增插件通过宿主服务访问数据的路径。
- 开发工具跨平台影响：无。未修改`Makefile`、脚本、代码生成入口、CI 或开发工具。
- DI 来源影响：未新增运行期依赖；删除`WASM`i18n host service 对`capability.Services.I18n()`的动态分发读取路径，源码插件能力目录 DI 不变。
- 测试策略：内部可执行协议行为使用 Go 单元测试和编译门禁覆盖；文档和规范使用`openspec validate`、静态检索和`git diff --check`覆盖；无前端 UI 或 E2E 资产变更。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/i18n.md`、`.agents/rules/testing.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、官网`AGENTS.md`和官网 Markdown 格式规则。

## Verification Record

- `openspec validate remove-dynamic-plugin-i18n-host-service --strict`：通过。
- `cd apps/lina-core && go test ./pkg/plugin/pluginbridge/... ./internal/service/plugin/internal/wasm -count=1`：通过。
- `cd apps/lina-core && go test ./internal/cmd -count=1`：通过。
- `cd apps/lina-core && go test ./internal/service/plugin/internal/catalog -count=1`：通过。
- `git diff --check`：通过。
- `git -C /Users/john/Workspace/github/linaproai/linapro-site diff --check -- apps/lina-site/docs/docs/2000-components/6000-domain-capabilities/6000-domain-capabilities.md apps/lina-site/docs/docs/2000-components/6000-domain-capabilities/7100-i18n.md`：通过。
- 静态检索`HostServiceI18n|CapabilityI18n|HostServiceMethodI18n|host:i18n|messages.translate|messages.keys.find|domainhostcall.I18n|Services.I18n|pluginbridge.Services.I18n`：动态插件协议和 dispatcher 无残留；仅源码插件 i18n 测试替身保留。

## Lina Review Record

- 审查范围：本变更创建的`openspec/changes/remove-dynamic-plugin-i18n-host-service/**`，动态插件`i18n`host service 删除涉及的 Go 文件，`apps/lina-core/pkg/plugin`中英文`README`，以及用户指定的两篇官网文档。既有未跟踪变更`openspec/changes/clarify-plugin-storage-files-boundary/**`不属于本次范围，未修改。
- 范围来源：`git status --short`、`git diff --name-only`、`git ls-files --others --exclude-standard`、官网仓库目标文件状态和`openspec status --change remove-dynamic-plugin-i18n-host-service --json`。
- 已读取规则文件：`AGENTS.md`、`.agents/rules/openspec.md`、`.agents/rules/documentation.md`、`.agents/rules/architecture.md`、`.agents/rules/plugin.md`、`.agents/rules/backend-go.md`、`.agents/rules/api-contract.md`、`.agents/rules/i18n.md`、`.agents/rules/testing.md`、`.agents/rules/cache-consistency.md`、`.agents/rules/data-permission.md`、官网`AGENTS.md`和官网 Markdown 格式规则。
- 发现的问题：审查中发现官网示例误用不存在的`pluginbridge.JSON`，已修正为 typed controller 可返回的 DTO 示例；修正后未发现阻塞问题。
- 规则域结论：OpenSpec、文档、架构、插件、后端 Go、`i18n`、测试均通过；API 契约未命中公开 HTTP DTO 或路由变更；缓存一致性、数据权限、开发工具跨平台无实现影响。
- E2E 质量审查：未触发。变更不涉及前端页面、路由、表单、表格、权限交互或端到端用户流程；使用 Go 单元测试、OpenSpec 校验、静态检索和文档格式检查覆盖。
