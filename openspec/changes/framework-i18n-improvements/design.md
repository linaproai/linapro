## Context

`framework-i18n-foundation` 已在 `apps/lina-core/internal/service/i18n/` 落地了"宿主三层资源 + 数据库覆写 + 业务内容多语言"完整体系,在 `zh-CN` / `en-US` 双语下功能完整。但通过 78 条反馈逐步发现的系统性问题,集中在**热路径性能、模块一致性、组件边界**三个维度:

- 当前 `Translate` 系列方法每次调用都会 `cloneFlatMessageMap` 整张运行时消息包,而调用方往往只需要 1 个 key——已知热路径(菜单列表、字典列表、公共配置)单次请求会触发数十次完整 clone。
- `runtimeBundleCache` 的 invalidate 是"全语言全扇区一锅清",任何插件启停或 DB 翻译导入都会击穿所有语言的缓存。
- `runtimeContentCache` 在 `InvalidateContentCache()` 时整张清空,业务内容数量增长后会出现击穿到 DB 的风险。
- 5 个 `*_i18n.go` 适配文件 (`menu_i18n.go` / `dict_i18n.go` / `sysconfig_i18n.go` / `jobmgmt_i18n.go` / `role.go`) 各自决定"何时翻译/何时跳过/用哪个 Translate*",`sysconfig_i18n.go` 甚至直接在 Go 内硬编码英文/中文 map(违反约定一)。
- `i18n_manage.go::isSourceTextBackedRuntimeKey` 把 `job.handler.*`、`job.group.default.*` 这些 jobmgmt 私有命名前缀漏到 i18n 包内部,形成反向依赖。
- `Service` 接口 18 个方法承担 5 类职责,业务模块测试要 stub 整个大接口。
- `apidoc/apidoc_i18n_loader.go` 与 `i18n/i18n_source.go` 各自维护一份 host/source-plugin/dynamic-plugin 资源遍历逻辑,共重复 ~280 行。
- 前端 `runtime-i18n.ts` 用裸 `fetch` 绕过 `requestClient`,失败语义和持久化缓存策略都不完整;`loadMessages` 把"运行时 bundle / 公共配置 / 三方库"用 `Promise.all` 一锅端,无差别失败处理。
- 整个框架在 `zh-CN` / `en-US` 这两种 LTR、无复数、半角数字的"近亲语言"下被验证;约定式翻译键、缺失检查、运行时聚合在引入真正不同的语言(RTL、多复数形式、不同数字字符集)时是否真的零业务代码改动,从未被压力测试。

由于本项目无历史兼容包袱,本次设计不保留旧 `Service` 接口签名、不维护旧 cache 结构,直接重构。

## Goals / Non-Goals

**Goals:**
- 让 `Translate` 热路径在 cache 命中时不再做大 map 克隆,单次查找接近常量时间。
- 让缓存失效具备扇区级精度:插件启停只清相关插件扇区,DB 导入只清 DB 扇区,语言切换不互相影响。
- 让运行时翻译包在 HTTP 层支持 ETag/304 协商,前端二次进入页面零网络切语言。
- 让 5 个业务模块的 `*_i18n.go` 收敛到 `LocaleProjector` 统一管控,消除"何时翻译"逻辑漂移。
- 让 `Service` 大接口拆分为多个职责清晰的小接口,业务模块只声明实际依赖。
- 让 `source-text` 命名空间从 i18n 包黑名单改为业务模块显式注册,消除反向依赖。
- 让 `apidoc` 与 runtime bundle 共用同一份 `ResourceLoader`,消除 ~280 行重复实现。
- 让前端 `loadMessages` 三件事按各自失败语义独立处理,带持久化缓存兜底。
- 让 WASM 自定义节解析归属 `pluginbridge`,i18n 不再反向依赖 WASM 文件格式。
- 让阿拉伯语 `ar-SA` 作为第三种语言成为压力测试基线,验证"约定式翻译键 + 缺失检查"在加新语言时真的零业务代码改动。
- 让基础 RTL (`<html dir>` + antd `ConfigProvider`) 跟随语言切换自动生效,阿语下页面"内容正确、能用"。

**Non-Goals:**
- 不在本次范围内做完整 RTL 设计语言:图标镜像、抽屉/通知滑入方向、表格固定列翻转、CSS logical properties 全面替换、菜单展开方向。
- 不在本次范围内引入用户级语言偏好(`sys_user.preferred_locale`)。
- 不在本次范围内提供国际化可视化管理后台页面;现有导入/导出/缺失/诊断 API 已足够。
- 不引入自动机器翻译、AI 翻译辅助或外部翻译平台对接。
- 不重新设计 `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` 三张表的 schema(基础结构在 foundation 已稳定)。

