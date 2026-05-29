## Context

现有插件资源体系已经有 `HostServices.Manifest()` / `manifest.get`，但历史规格和实现仍将其定位为 `metadata.yaml` 和声明型 YAML 的读取器。源码插件实现会拒绝 `config/`、`sql/`、`i18n/`，动态插件授权和 active release 资源投影也做了同样过滤。

这与插件目录规范不一致。插件通用资源目录明确要求所有插件维护 `plugin.yaml` 和 `manifest/`，其中 `manifest/sql/`、`manifest/i18n/`、`manifest/config/` 都是插件自身打包资源。源码插件作为随宿主编译的受信任扩展，应能读取自身嵌入资源；动态插件则应能读取自己 artifact 中声明并被宿主授权的资源。

本次变更不考虑兼容旧插件 API。历史 `Metadata` 概念只作为需要清理的旧命名处理，不提供 deprecated alias。

## Goals / Non-Goals

**Goals:**

- 将插件侧资源读取的唯一公开契约统一为 `Manifest()` / `manifest.get`。
- 删除或清理旧 `Metadata` 服务、`metadata` host service、`metadata.get` 或等价插件可见语义。
- 将 `metadata.yaml` 处理为 `manifest/metadata.yaml` 下的普通可选资源，保留“不要求占位文件”的历史设计。
- 允许源码插件读取自身 `manifest/` 下任意实际打包资源。
- 允许动态插件读取 active release artifact 中实际携带且经 `hostServices` 授权的 `manifest/` 资源。
- 保留 SQL、i18n 和 config 的专用生命周期管线，明确 `Manifest()` 只提供原文读取能力，不改变资源生效机制。
- 完整更新规格、注释、测试和示例，避免继续把 `Manifest()` 描述成 `Metadata` 或声明型 YAML 读取器。

**Non-Goals:**

- 不新增目录结构、插件资源目录或额外清单字段。
- 不新增 HTTP API、前端页面、数据库表或 DAO。
- 不改变插件 SQL 安装/卸载/mock 执行顺序、事务、账本或方言转译。
- 不改变 i18n 资源加载、缓存失效、语言治理或 apidoc 翻译治理。
- 不改变 `HostServices.Config()` 作为插件运行期有效配置读取入口的语义。
- 不处理宿主框架自身 `apps/lina-core/manifest/config/metadata.yaml` 的系统信息、OpenAPI、release tag 等宿主元数据逻辑。
- 不重命名与插件资源读取无关的内部 `Metadata` 用法，例如动态路由元数据、审计元数据、错误元数据、表元数据、cron 展示 metadata。

## Decisions

### 决策 1：`Manifest()` 成为插件 `manifest/` 原始资源视图

`Manifest()` 的 owner 是 `apps/lina-core/pkg/plugin/capability/manifest` 和 `pluginbridge` 的 `manifest` host service。调用方是源码插件和动态插件 guest。被调用方是宿主插件资源读取适配器和动态 active release resource view。

新契约如下：

- 路径参数始终是相对当前插件 `manifest/` 根目录的 slash path。
- 返回值是文件原始字节；是否按 YAML、JSON、SQL 或文本解析由插件内部决定。
- `Scan` 只表达“对 YAML 资源做便捷绑定”，不意味着 `Manifest()` 只支持 YAML。
- `metadata.yaml` 没有特殊服务身份，只是可选文件。

备选方案是新增 `Resources()` 或 `ManifestFS()`。不采用，因为当前已有 `Manifest()` host service 和 guest client，职责名称与插件目录事实一致，新增并列服务会增加心智负担。

### 决策 2：不保留 `Metadata` 兼容层

任何插件可见的 `Metadata()`、`MetadataService`、`metadata.get`、`service: metadata` 或等价读取能力都必须删除或迁移到 `Manifest()`。若当前代码中没有独立公开 `Metadata` 服务，也必须通过静态检索确认没有遗留命名、注释、测试或示例继续表达旧服务。

备选方案是保留 deprecated alias。用户已明确项目无历史负担且不要考虑兼容性；保留 alias 会继续制造“双入口”并削弱框架契约清晰度。

### 决策 3：源码插件和动态插件采用不同信任边界

源码插件随宿主编译和交付，是受信任扩展。`HostServices.ForPlugin(id).Manifest()` 可读取该插件自身嵌入文件系统或开发目录中 `manifest/` 下的任意真实文件。限制只保留资源作用域安全：不能空根读取，不能绝对路径，不能 URL，不能 Windows drive path，不能路径穿越，不能以 `manifest/` 前缀重复表达路径，不能跨插件或读取宿主 manifest。

动态插件通过 WASM artifact 接入，仍必须使用 `plugin.yaml` 中 `hostServices` 的 `service: manifest`、`methods: [get]` 和 `resources.paths` 声明授权边界。授权快照来自宿主确认后的当前 active release，不允许绕过授权读取未声明路径。

备选方案是让动态插件只要声明 `service: manifest` 就能读取全部 `manifest/`。不采用，因为动态插件不是源码插件信任边界，资源读取仍属于 host service 授权面。

### 决策 4：`config/sql/i18n` 可读，但不能通过 `Manifest()` 生效

`Manifest()` 可以读取以下资源原文：

- `config/config.yaml`、`config/config.example.yaml`
- `sql/*.sql`、`sql/uninstall/*.sql`、`sql/mock-data/*.sql`
- `i18n/<locale>/*.json`、`i18n/<locale>/apidoc/**/*.json`
- 其他插件自定义 manifest 资源

