## ADDED Requirements

### Requirement: 运行时可见文案必须有明确分类

系统 SHALL 对源代码中的字符串按运行时使用面分类，并对用户可见文案、用户交付物、用户展示投影、开发者诊断、运维日志和用户数据采用不同治理策略。用户可见文案、用户交付物和用户展示投影 MUST 通过运行时 i18n 资源或后端本地化投影输出；运维日志 MUST 使用稳定英文和结构化字段；用户输入和外部系统原文 MUST 保持原值，不得被自动翻译。

#### Scenario: 用户可见错误不得直接返回硬编码中文

- **WHEN** 后端业务服务需要返回用户可见错误
- **THEN** 错误 MUST 携带稳定错误码、运行时翻译键、参数和英文源文案
- **AND** 统一响应 MUST 按请求语言输出本地化 `message`
- **AND** 业务服务不得直接通过 `gerror.New("字典类型不存在")` 这类中文自由文本构造错误

#### Scenario: 运维日志保持稳定英文

- **WHEN** 后端服务记录仅用于排障、指标或启动期诊断的日志
- **THEN** 日志模板 MUST 使用稳定英文
- **AND** 日志 MUST 通过结构化字段记录错误码、资源 ID、插件 ID、路径或底层错误
- **AND** 日志不得依赖当前请求语言生成本地化文案

#### Scenario: 用户数据保持原值

- **WHEN** 字符串来自用户输入、外部接口返回、数据库业务名称或文件内容
- **THEN** 系统 MUST 原样保存和返回该字符串
- **AND** 系统不得尝试把该字符串当作翻译键自动翻译

### Requirement: 统一响应必须输出结构化错误字段

系统 SHALL 扩展统一 JSON 响应错误载荷，使运行时消息错误在保留本地化 `message` 的同时输出机器可读 `errorCode`、`messageKey` 和 `messageParams`。前端、插件和测试 MUST 以 `errorCode` 或 `messageKey` 判断错误语义，不得依赖自然语言 `message` 做业务判断。

#### Scenario: 结构化业务错误响应包含本地化消息和稳定字段

- **WHEN** 请求以 `Accept-Language: zh-CN` 触发 `USER_NOT_FOUND` 错误
- **THEN** 响应 JSON MUST 包含 `message`，其值为简体中文本地化结果
- **AND** 响应 JSON MUST 包含 `errorCode: "USER_NOT_FOUND"`
- **AND** 响应 JSON MUST 包含对应的 `messageKey`
- **AND** 响应 JSON MUST 包含用于格式化消息的 `messageParams`

#### Scenario: 同一错误在英文请求下返回英文展示文案

- **WHEN** 请求以 `Accept-Language: en-US` 触发同一个 `USER_NOT_FOUND` 错误
- **THEN** 响应 JSON 中的 `errorCode` 和 `messageKey` MUST 与 `zh-CN` 请求一致
- **AND** 响应 JSON 中的 `message` MUST 为英文展示文案
- **AND** 响应不得包含中文硬编码回退文案

#### Scenario: 非结构化错误仍通过现有本地化兜底处理

- **WHEN** 旧代码或第三方库返回非结构化错误
- **THEN** 统一响应 MUST 继续通过现有 `LocalizeError` 逻辑尝试翻译
- **AND** 若无法翻译，系统 MUST 返回受控 fallback 文案
- **AND** 新增业务错误不得继续使用非结构化自由文本作为主要实现方式

### Requirement: 后端业务错误必须使用运行时消息资源

后端宿主和源码插件的业务错误 SHALL 使用宿主或插件自己的运行时语言包维护翻译文本。宿主错误键 MUST 写入 `apps/lina-core/manifest/i18n/<locale>/*.json`；插件错误键 MUST 写入对应 `apps/lina-plugins/<plugin-id>/manifest/i18n/<locale>/*.json`。运行时错误文案 MUST NOT 写入或复用 `manifest/i18n/<locale>/apidoc` 资源。

#### Scenario: 宿主业务错误资源完整

