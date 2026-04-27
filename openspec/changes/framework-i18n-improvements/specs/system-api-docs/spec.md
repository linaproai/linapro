## ADDED Requirements

### Requirement: 接口文档翻译资源加载必须复用统一的 ResourceLoader
系统 SHALL 让 `apidoc` 翻译资源的加载链路通过 `pkg/i18nresource` 包提供的统一 `ResourceLoader` 完成,不得在 `apidoc` 包内维护一份独立的"宿主嵌入资源 → 源码插件嵌入资源 → 动态插件运行时资源"遍历实现,也不得为复用加载器而反向依赖 `internal/service/i18n`。`apidoc` 链路 MUST 通过 `ResourceLoader` 的配置参数声明 `Subdir = "manifest/i18n/apidoc"` 与 `PluginScope = RestrictedToPluginNamespace`,并保留多文件、层级 JSON 和扁平 dotted key 三种维护方式;系统在合并时仍归一化为稳定结构化 key。

#### Scenario: apidoc 与运行时 bundle 共用资源加载器实现
- **WHEN** 系统加载某语言的 `apidoc` 翻译资源
- **THEN** 加载流程通过 `i18nresource.ResourceLoader` 完成宿主嵌入资源、源码插件嵌入资源、动态插件运行时资源的发现与合并
- **AND** `apidoc` 包内不存在与 `i18n` 包重复的目录遍历或 `wasm` 解析逻辑

#### Scenario: 插件命名空间隔离仍然生效
- **WHEN** 源码插件的 `apidoc` 翻译资源声明键 `plugins.<plugin-id>.routes.*`
- **THEN** 该插件的资源只被允许贡献以 `plugins.<plugin-id>.` 为前缀的键
- **AND** 系统忽略其他插件命名空间或宿主命名空间下的键

### Requirement: 接口文档必须支持阿拉伯语展示
系统 SHALL 在新增 `ar-SA` 作为内置语言后,支持系统接口文档以阿拉伯语展示。宿主与所有源码插件、动态插件 MUST 提供 `manifest/i18n/apidoc/ar-SA.json` 或 `manifest/i18n/apidoc/ar-SA/**/*.json` 翻译资源,确保 `/api.json?lang=ar-SA` 返回的接口分组、摘要、描述、参数说明均按阿拉伯语展示。`en-US` 接口文档继续直接使用 API DTO 中的英文源文案,不依赖 `ar-SA` 资源。

#### Scenario: 阿拉伯语环境加载接口文档
- **WHEN** 管理员在 `ar-SA` 环境下打开系统接口文档,或请求 `/api.json?lang=ar-SA`
- **THEN** 宿主、源码插件与动态插件的路由分组、接口摘要、接口描述、请求参数与响应参数说明按阿拉伯语展示
- **AND** 缺失翻译时回退到 API DTO 中维护的英文源文案,不显示空白或翻译键
