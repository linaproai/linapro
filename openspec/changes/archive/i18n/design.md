# Design

## Resource Model

国际化按来源和生命周期分三层治理：

- 静态界面文案由默认管理工作台和共享前端包的本地资源维护。
- 动态元数据由后端按请求语言投影，覆盖菜单、字典、系统参数、系统信息、角色、任务、插件元数据等宿主治理数据。
- 业务内容由所属业务模块自行设计多语言内容模型，i18n 基础服务不反向感知业务实体。

宿主、源码插件和动态插件都以文件作为翻译资源单一事实来源。运行时 UI 资源位于`manifest/i18n/<locale>/`，`API`文档资源位于`manifest/i18n/<locale>/apidoc/**/*.json`。宿主不创建或依赖`sys_i18n_locale`、`sys_i18n_message`、`sys_i18n_content`等运行时翻译持久化表。

动态元数据翻译键从稳定业务锚点推导，例如`menu.<menu_key>.title`、`dict.<dict_type>.<value>.label`、`config.<config_key>.name`、`plugin.<plugin_id>.description`、`role.builtin.<key>.name`和`error.<domain>.<case>`。源码拥有的翻译命名空间通过显式注册表声明，缺失检查不能在 i18n 基础服务里硬编码业务模块名单。

## Locale Governance

请求语言由`LocaleResolver`统一解析，优先级为`lang`查询参数、`Accept-Language`请求头和默认配置中的`i18n.default`。解析结果写入业务上下文，供控制器、服务、插件桥接、导入导出和运行时翻译包聚合共用。

默认交付语言收敛为`zh-CN`与`en-US`。新增语言只需补充宿主、插件、`apidoc`与前端资源，并按需更新默认配置元数据，不得新增 Go 枚举、SQL seed 或前端硬编码语言清单。所有中文浏览器语言标签默认映射到`zh-CN`，默认工作台固定`ltr`方向；`i18n.enabled=false`时隐藏语言切换并只使用默认语言。

## Runtime Bundle And Cache

运行时翻译包接口按语言返回宿主、项目资源、源码插件和启用中的动态插件聚合消息，并输出前端可直接消费的嵌套对象。语言列表接口返回多语言开关、默认语言标记、显示名、原生名和固定`ltr`方向。

翻译包响应使用`ETag: "<locale>-<bundleVersion>"`与`Cache-Control: private, must-revalidate`。前端携带匹配`If-None-Match`时返回`304 Not Modified`且不带消息体。任何扇区缓存失效都会推动 bundle version 递增。

运行时缓存按"语言 x 扇区"分层：host、source-plugin、dynamic-plugin 和 merged。失效必须显式传入语言、扇区、插件 ID 或业务内容范围，不允许普通业务路径清空所有语言和所有扇区。`Translate`、`TranslateSourceText`、`TranslateOrKey`等单值热路径命中缓存时只读锁查找，不克隆整包消息；只有导出或构建完整消息集合的方法才允许克隆。

## Service Interfaces And Shared Loader

`i18n.Service`拆分为四类接口：

- `LocaleResolver`负责请求语言和上下文语言。
- `Translator`负责翻译查找与错误本地化。
- `BundleProvider`负责运行时翻译包、语言列表和版本输出。
- `Maintainer`负责导出、缺失检查、资源诊断和缓存失效。

业务模块按最小依赖声明接口，测试只需 mock 实际使用的能力。共享`pkg/i18nresource.ResourceLoader`统一宿主、源码插件、动态插件和`apidoc`资源遍历；运行时 UI 与`apidoc`通过不同配置实例复用同一加载器。动态插件`WASM`自定义 section 读取由`pkg/pluginbridge`提供，i18n、apidoc 和插件运行时不各自维护解析实现。

`framework-i18n-foundation`的最终需求全文与当前主规范一致，归档分组只保留本节中的架构原因、服务拆分和资源治理摘要，不再保存重复`spec.md`副本。

## Message Governance

运行时字符串分为用户消息、用户交付物、用户投影、开发者诊断、运维日志和用户数据。用户消息、交付物和投影必须通过运行时资源或后端本地化投影输出；开发者诊断和运维日志使用稳定英文与结构化字段；用户输入和外部系统文本保持原值。

宿主与插件通过`bizerr`输出结构化业务错误，统一携带`code`、`message`、`errorCode`、`messageKey`和`messageParams`。调用端以`errorCode`或`messageKey`判断语义，前端优先用`messageKey/messageParams`渲染，无法渲染时使用后端已本地化的`message`。

导入导出在请求入口解析语言并复用翻译结果，表头、工作表名、枚举文本和失败原因不得在行循环中重复构建语言包。插件桥接、WASM host call、host service 和插件生命周期错误返回稳定错误码或消息键，协议层默认诊断使用英文。

## API Documentation Localization

`/api.json`的请求语言上下文通过`lang`参数和`Content-Language`接入。API 文档本地化使用独立的`manifest/i18n/<locale>/apidoc/**/*.json`资源，与运行时 UI 翻译资源解耦。`en-US`直接使用英文源文本，`manifest/i18n/en-US/apidoc/**/*.json`文件保持空对象占位。

