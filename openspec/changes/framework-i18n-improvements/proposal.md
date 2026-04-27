## Why

`framework-i18n-foundation` 已经把宿主、插件、前端、数据库四层 i18n 基础能力跑通,在 `zh-CN` 与 `en-US` 双语环境下功能完整。但落地过程中通过 78 条反馈逐步暴露出三类系统性问题:

1. **性能层**:`Translate` 热路径在每次调用时都会克隆 800+ key 的整张运行时消息包 (一次菜单列表请求 = ~100 次大 map 分配);缓存粒度是"全语言全扇区一锅清",任何插件启停或翻译导入都会击穿所有语言;运行时翻译包接口没有 ETag/版本戳,前端切语言每次都全量传输。
2. **一致性层**:5 个 `*_i18n.go` 适配文件每个都重新决定"何时翻译/何时跳过/用哪个 Translate*",导致 `dict` 在默认语言时 `return`、`menu` 一律 `Translate`、`sysconfig` 直接在 Go 里硬编码英文/中文 map(违反约定式翻译键决策);初版方案试图通过中心化 `LocaleProjector` 消除重复,但会让 i18n 基础服务反向耦合业务实体与业务保护规则,因此改为"业务模块拥有投影规则、i18n 只提供底层能力";`isSourceTextBackedRuntimeKey` 把 jobmgmt 的 `job.handler.*` 命名前缀漏到 i18n 包里形成反向依赖;`Service` 接口承担 5 类职责共 18 个方法,业务模块只用前两类却必须 stub 整个接口。
3. **边界层**:`apidoc` 与 runtime bundle 各自维护 ~280 行结构相同的 host/plugin/dynamic 加载器和缓存;前端 `loadMessages` 把"运行时 bundle / 公共配置 / 三方库"三件失败语义不同的事用 `Promise.all` 一锅端;WASM 自定义节解析在 i18n 包里重复实现,让 i18n 反向依赖 WASM 文件格式。

同时,当前框架只在 `zh-CN`/`en-US` 这两种 LTR、无复数、半角数字的"近亲语言"下被验证过;"约定式翻译键 + 缺失检查"在引入真正不同的语言时是否真的零业务代码改动,从未被压力测试。本次变更引入阿拉伯语 (`ar-SA`) 作为试金石,既覆盖更大用户群,也借助 RTL/复数/数字格式暴露所有"语言只是 UTF-8 字符串"的隐藏假设。

## What Changes

### P1 性能优化
- 重写 `Translate`/`TranslateSourceText`/`TranslateOrKey`/`TranslateWithDefaultLocale` 的热路径,改为直接从缓存读取,不再克隆整张消息包;仅 `BuildRuntimeMessages` 等需要把消息交给前端的方法保留克隆语义。
- 把 `runtimeBundleCache` 重构为按 `locale + 扇区(host/plugin/dynamic/db)` 分层,失效时只清相关扇区或语言,不再整张清空;`runtimeContentCache` 同步按 `business_type + locale` 分层失效。
- 运行时翻译包接口 `/i18n/runtime/messages` 输出 `ETag` (基于 locale + bundleVersion),并支持 `If-None-Match` 304 协商;后端维护 `bundleVersion` 原子计数器,任何 invalidate 触发自增。
- 前端 `runtime-i18n.ts` 从裸 `fetch` 切回 `requestClient`,接入鉴权/错误/降级链;在内存缓存基础上叠加 `localStorage` 持久化,二次进入页面零网络切语言。

### P2 一致性收敛
- 收敛 `menu_i18n.go`/`dict_i18n.go`/`sysconfig_i18n.go`/`jobmgmt_i18n.go`/`role.go` 内的投影规则,但规则必须留在各业务模块边界内;`internal/service/i18n` 只暴露 `ResolveLocale`、`Translate`、`TranslateSourceText` 等底层能力,不得提供按业务实体命名的投影器。
- 删除 `sysconfig_i18n.go` 内的 `englishLabels`/`chineseLabels` Go map,改为 `config.field.<name>` 走 `manifest/i18n/<locale>.json`,验证决策一(约定式翻译键)能真正覆盖到导出/导入表头这种边界。
- `i18n` 包引入 `RegisterSourceTextNamespace(prefix, reason)` 显式注册表;`isSourceTextBackedRuntimeKey` 黑名单从 `i18n_manage.go` 移除,jobmgmt 在自身 `init` 注册自己的命名空间。
- `i18n.Service` 大接口拆分为 `LocaleResolver` / `Translator` / `BundleProvider` / `ContentProvider` / `Maintainer` 五个小接口;`serviceImpl` 统一实现,业务模块的 `i18nSvc` 字段只声明实际依赖的小接口。

### P3 边界整理
- 抽出 `pkg/i18nresource` 通用资源加载器,接受 `Subdir` 与 `PluginScopeNamespace` 配置;`apidoc_i18n_loader.go` 与 `i18n.go` 的资源加载薄壳共用同一份实现,删除 ~280 行重复代码,但不让 apidoc 反向依赖 `internal/service/i18n`。
- 前端 `loadMessages` 拆分:运行时 bundle 失败 → 命中持久化缓存或回退;公共配置失败 → fire-and-forget 不阻塞;三方库 locale → 必须等待。
- WASM 自定义节解析 `parseWasmCustomSectionsForI18N` 与 `readWasmULEB128ForI18N` 提到 `pkg/pluginbridge/pluginbridge_wasm_section.go`,i18n 包仅调用 `pluginbridge.ReadCustomSection(content, name)`。

