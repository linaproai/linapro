## Context

当前 `framework-i18n-foundation` 和 `framework-i18n-improvements` 已经为宿主提供了请求语言解析、运行时语言包聚合、缓存优化、ETag 协商、源码插件/动态插件资源合并和缺失翻译检查。统一响应中间件也已经通过 `i18n.LocalizeError(ctx, err)` 按请求语言翻译错误文本。

本次盘点发现的问题不在“有没有翻译基础设施”，而在“运行时可见文案是否进入了基础设施”。现在大量业务逻辑仍把中文或中英混排字符串直接写进 `gerror.New`、导入失败 `Reason`、Excel 表头、插件桥接错误、插件生命周期 `Message`、前端页面 `title/label`。这些字符串有些会直接返回给前端，有些会写入导出文件，有些会通过插件协议返回给开发者或插件运行时。详见 `runtime-message-i18n-audit.md`。

现有关键约束：

- 项目没有历史兼容包袱，可以直接调整统一响应错误字段、插件桥接错误载荷和内部错误构造方式。
- OpenAPI/API DTO 源文本必须保持英文，不能把 API 文档 i18n 资源和运行时 UI i18n 资源混用。
- 运行时翻译内容以 `manifest/i18n/<locale>/*.json` 和插件自己的 `manifest/i18n/<locale>/*.json` 为事实源，不重新引入数据库翻译表。
- 后端热路径必须复用已有运行时翻译缓存，不能在错误处理、列表投影或批量导出循环中重建整包语言资源。
- 用户输入、用户维护的业务名称和外部系统返回的原始错误不是框架可翻译文案，必须原样保留或作为参数嵌入本地化模板。

## Goals / Non-Goals

**Goals:**

- 建立运行时文案分类规则，明确哪些字符串必须 i18n、哪些必须保持英文日志、哪些属于用户数据或开发者诊断。
- 把后端业务错误从“自由文本”升级为结构化消息：稳定错误码、翻译键、参数、英文源文案和 HTTP/gcode 语义。
- 让统一响应和前端请求拦截器优先消费结构化错误字段，同时保留本地化后的 `message` 方便简单调用方显示。
- 清理用户、字典、系统参数、插件、任务、文件、通知、组织、监控等模块中的高风险中文硬编码。
- 让导入模板、导入失败原因、导出 sheet 名、表头和枚举值按请求语言一次性解析并复用。
- 让插件桥接和宿主服务错误具备稳定机器码，开发者诊断默认使用英文，管理端展示通过 i18n key 本地化。
- 为前端页面、插件前端和请求错误展示补齐 `$t` / 运行时语言包消费规则。
- 增加自动化扫描与缺失翻译测试，避免后续新增硬编码文案。

**Non-Goals:**

- 不引入线上翻译编辑后台、机器翻译或第三方翻译平台。
- 不把运维日志按用户语言本地化；日志使用稳定英文和结构化字段。
- 不翻译用户输入内容、数据库中用户维护的名称、外部系统返回原文和 SQL/协议/文件路径等技术参数。
- 不在本次方案中支持 RTL 布局或用户级语言偏好。
- 不通过修改旧 SQL seed 来注册语言；语言继续由 `manifest/i18n/<locale>/` 目录和默认配置元数据驱动。

## Decisions

### 决策一：运行时文案按使用面分类治理

**选择**：把代码中的字符串分为六类，并为每类指定处理策略：

- `UserMessage`：API 错误、业务校验、前端 toast、管理端结果提示，必须用错误码或翻译键。
- `UserArtifact`：Excel 表头、sheet 名、导入模板示例、导入失败原因、导出枚举值，必须按请求语言渲染。
- `UserProjection`：菜单、字典、角色、任务、审计/操作日志等后端拥有展示数据，必须用稳定业务键投影。
- `DeveloperDiagnostic`：插件协议、WASM host call、清单校验、CLI 诊断，必须有稳定机器码，默认英文源文案；进入管理端时再本地化。
- `OpsLog`：服务端日志和指标，使用英文和结构化字段，不参与运行时 i18n。
- `UserData`：用户输入、外部系统内容、数据库业务值，原样保存和返回，不自动翻译。

