---
name: lina-review
description: >-
  Code and spec review for the OpenSpec workflow. Automatically triggered after /opsx:apply tasks
  complete, after /opsx:feedback tasks complete, and before /opsx:archive runs. Also use when the
  user requests a code review, spec compliance check, or invokes /lina-review explicitly.
compatibility: 依赖 OpenSpec CLI、GoFrame v2 技能、lina-e2e 技能。
---

# Lina 审查

`OpenSpec` 开发工作流的结构化代码与规范审查。

**规范来源**：`AGENTS.md` 是所有审查标准的唯一事实来源。

**交互语言**：与用户交互的内容语言（如审查结果展示）以用户上下文使用的语言为准，用户使用英文则使用英文，用户使用中文则使用中文。

---

## 触发条件

**自动触发：**
- `/opsx:apply` 每个任务完成后
- `lina-feedback` 每个任务完成后
- `/opsx:archive` 执行前

**手动触发：**
- 用户明确请求："审查代码"、"检查规范合规性"、"/lina-review"

---

## 审查工作流

### 1. 确定审查范围

确定需要审查的内容：

1. **任务完成后** — 审查已完成任务修改/创建的文件
2. **归档前** — 审查当前 `OpenSpec` 变更中的所有变更
3. **手动调用** — 请用户指定范围或使用当前变更

**范围收集的强制规则：**

1. 从仓库状态开始，而非仅依赖 `git diff`：
   ```bash
   git status --short
   git ls-files --others --exclude-standard
   ```
2. 将**所有已跟踪和未跟踪的变更**视为审查候选项，包括：
   - 已暂存文件
   - 未暂存文件
   - 显示为 `??` 的未跟踪文件
   - 显示为 `?? path/` 的未跟踪目录
3. 当 `git status --short` 报告未跟踪目录时，在审查前将其展开为具体文件：
   ```bash
   find <path> -type f
   # 或优先使用：
   rg --files <path>
   ```
4. 如果任务运行了代码生成器（如 `make ctrl`、`make dao`、`codegen` 脚本）或产生了新测试文件，即使它们未出现在 `git diff` 中，也必须显式将生成的未跟踪文件纳入审查范围。
5. `git diff` 仅可作为状态收集后的辅助缩小工具。它**永远不足以**单独定义审查范围。

运行 `openspec status --change "<name>" --json` 了解当前变更状态。

### 2. 加载关键规范

读取 `AGENTS.md` 加载所有规范。这是唯一事实来源。

### 3. 后端代码审查

**触发条件**：任何 `*.go` 文件的变更

1. 调用 `goframe-v2` 技能检查 `GoFrame` 框架规范
2. 对照 `AGENTS.md` 后端代码规范进行检查

### 4. RESTful API 审查

**触发条件**：任何 `API` 端点变更

对照 `AGENTS.md` `API` 设计规范进行检查，包括：
1. `HTTP` 方法和资源路径是否符合 `RESTful` 规则
2. `API DTO` 文档元数据的完整性
3. `API` 文档国际化合规性：
   - `g.Meta` 和手写的 `DTO` 文档标签必须使用可读的英文源文本，禁止使用中文源文本或不透明的国际化占位符
   - 接口文档本地化必须使用专用的 `apidoc i18n JSON` 资源，与运行时前端 `UI` 的 `i18n` 语言包隔离
   - 必须使用稳定的结构化 `apidoc` 键而非源文本键；宿主 `core.*` `apidoc` 键保留在 `lina-core` 资源中，插件 `plugins.*` `apidoc` 键保留在各插件自己的 `manifest/i18n/<locale>/apidoc/**/*.json` 中
   - `en-US` `apidoc JSON` 文件保留为空占位符（英文文档直接使用源文本）
   - 禁止为 `internal.model.entity.*` 等生成的 `Schema` 元数据添加服务层中英文回退映射或 `apidoc JSON` 映射
   - 生成的元数据按数据源原文展示；`eg/example` 示例值不翻译且不纳入 `apidoc i18n` 资源
   - 必须包含测试或审查检查以防止英文源文本变更时遗漏非英文 `apidoc` 翻译

### 5. 项目规范审查

**触发条件**：任何实现变更

对照 `AGENTS.md` 架构设计规范和代码开发规范进行检查，按功能职责分为以下三类审查。

---

#### 后端代码规范审查

**触发条件**：任何 `*.go` 文件的变更

**文件命名规范**

