## Context

本次审查使用 Go 源码 CJK 字符扫描与风险面人工分类，发现当前后端仍存在三类重要残留：

- 手写、非测试、非生成 Go 文件中约 71 个文件、780 处中文字符串字面量。
- DAO/DO/Entity 生成文件中仍有大量中文注释和 `description` tag，其中 Entity schema 约 37 个文件、406 处中文 `description`。
- API DTO 的 `summary`、`dc`、`description`、`eg` 未发现中文源文本，当前 API DTO 文档源文本整体合规。

当前项目已经具备运行时 i18n 资源、`bizerr` 结构化错误、统一响应中间件、插件资源加载和 `zh-CN/en-US/zh-TW` 多语言资源目录。本变更不重新设计基础设施，而是把审查发现的后端残留中文硬编码按风险面落地清理。

需要特别处理与活跃变更 `runtime-message-i18n-governance` 的边界：该变更负责运行时消息治理基础设施和部分模块落地；本变更专门跟踪“当前后端残留中文硬编码清理”，作为更具体的收口迭代，任务必须以扫描清单和清理结果为验收依据。

## Goals / Non-Goals

**Goals:**

- 建立本次后端中文硬编码审查清单，并给每类字符串明确处理策略。
- 调用端可见错误统一改为 `bizerr` 或等价结构化错误，响应中保留稳定 `errorCode`、`messageKey`、`messageParams` 和按请求语言本地化的 `message`。
- 导出文件、树节点、状态 fallback、演示提示、系统信息展示等用户可见文案改为运行时 i18n 或结构化字段。
- 插件桥接、插件文件系统、插件数据库、WASM host service、catalog/runtime 校验等开发者诊断统一使用英文稳定文案，并在进入用户界面或 API 响应边界时结构化包装。
- 明确生成文件中文注释和 schema `description` 的治理路径：不手动修改生成物，通过 SQL 注释或生成源改为英文。
- 扩展硬编码扫描和 allowlist，确保后续新增高风险中文硬编码会被阻断或明确豁免。

**Non-Goals:**

- 不引入新的 i18n 存储表、线上翻译后台、机器翻译或第三方翻译服务。
- 不修改 API DTO 文档源文本规则；API DTO 仍必须直接使用英文源文本，apidoc 翻译继续使用专用 `manifest/i18n/<locale>/apidoc/**/*.json`。
- 不翻译用户输入、数据库中用户维护的业务名称、外部系统原始错误、SQL 片段、路径、权限标识、插件 ID、协议字段名等技术值。
- 不手工编辑 `dao/do/entity` 生成文件。
- 不把插件运行时文案集中写入 `lina-core` 语言包。

## Decisions

### 决策一：先分类再改造，不做 CJK 字符一刀切

**选择**：每个命中的中文字符串必须归入以下分类之一，并按分类处理：

| 分类 | 典型位置 | 处理策略 |
| ---- | -------- | -------- |
| 调用端可见错误 | `gerror.New*`、`gerror.Wrap*`、手写 JSON error response | 改为模块 `*_code.go` 中的 `bizerr` 定义；补齐运行时 `error.json` |
| 用户可见投影 | 树节点 label、状态 label、运行时配置 disabled reason、系统信息字段 | 使用运行时 i18n key 或返回结构化值交给前端格式化 |
| 用户交付物 | Excel 表头、sheet 名、导入失败原因、导出枚举值 | 在业务服务层按请求 locale 本地化；循环内复用预解析结果 |
| 开发者诊断 | 插件桥接、WASM、manifest 校验、codec 解析、启动期 panic | 默认改为英文稳定文案；若穿透到用户响应边界则包装为结构化错误 |
| 生成源 | DAO/DO/Entity 注释、Entity `description` | 不手改生成物；通过 SQL 注释或生成源治理 |
| 测试/fixture | `_test.go`、测试 JSON、断言用样例 | 可保留，但扫描 allowlist 必须说明用途 |

