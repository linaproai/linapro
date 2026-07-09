## Why

LinaPro 的国际化必须作为宿主级基础设施，而不是由默认工作台、业务模块和插件分别维护语言解析、资源加载、错误文案和动态元数据投影。此前菜单、字典、系统参数、系统信息、插件元数据、后端错误、导入导出内容和工作台页面仍存在单语言文本、硬编码中文、中英混杂、缓存失效粒度粗和资源真源不清的问题。

接口文档页面在接口数量较多时缺少加载状态反馈，用户看到空白页面易误判为加载失败；中文语言环境下部分接口标题仍显示英文，说明 `zh-CN` 的接口文档本地化资源覆盖存在缺口——OpenAPI 操作级本地化键推断在 GET/DELETE 等静态接口共享 `dc` 描述时产生歧义删除，以及启用 `i18n` 的源码插件路径被误识别为动态插件路径。

默认交付还曾同时维护简体中文、繁体中文和英文三套资源，放大了宿主、插件、前端运行时语言包和`API`文档资源的同步成本。框架默认基线只需要长期维护`zh-CN`与`en-US`，同时保留项目按资源目录和默认配置扩展新语言的能力。

## What Changes

- 建立静态界面文案、后端动态元数据、业务内容三层国际化模型，明确业务模块在自身边界维护投影规则。
- 将翻译资源收敛为文件单一事实来源：宿主、源码插件和动态插件通过`manifest/i18n/<locale>/`维护运行时资源，通过`manifest/i18n/<locale>/apidoc/**/*.json`维护`API`文档资源。
- 统一语言解析优先级：`lang`查询参数、`Accept-Language`请求头、默认语言；不可用语言回退到默认语言。
- 默认内置语言收敛为`zh-CN`与`en-US`，中文浏览器语言标签统一映射到`zh-CN`，默认工作台固定`ltr`方向。
- 提供运行时翻译包和语言列表接口，聚合宿主、项目资源、源码插件和启用中的动态插件资源，并支持`ETag`与`If-None-Match`协商。
- 将运行时翻译缓存按"语言 x 扇区"分层，失效必须显式传入作用域，单值翻译热路径不得克隆整包消息。
- 将`i18n.Service`拆分为`LocaleResolver`、`Translator`、`BundleProvider`和`Maintainer`，并抽取共享`ResourceLoader`复用运行时 UI 与`apidoc`资源加载。
- 修复接口文档页面加载体验：iframe 内新增 Loading 占位，按`lang`参数维护`zh-CN`和`en-US`文案，Stoplight Elements 完成渲染后自动隐藏。
- 修复接口文档中文标题本地化：静态接口共享相同`dc`描述时不再歧义删除索引，改为每个静态路由记录一次 DTO 稳定 key；源码插件路径优先使用 DTO 稳定 key，不因路径位于`/x/<plugin-id>/...`而回退到动态插件路径派生键。
- 建立结构化业务错误、消息分类、导入导出本地化、插件桥接错误和硬编码文案扫描治理。
- 本地化默认管理工作台启动语言、运行时语言包加载、公共前端配置、动态菜单、路由标题、英文布局和项目定位文案。

## Capabilities

### New Capabilities

- `i18n-infrastructure`
- `message-governance`
- `other-module-localization`
- `management-workbench-i18n`
- `project-positioning-governance`
- `readme-localization-governance`

### Modified Capabilities

- `framework-i18n-foundation`
- `plugin-governance`

## Impact

- 影响宿主 i18n 服务、运行时翻译包、`bizerr`、导入导出、前端语言切换、插件语言资源生命周期、`API`文档本地化、接口文档页面加载体验和 README 镜像治理。
- 接口文档本地化修复影响`apidoc` service 内部构建逻辑：静态路由 DTO 稳定 key 记录、`dc`歧义消除和源码插件路径识别；不改变业务 API 契约或 OpenAPI 响应结构。
- 交叉影响菜单、字典、配置、定时任务、系统信息、系统 API 文档、工作台首页、登录页、演示写保护和数据库初始化命令；这些能力的完整契约由对应 owner 分组或`openspec/specs`承载。
- 默认交付不再提供`zh-TW`专项资源与验收基线；项目仍可按资源目录约定自行新增语言。
