## 新增需求

### 需求：API 文档翻译资源加载必须复用统一的 ResourceLoader
系统 SHALL 让 `apidoc` 翻译资源加载管线通过 `pkg/i18nresource` 包提供的统一 `ResourceLoader` 完成，不得在 `apidoc` 包中维护独立的"宿主嵌入资源 -> 源码插件嵌入资源 -> 动态插件运行时资源"遍历实现，也不得反向依赖 `internal/service/i18n` 来复用加载器。`apidoc` 管线必须通过 `ResourceLoader` 配置参数声明 `Subdir = "manifest/i18n"`、`LocaleSubdir = "apidoc"`、`LayoutMode = LocaleSubdirectoryRecursive` 和 `PluginScope = RestrictedToPluginNamespace`，并保留多文件、分层 JSON 和扁平点分键三种维护方式；系统在合并时仍规范化为稳定结构化键。

#### 场景：apidoc 和运行时包共享资源加载器实现
- **当** 系统为某语言加载 `apidoc` 翻译资源时
- **则** 加载过程通过 `i18nresource.ResourceLoader` 完成宿主嵌入资源、源码插件嵌入资源和动态插件运行时资源的发现和合并
- **且** `apidoc` 包与 `i18n` 包相比不存在重复的目录遍历或 `wasm` 解析逻辑

#### 场景：插件命名空间隔离仍然生效
- **当** 源码插件的 `apidoc` 翻译资源声明了 `plugins.<plugin-id>.routes.*` 键时
- **则** 该插件的资源仅允许贡献以 `plugins.<plugin-id>.` 为前缀的键
- **且** 系统忽略来自其他插件命名空间或宿主命名空间的键

### 需求：API 文档必须支持繁体中文显示
系统 SHALL 在添加 `zh-TW` 作为内置语言后支持系统 API 文档以繁体中文显示。宿主和所有源码插件、动态插件必须提供 `manifest/i18n/zh-TW/apidoc/**/*.json` 翻译资源，确保 `/api.json?lang=zh-TW` 返回的 API 分组、摘要、描述和参数描述均以繁体中文显示。`en-US` API 文档继续直接使用 API DTO 中的英文源文本，不依赖 `zh-TW` 资源。

#### 场景：在繁体中文环境加载 API 文档
- **当** 管理员在 `zh-TW` 环境下打开系统 API 文档，或请求 `/api.json?lang=zh-TW` 时
- **则** 宿主、源码插件和动态插件的路由分组、API 摘要、API 描述、请求参数和响应参数描述均以繁体中文显示
- **且** 翻译缺失时回退到 API DTO 中维护的英文源文本，不显示空白或翻译键