对每个后端服务，组件主文件可命名为 `<component>.go`；该组件目录中的每个非主 `Go` 源文件必须使用 `<component>_*.go` 前缀，测试文件也必须遵循相同的前缀规范。

**错误码使用（bizerr）**

对每个可能向 `API` 调用方、插件调用方、源码插件后端调用方、`WASM` 宿主服务调用方或其他外部调用边界返回错误的后端变更，还需执行**面向调用方的 bizerr 审查**：
1. 可能进入响应载荷的业务错误、授权错误、校验错误和用户可见的失败原因，必须通过 `apps/lina-core/pkg/bizerr`（`bizerr.NewCode`、`bizerr.WrapCode` 或等价的本地封装）创建或包装。
2. 每个模块必须在其 `*_code.go` 文件中定义可复用的 `bizerr.Code` 值；调用点必须引用这些变量，禁止硬编码机器错误码、运行时 `i18n` 键、原始数字码或用户可见的消息文本。
3. 当返回的错误可能到达统一 `HTTP` 响应、插件桥接/宿主服务响应、源码插件 `API` 响应或其他调用方可见的载荷且未被包装为 `bizerr` 时，标记直接使用的 `gerror.New/Newf/NewCode/NewCodef/Wrap/Wrapf/WrapCode/WrapCodef`、`errors.New` 和 `fmt.Errorf`。
4. 仅在启动/测试代码、内部开发诊断、在边界前被包装的纯低层技术原因，或永远不会作为调用方可见接口错误出现的运维日志路径中，才允许直接使用 `gerror`、`errors` 或 `fmt.Errorf`。

**日志上下文传播**

对每个涉及日志记录、辅助函数、清理路径、异步回调、中间件、控制器或服务层调用链的后端变更，还需执行**日志上下文传播审查**：
1. 后端宿主代码和源码插件后端代码必须将上游 `context.Context` 沿调用链传递到任何可能打印日志的底层方法。
2. 调用项目日志器时（如 `logger.Info(ctx, ...)`、`logger.Warningf(ctx, ...)`、`logger.Errorf(ctx, ...)` 等方法），当存在业务/请求上下文时必须使用传播的 `ctx`，以保留追踪、请求、租户和用户元数据。
3. 当上游 `ctx` 可用或可通过函数签名传递时，标记在日志调用处使用的 `context.Background()` 或 `context.TODO()`。
4. `context.Background()` 仅在真正的根上下文中可接受，包括启动期、进程级初始化、测试构造或确实没有请求/业务上下文的代码路径；此类使用不得掩盖嵌套调用链中可避免的上下文中断。

---

#### 国际化（i18n）治理审查

**触发条件**：任何涉及 `i18n` 资源、前端文案、`manifest/i18n`、字典枚举或语言配置的变更

**字典与语言包单一来源**

对每个引入或修改枚举业务值（状态、类型、作用域、模式、来源等）的前端变更，还需执行**字典与语言包单一来源审查**：
1. 如果后端 `sys_*` 字典已拥有相同的枚举（通过 `manifest/sql` 或运行时字典注册），前端必须通过 `useDictStore().getDictOptions(...)` / `getDictOptionsAsync(...)` 消费该字典，禁止在 `frontend/pages/data.ts`、`*.vue` 表单 Schema 或同级文件中维护包含 `$t(...)` 标签的并行 `options: [...]` 字面量数组。仅在后端字典不拥有该枚举时，才可接受静态 `options` 数组。
2. 当同一字段在表格列、搜索表单、创建/编辑表单和任何预览/详情界面中都渲染为 `DictTag` 时，每个界面必须使用同一个字典作为唯一事实来源。标记在同一字典的 `DictTag` 消费方旁边注入了字面量 label/value 对的界面。
3. 共享的前端 `pages.*` 命名空间（如 `apps/lina-vben/apps/web-antd/src/locales/langs/<locale>/pages.json`）不得携带与 `sys_*` 字典标签重复的翻译。如果宿主 UI 合理地渲染了插件字典拥有的枚举，宿主后端（如 `usermsg`、`notify`）必须在其 API 响应上提供本地化的标签字段，宿主前端必须直接消费该标签；禁止添加镜像字典标签的 `pages.status.<enum>` 键作为跨模块耦合的变通方案。
4. 当后端拥有的数据字段需要在前端本地化展示时，优先在后端服务/API 输出上添加本地化字段（如 `typeLabel`、`statusLabel`）并直接消费。前端不得维护镜像字典语义的 `type === N ? $t(...) : $t(...)` 映射辅助函数。
5. 如果变更删除了前端 `options` 字面量，还需确认同一变更中删除了所有孤立的 `pages.*` 键、回退数组和 `getXxxLabel` 辅助函数，避免遗留失效的翻译键。