**原因**：同一字符串不能同时承担日志检索、前端展示和协议状态判断。先分类可以避免把所有字符串都塞进 i18n，也避免把用户数据误翻译。

**备选方案**：全量扫描中文并全部替换成翻译键。未采用，因为会误伤注释、测试、用户示例、协议常量和运维日志，维护成本高且语义不清。

### 决策二：新增结构化运行时消息错误模型

**选择**：在 `apps/lina-core/pkg/bizerr` 引入统一的业务错误构造能力。`bizerr` 只定义业务语义、运行时 i18n 元数据和最接近的 GoFrame 类型错误码，不再为每个业务错误分配自定义整型响应码。统一响应中的 `code` 字段回归 GoFrame `gcode.Code` 的类型码，例如 `CodeInvalidParameter`、`CodeNotFound`、`CodeNotAuthorized`；具体业务语义通过 `errorCode`、`messageKey` 和 `messageParams` 表达。

```go
type Meta struct {
    ErrorCode string     // 机器可读业务语义码，如 USER_NOT_FOUND
    MessageKey string    // 运行时 i18n key，如 error.user.not.found
    Fallback  string     // 英文源文案
    TypeCode  gcode.Code // GoFrame 类型错误码，如 gcode.CodeNotFound
}
```

每个组件在独立的 `*_code.go` 文件中集中定义业务错误，例如：

```go
var CodeDictTypeExists = bizerr.MustDefine(
    "DICT_TYPE_EXISTS",
    "Dictionary type already exists",
    gcode.CodeInvalidParameter,
)
```

业务代码只引用组件错误定义，不直接写 `errorCode`、`messageKey` 或裸数字：

```go
return 0, bizerr.NewCode(CodeDictTypeExists)
```

所有会进入 HTTP API、插件调用、源码插件后端接口、WASM host service 或其他调用端响应载荷的接口错误，都必须通过 `bizerr.NewCode`、`bizerr.WrapCode` 或等价封装创建/包装。`bizerr` 内部使用 GoFrame `gerror` 创建实际错误对象，因此仍保留堆栈、cause 和 `gcode.Code` 能力；直接 `gerror.New*`、`errors.New` 或 `fmt.Errorf` 仅允许作为不会暴露给调用端的内部诊断，或作为被 `bizerr.WrapCode` 立即包装的底层 cause。

业务语义标识必须按模块命名空间治理。宿主模块使用 `<MODULE>_<CASE>` 形式，例如 `USER_NOT_FOUND`、`DICT_TYPE_EXISTS`；插件模块使用 `<PLUGIN>_<MODULE>_<CASE>` 或等价前缀，例如 `ORG_CENTER_DEPT_NOT_FOUND`。同一模块的业务错误定义集中维护在该模块的 `*_code.go`，禁止业务调用点直接写字符串，禁止跨模块复用语义不一致的业务错误定义。

统一响应中间件识别该元数据后输出：

```json
{
  "code": 65,
  "message": "用户不存在",
  "errorCode": "USER_NOT_FOUND",
  "messageKey": "error.user.not.found",
  "messageParams": {"id": 12},
  "data": null
}
```

`code` 是 GoFrame 类型错误码，用于表达错误类别；`message` 是服务端按请求语言解析后的展示文本；`errorCode` 和 `messageKey` 供前端、测试、插件或第三方调用方做稳定业务语义判断。

**原因**：当前 `LocalizeError` 只能把 `err.Error()` 当作 key 或源文本翻译，无法稳定携带机器码、参数和 fallback，也无法区分用户可见文案与开发者诊断。结构化模型保留现有中间件入口，同时补齐契约。

**备选方案**：继续约定 `gerror.New("error.user.notFound")`。未采用，因为纯字符串 key 无法强制参数完整性，也无法防止业务继续写中文自由文本。

