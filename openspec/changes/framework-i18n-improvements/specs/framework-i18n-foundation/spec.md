## ADDED Requirements

### Requirement: 宿主翻译服务接口必须按职责拆分为多个小接口
宿主系统 SHALL 把 i18n 翻译服务接口拆分为 `LocaleResolver`、`Translator`、`BundleProvider`、`ContentProvider`、`Maintainer` 五个小接口。每个小接口 MUST 仅承担一类职责:`LocaleResolver` 解析请求语言与上下文语言;`Translator` 提供翻译查找与错误本地化;`BundleProvider` 输出运行时翻译包与语言列表;`ContentProvider` 处理业务内容多语言;`Maintainer` 提供导入、导出、缺失检查、诊断与缓存失效。`Service` 类型 MUST 是这五个小接口的组合,业务模块持有的 `i18nSvc` 字段类型 SHALL 收敛为实际依赖的最小接口而非整个 `Service`。

#### Scenario: 业务模块只声明实际依赖的小接口
- **WHEN** `menu` / `dict` / `sysconfig` / `jobmgmt` 等模块需要本地化翻译能力
- **THEN** 模块在自身结构体内将 `i18nSvc` 字段声明为 `LocaleResolver` 与 `Translator` 的最小组合
- **AND** 模块单元测试可以仅 mock 这两个小接口而不必 stub 维护类方法

#### Scenario: 控制器测试通过最小接口替身完成
- **WHEN** 测试 `i18n` 管理类控制器(导入、导出、诊断、缺失检查)
- **THEN** 测试可以单独 mock `Maintainer` 接口
- **AND** 不需要同时为 `Translator` / `BundleProvider` 提供占位实现

### Requirement: 宿主必须默认启用阿拉伯语作为第三种内置语言
宿主系统 SHALL 在 `sys_i18n_locale` 表中以 seed DML 形式启用 `ar-SA` 作为第三种内置语言,且默认 `is_default=0`、`status=1`。运行时语言列表、缺失翻译检查、覆写来源诊断、运行时翻译包接口、ETag 协商、前端持久化缓存与基础 RTL 切换 MUST 自动覆盖 `ar-SA`,无需业务模块新增任何代码。宿主、源码插件与动态插件 MUST 同步提供 `ar-SA` 的运行时 UI 翻译资源与 apidoc 翻译资源,以使 `CheckMissingMessages` 在交付状态下不返回非空结果。

#### Scenario: 启动后阿拉伯语自动加入运行时语言列表
- **WHEN** 项目执行 `make init` 完成数据库初始化后启动服务
- **THEN** `/i18n/runtime/locales` 接口返回的语言列表中包含 `ar-SA`
- **AND** `ar-SA` 标记为非默认语言但启用状态

#### Scenario: 加新语言时不需要业务模块改代码
- **WHEN** 仅新增 `ar-SA` 的 `manifest/i18n/ar-SA.json` 与 `manifest/i18n/apidoc/ar-SA/*.json` 资源
- **AND** 仅在 `sys_i18n_locale` 中启用 `ar-SA`
- **THEN** 菜单、字典、配置、定时任务、插件、角色、系统信息等动态元数据自动按阿语返回本地化结果
- **AND** 业务模块代码、`i18n` 服务实现源码与翻译键约定均无需修改

### Requirement: 默认管理工作台必须随语言切换自动调整文档基础方向
默认管理工作台 SHALL 维护一个 RTL 语言注册表,目前包含 `ar-SA`,以便后续扩展 `he-IL`、`fa-IR` 等语言时仅需登记。语言切换时,工作台 MUST 同步设置 `<html dir>` 为 `rtl` 或 `ltr`,并把对应方向注入 `Ant Design Vue` 的 `ConfigProvider`,使组件库按当前方向工作。在 RTL 语言下,工作台的视觉镜像偏差(图标方向、抽屉滑出方向、表格固定列翻转、菜单展开方向)允许存在,但页面"内容正确、可用"。完整 RTL 设计语言不在本能力范围内。

#### Scenario: 切换到阿拉伯语时 html 方向变化
- **WHEN** 用户在默认管理工作台中将语言切换为 `ar-SA`
- **THEN** `document.documentElement` 的 `dir` 属性变为 `rtl`
- **AND** `Ant Design Vue` 的 `ConfigProvider` 接收到 `direction="rtl"`
- **AND** 切换回 `zh-CN` 或 `en-US` 时 `dir` 恢复为 `ltr`

#### Scenario: 阿拉伯语下页面内容可用即视为合格
- **WHEN** 用户在阿拉伯语环境下打开框架默认交付的列表页、抽屉与弹窗
- **THEN** 页面文案按阿语展示且布局未阻断核心操作
- **AND** 不要求图标、抽屉滑入方向或表格固定列做完全镜像

