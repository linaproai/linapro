## Why

当前插件侧 `Manifest()` 能力仍保留旧 `Metadata` 读取心智：默认围绕 `metadata.yaml` 和声明型 `*.yaml` 文件设计，并排除 `manifest/config/`、`manifest/sql/`、`manifest/i18n/` 等插件通用资源目录。这使源码插件和动态插件无法用同一个稳定能力读取自身打包的完整 `manifest/` 原始资源，也让历史 `metadata` 语义与插件目录规范长期并存。

项目是全新项目，不需要保留旧兼容层。本次变更将插件资源读取统一为 `Manifest()`，把 `metadata.yaml` 降级为普通可选资源，并完整处理历史设计中关于路径安全、动态授权、active release 绑定、SQL/i18n/config 专用生命周期管线和历史命名的所有影响。

## What Changes

- **BREAKING**：移除任何插件可见的旧 `Metadata` 服务、`metadata` host service、`metadata.get` 或等价公开读取语义；插件统一使用 `Manifest()` / `manifest.get`。
- **BREAKING**：`Manifest()` 不再被定义为“声明型 YAML 元数据读取器”，而是插件自有 `manifest/` 目录的只读原始资源视图。
- **BREAKING**：源码插件可通过 `HostServices.Manifest()` 读取自身嵌入或开发目录中 `manifest/` 下任意真实文件，包括 `config/`、`sql/`、`i18n/`，但仍不得跨插件、跨目录、读取宿主 manifest、读取 URL 或读取任意文件系统路径。
- **BREAKING**：动态插件可通过 `manifest.get` 读取当前 active release artifact 携带且在 `plugin.yaml` 的 `hostServices` 授权快照中声明的 `manifest/` 相对路径，包括 `config/`、`sql/`、`i18n/`。
- 保留 `metadata.yaml` 的历史处理逻辑，但仅作为 `manifest/metadata.yaml` 下的普通可选资源；系统不得要求插件提交占位 `metadata.yaml`。
- 保留 SQL、i18n 和 config 的专用生命周期管线：`Manifest()` 只提供只读原文，不执行 SQL、不加载翻译、不覆盖运行期配置解析和生效逻辑。
- 更新动态插件打包和 active release 资源视图，使 `manifest/` 下实际打包资源按原始路径投影给 `Manifest()`，不再只投影 `*.yaml` 或排除专用目录。
- 清理历史命名、注释、测试和规格中把 `Manifest()` 等同于 `metadata` 或“声明型资源”的表述，避免新插件继续依赖旧概念。

## Capabilities

### New Capabilities

- 无。

### Modified Capabilities

- `plugin-manifest-lifecycle`：修改 `HostServices.Manifest()` 的读取范围、路径治理和专用目录语义，移除旧 `Metadata` 服务语义。
- `plugin-runtime-loading`：修改动态插件 active release 资源视图，要求完整投影 artifact 中 `manifest/` 下的授权资源。
- `plugin-embed-snapshot-packaging`：修改动态插件打包资源语义，要求完整保留 `manifest/` 下资源路径，不再以 `metadata.yaml` 或声明型 YAML 为中心。

## Impact

- 影响 `apps/lina-core/pkg/plugin/capability/manifest` 的源码插件 manifest 读取实现、接口注释和测试。
- 影响 `apps/lina-core/pkg/plugin/capability/guest` 的动态插件 guest manifest client 注释和测试语义。
- 影响 `apps/lina-core/pkg/plugin/pluginbridge/internal/hostservice` 和 `protocol` 中 `manifest` host service 的授权路径校验。
- 影响 `apps/lina-core/internal/service/plugin/internal/runtime` 的 dynamic artifact manifest resource projection。
- 影响 `apps/lina-core/internal/service/plugin/internal/wasm` 的 `manifest.get` host call 授权和执行测试。
- 影响动态插件构建、artifact 解析、资源计数和示例插件 `plugin.yaml` 中 `hostServices` 的 `manifest` 资源声明。
- 不新增 HTTP API、数据库表、DAO、前端页面或运行时写入能力。
- `i18n` 影响：允许 `Manifest()` 读取插件打包的 `manifest/i18n/` 原文，但不改变 i18n 资源发现、加载、缓存失效和翻译治理边界。
- 缓存影响：动态插件 manifest 资源视图仍绑定 active release checksum 或 generation，并沿用 `plugin-runtime` 缓存失效机制；不新增独立缓存域。
- 数据权限影响：不新增业务数据读写、列表、详情、导出、聚合、租户或组织数据可见性逻辑。
- SQL 影响：不新增或修改 SQL 迁移；允许读取 SQL 文件原文但不改变插件 SQL 执行、账本、事务和方言转译管线。
