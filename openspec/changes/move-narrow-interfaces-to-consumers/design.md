# 设计

## 背景

当前部分 service 主文件同时承担组件默认契约入口和多个“分类接口”入口。分类接口如果只被同包`Service`内嵌，并没有提供真正的独立替换边界，反而要求阅读者先理解多个抽象名称再回到完整服务能力。另一类接口是消费方为了构造依赖所需的窄方法集合，例如插件 host config 只需要`GetRaw`，插件 auth adapter 只需要租户令牌签发能力。这类接口放在生产者包中，会把消费方语义扩散到生产者组件。

## 目标

- 生产者组件的主文件只保留默认`Service`、稳定跨组件契约和必要构造入口。
- 消费方优先复用目标组件已有`Service`或稳定契约；仅当完整契约不能清晰表达消费边界、会扩大不稳定实现面或用于适配特殊运行期边界时，才在消费方包内定义窄依赖接口。
- 不改变运行时行为、方法名、业务错误或数据库结构。
- 反馈修正允许删除无业务入口的 i18n 管理诊断 API，并收敛源码插件 i18n capability 方法集合。

## 非目标

- 不重命名`GetJwtExpire`、`GetSessionTimeout`等仍有业务必要的方法。
- 不拆分或重构具体业务实现逻辑。
- 不新增通用 DI 容器、聚合依赖结构体或 service locator。
- 不调整数据库、前端运行时加载协议或前端页面。

## 反馈修正：`i18n.Service`契约收敛

`i18n`上一轮合并生产者侧分类接口后仍保留过多方法，根因是将语义包装方法、运行时语言列表的配套查询方法、HTTP 缓存 freshness 包装方法、维护诊断方法和插件消息搜索辅助能力都继续放在核心`Service`契约上。它们要么可以由更通用的方法直接表达，要么只服务已经没有业务入口的管理诊断功能，继续保留会让宿主核心国际化服务承担不必要的复杂度。

本次反馈修正采用以下收敛规则：

- `TranslateSourceText`、`TranslateOrKey`和`TranslateWithDefaultLocale`不再作为`Service`方法暴露；调用方通过`Translate(ctx, key, fallback)`表达源码文案兜底、key 兜底或默认语言上下文兜底。
- `BundleVersion`并入`BundleRevision(ctx, locale)`；`BundleRevision`负责在读取 revision 前执行集群 freshness 检查并返回错误，控制器不再单独调用`EnsureRuntimeBundleCacheFresh`。
- `ListRuntimeLocales`和`IsMultiLanguageEnabled`并入`RuntimeLocales(ctx, locale)`，一次返回语言切换状态和语言描述符列表，保持 HTTP 响应结构不变。
- `ExportMessages`、`CheckMissingMessages`和`DiagnoseMessages`作为无业务入口的管理诊断能力删除，对应 HTTP API、DTO、控制器、路由和 apidoc 翻译资源同步移除。
- 源码插件 i18n capability 删除`FindMessageKeys`，避免为了插件搜索便利函数反向要求核心 i18n 服务保留`ExportMessages`。

该修正改变 i18n service 和源码插件公开 capability 方法集合，但不改变运行时翻译包 API、运行时语言列表 API、数据库结构或前端运行时加载协议。

## 接口 owner 规则

接口按 owner 分为三类：

| 类型 | 定义位置 | 示例 |
| ---- | ---- | ---- |
| 生产者完整契约 | 生产者组件主文件 | `config.Service`、`role.Service`、`i18n.Service` |
| 稳定产品/运行期契约 | 能力 owner 所在包 | `session.Store`、`jobmgmt.Scheduler` |
| 消费方窄依赖 | 消费方包，靠近构造函数或适配器 | 插件 host config 的 raw reader、插件能力宿主的 token issuer、启动装配的 token issuer |

当一个窄接口只服务单个消费者或少数同一消费场景，且复用目标组件已有`Service`会增加不必要依赖面或模糊 owner 时，应放在消费者侧。生产者不应为了“可能有消费者”预先创建多组难以理解的导出接口。若目标组件已有`Service`已经是清晰、稳定、可复用的 owner 契约，消费方应直接复用该`Service`，避免为单个方法重复声明本地接口。

## 实现路径

1. 在 OpenSpec 增量规范中补充消费者侧窄接口规则。
2. 合并仅用于生产者同包自组合的分类接口：
   - `dict`：`TypeService`、`DataService`、`ImportExportService`、`LookupService`。
   - `jobmgmt`：`GroupService`、`JobService`、`LogService`。
   - `middleware`：`HTTPMiddleware`、`RuntimeSupport`。
   - `plugin`根 service：仅用于内嵌的功能分类接口。
   - `role`：角色查询、写入、菜单、用户、权限和访问快照分类接口。
3. 将跨包窄接口移动到消费者：
   - `config.RawReader`迁移为插件 host config/plugin config 消费方本地接口。
   - `auth.SessionRevoker`和`auth.TenantTokenIssuer`迁移为插件 host service、capabilityhost和启动装配消费方接口。
   - `role.RoleAccessSnapshotService`迁移为插件 host service/capabilityhost消费方接口。
   - `i18n`生产者侧分片接口合并回`i18n.Service`；普通控制器优先直接依赖`i18n.Service`，仅在插件适配器、中间件响应边界或测试替身成本明显需要收敛时定义消费方本地接口。
4. 使用静态检索和 Go 编译门禁确认旧生产者侧接口不再被引用。

## 影响分析

- `i18n`影响：本次只移动 Go 接口定义，不新增或修改运行时用户可见文案、API 文档源文本、插件清单或语言包资源。
- 缓存一致性影响：不改变配置、权限、插件、i18n 或 session 缓存的权威数据源、失效机制、跨实例同步或共享实例策略。
- 数据权限影响：不改变任何数据读写、查询过滤、租户/组织边界或插件数据访问路径。
- 开发工具跨平台影响：不修改脚本、Makefile、CI 或生成工具。
- 测试策略：属于内部接口治理与编译期契约收敛，无运行时行为变化；使用 OpenSpec 严格校验、静态检索和 Go 包编译门禁验证。