### 决策三：语言资源按 locale 目录和语义域拆分，不新增数据库表

**选择**：所有新增运行时文案写入宿主或插件自己的 `manifest/i18n/<locale>/*.json`。语言编码作为第一层目录，运行时资源与 API 文档资源在同一语言目录下分区维护：

```text
manifest/i18n/
  en-US/
    framework.json
    menu.json
    dict.json
    config.json
    error.json
    artifact.json
    public-frontend.json
    apidoc/
      common.json
      core-api-user.json
  zh-CN/
    framework.json
    menu.json
    dict.json
    config.json
    error.json
    artifact.json
    public-frontend.json
    apidoc/
      common.json
      core-api-user.json
```

源码插件采用相同布局，资源仍归属于插件目录：

```text
apps/lina-plugins/<plugin-id>/manifest/i18n/
  en-US/
    plugin.json
    menu.json
    error.json
    page.json
    artifact.json
    apidoc/
      plugin-api-main.json
  zh-CN/
    plugin.json
    menu.json
    error.json
    page.json
    artifact.json
    apidoc/
      plugin-api-main.json
```

运行时 JSON 文件按语义域分类：

| 文件 | 内容范围 |
| ---- | -------- |
| `framework.json` | 框架名称、框架描述、语言名称、宿主通用展示文案 |
| `menu.json` | 宿主或插件菜单、路由标题、导航相关文案 |
| `dict.json` | 字典类型名称、字典项标签、后端内建枚举展示文案 |
| `config.json` | 系统配置项名称、描述、分组、公开前端配置展示文案 |
| `error.json` | `bizerr` 业务错误、校验错误、调用端可见失败原因 |
| `artifact.json` | 导入导出 sheet 名、表头、模板字段、失败原因、交付物枚举展示 |
| `public-frontend.json` | 登录页、公开页、默认工作台公共文案 |
| `plugin.json` | 插件名称、插件描述、插件元数据和插件通用文案 |
| `page.json` | 插件前端页面标题、表单、表格、按钮、提示文案 |

当某个语义域过大时，可以继续按业务模块拆分，例如 `user.json`、`job.json`、`plugin-lifecycle.json`，但文件名必须表达业务含义，不使用 `00-`、`10-` 这类数字顺序前缀。加载顺序只用于保证结果确定性，不表达业务覆盖关系；同一 locale 内出现重复 key 时，校验必须失败。

运行时和 API 文档资源的加载边界如下：

- 运行时 i18n 只加载 `manifest/i18n/<locale>/*.json`，不递归进入 `apidoc/`。
- API 文档 i18n 只加载 `manifest/i18n/<locale>/apidoc/**/*.json`。
- 宿主 API 文档文件使用 `common.json` 和 `core-api-<module>.json` 命名。
- 插件 API 文档文件使用 `plugin-api-<module>.json` 或 `<plugin-id>-api-<module>.json` 命名。
- `en-US` API 文档仍以 DTO 英文源文本为事实源，`apidoc` 下只保留必要空占位或非源文本补充，不重新维护英文映射。

运行时 key 继续按既有命名空间治理：

- 宿主错误：`error.<domain>.<case>`
- 宿主导入导出、模板和交付物：`artifact.<module>.<section>.<field>`
- 宿主业务投影和字典内建文案：按既有运行时命名空间维护，例如 `dict.<domain>...`、`user.<domain>...`
- 前端页面：沿用 `pages.<module>...`
- 插件文案：`plugin.<plugin-id>...`

英文源文案保留在代码 fallback 中，用于开发者理解和缺失资源时兜底；`zh-CN`、`en-US`、`zh-TW` 目录必须都包含对应运行时键。`apidoc` 资源继续只服务接口文档，不参与运行时错误和 UI 展示。动态插件源码资源可以拆分维护，但打包为 Wasm runtime artifact 时仍按 locale 合并为运行时 i18n asset 和 apidoc i18n asset，避免改变运行时协议形态。