**国际化影响面**

对每个功能变更，还需执行**国际化影响审查**：
1. 识别变更是否新增、修改或删除了用户可见的文本、菜单、路由、按钮、表单字段、表格列、状态标签、提示信息、校验/错误信息、`API` 文档、插件清单、种子数据标签或其他本地化内容。
2. 验证对应的 `i18n JSON` 资源是否已按需新增、更新或删除，包括前端运行时语言包、宿主/插件 `manifest/i18n` 资源以及专用的 `apidoc i18n JSON` 文件。
3. 标记硬编码的用户界面文本、缺失的翻译键、功能删除后遗留的失效/孤立翻译条目，以及未明确评估 i18n 影响的变更。
4. 如果变更无 `i18n` 影响，要求审查结果中明确说明该结论。

**硬编码双语映射**

对每个本地化导出/导入表头、字段标签或元数据投影的后端变更，还需执行**硬编码双语映射审查**：
1. 业务模块不得维护 `englishLabels`、`chineseLabels` 或等价的 `locale-to-label` 表等 `Go` 映射来处理用户可见文本。
2. 此类标签必须通过运行时 i18n 键（如 `config.field.<name>`）和对应的宿主/插件 `manifest/i18n/<locale>/*.json` 资源解析，包括宿主嵌入交付资源时的打包清单副本。

**源文本命名空间注册**

对每个涉及源文本支撑的缺失消息行为的变更，还需执行**源文本命名空间注册审查**：
1. `apps/lina-core/internal/service/i18n/` 在判断键是否由源文本支撑时，不得硬编码 `job.handler.` 或 `job.group.default.` 等业务拥有的前缀。
2. 拥有源文本支撑运行时键的业务模块必须通过其自身的包初始化或装配路径调用 `i18n.RegisterSourceTextNamespace(prefix, reason)` 注册命名空间。
3. 缺失消息测试必须覆盖未注册命名空间仍出现在缺失结果中以及注册命名空间从缺失结果中消失两种情况。

**国际化基础层边界**

对每个添加本地化投影逻辑的后端变更，还需执行**国际化基础层边界审查**：
1. `apps/lina-core/internal/service/i18n/` 是基础翻译服务，不得为菜单、字典、系统配置、任务管理、角色、用户消息、通知或插件运行时等模块拥有业务实体投影规则、业务保护规则或业务翻译键派生逻辑。
2. 标记任何命名为 `ProjectMenu`、`ProjectDictType`、`ProjectBuiltinJob`、`ProjectMeta` 等接受业务实体来修改显示字段的 i18n 包 API。
3. 业务模块必须将自身的本地化投影规则保持在其包边界内，仅依赖窄接口的 i18n 能力（如 `ResolveLocale`、`Translate`、`TranslateSourceText`）。
4. apidoc 和运行时 i18n 共用的资源加载工具必须放在稳定的公共组件中（如 `pkg/i18nresource`）；如果会使无关服务依赖运行时 i18n 服务包，则不要放在 `internal/service/i18n` 中。

**最小化 i18n 接口依赖**

对每个涉及 `i18n` 服务消费者的后端变更，还需执行**最小化 i18n 接口依赖审查**：
1. 业务服务、控制器、中间件和插件适配器在仅使用方法子集时，不得将字段声明为 `i18n.Service` / `i18nsvc.Service` / `internali18n.Service`。
2. 当调用方仅使用一两个方法时，优先使用包级别的窄接口；否则依赖导出的小接口 `LocaleResolver`、`Translator`、`BundleProvider`、`ContentProvider` 和 `Maintainer`，或这些接口的显式组合。
3. 完整的 `i18n.Service` 组合体仅保留给构造函数、服务工厂和确实需要完整接口的罕见集成点；此类使用必须在审查中说明理由。

**运行时 i18n 缓存卫生**