## Decisions

### 决策一:`Translate` 热路径直接读 cache,放弃克隆

**选择**:把 `Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` 改为直接对 `runtimeBundleCache` 持读锁后查值,不再走 `buildRuntimeMessageCatalog → cloneFlatMessageMap`。仅 `BuildRuntimeMessages` (输出给前端) 与 `ExportMessages` 保留克隆语义。

**原因**:
- `Translate` 99% 调用路径只读 1 个 key,clone 整张 800+ key 的 map 是纯浪费。
- `map[string]string` 的并发只读是安全的,只要写入路径(invalidate / cache 重建) 用 `sync.RWMutex` 保护。
- 业务模块拿到字符串后并不会修改它,clone 的"防御性"不成立。

**约束**:
- cache 重建必须先在临时 map 中完成,再 atomic 替换缓存条目;禁止"边清边写"。
- 任何返回 `map[string]string` 给外部模块的接口仍必须 clone 一次。

**备选方案**:
- 改为 `sync.Map`。未采用,因为运行时 cache 在 miss 时整张重建、命中时全表只读,典型读多写少场景,`RWMutex + map` 更直接也更易调试。

### 决策二:`runtimeBundleCache` 重构为按 locale + 扇区分层

**选择**:cache 结构从

```
bundles map[locale]map[key]value
```

改为

```
type localeCache struct {
    host       map[string]string                          // 不可变,启动期一次性加载
    plugins    map[pluginID]map[string]string             // 源码插件,源码注册表变化时刷新
    dynamic    map[pluginID]map[string]string             // 动态插件,plugin lifecycle 钩子刷新
    db         map[string]string                          // sys_i18n_message,导入/管理操作刷新
    merged     map[string]string                          // 按优先级合并后的视图,任何子层变化即作废
    mergedAt   atomic.Uint64                              // 与全局 bundleVersion 关联
}
type runtimeCache struct {
    sync.RWMutex
    locales map[string]*localeCache
    version atomic.Uint64                                 // bundleVersion,用于 ETag
}
```

`Translate` 优先读 `merged`;`merged` 缺失时按 db > dynamic > plugins > host 顺序合并并填充。
失效粒度:
- DB 导入 → `locales[L].db = nil; locales[L].merged = nil`
- 动态插件启停 → 只清相关 plugin ID 在所有 locale 的 dynamic 子层与 merged
- 源码插件注册表变化 → plugins 与 merged 失效
- 启用/停用 locale → 整体清该 locale

每次 invalidate 触发 `version.Add(1)`,用于驱动前端 ETag 协商。

**原因**:
- 当前"一锅清"在多租户/活跃管理场景下会形成可观测的性能凹陷。
- 分层结构把"谁拥有该层"和"何时失效"做成 1:1 映射,排查问题清晰。
- `merged` 视图保证 `Translate` 的命中路径仍是 O(1) map 查找。

**备选方案**:
- 不引入 `merged` 视图,每次 `Translate` 都按四层依次查找。未采用,因为多语言下叠加 4 次 map 查找的常数远大于一次。

### 决策三:运行时翻译包接口加 ETag 与持久化缓存

**选择**:
- 后端:`/i18n/runtime/messages` 响应头新增 `ETag: "<locale>-<bundleVersion>"` 与 `Cache-Control: private, must-revalidate`;请求带 `If-None-Match` 时若匹配则返回 `304 Not Modified` (空 body)。
- 前端:`runtime-i18n.ts` 改走 `requestClient`,在 `localStorage` 持久化 `{locale, etag, messages, savedAt}`;切语言时先用持久化缓存快速渲染,后台异步带 `If-None-Match` 协商;命中 304 直接保持持久化数据。
- 持久化缓存 TTL 默认 7 天,超过后强制重新拉取(防止用户长期不重启浏览器导致版本漂移过大)。

**原因**:
- 运行时翻译包的本质特征是"绝大多数请求拿到完全一致的内容",是 ETag 的最佳场景。
- 持久化缓存解决"二次登录全量传输 ~80KB"的体验问题,用户切语言瞬时生效。
- TTL 兜底防止缓存与实际版本长期偏离。