### Requirement: 翻译资源加载器必须在宿主与插件、UI 与 apidoc 之间共用
宿主系统 SHALL 在 `pkg/i18nresource` 包内提供统一的 `ResourceLoader` 组件,接受 `Subdir`、`PluginScope`、`LayoutMode` 等配置参数,集中实现"宿主嵌入资源 → 源码插件资源 → 动态插件资源"的发现与加载逻辑。运行时 UI 翻译资源加载与 API 文档翻译资源加载 MUST 通过不同 `ResourceLoader` 实例完成,不得各自维护重复实现,也不得让 API 文档模块为复用加载器而反向依赖运行时 `internal/service/i18n` 包。源码插件的 apidoc 命名空间隔离 MUST 由 `ResourceLoader` 配置而非额外重复代码完成。

#### Scenario: 运行时 bundle 与 apidoc 共享同一资源加载器实现
- **WHEN** 系统加载运行时 UI 翻译资源或 apidoc 翻译资源
- **THEN** 两条链路通过同一份 `i18nresource.ResourceLoader` 实现完成宿主、源码插件与动态插件的资源遍历
- **AND** apidoc 链路通过 `PluginScope=RestrictedToPluginNamespace` 配置约束插件命名空间
- **AND** 运行时 UI 链路通过 `PluginScope=Open` 配置允许插件贡献任意键

## MODIFIED Requirements

### Requirement: 宿主必须提供运行时翻译包分发能力
宿主系统 SHALL 提供运行时翻译包接口与语言列表接口,按语言返回聚合后的消息资源以及当前可用的语言描述信息,供默认管理工作台和宿主嵌入式插件页面加载。运行时翻译包 MUST 能同时包含宿主、项目和当前已启用插件的国际化消息,并在输出时转换为前端可直接消费的嵌套消息对象。运行时翻译包接口 MUST 在响应中输出 `ETag` 头,值由当前语言与运行时翻译包版本派生且在版本变化时必然不同;系统 MUST 接收请求中的 `If-None-Match` 头,匹配时返回 `304 Not Modified` 且不携带消息体。任何扇区缓存失效 MUST 触发运行时翻译包版本自增,确保同语言下不同 bundle 内容拥有不同 ETag。

#### Scenario: 默认工作台加载运行时翻译包
- **WHEN** 前端以 `zh-CN` 请求运行时翻译包
- **THEN** 宿主返回该语言下的聚合消息集合
- **AND** 结果中包含宿主资源、项目级资源以及已启用插件资源的合并结果
- **AND** 响应中包含 `ETag` 头

#### Scenario: 前端获取宿主支持的语言列表
- **WHEN** 前端请求运行时语言列表
- **THEN** 宿主返回当前支持的语言编码、默认语言标记、展示名称与原生名称
- **AND** 展示名称按当前请求语言返回,原生名称保持为对应语言自身文案
- **AND** 列表中包含 `zh-CN`、`en-US` 与 `ar-SA`

#### Scenario: 运行时语言包支持层级维护并保持扁平治理
- **WHEN** 宿主从文件、插件或数据库加载翻译资源
- **THEN** 宿主允许运行时 UI 文件资源使用层级 JSON 或扁平 dotted key 编写
- **AND** 宿主内部统一以扁平 key 形式聚合和覆写消息
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
宿主系统 SHALL 提供语言启停、翻译消息导入导出、缺失翻译检查和资源来源诊断能力,以支持交付项目长期维护多语言资源。交付项目和插件新增国际化资源时 MUST 遵循统一目录约定和翻译键规范。缓存失效 MUST 提供按"语言 × 扇区(host / source-plugin / dynamic-plugin / database)"维度的精细控制,不得在任意单点变更时整张清空所有语言所有扇区的缓存。运行时业务内容缓存 MUST 至少支持按 `business_type` 与 `locale` 维度作废,不得整张清空。所有"该键由代码源拥有"的判定 MUST 通过命名空间显式注册表完成,系统不得在 i18n 包内硬编码任何具体业务模块的命名前缀。

#### Scenario: 导出某语言的翻译资源
- **WHEN** 管理员或交付工具请求导出 `en-US` 的翻译资源
- **THEN** 系统返回该语言的聚合消息结果或可导入文件
- **AND** 导出内容能够用于后续批量维护和重新导入

#### Scenario: 检查缺失翻译键
- **WHEN** 管理员或交付工具执行缺失翻译检查
- **THEN** 系统返回当前语言相对默认语言缺失的翻译键列表
- **AND** 结果中能够定位缺失键所属的宿主模块、项目资源或插件资源范围
- **AND** 注册为代码源命名空间的翻译键不出现在缺失结果中

#### Scenario: DB 导入仅清除该语言的数据库扇区
- **WHEN** 管理员通过 `ImportMessages` 写入某一语言的若干翻译键覆写
- **THEN** 系统只清除该语言的数据库扇区缓存与合并视图
- **AND** 其他语言与本语言的非数据库扇区缓存保持有效
- **AND** 该语言的运行时翻译包版本自增,前端下次协商必能感知

#### Scenario: 业务内容缓存按业务类型失效
- **WHEN** 某业务模块写入或更新某 `business_type` 的内容
- **THEN** 系统仅清除该 `business_type` 关联的内容缓存条目
- **AND** 其他 `business_type` 与其他语言的内容缓存保持有效