**原因**：同样含中文的字符串可能是用户提示、开发诊断、生成元数据或测试样例。先分类可以避免误翻译用户数据，也避免把运维日志和插件协议错误本地化成不稳定文案。

**备选方案**：直接把所有中文字符串替换成 i18n key。未采用，因为会误伤测试 fixture、生成文件、协议字段和开发诊断。

### 决策二：调用端可见错误优先治理为 `bizerr`

**选择**：凡是可能进入 HTTP API、源码插件后端 API、动态插件路由、WASM host service、插件宿主服务或统一响应载荷的业务错误，必须使用 `bizerr.NewCode`、`bizerr.WrapCode` 或模块本地等价封装。每个模块在自己的 `*_code.go` 中集中定义错误码和英文 fallback，调用点不得直接硬编码中文错误文本、机器错误码或翻译键。

优先清理的已知文件包括：

- `apps/lina-core/internal/controller/plugin/plugin_v1_resource_list.go`
- `apps/lina-plugins/content-notice/backend/internal/service/notice/notice.go`
- `apps/lina-plugins/org-center/backend/internal/service/dept/dept.go`
- `apps/lina-plugins/org-center/backend/internal/service/post/post.go`
- `apps/lina-plugins/monitor-loginlog/backend/internal/service/loginlog/loginlog.go`
- `apps/lina-plugins/monitor-operlog/backend/internal/service/operlog/operlog.go`
- `apps/lina-plugins/plugin-demo-source/backend/internal/service/demo/*.go`
- `apps/lina-plugins/plugin-demo-dynamic/backend/internal/service/dynamic/dynamic_demo_record.go`

**原因**：统一响应中间件已经支持结构化错误本地化。继续返回中文自由文本会绕过 `errorCode/messageKey/messageParams`，导致前端、插件调用方和自动化测试无法稳定识别错误语义。

**备选方案**：继续让 `LocalizeError` 使用错误文本当翻译 key 兜底。未采用，因为中文文本不能作为稳定业务契约，也无法携带命名参数。

### 决策三：用户可见投影和交付物在业务边界本地化

**选择**：树节点、导出表头、状态 fallback、演示提示、源码插件摘要、运行时配置原因、系统运行时长等用户可见内容，由拥有该业务语义的模块负责本地化。公共工具层如 `excelutil` 继续只做文件操作，不感知 i18n。

优先清理的已知位置包括：

- `org-center/post.go`：`未分配部门`、岗位导出表头、`正常/停用`。
- `monitor-loginlog/loginlog.go`：`成功/失败` 登录状态 fallback、导出表头 fallback。
- `monitor-operlog/operlog.go`：操作类型和操作结果 fallback、导出表头 fallback。
- `demo-control/middleware_guard.go`：演示模式拦截消息。
- `plugin-demo-source/demo_summary.go`：示例页摘要文案。
- `config/config_cron.go`：shell 模式不支持原因。
- `sysinfo/sysinfo.go`：运行时长单位。

**原因**：这些文案直接面向管理端用户或导出文件使用者。它们必须随请求语言变化，并且不能和字典、运行时配置、插件资源形成双重事实源。

**备选方案**：在前端对后端返回的中文 label 二次翻译。未采用，因为导出文件、源码插件 API 和后端投影也需要一致语言，且前端无法可靠判断哪些后端字段是翻译 key。

### 决策四：插件平台内部诊断改为英文，用户边界再结构化

**选择**：`pkg/pluginbridge`、`pkg/pluginfs`、`pkg/plugindb`、`internal/service/plugin/internal/**` 中的协议解析、路径校验、manifest 校验、WASM 初始化、host service codec 错误默认使用英文开发者诊断。若这些错误会通过插件管理 API、动态插件路由或宿主服务响应暴露给管理端用户，则在边界包装为 `bizerr` 或结构化插件错误。

**原因**：插件平台错误通常服务插件开发者、日志检索和运行时调试，语言必须稳定；但管理端展示仍要本地化。把协议层诊断和展示层错误分离，可以避免用自然语言做协议判断。