- **WHEN** 宿主新增 `error.dict.type.exists` 错误键
- **THEN** `zh-CN`、`en-US` 和 `zh-TW` 宿主运行时语言包 MUST 都包含该键
- **AND** 缺失翻译检查 MUST 在任一目标语言缺失时失败

#### Scenario: 插件业务错误资源归属插件

- **WHEN** `org-center` 插件新增 `plugin.org-center.error.deptNotFound` 错误键
- **THEN** 该键 MUST 写入 `apps/lina-plugins/org-center/manifest/i18n/<locale>/*.json`
- **AND** 不得把插件运行时错误键集中写入 `lina-core` 的运行时语言包

### Requirement: 业务错误语义必须按模块命名空间治理

后端宿主、插件平台组件和源码插件 SHALL 按业务模块维护业务错误语义标识。业务错误定义 MUST 放在所属模块的 `*_code.go` 文件中；业务实现 MUST 只引用该文件中定义的 `bizerr.Code` 变量，不得在业务调用点硬编码机器错误码、翻译键或跨模块借用错误定义。所有可能返回给 HTTP API、插件调用、源码插件后端接口、WASM host service 或其他调用端响应载荷的接口错误 MUST 通过 `bizerr.NewCode`、`bizerr.WrapCode` 或等价封装创建/包装。HTTP 响应中的 `code` MUST 使用 GoFrame `gcode.Code` 类型错误码表达错误类别，具体业务语义 MUST 由 `errorCode`、`messageKey` 和 `messageParams` 表达。

#### Scenario: 宿主业务模块使用独立业务语义命名空间

- **WHEN** 用户模块新增 `USER_EMAIL_EXISTS` 错误
- **THEN** 该错误 MUST 定义在 `internal/service/user/user_code.go`
- **AND** `errorCode` MUST 使用用户模块前缀，例如 `USER_EMAIL_EXISTS`
- **AND** 响应 `code` MUST 使用最接近的 GoFrame 类型错误码，例如 `gcode.CodeInvalidParameter`
- **AND** 角色、菜单、字典等其他模块不得复用该业务语义标识

#### Scenario: 错误码定义与业务使用解耦

- **WHEN** 字典模块业务代码需要返回“字典类型已存在”错误
- **THEN** 业务代码 MUST 使用 `bizerr.NewCode(CodeDictTypeExists)` 或等价封装
- **AND** 业务代码不得写入 `"DICT_TYPE_EXISTS"`、`"error.dict.type.exists"` 或裸数字错误码
- **AND** 响应 `code` MUST 来自该定义绑定的 GoFrame 类型错误码

#### Scenario: 调用端可见错误必须由 bizerr 承载结构化元数据

- **WHEN** 控制器、中间件、业务服务或插件 host service 需要返回参数错误、鉴权错误、业务校验失败或用户可见失败原因
- **THEN** 该错误 MUST 使用所属模块 `*_code.go` 中定义的 `bizerr.Code`
- **AND** 该错误 MUST 通过 `bizerr.NewCode`、`bizerr.WrapCode` 或等价封装返回
- **AND** 业务路径不得直接返回 `gerror.New("请选择要导入的文件")`、`gerror.NewCode(gcode.CodeInvalidParameter, "error.xxx")`、`errors.New(...)` 或 `fmt.Errorf(...)` 作为调用端可见接口错误
- **AND** 低层技术错误只有在返回边界前被 `bizerr.WrapCode` 包装为业务语义错误时才允许作为 cause 传递

#### Scenario: 源码插件内部模块也按命名空间划分

- **WHEN** `org-center` 插件新增部门模块和岗位模块错误
- **THEN** 部门模块错误 MUST 使用 `ORG_CENTER_DEPT_*` 或等价模块前缀
- **AND** 岗位模块错误 MUST 使用 `ORG_CENTER_POST_*` 或等价模块前缀
- **AND** 插件错误定义不得写入宿主 `lina-core` 的错误定义文件

### Requirement: 导入导出交付物必须按请求语言渲染

系统 SHALL 让 Excel 导出、导入模板、导入失败原因、sheet 名、表头、状态、性别、操作类型和操作结果等用户交付物文案按请求语言渲染。导入导出流程 MUST 在请求级别解析 locale，并复用本模块所需翻译结果；不得在批量行循环中重复构建运行时语言包。

