## 1. P1 性能优化:翻译热路径与缓存分层

- [x] 1.1 重写 `apps/lina-core/internal/service/i18n/i18n.go` 中 `Translate` / `TranslateSourceText` / `TranslateOrKey` / `TranslateWithDefaultLocale` 的实现为"持读锁直接读 cache",删除单 key 查找路径上的 `cloneFlatMessageMap` 调用
- [x] 1.2 保留 `BuildRuntimeMessages` 与 `ExportMessages` 的克隆语义,在内部新增 `lookupBundleKey(locale, key)` 工具方法,统一管理 cache 命中读取
- [x] 1.3 重构 `runtimeBundleCache` 结构为按 `locale × 扇区(host / source-plugin / dynamic-plugin)` 分层,新增 `mergedView` 视图与 `bundleVersion` 原子计数器
- [x] 1.4 重构 `InvalidateRuntimeBundleCache` 接受 `InvalidateScope` 参数,提供按 locale、按扇区、按插件 ID 的精细失效;迁移所有调用点为显式 scope
- [x] 1.5 删除运行时内容翻译缓存设计,确认本迭代不再提供 `sys_i18n_content` 业务内容翻译表
- [x] 1.6 补充 `Translate` 单次/批量调用的基准测试(`testing.B`),目标:命中 cache 时单次调用 < 100ns
- [x] 1.7 补充分层失效单元测试,覆盖"宿主资源只清宿主扇区"、"插件启停只清动态扇区"、"源码插件注册只清源码插件扇区"三种场景
- [x] 1.8 在 `lina-review` 技能中新增审查规则:禁止业务模块在 i18n 包外 clone 运行时消息包;`InvalidateRuntimeBundleCache` 必须显式传入 scope

## 2. P1 性能优化:运行时翻译包 ETag 协商

- [x] 2.1 在 `i18n.Service` 中新增 `BundleVersion()` 方法返回当前运行时翻译包版本号,任何扇区 invalidate 必须自增该版本
- [x] 2.2 修改 `apps/lina-core/internal/controller/i18n/i18n_v1_runtime_messages.go`,响应中输出 `ETag: "<locale>-<version>"` 与 `Cache-Control: private, must-revalidate`
- [x] 2.3 在该控制器中实现 `If-None-Match` 协商,匹配时返回 `304 Not Modified` 且不带消息体
- [x] 2.4 补充 `apps/lina-core/internal/controller/i18n/i18n_v1_runtime_test.go` 单元测试覆盖 ETag 输出、304 响应、版本变化后 ETag 必然不同三种路径
- [x] 2.5 创建 E2E 测试用例 `TC0124-runtime-i18n-etag.ts`,验证后端 ETag 与 304 协商在多语言切换流程中正确工作(繁体中文相关验证延后到第 12 节 TC0127-TC0129 完成)

## 3. P1 性能优化:前端 RequestClient 与持久化缓存

- [x] 3.1 重写 `apps/lina-vben/apps/web-antd/src/runtime/runtime-i18n.ts`:用 `requestClient` 替换裸 `fetch`,补充 Bearer 注入、错误降级、重试链
- [x] 3.2 在 `runtime-i18n.ts` 中新增 `localStorage` 持久化层:`linapro:i18n:runtime:<locale>` key,值为 `{etag, messages, savedAt}`,TTL 7 天
- [x] 3.3 实现"持久化命中即渲染、后台带 If-None-Match 协商、304 不更新"的快速路径
- [x] 3.4 改造 `apps/lina-vben/apps/web-antd/src/locales/index.ts` 中 `loadMessages` 为按失败语义拆分:运行时 bundle 失败 → 持久化兜底 + 用户提示;公共配置失败 → fire-and-forget;三方库 locale → 必须等待
- [x] 3.5 补充 `runtime-i18n.test.ts` 单元测试覆盖持久化命中、TTL 过期强制刷新、304 路径、网络异常降级四个场景
- [x] 3.6 补充 `loadMessages` 单元测试覆盖三件事独立失败语义

## 4. P2 一致性收敛:业务模块本地化投影边界