**备选方案**：插件协议层直接返回按请求语言本地化的 message。未采用，因为动态插件 guest 不应依赖宿主某个用户的语言环境来判断错误。

### 决策五：生成文件中文通过生成源治理

**选择**：不手动修改 `internal/dao`、`internal/model/do`、`internal/model/entity` 下的生成文件。若生成的 Entity `description` 会进入 OpenAPI schema 并显示中文，应修改对应 SQL 表/字段注释或 codegen 输入源，重新运行 `make dao` 生成英文元数据。

**原因**：手工修改生成物会在下一次 `gf gen dao` 后丢失，也违反项目生成代码规范。OpenAPI 规则也明确禁止在 apidoc service 层为生成 schema 做临时中英转换表。

**备选方案**：在 apidoc 渲染层对生成 schema 做翻译映射。未采用，因为会把数据源问题隐藏到文档层，且需要维护不稳定的中文到英文转换表。

### 决策六：扫描工具以高风险语义位置为阻断目标

**选择**：扩展 `hack/tools/runtime-i18n` 或等价工具，至少覆盖 Go 高风险模式：

- `gerror.New*`、`gerror.Wrap*`、`fmt.Errorf`、`errors.New` 中的中文字符串。
- `Message`、`Reason`、`Fallback`、`Label`、`Title`、`DisabledReason` 等用户可见字段赋值。
- 导出表头数组、状态/类型 label map、树节点 label 构造。
- 插件 bridge/host service/manifest 校验错误构造。

扫描允许配置 allowlist，但每条豁免必须说明分类、文件、原因和过期条件。生成文件和 `_test.go` 默认可单独统计，不作为本次实现阻断对象。

**原因**：纯 `rg "\p{Han}"` 噪音过大；高风险位置扫描更适合作为长期门禁。

**备选方案**：只在 code review 中人工检查。未采用，因为当前残留分布较广，单靠审查容易回归。

## Risks / Trade-offs

- **风险：与活跃 `runtime-message-i18n-governance` 范围重叠** → 本变更只跟踪本次审查发现的后端残留清理；基础设施变更复用现有能力，不重复设计。
- **风险：一次性清理 71 个文件影响面过大** → 按“调用端错误、用户交付物、插件平台诊断、生成源、扫描门禁”分批提交和测试。
- **风险：插件运行时错误从中文改为英文后测试断言失败** → 测试改为断言 `errorCode/messageKey` 或英文开发者诊断，不再断言中文自由文本。
- **风险：导出本地化影响性能** → 导出前预解析表头和枚举 label，行循环内只使用 map 查找，不重复构建翻译包。
- **风险：扫描 allowlist 变成永久豁免** → allowlist 需要分类、原因和可验证边界；任务中必须输出剩余 allowlist 清单。

## Migration Plan

1. 固化审查清单和分类，生成当前基线统计，记录必须清理和可保留范围。
2. 先清理调用端可见业务错误，补齐 `bizerr` 定义和语言资源，更新对应单元测试。
3. 清理用户可见投影和导出文案，补齐运行时 i18n key，验证 `zh-CN/en-US/zh-TW` 展示。
4. 清理插件平台内部中文诊断，统一改为英文；对用户边界补结构化包装。
5. 处理生成文件来源，必要时修改 SQL 注释并重新生成 DAO/DO/Entity。
6. 扩展扫描工具和 allowlist，接入本地验证入口。
7. 运行 Go 单元测试、相关 E2E、硬编码扫描和 `lina-review`。

## Open Questions

- 是否把所有插件平台内部诊断都立即接入稳定机器码，还是本次先统一英文文案、仅用户边界使用 `bizerr`？建议本次采用后者，降低协议面改动风险。
- Entity schema 中文 `description` 是否全部需要在本次迭代改 SQL 注释，还是先只处理会进入 OpenAPI 的核心宿主表？建议先以会进入接口文档或用户可见文档的 schema 为优先级。