#### Scenario: 用户导出在英文请求下输出英文表头

- **WHEN** 用户以 `Accept-Language: en-US` 请求用户列表导出
- **THEN** 导出的 Excel 表头 MUST 使用英文文案
- **AND** 性别、状态等枚举展示 MUST 使用英文文案
- **AND** 导出文件不得包含 `用户名`、`正常`、`停用` 这类中文硬编码表头或状态

#### Scenario: 字典导入失败原因按请求语言返回

- **WHEN** 用户以 `Accept-Language: zh-TW` 导入包含无效字典类型的 Excel
- **THEN** 导入结果中的失败原因 MUST 使用繁体中文文案
- **AND** 失败原因 MUST 保留行号、字段名和非法值等参数
- **AND** 底层技术错误不得直接拼接到用户展示文案中造成中英混排

#### Scenario: 导出循环不得重复构建语言包

- **WHEN** 系统导出 10000 行数据
- **THEN** 翻译资源 MUST 在进入导出流程时按模块或 key 集合解析并缓存到请求级上下文
- **AND** 行循环内 MUST 只做已解析翻译结果查找、参数格式化或用户数据写入

### Requirement: 插件桥接和宿主服务错误必须稳定且可本地化

插件桥接协议、WASM host call、host service 调用、插件清单校验、插件资源校验和动态插件运行时错误 SHALL 返回稳定状态码或错误码，并为进入管理端展示的错误提供运行时翻译键。协议层默认开发者诊断文案 MUST 使用英文，管理端展示 MUST 通过请求语言本地化。

#### Scenario: Host call 协议错误包含稳定错误码

- **WHEN** 动态插件调用 host service 时提交非法 network URL
- **THEN** host call 响应 MUST 包含稳定状态码或错误码
- **AND** 开发者诊断文案 MUST 使用英文源文案
- **AND** 管理端展示该错误时 MUST 通过 `messageKey` 或错误码映射为当前语言文案

#### Scenario: 插件清单校验错误不再中英混排

- **WHEN** 插件清单校验发现菜单 key 不合法
- **THEN** 校验错误 MUST 使用稳定错误码和参数描述插件 ID、字段名、实际值和期望规则
- **AND** 用户可见展示 MUST 通过运行时翻译资源生成
- **AND** 不得返回 `插件菜单 key 必须使用当前插件前缀 plugin:<id>:*` 这类中英混排自由文本

### Requirement: 插件生命周期和升级结果必须使用消息键

插件安装、卸载、启用、停用、自动启用、源码插件升级和动态插件发布治理结果 SHALL 返回或存储稳定 `messageKey`、`messageParams` 和 `errorCode`。若接口仍需要简单展示字段，`message` MUST 按请求语言或命令 locale 从结构化字段渲染得到。

#### Scenario: 源码插件无需升级时返回结构化结果

- **WHEN** 源码插件当前生效版本等于发现版本
- **THEN** 升级结果 MUST 包含表示无需升级的稳定 `messageKey`
- **AND** 结果 MUST 包含插件 ID、当前版本和发现版本等参数
- **AND** 不得只返回 `当前源码插件已是最新版本，无需升级。` 这类固定中文字符串

#### Scenario: 插件生命周期失败可按语言展示

- **WHEN** 插件安装失败并被管理端展示
- **THEN** 后端 MUST 返回稳定错误码和翻译键
- **AND** 前端 MUST 使用当前语言展示本地化失败原因
- **AND** 日志 MUST 保留英文诊断和结构化参数便于排障

### Requirement: 前端用户可见文案必须通过 i18n 渲染

默认管理工作台和插件前端 SHALL 对页面标题、表单标签、表格列、空状态、提示、确认框、toast、tooltip 和时间单位等用户可见文案使用 `$t` 或运行时语言包。前端请求错误拦截器 MUST 优先消费后端返回的 `messageKey/messageParams`，其次才使用后端已本地化 `message`。

#### Scenario: 监控页面切换语言后所有静态标签变化