- [x] 4.1 删除原计划的 `apps/lina-core/internal/service/i18n/i18n_projector.go` 中心投影器方案,确认该方案由本任务组引入且会让 i18n 基础服务反向耦合业务实体与业务保护规则
- [x] 4.2 在各业务模块自己的 `*_i18n.go` 中保留"何时翻译 / 何时跳过 / 用哪个 Translate*"投影决策,`i18n` 包仅提供 `ResolveLocale` / `Translate` / `TranslateSourceText` 等底层能力
- [x] 4.3 改造 `apps/lina-core/internal/service/menu/menu_i18n.go`,菜单翻译键推导由 menu 模块拥有
- [x] 4.4 改造 `apps/lina-core/internal/service/dict/dict_i18n.go`,字典默认语言跳过策略与 `dict.*` 键约定由 dict 模块拥有
- [x] 4.5 改造 `apps/lina-core/internal/service/sysconfig/sysconfig_i18n.go`,配置投影与字段表头翻译键由 sysconfig 模块拥有
- [x] 4.6 改造 `apps/lina-core/internal/service/jobmgmt/jobmgmt_i18n.go`,内置任务和默认任务组保护规则由 jobmgmt 模块拥有
- [x] 4.7 改造 `apps/lina-core/internal/service/role/role.go`,内置 admin 角色投影规则由 role 模块拥有
- [x] 4.8 改造 `apps/lina-core/internal/service/plugin/internal/runtime/registry.go`,插件元数据投影规则由 plugin runtime 模块拥有
- [x] 4.9 补充/保留业务模块本地化投影测试,覆盖默认语言跳过、内置受保护记录翻译、用户记录保持原值三种场景

## 5. P2 一致性收敛:删除 sysconfig 硬编码标签 map

- [x] 5.1 在 `apps/lina-core/manifest/i18n/zh-CN.json` 与 `apps/lina-core/manifest/i18n/en-US.json` 补齐 `config.field.name` / `config.field.key` / `config.field.value` / `config.field.remark` / `config.field.createdAt` / `config.field.updatedAt`
- [x] 5.2 改造 `sysconfig_i18n.go::buildLocalizedImportTemplateHeaders` 与 `buildLocalizedExportHeaders`,删除 `englishLabels` / `chineseLabels` Go map,改为通过 `i18nSvc.Translate(ctx, "config.field."+name, fallback)` 解析
- [x] 5.3 删除 `localizedConfigFieldLabel` 内对 `ResolveLocale == "en-US"` 的硬编码判断
- [x] 5.4 创建 E2E 测试用例 `hack/tests/e2e/settings/config/TC0125-config-export-headers-via-i18n-keys.ts`,验证导出表头随语言切换变化且不依赖 Go map
- [x] 5.5 在 `lina-review` 技能中新增审查规则:`apps/lina-core/internal/service/sysconfig/` 与其他业务模块禁止维护英文/中文文案 Go map

## 6. P2 一致性收敛:source-text 命名空间显式注册

- [x] 6.1 在 `apps/lina-core/internal/service/i18n/i18n_source_text_namespace.go` 新建 `RegisterSourceTextNamespace(prefix, reason string)` 与查询函数;数据存储为包级 `sync.RWMutex` 保护的 `map[string]string`
- [x] 6.2 删除 `apps/lina-core/internal/service/i18n/i18n_manage.go::isSourceTextBackedRuntimeKey` 内的硬编码,改为查询命名空间注册表
- [x] 6.3 在 `apps/lina-core/internal/service/jobmeta` 或 `jobmgmt` 包内新增 `init()` 注册 `job.handler.` 与 `job.group.default.` 命名空间
- [x] 6.4 补充单元测试覆盖"未注册命名空间不豁免缺失检查"与"已注册命名空间从缺失结果中消失"两种场景
- [x] 6.5 在 `lina-review` 技能中新增审查规则:`apps/lina-core/internal/service/i18n/` 包内禁止以 `job.handler.` / `job.group.` 等业务命名前缀做硬编码判定

## 7. P2 一致性收敛:Service 接口拆分