### 第三种语言:阿拉伯语作为压力测试
- `sys_i18n_locale` 注册启用 `ar-SA`;运行时语言列表、语言切换、缺失翻译检查、覆写来源诊断全链路必须自动覆盖到 `ar-SA`,不需要业务模块改任何代码。
- 宿主 `manifest/i18n/ar-SA.json` 与所有源码插件 `manifest/i18n/<plugin-id>/manifest/i18n/ar-SA.json` 补齐翻译;前端 `apps/lina-vben/apps/web-antd/src/locales/langs/ar-SA/*.json` 与 `packages/locales/src/langs/ar-SA/*.json` 补齐对应静态翻译。
- `apidoc i18n` 同步加 `ar-SA`;插件 apidoc 翻译资源同步补齐。
- 复数/数字格式化以 `ICU MessageFormat` 风格在前端约定 API,首期约束在 `count` 类文案上(列表统计、批量删除提示),不强求所有页面立刻使用。

### 基础 RTL(纳入本次)
- `<html dir>` 在切语言时自动跟随 (`ar-SA` → `rtl`,其他 → `ltr`);`Ant Design Vue` 的 `ConfigProvider` 同步注入 `direction="rtl"`。
- 验收标准:阿语下页面"内容正确、能用",布局允许有镜像偏差;LTR 页面无回归。

### 范围外(后续独立变更)
- 完整 RTL 设计语言:图标镜像、抽屉/通知滑入方向、表格固定列翻转、CSS logical properties 全面替换、菜单展开方向。
- 用户级语言偏好沉淀到 `sys_user`。
- 国际化可视化管理后台页面(导入导出 API 已存在)。
- 自动机器翻译/AI 翻译辅助。

## Capabilities

### New Capabilities
- `framework-i18n-runtime-performance`: 运行时翻译查找的零拷贝热路径、按扇区分层的缓存失效策略,以及前端 ETag/304 协商与持久化缓存能力。
- `framework-i18n-module-projection`: 业务模块边界内的本地化投影规则,封装翻译键推导、跳过策略与 fallback 选择,同时禁止 i18n 基础服务反向感知业务实体。
- `framework-i18n-source-text-registry`: 代码源文案命名空间显式注册机制,使缺失检查、诊断和导出能识别"由代码源拥有的翻译键"而无需 i18n 包反向感知具体业务模块。

### Modified Capabilities
- `framework-i18n-foundation`: 翻译服务接口拆分为 `LocaleResolver` / `Translator` / `BundleProvider` / `ContentProvider` / `Maintainer` 多个小接口;运行时翻译包接口新增 ETag/304 协商语义;新增 `ar-SA` 作为内置启用语言;基础 RTL 切换语义纳入语言切换链路;翻译资源加载在宿主、源码插件、动态插件之间共用统一资源加载器。
- `system-api-docs`: API 文档翻译资源加载改用统一资源加载器,同步覆盖 `ar-SA`。
- `plugin-runtime-loading`: WASM 自定义节读取能力提升为 `pluginbridge` 公共能力;插件运行时与 i18n 共用同一份解析器。
- `config-management`: 配置导入/导出表头改为通过翻译键解析,删除后端硬编码 `englishLabels`/`chineseLabels` Go 映射。
- `plugin-manifest-lifecycle`: 插件清单与生命周期需要在新增语言时自动覆盖,而无需修改宿主代码或插件代码。

## Impact

- **后端能力**:重写 `apps/lina-core/internal/service/i18n/` 内的 `Translate*` 热路径与缓存层;拆分 `Service` 接口;调整 5 个 `*_i18n.go` 适配文件并删除中心化投影器;抽出 `pkg/i18nresource`、`pkg/pluginbridge/pluginbridge_wasm_section.go`;`apidoc_i18n_loader.go` 与 `i18n.go` 的资源加载薄壳共用 loader;`config-management` 删除 Go 硬编码标签 map。
- **数据库**:`sys_i18n_locale` 新增一行 `ar-SA` 启用记录(seed DML);无新表、无 schema 改动。
- **前端能力**:`runtime-i18n.ts` 改走 `requestClient` 并接入 `localStorage`;`loadMessages` 拆分失败语义;`<html dir>` 与 antd `ConfigProvider` 接入 RTL 切换;新增 `ar-SA` 静态语言包。
- **资源文件**:宿主 `manifest/i18n/ar-SA.json`、`manifest/i18n/apidoc/ar-SA/*.json`;每个源码插件的 `manifest/i18n/ar-SA.json` 与 `manifest/i18n/apidoc/ar-SA/*.json`;前端 `packages/locales/src/langs/ar-SA/*.json` 与 `apps/web-antd/src/locales/langs/ar-SA/*.json`。
- **测试**:后端补充 `Translate` 热路径基准测试与缓存分层失效单元测试;前端补充运行时 ETag/持久化缓存单元测试;新增 E2E 用例覆盖 `ar-SA` 下的语言切换、关键页面文本完整性、`<html dir>` 切换断言;运行 `lina-review` 对前后端一致性校验。
- **交付与维护**:`README.md` 与 `manifest/i18n/README.md` 等文档同步说明 ETag、新增语言流程与 RTL 边界;`OpenSpec specs` 同步更新 `framework-i18n-foundation` 主规范的对应条款。
