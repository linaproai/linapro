## ADDED Requirements

### Requirement: 宿主翻译服务接口必须按职责拆分为多个小接口
宿主系统 SHALL 把 i18n 翻译服务接口拆分为 `LocaleResolver`、`Translator`、`BundleProvider`、`Maintainer` 四个小接口。每个小接口 MUST 仅承担一类职责:`LocaleResolver` 解析请求语言与上下文语言;`Translator` 提供翻译查找与错误本地化;`BundleProvider` 输出运行时翻译包与语言列表;`Maintainer` 提供资源导出、缺失检查、来源诊断与缓存失效。`Service` 类型 MUST 是这四个小接口的组合,业务模块持有的 `i18nSvc` 字段类型 SHALL 收敛为实际依赖的最小接口而非整个 `Service`。宿主核心 i18n 服务不得提供数据库翻译覆写导入或通用业务内容多语言持久化接口。

#### Scenario: 业务模块只声明实际依赖的小接口
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` 等模块需要本地化翻译能力
- **THEN** 模块在自身结构体内将 `i18nSvc` 字段声明为 `LocaleResolver` 与 `Translator` 的最小组合
- **AND** 模块单元测试可以仅 mock 这两个小接口而不必 stub 维护类方法

#### Scenario: 控制器测试通过最小接口替身完成
- **WHEN** 测试 `i18n` 管理类控制器(导出、诊断、缺失检查)
- **THEN** 测试可以单独 mock `Maintainer` 接口
- **AND** 不需要同时为 `Translator` / `BundleProvider` 提供占位实现

### Requirement: 宿主必须通过资源约定与默认配置发现内置语言
宿主系统 SHALL 从 `manifest/i18n/<locale>.json` 文件自动发现内置运行时语言,并通过默认配置文件中的 `i18n` 配置段维护默认语言、多语言开关、展示排序和原生名等少量不可由文件名推导的元数据。新增内置语言时 MUST 只需要新增对应语言的运行时 JSON、apidoc JSON、插件 JSON 和可选默认配置元数据,不得要求新增后端 Go 语言枚举、SQL seed 或前端 TS 语言清单。运行时文本方向按当前宿主约定固定为 `ltr`,不得要求在配置中维护 `direction`。运行时语言列表、缺失翻译检查、资源来源诊断、运行时翻译包接口、ETag 协商与前端持久化缓存 MUST 自动覆盖 `zh-TW`。

#### Scenario: 启动后繁体中文由资源文件自动加入运行时语言列表
- **WHEN** 项目存在 `manifest/i18n/zh-TW.json`
- **AND** 默认配置文件 `i18n.locales` 中为 `zh-TW` 提供 `nativeName`
- **AND** 服务启动后前端请求运行时语言列表
- **THEN** `/i18n/runtime/locales` 接口返回的语言列表中包含 `zh-TW`
- **AND** `zh-TW` 标记为非默认语言
- **AND** `zh-TW` 的方向字段为固定值 `ltr`

#### Scenario: 加新语言时不需要修改 Go、SQL 或前端 TS 语言清单
- **WHEN** 交付项目新增某语言的 `manifest/i18n/<locale>.json` 与 `manifest/i18n/apidoc/<locale>/**/*.json` 资源
- **AND** 源码插件与动态插件按同一目录约定新增该语言资源
- **AND** 如需控制默认语言、排序、原生名或启停状态,仅修改默认配置文件中的 `i18n` 配置段
- **THEN** 菜单、字典、配置、定时任务、插件、角色、系统信息等动态元数据自动按该语言返回本地化结果
- **AND** 运行时语言列表自动包含该语言
- **AND** 不需要修改后端 Go 常量、SQL seed、前端 `SUPPORT_LANGUAGES` 或第三方 locale switch 分支

#### Scenario: 关闭多语言后只使用默认语言
- **WHEN** 默认配置文件中 `i18n.enabled` 为 `false`
- **AND** 用户浏览器或请求参数传入非默认语言
- **THEN** 宿主请求语言解析结果回退到 `i18n.default`
- **AND** `/i18n/runtime/locales` 响应标记多语言开关为关闭,并仅返回默认语言描述
- **AND** 默认管理工作台隐藏语言切换按钮,按默认语言加载静态语言包、运行时翻译包和公共前端配置

#### Scenario: 从 locales 列表移除语言即停用该语言
- **WHEN** 项目存在多个 `manifest/i18n/<locale>.json` 资源
- **AND** 默认配置文件中 `i18n.locales` 只列出其中一部分语言
- **THEN** `/i18n/runtime/locales` 只返回 `i18n.locales` 中列出的语言
- **AND** 请求未列出的语言时按 `i18n.default` 回退

### Requirement: 默认管理工作台必须保持固定 LTR 文档方向
默认管理工作台 SHALL 按当前宿主约定固定使用 `ltr` 文档方向。语言切换时,工作台 MUST 同步设置 `<html dir>` 为 `ltr`,并把 `direction="ltr"` 注入 `Ant Design Vue` 的 `ConfigProvider`。前端不得维护静态 RTL 语言注册表,也不得要求新增语言时修改方向相关 TypeScript 分支。

#### Scenario: 切换到繁体中文时 html 方向仍为 LTR
- **WHEN** 用户在默认管理工作台中将语言切换为 `zh-TW`
- **THEN** `document.documentElement` 的 `dir` 属性保持为 `ltr`
- **AND** `Ant Design Vue` 的 `ConfigProvider` 接收到 `direction="ltr"`
- **AND** 切换回 `zh-CN` 或 `en-US` 时 `dir` 仍保持为 `ltr`

#### Scenario: 繁体中文下页面文案完整即视为合格
- **WHEN** 用户在繁体中文环境下打开框架默认交付的列表页、抽屉与弹窗
- **THEN** 页面文案按繁体中文展示且布局未阻断核心操作
- **AND** 不要求提供 RTL 镜像布局

### Requirement: 翻译资源加载器必须在宿主与插件、UI 与 apidoc 之间共用
宿主系统 SHALL 在 `pkg/i18nresource` 包内提供统一的 `ResourceLoader` 组件,接受 `Subdir`、`PluginScope`、`LayoutMode` 等配置参数,集中实现"宿主嵌入资源 → 源码插件资源 → 动态插件资源"的发现与加载逻辑。运行时 UI 翻译资源加载与 API 文档翻译资源加载 MUST 通过不同 `ResourceLoader` 实例完成,不得各自维护重复实现,也不得让 API 文档模块为复用加载器而反向依赖运行时 `internal/service/i18n` 包。源码插件的 apidoc 命名空间隔离 MUST 由 `ResourceLoader` 配置而非额外重复代码完成。

#### Scenario: 运行时 bundle 与 apidoc 共享同一资源加载器实现
- **WHEN** 系统加载运行时 UI 翻译资源或 apidoc 翻译资源
- **THEN** 两条链路通过同一份 `i18nresource.ResourceLoader` 实现完成宿主、源码插件与动态插件的资源遍历
- **AND** apidoc 链路通过 `PluginScope=RestrictedToPluginNamespace` 配置约束插件命名空间
- **AND** 运行时 UI 链路通过 `PluginScope=Open` 配置允许插件贡献任意键

## MODIFIED Requirements

### Requirement: 宿主必须提供运行时翻译包分发能力
宿主系统 SHALL 提供运行时翻译包接口与语言列表接口,按语言返回聚合后的消息资源以及当前可用的语言描述信息,供默认管理工作台和宿主嵌入式插件页面加载。运行时翻译包 MUST 能同时包含宿主、源码插件和当前已启用动态插件的国际化消息,并在输出时转换为前端可直接消费的嵌套消息对象。运行时翻译包接口 MUST 在响应中输出 `ETag` 头,值由当前语言与运行时翻译包版本派生且在版本变化时必然不同;系统 MUST 接收请求中的 `If-None-Match` 头,匹配时返回 `304 Not Modified` 且不携带消息体。任何扇区缓存失效 MUST 触发运行时翻译包版本自增,确保同语言下不同 bundle 内容拥有不同 ETag。

#### Scenario: 默认工作台加载运行时翻译包
- **WHEN** 前端以 `zh-CN` 请求运行时翻译包
- **THEN** 宿主返回该语言下的聚合消息集合
- **AND** 结果中包含宿主资源、源码插件资源以及已启用动态插件资源的合并结果
- **AND** 响应中包含 `ETag` 头

#### Scenario: 前端获取宿主支持的语言列表
- **WHEN** 前端请求运行时语言列表
- **THEN** 宿主返回多语言开关、当前支持的语言编码、默认语言标记、展示名称、原生名称与固定 LTR 文本方向
- **AND** 展示名称按当前请求语言返回,原生名称保持为对应语言自身文案
- **AND** 列表中包含 `zh-CN`、`en-US` 与 `zh-TW`

#### Scenario: 运行时语言包支持层级维护并保持扁平治理
- **WHEN** 宿主从文件、源码插件或动态插件加载翻译资源
- **THEN** 宿主允许运行时 UI 文件资源使用层级 JSON 或扁平 dotted key 编写
- **AND** 宿主内部统一以扁平 key 形式聚合消息
- **AND** 运行时接口返回给前端的结果为嵌套对象结构,以便直接并入前端 `vue-i18n` 消息树

#### Scenario: 插件停用后不再暴露其翻译资源
- **WHEN** 某个插件被停用或卸载
- **THEN** 后续运行时翻译包结果中不再包含该插件贡献的翻译消息
- **AND** 其他宿主与已启用插件资源保持可用
- **AND** 系统对应触发该插件相关扇区的缓存失效,运行时翻译包版本自增

#### Scenario: 同一翻译包二次请求返回 304
- **WHEN** 前端首次请求运行时翻译包后保留响应中的 `ETag`
- **AND** 后端在两次请求之间未发生缓存失效
- **AND** 前端在第二次请求中携带 `If-None-Match` 等于上次的 `ETag`
- **THEN** 后端返回 `304 Not Modified` 且不携带消息体

### Requirement: 宿主必须提供国际化维护与校验能力
宿主系统 SHALL 提供翻译资源导出、缺失翻译检查和资源来源诊断能力,以支持交付项目在开发期维护多语言资源。交付项目和插件新增国际化资源时 MUST 遵循统一目录约定和翻译键规范。缓存失效 MUST 提供按"语言 × 扇区(host / source-plugin / dynamic-plugin)"维度的精细控制,不得在任意单点变更时整张清空所有语言所有扇区的缓存。所有"该键由代码源拥有"的判定 MUST 通过命名空间显式注册表完成,系统不得在 i18n 包内硬编码任何具体业务模块的命名前缀。宿主核心不得创建或依赖 `sys_i18n_locale`、`sys_i18n_message`、`sys_i18n_content` 三张运行时 i18n 持久化表;翻译内容的唯一事实源是开发期 JSON/YAML 资源。

#### Scenario: 导出某语言的翻译资源
- **WHEN** 管理员或交付工具请求导出 `en-US` 的翻译资源
- **THEN** 系统返回该语言的聚合消息结果
- **AND** 导出内容能够用于离线校对并回写到对应 JSON 资源文件

#### Scenario: 检查缺失翻译键
- **WHEN** 管理员或交付工具执行缺失翻译检查
- **THEN** 系统返回当前语言相对默认语言缺失的翻译键列表
- **AND** 结果中能够定位缺失键所属的宿主模块、项目资源或插件资源范围
- **AND** 注册为代码源命名空间的翻译键不出现在缺失结果中

#### Scenario: 插件资源变更仅清除受影响的插件扇区
- **WHEN** 某动态插件被启用、停用、卸载或升级
- **THEN** 系统仅清除该插件相关的动态插件扇区缓存与合并视图
- **AND** 其他语言、宿主资源与未受影响插件资源继续保持有效
- **AND** 该语言的运行时翻译包版本自增,前端下次协商必能感知