- [x] 7.1 在 `apps/lina-core/internal/service/i18n/i18n.go` 中拆分接口为 `LocaleResolver` / `Translator` / `BundleProvider` / `Maintainer`,`Service` 改为这四个小接口的组合
- [x] 7.2 收敛 `menu` / `dict` / `sysconfig` / `jobmgmt` / `role` / `usermsg` 模块的 `i18nSvc` 字段类型为最小依赖接口(多数情况下是 `LocaleResolver + Translator` 的组合)
- [x] 7.3 收敛 `apidoc` 服务的 `i18nSvc` 字段为 `LocaleResolver + Translator`
- [x] 7.4 收敛 `controller/i18n/` 内各控制器字段类型,管理类控制器只持有 `Maintainer`,运行时控制器持有 `BundleProvider + LocaleResolver`
- [x] 7.5 更新所有相关单元测试的 mock,改为只 stub 实际依赖的小接口
- [x] 7.6 在 `lina-review` 技能中新增审查规则:业务模块字段类型应优先声明最小接口,禁止默认声明完整 `Service`

## 8. P3 边界整理:统一 ResourceLoader

- [x] 8.1 在 `apps/lina-core/pkg/i18nresource/` 新建稳定公共 `ResourceLoader` 组件,接受 `Subdir` / `PluginScope` / `LayoutMode` 配置参数,避免 apidoc 反向依赖 `internal/service/i18n`
- [x] 8.2 实现 `LoadHostBundle(ctx, locale)` / `LoadSourcePluginBundles(ctx, locale)` / `LoadDynamicPluginBundles(ctx, locale, releases)` 三个职责清晰的方法
- [x] 8.3 改造 `apps/lina-core/internal/service/i18n/i18n.go` 使用 `i18nresource.ResourceLoader{Subdir: "manifest/i18n", PluginScope: Open}` 替换重复实现
- [x] 8.4 改造 `apps/lina-core/internal/service/apidoc/apidoc_i18n_loader.go` 使用 `i18nresource.ResourceLoader{Subdir: "manifest/i18n/apidoc", PluginScope: RestrictedToPluginNamespace}` 替换重复实现
- [x] 8.5 删除两侧重复的目录遍历、ULEB128 解码、`wasm` 节解析代码,实现重复实现收敛
- [x] 8.6 补充 `ResourceLoader` 单元测试覆盖宿主、源码插件、动态插件三种来源以及插件命名空间隔离

## 9. P3 边界整理:WASM 解析提到 pluginbridge

- [x] 9.1 在 `apps/lina-core/pkg/pluginbridge/pluginbridge_wasm_section.go` 新建 `ReadCustomSection(content []byte, name string)` 与 `ListCustomSections(content []byte)` 公共函数,迁移 `wasm` 文件头校验、节遍历与 ULEB128 解码逻辑
- [x] 9.2 将 `pluginbridge.WasmSection*` 节名常量集中维护在 `pluginbridge` 包
- [x] 9.3 删除 `apps/lina-core/internal/service/i18n/i18n_plugin_dynamic.go` 内的 `parseWasmCustomSectionsForI18N` / `readWasmULEB128ForI18N` 私有函数,改为调用 `pluginbridge.ReadCustomSection`
- [x] 9.4 调整 `apidoc` 包内动态插件 apidoc 资源加载流程,统一改用 `pluginbridge.ReadCustomSection`
- [x] 9.5 补充 `pluginbridge` WASM 单元测试覆盖正常节读取、文件头错误、ULEB128 越界三种场景
- [x] 9.6 验证插件运行时 `pkg/pluginhost` 与 i18n 共用同一份解析路径无回归

## 10. 繁体中文接入:数据与资源

- [x] 10.1 在默认配置文件 `i18n` 段中通过 JSON 文件发现语言并维护默认语言、排序与原生名,不再通过 SQL seed 启用 `zh-TW`
- [x] 10.2 删除本迭代 `zh-TW` SQL seed 依赖,确认新增内置语言不需要改宿主 `manifest/sql/` 文件
- [x] 10.3 创建 `apps/lina-core/manifest/i18n/zh-TW.json` 完整覆盖宿主运行时 UI 翻译键,与 `zh-CN.json` 键集合一致
- [x] 10.4 创建 `apps/lina-core/manifest/i18n/apidoc/zh-TW.json` 与 `apps/lina-core/manifest/i18n/apidoc/zh-TW/*.json` 完整覆盖宿主接口文档翻译键
- [x] 10.5 在每个源码插件目录(`org-center` / `monitor-online` / `monitor-loginlog` / `monitor-operlog` / `monitor-server` / `content-notice` / `plugin-demo-source` / `plugin-demo-dynamic` / `demo-control`)新增 `manifest/i18n/zh-TW.json` 与对应 `manifest/i18n/apidoc/zh-TW/*.json`
- [x] 10.6 创建 `apps/lina-vben/packages/locales/src/langs/zh-TW/{authentication.json,common.json,preferences.json,profile.json,ui.json}` 五个静态语言包文件
- [x] 10.7 创建 `apps/lina-vben/apps/web-antd/src/locales/langs/zh-TW/{demos.json,page.json,pages.json}` 三个项目级语言包文件
- [x] 10.8 在 `apps/lina-vben/apps/web-antd/src/locales/index.ts` 中改为按语言编码约定推导 dayjs / antd locale,避免新增语言时继续添加 `case '<locale>'` 分支
- [x] 10.9 运行 `CheckMissingMessages(locale='zh-TW')` 确认返回 `total=0`(注册的代码源命名空间豁免)