**原因**：现有 i18n 改进已经明确去掉运行时翻译数据库双源模型。继续使用文件资源可以保持部署、缓存和缺失检查简单；按 locale 目录拆分后，新增语言时只需要新增一个语言目录，维护者可以在同一目录下同时处理运行时文案和 API 文档文案。

**备选方案**：保留 `manifest/i18n/<locale>.json` 单文件。未采用，因为运行时 key 数量增长后文件过大，查找、审查和冲突定位成本持续上升。

**备选方案**：把错误消息写入字典或配置表。未采用，因为错误消息属于代码契约，和运行时可编辑业务配置生命周期不同。

### 决策四：导入导出使用批量本地化上下文

**选择**：为导入导出建立请求级 `MessageRenderer` 或等价窄接口，进入导出/导入流程时一次性解析当前 locale 和本模块所需 key，循环内只做 map 查找和参数格式化。`excelutil` 继续只负责 Excel 文件操作，不直接依赖 i18n service；业务服务负责传入已本地化的 sheet 名、表头、枚举文本和失败原因。

**原因**：导出可能有 10000 行，如果每个单元格都调用完整翻译链会放大锁竞争和字符串处理成本。批量预取让本地化成本与字段数量相关，而不是与行数相关。

**备选方案**：在 `excelutil.SetCellValue` 中自动翻译字符串。未采用，因为 Excel 工具层无法知道字符串是翻译键、用户数据还是技术值，会误翻译。

### 决策五：插件桥接错误分成协议层和展示层

**选择**：插件桥接、WASM host call 和 host service 协议返回稳定状态码与错误码，默认错误源文案使用英文开发者诊断；当这些错误进入管理端 UI 或宿主统一响应时，再用 `messageKey` 和 locale 渲染。动态插件 guest 返回的 JSON 错误也应支持 `errorCode/messageKey/messageParams/message`，宿主透传时保留结构化字段。

**原因**：插件协议是开发者和运行时之间的稳定契约，不能依赖某种自然语言字符串判断错误；但管理端用户仍需要本地化提示。

**备选方案**：host call payload 直接返回本地化中文或英文。未采用，因为 guest 插件不一定和当前管理端用户同语言，协议层本地化会造成调试与测试不稳定。

### 决策六：前端请求拦截器优先消费结构化字段

**选择**：默认工作台请求错误处理顺序调整为：

1. 如果后端返回 `messageKey`，前端用 `$t(messageKey, messageParams)` 渲染。
2. 否则使用后端已按请求语言本地化的 `message`。
3. 否则使用请求库默认错误文本。

页面级文案必须使用 `$t` 或运行时语言包，禁止在 `title`、`label`、`placeholder`、`Modal.confirm`、`message.*` 等用户可见位置直接写中文或中英混排字符串。

**原因**：后端统一响应会给简单调用方提供本地化 `message`，但前端拥有更完整的语言运行时和动态切换能力，优先 `messageKey` 能避免页面语言已切换但旧请求消息仍停留在旧语言。

**备选方案**：前端完全信任后端 `message`。未采用，因为前端已经有运行时语言包和 `$t`，结构化字段更利于语言切换和测试。

### 决策七：日志保持英文，审计展示按语言投影

**选择**：`logger.*` 调用中的运维日志统一使用英文固定模板和结构化字段；业务错误进入日志时记录 `errorCode/messageKey` 和关键参数，不记录本地化后的用户展示文案作为唯一信息。操作日志、登录日志、任务日志和插件升级结果等面向用户展示的数据存储稳定代码和参数，列表、详情和导出时按请求语言投影。

**原因**：日志是机器检索和跨团队排障资产，按用户语言本地化会降低稳定性。审计展示是用户界面，应本地化。

**备选方案**：日志也按请求语言输出。未采用，因为异步任务、启动期流程和多用户共享日志没有可靠的单一请求语言。

### 决策八：扫描门禁以“运行时可见位置”为主，不做裸字符一刀切