**备选方案**:
- 用 `Last-Modified`。未采用,bundle 版本变化由配置导入和插件启停驱动,二者都非"文件时间戳"语义,`ETag` 更准确。
- 后端写 Redis 集中保存 bundle 版本。未采用,bundle 缓存本来就是进程内的 `runtimeBundleCache`,版本号也应该是同一个对象的属性,不需要外部存储。

### 决策四:抽出 `i18n.LocaleProjector`,消除 5 个 `*_i18n.go` 的重复决策

**选择**:在 `apps/lina-core/internal/service/i18n` 包下新增 `projector.go`,提供:

```go
type LocaleProjector interface {
    ProjectMenu(ctx context.Context, m *entity.SysMenu)
    ProjectDictType(ctx context.Context, t *entity.SysDictType)
    ProjectDictData(ctx context.Context, d *entity.SysDictData)
    ProjectConfig(ctx context.Context, c *entity.SysConfig)
    ProjectBuiltinJob(ctx context.Context, j *entity.SysJob)
    ProjectJobGroup(ctx context.Context, g *entity.SysJobGroup)
    ProjectBuiltinRole(ctx context.Context, r *entity.SysRole)
    ProjectPluginMeta(ctx context.Context, p *PluginMeta)
}
```

每个 `Project*` 方法封装该实体的翻译键推导、跳过策略、源文案选择。`*_i18n.go` 只剩 1-2 行调用;`menu/menu.go` 等业务方法持有 `LocaleProjector` 字段。

**原因**:
- 5 个适配文件原本各写 30-150 行,现在只剩调用入口,审查面缩到一半以下。
- "何时翻译/何时跳过"集中决策,加新模块只需扩展 projector,业务模块零代码。
- `LocaleProjector` 内部可以引用更细粒度的小接口(决策五),与"决策一缓存重构"自然对齐。

**约束**:
- `LocaleProjector` 不允许暴露按 entity 类型分散的 *Translate 方法重载,只暴露语义化的 `Project*`。
- 默认语言下的"跳过翻译"策略统一在 projector 内做,业务模块不再判断 `ResolveLocale == DefaultLocale`。

**备选方案**:
- 在每个业务模块自己包内放 `localizer`。未采用,因为正是当前问题的根源——决策权一旦下放就会漂移。

### 决策五:`Service` 大接口拆分为五个小接口

**选择**:

```go
type LocaleResolver interface {
    ResolveRequestLocale(*ghttp.Request) string
    ResolveLocale(context.Context, string) string
    GetLocale(context.Context) string
}
type Translator interface {
    Translate(ctx context.Context, key, fallback string) string
    TranslateSourceText(ctx context.Context, key, sourceText string) string
    TranslateOrKey(ctx context.Context, key string) string
    TranslateWithDefaultLocale(ctx context.Context, key, fallback string) string
    LocalizeError(ctx context.Context, err error) string
}
type BundleProvider interface {
    BuildRuntimeMessages(ctx context.Context, locale string) map[string]any
    ListRuntimeLocales(ctx context.Context, locale string) []LocaleDescriptor
    BundleVersion() uint64
}
type ContentProvider interface {
    GetContent(ctx context.Context, in ContentLookupInput) (ContentLookupOutput, error)
    ListContentVariants(ctx context.Context, businessType, businessID, field string) ([]ContentVariant, error)
}
type Maintainer interface {
    ExportMessages(ctx context.Context, locale string, raw bool) MessageExportOutput
    CheckMissingMessages(ctx context.Context, locale, prefix string) []MissingMessageItem
    DiagnoseMessages(ctx context.Context, locale, prefix string) []MessageDiagnosticItem
    ImportMessages(ctx context.Context, in MessageImportInput) (MessageImportOutput, error)
    InvalidateRuntimeBundleCache()
    InvalidateContentCache()
}
```

`serviceImpl` 同时实现这五个接口;`New()` 返回值类型仍是 `Service`(为了向其他包提供完整能力),但 `Service` 改为 `interface { LocaleResolver; Translator; BundleProvider; ContentProvider; Maintainer }` 的组合。业务模块字段类型改为最小的小接口(如 `menu.serviceImpl.translator Translator`)。

**原因**:
- 业务模块测试 stub 从 18 个方法降到 5 个或更少。
- 阅读 `menu` 代码时,只看 `Translator` 就能理解它需要什么能力,`Maintainer` 与它无关。

**备选方案**:
- 保持单一 `Service` 接口。未采用,Go 习惯是"小接口大实现",拆分不影响实现复杂度,只优化使用面。