## 11. 固定 LTR 方向接入

- [x] 11.1 在 `apps/lina-vben/packages/locales/src/i18n.ts` 中固定运行时语言方向为 `ltr`,不维护静态方向语言注册表
- [x] 11.2 在 `setI18nLanguage(locale)` 中固定设置 `document.documentElement.dir = 'ltr'`,并保留响应式 `direction` 状态供组件使用
- [x] 11.3 在 `apps/lina-vben/apps/web-antd/src/bootstrap.ts` 与 `App.vue`(或对应根组件)中将固定 `ltr` 方向注入 `Ant Design Vue` 的 `ConfigProvider`
- [x] 11.4 验证语言切换时 `<html dir>` 与 `ConfigProvider direction` 始终保持 `ltr`
- [x] 11.5 创建 E2E 测试用例 `TC0126-traditional-chinese-ltr-direction-switch.ts`,验证 `<html dir>` 与 antd 组件方向在多语言切换时保持 `ltr`
- [x] 11.6 在 `lina-review` 技能中新增范围说明:本次变更不支持 RTL 设计语言

## 12. 繁体中文回归与压力测试

- [x] 12.1 创建 E2E 测试用例 `TC0127-traditional-chinese-page-content-audit.ts`,在繁体中文下逐页访问框架默认交付的菜单路由,确认无简体中文/英文残留
- [x] 12.2 创建 E2E 测试用例 `TC0128-traditional-chinese-plugin-pages.ts`,覆盖源码插件页面与动态插件示例页在繁体中文下的展示
- [x] 12.3 创建 E2E 测试用例 `TC0129-traditional-chinese-apidoc.ts`,验证 `/api.json?lang=zh-TW` 返回的接口文档分组、摘要、参数说明按繁体中文展示
- [x] 12.4 在 CI / 本地 `make test` 链路中确保 `CheckMissingMessages(locale='zh-TW')` 阈值与 `en-US` 一致(均为 `total=0`),缺失即阻断
- [x] 12.5 编写"加新语言流程"文档草稿,放入 `apps/lina-core/manifest/i18n/README.md` 与中文镜像

## 13. 性能验证与基准测试

- [x] 13.1 编写 `Translate` 热路径基准测试 `apps/lina-core/internal/service/i18n/i18n_bench_test.go`,覆盖单 key 查找、批量 100 次查找、cache miss 重建三种场景
- [x] 13.2 验证基准结果:命中 cache 时单次 `Translate` < 100ns;批量 100 次相对改造前下降 ≥ 80%
- [x] 13.3 验证运行时翻译包接口在 ETag 命中 304 时响应体为空、Content-Length 为 0
- [x] 13.4 验证前端二次进入页面切语言时 Network 面板看到 `304` 而非 `200`
- [x] 13.5 在 PR 描述或 review 报告中附上前后对比基准数据

## 14. 文档与审查

- [x] 14.1 更新 `apps/lina-core/manifest/i18n/README.md` 与中文镜像 `README.zh_CN.md`,说明 ETag 协商、新增语言流程、源码命名空间注册三件事
- [x] 14.2 更新 `apps/lina-vben/apps/web-antd/src/locales/README.md` 与中文镜像,说明固定 LTR 方向、持久化缓存策略
- [x] 14.3 更新根 `CLAUDE.md` 中"i18n 持续治理要求"段落,补充"加新语言时禁止修改 Go 代码"与"运行时 cache 失效必须显式 scope"两条
- [x] 14.4 调用 `lina-review` 技能对 P1 / P2 / P3 / 繁体中文 + 固定 LTR 四组改动分别完成代码与规范审查
- [x] 14.5 运行 `make test` 全套 E2E 通过(含新增 `TC0124` ~ `TC0129`)
- [x] 14.6 修复审查与 E2E 中暴露的全部阻断问题后,确认 `openspec validate framework-i18n-improvements` 通过