**选择**：新增硬编码文案扫描脚本，使用 Go AST 和前端 AST/ESLint 规则识别高风险位置：

- Go：`gerror.New*`、`gerror.Wrap*`、`panic(gerror...)`、`Reason/Message/Fallback` 字段、导出表头数组、状态文本映射、插件桥接错误构造。
- Vue/TypeScript：`title/label/placeholder`、模板文本节点、`message.*`、`notification.*`、`Modal.confirm`、表格列定义。
- 允许清单：注释、测试 fixture、用户示例数据、技术单位、协议常量、英文运维日志。

扫描结果进入 CI 或本地 `make` 检查；新增例外必须在 allowlist 中说明原因和归属分类。

**原因**：单纯 `rg "\p{Han}"` 会产生大量误报，无法长期执行。基于语义位置扫描可以作为可维护门禁。

**备选方案**：只依赖 code review。未采用，因为当前硬编码分布很广，仅人工审查容易回归。

## Risks / Trade-offs

- **风险：错误响应字段变化影响前端和插件调用方** -> 项目无兼容包袱，可以直接升级统一响应；同时保留 `message` 字段降低调用成本。
- **风险：翻译键数量快速增加** -> 使用命名空间规范和缺失检查；按 locale 目录和语义域拆分资源文件，插件文案放在插件自己的 manifest。
- **风险：批量导出本地化影响性能** -> 使用请求级批量预取和 map 查找；禁止循环内构建 bundle。
- **风险：扫描规则误报或漏报** -> 第一阶段以 warn/report 输出生成清单，稳定后改为阻断；allowlist 必须带分类和原因。
- **风险：开发者诊断和用户提示混淆** -> 插件桥接协议默认英文开发者文案，管理端展示另走 `messageKey`。
- **风险：底层数据库/驱动错误不可翻译** -> 只翻译外层业务模板，把底层错误作为技术参数保留或在用户提示中隐藏。

## Migration Plan

1. 新增结构化错误和消息渲染基础组件，接入统一响应中间件，保持现有 `LocalizeError` 作为非结构化错误兜底。
2. 更新前端请求拦截器，支持 `messageKey/messageParams/errorCode`，并补齐页面级 `$t` 规则。
3. 按优先级清理宿主核心 API 错误：用户、字典、系统参数、用户消息、文件、任务、插件生命周期。
4. 清理导入导出：用户、字典、系统参数、岗位、操作日志等表头、sheet、枚举值和失败原因。
5. 清理插件平台：pluginbridge、pluginfs、plugindb、catalog、runtime、wasm host service 错误码和英文开发者诊断。
6. 清理源码插件业务错误和插件前端文案，资源写入插件自己的 `manifest/i18n/<locale>/*.json`。
7. 重组宿主和插件 i18n 资源目录，把运行时资源迁移到 `manifest/i18n/<locale>/*.json`，把 API 文档资源迁移到 `manifest/i18n/<locale>/apidoc/**/*.json`。
8. 建立扫描脚本和测试门禁，新增缺失翻译、错误本地化、Excel 语言和前端页面 i18n 测试。
9. 运行完整后端测试、前端构建和相关 E2E；确认 `zh-CN`、`en-US`、`zh-TW` 下关键错误和导出内容一致受控。

回滚策略：本迭代主要是代码与资源治理，不涉及数据库迁移。若某批清理引入问题，可按模块回退对应错误构造和语言资源；统一响应新增字段可保留，不影响只读取 `message` 的调用方。

## Open Questions

- 插件桥接 protobuf 结构是否直接扩展 `errorCode/messageKey/messageParams` 字段，还是先在 payload 中承载 JSON 错误对象，需要在实施时结合当前 ABI 兼容策略确认。
- 命令行输出的 locale 来源使用环境变量、配置默认语言还是显式参数，需要在处理 `cmd_*` 文件时确认。
- 对底层数据库错误在用户提示中展示多少细节，需要按安全性和可诊断性在具体模块中取舍。