### 决策六:`source-text` 命名空间显式注册

**选择**:在 `i18n` 包内新增 `RegisterSourceTextNamespace(prefix, reason string)`;`isSourceTextBackedRuntimeKey` 改为读取注册表;`jobmgmt` / 未来其他业务模块在自己的 `init()` 中注册自己的命名空间。

```go
// jobmgmt/init.go
func init() {
    i18n.RegisterSourceTextNamespace("job.handler.", "code-owned cron handler display")
    i18n.RegisterSourceTextNamespace("job.group.default.", "code-owned default group")
}
```

**原因**:
- 消除 i18n 包对 jobmgmt 的反向依赖。
- 加新代码源文案模块时,只需在自己包内 `init` 一行,无需修改 i18n 包。

**备选方案**:
- 在 manifest 中静态声明命名空间。未采用,这是代码契约不是文件契约,与 Go 模块绑定更合适。

### 决策七:`apidoc` 与 runtime bundle 共用 `ResourceLoader`

**选择**:在 `apps/lina-core/internal/service/i18n` 下新增 `resourceloader.go`:

```go
type ResourceLoader struct {
    Subdir            string                              // "manifest/i18n" 或 "manifest/i18n/apidoc"
    PluginScope       PluginScopeMode                      // OpenScope | RestrictedToPluginNamespace
    LayoutMode        LayoutMode                           // FlatJSON | NestedDirectory
}

func (l *ResourceLoader) LoadHostBundle(locale string) map[string]string
func (l *ResourceLoader) LoadSourcePluginBundles(locale string) map[string]map[string]string
func (l *ResourceLoader) LoadDynamicPluginBundles(ctx context.Context, locale string, releases []ReleaseRef) map[string]map[string]string
```

`apidoc_i18n_loader.go` 与 `i18n_source.go` 改为薄壳,通过不同的 `ResourceLoader` 实例工作。

**原因**:
- 两侧逻辑结构完全相同,仅 `Subdir` 与"是否限制插件命名空间"不同。
- `ResourceLoader` 一处实现一处测试,避免双轨漂移。

**备选方案**:
- 把 apidoc 加载器并入 `i18n` 包统一管理。未采用,apidoc 翻译资源是文档专属域,不应该跟运行时 UI bundle 共享生命周期(已是 foundation 决策)。

### 决策八:WASM 自定义节解析提到 `pluginbridge`

**选择**:把 `i18n_plugin_dynamic.go` 内的 `parseWasmCustomSectionsForI18N` 与 `readWasmULEB128ForI18N` 移动到 `apps/lina-core/pkg/pluginbridge/wasm.go`:

```go
package pluginbridge

func ReadCustomSection(content []byte, name string) ([]byte, error)
func ListCustomSections(content []byte) (map[string][]byte, error)
```

`i18n` 与 plugin runtime 都通过 `pluginbridge.ReadCustomSection` 访问 WASM 节,i18n 包不再直接 import WASM 文件格式相关常量。

**原因**:
- WASM 文件格式是 `pluginbridge` 的天然职责;i18n 应该只关心翻译。
- 解析器集中后修复 bug、扩展节类型只需改一处。

**备选方案**:
- 提到 `internal/util/wasm`。未采用,违反 CLAUDE.md "禁止新增 internal/util 兜底目录" 规范。

### 决策九:前端 `loadMessages` 按失败语义拆分,带持久化兜底

**选择**:

```ts
async function loadMessages(lang) {
  const persisted = readPersistedRuntime(lang);
  let runtime = persisted?.messages ?? {};
  let nextRuntimeVersion = persisted?.etag ?? '';

  try {
    const fresh = await loadRuntimeLocaleMessagesViaRequestClient(lang, persisted?.etag);
    if (fresh.notModified) {
      // 304: 持久化数据有效
    } else {
      runtime = fresh.messages;
      nextRuntimeVersion = fresh.etag;
      writePersistedRuntime(lang, fresh);
    }
  } catch (err) {
    notifyDegraded('runtime-i18n', err);
    // runtime 保持持久化兜底
  }

  syncPublicFrontendSettings(lang).catch((err) => notifyDegraded('public-config', err));
  await loadThirdPartyMessage(lang);     // 必须等待
  return mergeMessages(appLocalesMap[lang] || {}, runtime);
}
```

**原因**:
- 三件事失败语义本就不同,统一 `Promise.all` 是过度耦合。
- 持久化兜底让弱网/离线场景下也能正确切换语言。