### Operation Key Resolution

OpenAPI 操作级本地化键推断遵循以下优先级：

1. 静态路由：优先使用请求 DTO 对应的稳定 key（通过构建阶段为每个静态路由记录一次 DTO key 实现）。即使多个静态接口共享相同的`g.Meta dc`描述，也不因歧义删除索引。内部`x-lina-apidoc-operation-key`字段仅用于本地化并在返回文档前删除。
2. 源码插件路径：接口挂载在`/x/<plugin-id>/...`命名空间下时，优先使用 DTO 稳定 key 和插件自有`apidoc`翻译资源，不因路径位于该命名空间而回退到动态插件路径派生键。
3. 动态插件路径：由路径派生 key。

### API Docs Loading Experience

接口文档页面由`apps/lina-vben/apps/web-antd/public/stoplight/apidocs.html`中独立 iframe 承载。iframe 内新增 Loading 占位，按`lang`参数维护`zh-CN`和`en-US`文案（不复用前端运行时语言包，保持接口文档资源隔离）。Stoplight Elements 完成侧边栏内容渲染后自动隐藏 Loading。长耗时增加提示；脚本加载失败增加错误提示。

## Default Workspace

默认管理工作台首次启动在没有已保存偏好和显式初始化语言时使用浏览器首选语言；中文类标签映射到`zh-CN`，其他语言默认`en-US`。前端运行时翻译包持久化到本地缓存，页面加载或语言切换时先用可用缓存渲染，再后台通过`ETag`协商刷新。

语言切换必须同步刷新静态语言包、运行时翻译包、公共前端配置、动态菜单、路由标题和嵌入式插件页面语言上下文。英文环境下，默认交付页面的标题、按钮、标签、表头、空态、系统生成节点、内置记录和确认弹窗不得残留中文；长英文文案通过布局、列宽、短文案或 tooltip 保持可读。

## Project And Document Governance

LinaPro 统一定位为"面向可持续交付的`AI`原生全栈框架"。仓库入口文档、系统信息页、`API`文档元数据和脚本输出都必须保持同一定位，不把默认管理工作台描述为唯一产品边界。

目录级主说明文档必须同时维护英文`README.md`和中文镜像`README.zh-CN.md`，二者结构与技术事实保持一致。OpenSpec 语言规则与 README 镜像规则共同保证 AI 读取文档时不会在不同入口得到冲突定位。

## Cross-Domain Impacts

- `config-management`承载配置名称、备注、导入导出表头和公共前端配置的本地化投影，当前契约由`openspec/specs/config-management/spec.md`承载，历史 owner 为`archive/system-config`。
- `menu-management`承载菜单标题、按钮标题、动态 route permission 和菜单治理的本地化投影，当前契约由`openspec/specs/menu-management/spec.md`承载，历史 owner 为`archive/user-management`。
- `dict-management`承载字典类型、字典项、`DictTag`和内置保护记录的本地化投影，当前契约由`openspec/specs/dict-management/spec.md`承载，历史 owner 为`archive/org-structure`。
- `cron-job-management`承载内置任务、任务组、执行日志元数据和执行确认治理，当前契约由`openspec/specs/cron-job-management/spec.md`承载，历史 owner 为`archive/scheduled-jobs`。
- `system-api-docs`承载英文 OpenAPI 源文本、独立`apidoc`翻译资源、API 文档定位和接口文档页面加载体验，当前契约由`openspec/specs/system-api-docs/spec.md`承载，历史 owner 为`archive/system-governance`。
- `system-info`承载系统信息页项目介绍和组件描述本地化，当前契约由`openspec/specs/system-info/spec.md`承载，历史 owner 为`archive/system-governance`。
- `dashboard-workbench`和`login-page-presentation`承载工作台首页和登录页可观察体验，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 分别为`archive/plugin-governance`与`archive/user-management`。
- `database-bootstrap-commands`承载`init`/`mock`显式确认和首错即停，当前契约由`openspec/specs/database-bootstrap-commands/spec.md`承载，历史 owner 为`archive/database-engine`。
- `core-host-boundary-governance`和`demo-control-guard`承载宿主边界与演示写保护，当前契约由`openspec/specs/<capability>/spec.md`承载，历史 owner 为`archive/host-plugin-boundary`。

## Risks And Boundaries

- 默认移除`zh-TW`降低即装即用语言覆盖，但能显著降低内建能力和插件示例维护成本；需要繁体中文的项目按资源目录自行扩展。
- 文件资源单一事实来源放弃运行时数据库热修翻译能力，但提升审计、缺失检查和交付可重复性。
- 分层缓存增加实现复杂度，但避免任意插件或资源变更触发全量清缓存，并为集群失效保留明确作用域。
- i18n 基础服务不得为了方便投影而依赖业务实体；业务模块必须在自身边界维护翻译键派生和保护规则。
- 接口文档 operation key 消除`dc`歧义删除后，内部标记字段`x-lina-apidoc-operation-key`必须在返回文档前清除，避免泄露实现细节。