对每个涉及宿主 `i18n` 服务或其调用方的变更，还需执行**运行时 i18n 缓存卫生审查**：
1. 热路径翻译调用（`Translate`、`TranslateSourceText`、`TranslateOrKey`、`TranslateWithDefaultLocale`）不得克隆运行时消息目录。标记任何在逐键查找路径上引入 `cloneFlatMessageMap` 或等价的全映射复制的代码；缓存返回只读合并视图，直接 `merged[key]` 访问是约定。
2. `apps/lina-core/internal/service/i18n/` 外的代码不得克隆 i18n 服务返回的运行时消息目录。服务负责在将映射交给外部消费者前进行克隆（`BuildRuntimeMessages`、`ExportMessages`）；业务模块和控制器必须将返回的映射视为只读。
3. 每次调用 `InvalidateRuntimeBundleCache` 必须传入显式的 `i18n.InvalidateScope`。标记任何省略 scope 或使用零值 scope 且无正当理由的调用 — 清除所有语言和所有扇区仅保留给完整的进程级重载路径，且必须包含注释说明为何无法使用更窄的范围。插件生命周期失效必须设置 `Sectors: []Sector{SectorDynamicPlugin}` 和 `DynamicPluginID`；数据库导入必须设置 `Sectors: []Sector{SectorDatabase}` 和受影响的 `Locales`。
4. 每次调用 `InvalidateContentCache` 必须传入显式的 `i18n.ContentInvalidateScope`。纯 `ContentInvalidateScope{}`（全量清除）仅允许在测试清理或完整重载路径中使用；生产调用方必须按 `BusinessType` 和/或 `Locale` 限定范围。
5. 任何贡献运行时缓存的新扇区必须注册在 `apps/lina-core/internal/service/i18n/i18n_cache.go`（`Sector` 枚举和 `mergeLocaleSectors` 中的合并顺序）。禁止在业务模块中引入临时扇区。

**固定 LTR 方向**

对每个涉及语言方向处理的前端变更，还需执行**固定 LTR 方向审查**：
1. 当前 `framework-i18n-improvements` 范围将 `<html dir>` 和 `Ant Design Vue` `ConfigProvider.direction` 固定为所有运行时语言的 `ltr`。
2. 默认配置的 `i18n` 部分不得暴露 `direction`，前端不得维护 `RTL_LOCALES` 或等价的方向注册。
3. 仍需标记语言切换后出现的不一致重叠、不可读文本、导航损坏或控件不可用等问题。

**语言注册单一来源**

对每个涉及内置运行时语言的变更，还需执行**语言注册单一来源审查**：
1. 内置运行时语言必须从包含直接运行时 `JSON` 文件的 `manifest/i18n/<locale>/` 目录中发现；默认配置的 `i18n` 部分仅可提供默认语言、`enabled`、排序、原生名称元数据和启用语言白名单。
2. 标记新增的按语言 `Go` `常量、SQL` 种子行、前端 `SUPPORT_LANGUAGES` 添加项、前端 `RTL_LOCALES` 添加项或语言特定的 `switch` 分支为严重问题，除非它们是包级别的通用回退规则而非项目语言注册。
3. 添加内置语言不得要求修改后端 `Go` 枚举代码、宿主 `SQL` 文件或前端 `TS` 语言列表。
4. 当 `i18n.enabled=false` 时，前端语言切换器必须隐藏，运行时语言解析必须使用 `i18n.default`。

---

#### 分布式缓存一致性审查

**触发条件**：任何引入、修改、失效、刷新或依赖缓存数据的设计或实现变更

对每个引入、修改、失效、刷新或依赖缓存数据的设计或实现变更，还需执行**分布式缓存一致性与可靠性审查**：
1. `OpenSpec` 设计、任务说明、实现或审查结论必须说明缓存的权威数据源、一致性模型、失效/刷新触发点、跨实例同步机制、最大可接受陈旧时间和故障降级策略。如果变更无缓存影响，需在审查结果中明确说明。
2. 缓存行为必须显式地由宿主现有的集群模式开关和拓扑抽象（`cluster.enabled`、`cluster.Service` 或包级别的包装接口）控制。单机模式下，进程内缓存配合本地失效是可接受的，不应要求分布式协调；集群模式下，同一缓存必须使用跨节点失效、共享修订号、事件/消息广播、共享分布式缓存、主节点协调或等价策略。
3. 本地进程内缓存不得被视为自身具有分布式一致性。标记在 `cluster.enabled=true` 时仅依赖进程内存、本地定时器或节点本地状态的缓存设计，或复制临时集群标志而非使用共享拓扑源的缓存控制分支。
4. 对于授权、配置、插件生命周期状态、租户隔离、字典、路由、i18n 资源、API 文档包和安全敏感元数据等关键运行时数据，缓存失效/更新路径必须是有范围的、幂等的、可重试或安全重复的，并可通过日志/指标/测试观测。标记广泛的全量清除失效，除非是有正当理由的完整重载路径。
5. 验证缓存更新与成功的权威数据源写入耦合，而非在事务仍可能回滚前发出。遗漏的失效事件、进程重启和过期的分布式条目必须有恢复路径（如 TTL、版本检查、重建/穿透读取、对账或显式重载）。
6. 测试或审查证据应在适当的风险级别覆盖变更的缓存行为，特别是单机与集群模式分支、多实例失效、有界陈旧性、重试/重建行为，以及缓存后端不可用时的面向调用方行为。