**备选方案**:
- 不引入持久化,每次必须等到运行时 bundle 拉取完才渲染。未采用,体验差且与 ETag 协商方案不匹配。

### 决策十:阿拉伯语 `ar-SA` 作为压力测试基线,基础 RTL 纳入本次

**选择**:
- `sys_i18n_locale` 通过新增的 `015-framework-i18n-foundation.sql` 兄弟文件 `017-framework-i18n-improvements.sql` 写入 `INSERT IGNORE INTO sys_i18n_locale` 一行 `ar-SA`,默认 `is_default=0`、`status=1`。
- 宿主与所有源码插件的 `manifest/i18n/ar-SA.json` 与 `manifest/i18n/apidoc/ar-SA/*.json` 必须补齐,否则 `CheckMissingMessages` 会返回非空。
- 前端 `packages/locales/src/langs/ar-SA/*.json`、`apps/web-antd/src/locales/langs/ar-SA/*.json` 补齐;dayjs 加载 `dayjs/locale/ar-sa`;antd `ConfigProvider` 通过 `direction` prop 跟随。
- `<html dir>` 在 `setI18nLanguage(locale)` 内根据 `isRtlLocale(locale)` 设置;阿语下 `dir=rtl`,其他 `dir=ltr`。
- 新增 `RtlAwareLocale` 注册表(目前仅 `ar-SA`),未来加 `he-IL`/`fa-IR` 时只需登记。
- 复数与数字格式:首期不在所有页面铺开,只在"统计计数 / 批量操作提示"两类约定 `$tn(key, count)` 钩子,`vue-i18n` 自带 `pluralization` 已能处理 `ar-SA` 的 6 种形式。

**原因**:
- 阿拉伯语真正考验的是 RTL、复数、数字格式三个隐藏假设;只有它能验证"加新语言零业务代码改动"的承诺是否真实。
- 基础 RTL(`html dir` + `ConfigProvider`)是单点开关,不引入这一项就连"页面能用"都做不到。
- 完整 RTL 设计语言是独立工作量级,本次禁入。

**备选方案**:
- 选 `ja-JP`。未采用,LTR + 无复数无法暴露隐藏假设,等于"再做一遍 en-US"。
- 一次性把完整 RTL 设计系统做完。未采用,会让本次变更范围爆炸,foundation 已经表明"主任务多 + 反馈多"的模式不健康。

## Risks / Trade-offs

- [Risk] 决策一移除 clone 后,如果有任何业务代码假设拿到的 map 可写并修改它,会污染缓存。→ Mitigation:codegen/lint 阶段不需要管,依靠改造期 grep `[]string\|map[string]string` 写操作并补单元测试;改造完毕后 `Translate` 系列只返回 `string`,根本没有 map 暴露给业务方。

- [Risk] 决策二的扇区缓存重构涉及 invalidate 路径多个调用点(plugin runtime、i18n manage、apidoc loader),迁移时容易漏掉。→ Mitigation:用 `Maintainer.InvalidateRuntimeBundleCache(scope ...InvalidateScope)` 替代裸调用,scope 由调用方显式传入;改造完成后用 `lina-review` 校验所有 invalidate 入口都带 scope 参数。

- [Risk] 决策三的 ETag 协商在前端持久化与后端 bundleVersion 不一致时(如某次 invalidate 没触发 version 自增),会导致用户看到陈旧翻译。→ Mitigation:每个 invalidate 路径都必须 `version.Add(1)`,审查规则覆盖;持久化 7 天 TTL 兜底;`reloadActiveLocaleMessages(force=true)` 仍可用于强制刷新。

- [Risk] 决策五拆接口可能导致下游模块大规模改类型签名,迁移成本大。→ Mitigation:`Service` 仍存在并是组合类型,业务模块改字段类型为更小接口是逐模块进行的可选优化,不阻塞功能交付;后续审查规则鼓励但不强制。

- [Risk] 决策十引入阿拉伯语后,翻译资源缺失会导致 `CheckMissingMessages` 与 E2E 巡检长期红色。→ Mitigation:阿语 manifest 补齐必须放在 tasks.md 中独立任务,且任何模块文案改动审查规则要求三语同步(`zh-CN` + `en-US` + `ar-SA`);CI 中 `CheckMissingMessages` 对 `ar-SA` 阈值与 `en-US` 一致,缺失即阻断。