但专用管线继续负责资源生效：

- 插件运行期有效配置仍通过 `HostServices.Config()` 读取。
- SQL 仍由插件生命周期、数据库、迁移账本和方言转译管线执行。
- i18n 仍由 i18n 资源发现、聚合和缓存失效管线加载。

这个设计把“读取原文”和“资源生效”拆开，避免因为开放读取能力而绕过治理。

### 决策 5：动态 active release 资源视图完整投影 `manifest/`

动态 artifact 解析和运行时资源视图必须保留 artifact 中 `manifest/` 下实际携带的资源，并以相对 `manifest/` 的路径投影给 `manifest.get`。不再按扩展名过滤为 `*.yaml`，也不再排除 `config/`、`sql/`、`i18n/`。

`HostServices.Config()` 读取默认配置时仍可复用 artifact 中的 `manifest/config/config.yaml` 作为默认配置来源，但这不是 `Manifest()` 的特殊分支。两者可以共享同一 active release resource snapshot，但语义不同。

### 决策 6：历史 `Metadata` 名称按语义分流处理

实现时必须逐项扫描 `Metadata` 命名，按以下规则处理：

| 类别 | 处理 |
| --- | --- |
| 插件资源读取服务、接口、host service、guest client、测试、注释中的 `Metadata` | 删除或迁移为 `Manifest` |
| `manifest/metadata.yaml` 文件示例 | 保留为普通可选资源示例，测试可继续覆盖但不能代表服务名 |
| 宿主 `apps/lina-core/manifest/config/metadata.yaml` | 保留，不属于插件 `Manifest()` |
| 插件发布快照同步 `SyncMetadata` | 若会造成歧义，重命名为 release/projection 同步语义；否则至少更新注释说明不是插件资源读取服务 |
| 动态路由、审计、错误、数据库表、OpenAPI schema metadata | 保留，不属于本变更 |
| `AddWithMetadata` cron 展示信息 | 保留，不属于插件 manifest 资源读取 |

## Risks / Trade-offs

- 动态插件授权声明变宽后误读敏感打包资源 → 动态插件仍必须显式声明 `resources.paths`，并由授权快照校验；源码插件受信任但仍限定自身插件目录。
- 插件作者误以为读取 `config/config.yaml` 等于读取运行期有效配置 → 文档、注释和测试明确 `Config()` 才是有效配置入口，`Manifest()` 只返回打包原文。
- 插件作者误以为读取 SQL 文件会触发执行 → 规格明确 SQL 执行只能由生命周期管线触发。
- 资源视图包含更多文件导致内存占用增加 → 动态 artifact 已经携带这些资源；本变更改变的是投影范围，不引入数据库查询或按请求扫描目录。若需要优化，应在 artifact 解析阶段复用 active release snapshot。
- 历史规格存在“专用目录不得被 Manifest 读取”的相反约束 → 本变更通过 `MODIFIED Requirements` 覆盖该约束，并保留“不得绕过专用管线”的治理目标。

## Migration Plan

1. 修改 OpenSpec 规格，明确 `Manifest()` 是完整 `manifest/` 只读资源视图。
2. 修改源码插件 manifest service 的路径规范化，移除 `config/sql/i18n` 排除逻辑，保留路径安全。
3. 修改动态插件 host service 授权路径规范化，允许 `config/sql/i18n`，保留路径安全和授权快照。
4. 修改动态 active release artifact resource projection，完整投影 `manifest/` 下资源。
5. 清理旧 `Metadata` 服务、注释、测试和示例命名；不提供兼容 alias。
6. 更新示例动态插件 `plugin.yaml`，使用 `service: manifest` 声明需要读取的资源。
7. 补充单元测试覆盖源码插件读取 `config/sql/i18n`、动态插件授权读取这些路径、未授权拒绝、路径穿越拒绝、artifact 资源投影完整性。
8. 运行变更包 Go 测试、静态检索、`openspec validate replace-plugin-metadata-with-manifest --strict` 和 `git diff --check`。

## 影响分析

- 宿主通用能力：本变更属于 `apps/lina-core` 插件宿主通用能力和插件扩展能力的契约收敛。
- 源码插件能力：源码插件读取自身 `manifest/` 原文能力扩大，但仍限定自身作用域。
- 动态插件能力：动态插件读取能力由 `hostServices` 授权快照控制，读取范围取决于当前 active release artifact。
- 工作台适配层和前端展示适配：无影响。
- 数据权限影响：无业务数据存在性、租户、组织或数据权限边界变化。
- 缓存一致性影响：复用既有 `plugin-runtime` active release 资源视图和 revision 失效机制；不新增缓存域。
- `i18n` 影响：可读取插件 i18n 原文，但不改变 i18n 资源加载、失效、翻译完整性和宿主/插件边界治理。
- SQL 影响：可读取插件 SQL 原文，但不改变 SQL 执行、迁移账本、事务或幂等规则。
- 开发工具跨平台影响：不修改脚本、CI、`linactl` 或构建入口；若实现阶段触及动态插件构建器，需要补充跨平台验证。
- DI 影响：预计不新增运行期依赖；继续复用启动期已配置的 manifest factory、WASM host service 和 active release resource snapshot。实现阶段若新增依赖，任务记录必须追溯 owner、创建位置、传递路径和共享实例策略。