### 6. SQL 规范审查

**触发条件**：`apps/lina-core/manifest/sql/`、`apps/lina-core/manifest/sql/mock-data/`、`apps/lina-plugins/**/manifest/sql/` 下新增或修改的文件，或相关交付文档中嵌入的 `SQL` 片段

对照 `AGENTS.md` `SQL` 文件管理规范进行检查，至少覆盖：
1. 文件命名、版本管理和单迭代单文件规则
2. 种子数据与模拟数据的分离
3. **幂等执行安全性** — `SQL` 必须可安全多次运行，不会出现重复对象错误或重复种子数据；验证是否使用了 `IF [NOT] EXISTS`、`IF EXISTS`、`INSERT IGNORE` 或等价的安全重入模式
4. **种子数据写入风格合规性** — 交付的 `SQL` 必须拒绝 `INSERT ... ON DUPLICATE KEY UPDATE`，且拒绝在种子/模拟/安装数据中显式写入 `AUTO_INCREMENT` `id` 列
5. `Schema`/数据变更是否仍与当前变更范围和部署路径匹配

### 7. E2E 测试审查

**触发条件**：`hack/tests/e2e/` 目录下新增或修改的 E2E 测试文件

1. 调用 `lina-e2e` 技能检查测试规范
2. 对照 `AGENTS.md` E2E 测试规范进行检查

### 8. 生成审查报告

```markdown
## Lina 审查报告

**变更：** <变更名称>
**范围：** <任务级 / 全部变更>
**审查文件数：** <数量>
**范围来源：** `git status --short` + `git ls-files --others --exclude-standard` + 任务/变更上下文

### 后端代码审查
✓ 全部通过 / ⚠ 发现 N 个问题

### RESTful API 审查
✓ 所有端点合规 / ⚠ 发现 N 个违规
✓ API 文档国际化合规 / ⚠ 发现 N 个 apidoc i18n 问题

### 后端代码规范审查
✓ 文件命名规范合规 / ⚠ 发现 N 个违规
✓ 错误码（bizerr）使用合规 / ⚠ 发现 N 个违规
✓ 日志上下文传播合规 / ⚠ 发现 N 个违规

### 国际化（i18n）治理审查
✓ 字典与语言包单一来源合规 / ⚠ 发现 N 个问题
✓ 国际化影响面已评估 / ⚠ 发现 N 个 i18n 遗漏
✓ i18n 基础层边界、接口依赖、缓存卫生、语言注册合规 / ⚠ 发现 N 个问题

### 分布式缓存一致性审查
✓ 已审查分布式缓存一致性 / ⚠ 发现 N 个缓存一致性问题

### SQL 审查
✓ 无 SQL 变更 / ✓ SQL 变更合规 / ⚠ 发现 N 个 SQL 问题

### E2E 测试审查
✓ 测试遵循规范 / ⚠ 发现 N 个问题

### 摘要
- **严重：** N（归档前必须修复）
- **警告：** N（建议修复）

### 建议操作
1. [具体操作及 AGENTS.md 引用]
```

---

## 问题严重性

| 级别 | 行为 |
|------|------|
| **严重** | 阻塞归档，必须修复 |
| **警告** | 显示但允许继续 |

---

## 审查触发点

| 工作流步骤 | 行为 |
|-----------|------|
| `/opsx:apply` 任务完成 | 审查，提供在下一任务前修复问题的选项 |
| `lina-feedback` 任务完成 | 审查，在标记完成前修复问题 |
| `/opsx:archive` | 审查所有变更，严重问题阻塞归档 |

---

## 硬性规则

- **`AGENTS.md` 是唯一事实来源** — 所有规范引用均指向它
- 仅检查与变更文件相关的类别
- 范围识别必须包含未跟踪文件和展开的未跟踪目录；永远不要仅依赖 `git diff`
- 警告不阻塞 — 仅严重问题阻塞归档
- 问题报告中包含文件路径和行号
- 简单问题提供自动修复选项