- **WHEN** 用户从 `zh-CN` 切换到 `en-US`
- **THEN** 服务器监控页面中的数据库信息、服务器信息、服务信息、磁盘列名、空状态和时间单位 MUST 切换为英文
- **AND** 页面不得继续显示硬编码中文标签

#### Scenario: 在线用户页面列名使用翻译键

- **WHEN** 用户打开在线用户页面
- **THEN** 查询表单标签和表格列标题 MUST 通过 `$t` 或运行时语言包渲染
- **AND** `data.ts` 中不得保留 `用户账号`、`登录账号`、`部门名称` 等中文硬编码标签

#### Scenario: 请求错误优先使用 messageKey

- **WHEN** 后端错误响应同时包含 `messageKey`、`messageParams` 和 `message`
- **THEN** 前端请求拦截器 MUST 优先使用 `$t(messageKey, messageParams)` 展示错误
- **AND** 当前端缺少该翻译键时 MUST fallback 到后端 `message`

### Requirement: 审计和操作展示必须存储稳定语义并按语言投影

操作日志、登录日志、任务日志、通知消息、插件升级结果和其他面向用户展示或导出的运行记录 SHALL 优先存储稳定类型码、状态码、翻译键和参数。列表、详情、导出和消息预览 MUST 按请求语言投影展示文案。系统不得只持久化已经本地化的中文展示值作为唯一语义来源。

#### Scenario: 操作日志导出按请求语言展示操作类型

- **WHEN** 用户以 `en-US` 导出操作日志
- **THEN** 操作类型和操作状态 MUST 根据稳定代码渲染为英文
- **AND** 导出结果不得依赖持久化的 `导出`、`成功`、`失败` 中文字符串

#### Scenario: 任务日志详情保留机器语义和本地化展示

- **WHEN** 用户查看任务日志详情
- **THEN** 后端或前端 MUST 使用任务状态码、处理器 key 和错误码投影展示文案
- **AND** 原始 stdout/stderr 或用户脚本输出 MUST 保持原值

### Requirement: 自动化治理必须阻断新增高风险硬编码文案

系统 SHALL 提供自动化扫描或测试门禁，识别 Go、Vue 和 TypeScript 中运行时可见位置的硬编码中文或中英混排文案。扫描 MUST 支持 allowlist，但每个例外 MUST 标明分类、原因和责任模块。新增或修改运行时翻译键时，缺失翻译检查 MUST 覆盖所有启用内置语言。

#### Scenario: Go 高风险字符串扫描失败

- **WHEN** 开发者在生产 Go 代码中新增 `gerror.New("部门不存在")`
- **THEN** 硬编码文案扫描 MUST 报告该位置
- **AND** 检查 MUST 要求改为结构化错误和运行时翻译键

#### Scenario: 前端高风险字符串扫描失败

- **WHEN** 开发者在 Vue 页面表格列中新增 `title: '操作'`
- **THEN** 前端扫描 MUST 报告该位置
- **AND** 检查 MUST 要求改为 `$t` 或运行时语言包键

#### Scenario: 允许的非运行时中文不阻断

- **WHEN** 中文字符串只出现在注释、测试 fixture、用户示例数据或明确标注的 allowlist 项中
- **THEN** 扫描规则允许放行该字符串
- **AND** allowlist 项 MUST 记录该字符串不属于运行时可见框架文案的原因

### Requirement: 本地化查找必须满足热路径性能约束

运行时错误本地化、列表投影、导入导出和前端语言切换 SHALL 复用现有运行时翻译缓存。系统 MUST NOT 在单个错误、单行导出、单个表格列渲染或单条日志投影时克隆或重建完整语言包。

#### Scenario: 单个错误本地化只执行缓存查找

- **WHEN** 统一响应中间件渲染一个结构化业务错误
- **THEN** 系统 MUST 通过当前 locale 的运行时缓存查找对应 key
- **AND** 不得为了该错误构建完整运行时消息包

#### Scenario: 批量列表投影复用同一请求语言上下文

- **WHEN** 后端返回包含 1000 条记录的操作日志列表
- **THEN** 系统 MUST 在同一请求中复用 locale 和翻译查找能力
- **AND** 不得为每条记录重复解析请求语言或加载语言资源