## Feedback

<!-- 用户对实施过程中暴露的问题的反馈,通过 lina-feedback 技能追加 -->

- [x] **FB-1**: `apps/lina-core/internal/service/i18n/` 新增文件缺少 `i18n_` 前缀,且 `lina-review` 未显式覆盖服务组件文件命名检查
- [x] **FB-2**: `LocaleProjector` 让 i18n 基础服务反向耦合菜单、字典、配置、任务、角色、插件元数据等业务规则,需要立即移除并重新评估后续方案边界
- [x] **FB-3**: 新增内置语言不应要求修改后端 Go 枚举、SQL seed 或前端 TS 语言清单,应由 `manifest/i18n/<locale>.json` 与简化 YAML 元数据驱动
- [x] **FB-4**: 移除 `sys_i18n_locale` / `sys_i18n_message` / `sys_i18n_content` 三张运行时 i18n 持久化表,将翻译内容收敛为开发期 JSON/YAML 资源唯一事实源
- [x] **FB-5**: 将语言元数据配置移到默认配置文件 `i18n` 段,移除 `direction` 配置并固定 LTR,同时消除后端 fallback 与前端第三方 locale 兜底中的多语言硬编码
- [x] **FB-6**: 为默认配置文件 `i18n` 段增加配置项注释,移除阿拉伯语资源并改为繁体中文 `zh-TW` 支持
- [x] **FB-7**: 版本 SQL 中包含清理数据库数据的 `DELETE` / `DROP` / `UPDATE` 语句,需要移除并确认同类脚本风险
- [x] **FB-8**: 将运行时语言元数据从独立语言配置文件合并到默认配置文件 `i18n` 段,新增 `enabled` 开关,并确保关闭多语言或移除 `locales` 项时后端语言解析与前端语言切换按钮按配置生效,补充 `TC0130` 回归覆盖
- [x] **FB-9**: 移除 `config_i18n.go`、`i18n_locale.go` 与 fallback 路径中的默认语言硬编码,确保默认语言和语言列表只来自默认配置文件 `i18n` 段
- [x] **FB-10**: 移除 `config_i18n.go` 对默认配置模板路径和嵌入文件读取的依赖,改为只读取系统已装载的 `i18n` 配置段
- [x] **FB-11**: `config.Service` 新增 `GetI18n` 后,`role` 单元测试 mock 未同步适配,导致后端全量单测编译失败
- [x] **FB-12**: `i18n.enabled=false` 与 `zh-TW` 导出表头覆盖仍偏 E2E,需要补充后端/前端单元保护并扩展参数导出 E2E 断言
- [x] **FB-13**: `zh-TW` apidoc 与 packed manifest 资源缺少单元级完整性保护,插件 apidoc 资源缺失可能被测试静默跳过
- [x] **FB-14**: runtime bundle cache 全量/按语言失效的版本递增缺少单测,`i18n.enabled=false` 缺少 controller 层默认语言返回形态覆盖
- [x] **FB-15**: 新增 i18n E2E 对插件状态、插件 apidoc、前端 ETag 流程、繁体登录页与 raw key 泄漏检测的覆盖仍不完整
- [x] **FB-16**: full E2E 暴露默认品牌 Logo 静态资源缺失,偏好抽屉与用户抽屉页面对象 locator 过宽导致验证不稳定
- [x] **FB-17**: full E2E serial 暴露动态插件安装前 cron 授权元数据无法从未启用 wasm 资源本地化,且安装并启用快捷授权链路对 cron host service 的 payload 覆盖不完整
- [x] **FB-18**: 默认品牌 Logo 已在系统配置数据中统一为 `/logo.png`,前端默认偏好与 E2E 断言不应继续使用 `/linapro-mark.png`
- [x] **FB-19**: `permission-display.ts` 中动态路由权限展示模板与片段词表仍维护在 TypeScript 中,需要收敛到前端 i18n JSON 资源并保留未知权限片段兜底展示