- [Risk] 基础 RTL 可能与 antd 已有组件 CSS 不兼容,出现部分组件视觉错乱。→ Mitigation:验收标准明确为"内容正确、能用",视觉镜像偏差可接受;未来"完整 RTL 设计语言"独立变更收口。

- [Trade-off] 决策七引入的 `ResourceLoader` 抽象会让 apidoc 与 runtime bundle 加载流程多一层间接调用。→ 接受,因为消除 ~280 行重复实现的收益远大于一层抽象的阅读代价。

- [Trade-off] 决策十不做完整 RTL 会让阿拉伯语用户在某些页面遇到"内容是阿语、布局还是 LTR"的混合体验。→ 接受,因为完整 RTL 设计是独立工程,本次明确范围外。

## Migration Plan

1. 性能优化(P1):
   - 重写 `Translate*` 不再 clone;补 benchmark 单元测试;补缓存命中率断言。
   - 重构 `runtimeBundleCache` 为分层结构;迁移所有 `InvalidateRuntimeBundleCache()` 调用点为 `InvalidateScope`。
   - 后端实现 `bundleVersion` 与 `ETag`;改 `runtime_messages.go` 控制器读 `If-None-Match`、写 `ETag`、304 返回。
   - 前端 `runtime-i18n.ts` 改走 `requestClient`;实现 `localStorage` 持久化与 304 协商;补单元测试。

2. 一致性收敛(P2):
   - 抽出 `LocaleProjector`;改造 5 个 `*_i18n.go` 为薄壳;删除 `sysconfig_i18n.go` 内 `englishLabels`/`chineseLabels` 并补对应 `config.field.*` 翻译键。
   - 实现 `RegisterSourceTextNamespace`;`jobmgmt` 在自己 `init()` 中注册;删除 `i18n_manage.go::isSourceTextBackedRuntimeKey` 黑名单。
   - 拆分 `Service` 接口为五个小接口;业务模块字段类型逐个收敛(menu/dict/sysconfig/jobmgmt/role/usermsg/apidoc/plugin)。

3. 边界整理(P3):
   - 抽出 `ResourceLoader`;`apidoc_i18n_loader.go` 与 `i18n_source.go` 共用;删除重复实现。
   - 前端 `loadMessages` 拆分失败语义;补单元测试覆盖弱网/超时降级。
   - WASM 自定义节解析提到 `pluginbridge`;i18n 与 plugin runtime 切换调用路径;删除 i18n 内 WASM 工具函数。

4. 阿拉伯语 + 基础 RTL:
   - 新增 `017-framework-i18n-improvements.sql` 启用 `ar-SA`(seed DML);`make init` 验证幂等。
   - 补齐宿主、所有源码插件的 `manifest/i18n/ar-SA.json` 与 `manifest/i18n/apidoc/ar-SA/*.json`。
   - 补齐前端 `packages/locales/src/langs/ar-SA/*.json` 与 `apps/web-antd/src/locales/langs/ar-SA/*.json`。
   - `setI18nLanguage(locale)` 集成 `<html dir>` 切换;antd `ConfigProvider` 接入 `direction`。
   - dayjs 注册 `ar-sa` locale。
   - `CheckMissingMessages` 对 `ar-SA` 阈值与 `en-US` 一致。
   - E2E 新增 `TC0124` 覆盖阿语下的语言切换、`<html dir>` 断言、关键页面文本完整性。

5. 验证与审查:
   - `make test` 全套 E2E 通过(含新增 `TC0124`)。
   - `lina-review` 对 P1/P2/P3/阿语四组改动分别审查。
   - benchmark 报告:`Translate` 单次 < 100ns(目标);热路径 100 次调用总时延下降 ≥ 80%。

## Open Questions

- 决策三的持久化 TTL 默认 7 天是否合适?是否需要做成可配置(`sys_config` 中暴露)?当前结论:首期固定 7 天,未来如有需要再开放配置。
- 决策十阿语 manifest 是由人工逐键翻译还是临时使用占位 + 后续校对?当前结论:首期由人工翻译关键路径(登录页、主导航、用户/角色/菜单/字典/参数/通知/调度/插件页面),其余允许同步留空,但 E2E 必须保证"在阿语下不显示中文/英文残留"——通过 `CheckMissingMessages` 强约束。
- 复数 API 是否引入自定义 `$tn` 钩子,还是直接用 vue-i18n 内置?当前倾向直接用 vue-i18n 的 `t` 复数语法(`{ count, plural, zero {...} one {...} other {...} }`),首期只在批量操作提示文案上落一两个示例,后续按需扩展。
