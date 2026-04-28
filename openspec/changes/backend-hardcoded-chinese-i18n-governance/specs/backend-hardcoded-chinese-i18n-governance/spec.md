## ADDED Requirements

### Requirement: 后端中文硬编码分类治理

系统 MUST 对后端 Go 源码中的中文字符串字面量进行分类治理，并区分调用端可见错误、用户可见投影、用户交付物、开发者诊断、生成源、测试 fixture 和用户数据示例。

#### Scenario: 扫描结果必须可分类
- **WHEN** 开发者运行后端中文硬编码扫描工具
- **THEN** 扫描结果 MUST 输出每个命中项所属分类或 allowlist 原因
- **AND** 未分类的手写非测试 Go 中文字符串 MUST 被标记为待处理问题

#### Scenario: 测试和生成文件不阻断业务清理
- **WHEN** 扫描命中 `_test.go`、`internal/dao`、`internal/model/do` 或 `internal/model/entity` 下的中文内容
- **THEN** 系统 MUST 单独统计这些命中项
- **AND** 系统 MUST NOT 要求开发者手工修改 DAO、DO 或 Entity 生成文件

### Requirement: 调用端可见错误必须结构化

所有可能进入 HTTP API、源码插件后端 API、动态插件路由、WASM host service、插件宿主服务或统一响应载荷的业务错误、鉴权错误、参数错误和用户可见失败原因 MUST 使用 `bizerr` 或等价结构化错误表达。

#### Scenario: 业务错误返回给 HTTP 调用方
- **WHEN** 后端业务服务返回调用端可见错误
- **THEN** 统一响应 MUST 包含稳定 `errorCode`
- **AND** 统一响应 MUST 包含 `messageKey`
- **AND** 统一响应 MUST 按请求语言返回本地化 `message`
- **AND** 业务调用点 MUST NOT 直接返回中文 `gerror.New*`、`errors.New` 或 `fmt.Errorf`

#### Scenario: 模块定义自己的错误码
- **WHEN** 一个模块新增可见业务错误
- **THEN** 该错误 MUST 在模块自己的 `*_code.go` 中集中定义
- **AND** 定义 MUST 包含英文 fallback、稳定机器错误码和运行时 i18n key
- **AND** 调用点 MUST 通过定义好的错误变量创建或包装错误

### Requirement: 插件错误和语言资源必须归属插件

源码插件后端新增或修改用户可见错误、导出文案、页面摘要、演示提示或业务投影文案时，相关运行时 i18n 资源 MUST 放在该插件自己的 `manifest/i18n/<locale>/*.json` 中。

#### Scenario: 插件业务错误本地化
- **WHEN** `org-center`、`content-notice`、`monitor-loginlog`、`monitor-operlog`、`plugin-demo-source`、`plugin-demo-dynamic` 或其他源码插件新增业务错误
- **THEN** 错误定义 MUST 使用插件命名空间下的稳定错误码和 message key
- **AND** `zh-CN`、`en-US`、`zh-TW` 翻译资源 MUST 在该插件目录下维护
- **AND** lina-core 的运行时语言包 MUST NOT 集中承载该插件的业务错误翻译

#### Scenario: 插件导出文案本地化
- **WHEN** 插件生成 Excel、CSV 或其他用户交付物
- **THEN** 表头、sheet 名和枚举展示值 MUST 使用插件运行时 i18n 资源
- **AND** 用户输入或数据库业务内容 MUST 原样输出

### Requirement: 用户可见投影和交付物必须按语言渲染

后端拥有的用户可见展示字段、导出表头、导入模板字段、导入失败原因、状态 fallback 和运行时配置展示原因 MUST 根据请求语言渲染，或返回结构化值由前端渲染。

#### Scenario: 导出文件随语言变化
- **WHEN** 用户在不同运行时语言下触发同一个导出接口
- **THEN** 导出文件中的系统表头和系统枚举展示值 MUST 按当前请求语言输出
- **AND** 导出文件中的用户数据 MUST 保持原值

#### Scenario: 后端投影字段随语言变化
- **WHEN** 用户请求部门树、岗位树、系统信息或运行时配置展示字段
- **THEN** 后端系统生成的 label、单位、原因说明 MUST 按当前请求语言输出
- **OR** 后端 MUST 返回结构化数值和 code，让前端按当前语言渲染

### Requirement: 插件平台开发者诊断必须稳定

插件桥接、插件文件系统、插件数据库、WASM host service、插件 catalog/runtime 校验等开发者诊断错误 MUST 使用稳定英文源文案；当这些错误进入用户界面或调用端响应边界时，MUST 被包装为结构化错误或结构化插件错误载荷。

#### Scenario: 插件协议解析失败
- **WHEN** 插件 bridge codec 或 host service codec 解析协议失败
- **THEN** 内部诊断错误 MUST 使用英文稳定文案
- **AND** 协议调用方 MUST NOT 依赖中文自然语言判断错误类型

#### Scenario: 插件管理接口暴露平台错误
- **WHEN** 插件平台内部错误通过插件管理 API 暴露给管理端用户
- **THEN** 响应 MUST 携带稳定错误码或 message key
- **AND** 用户展示文案 MUST 可按运行时语言本地化

### Requirement: 生成 schema 中文必须从生成源治理

系统 MUST NOT 手工修改 DAO、DO 或 Entity 生成文件中的中文注释或 `description` tag；当生成 schema 需要进入 OpenAPI 或用户可见文档时，MUST 修改 SQL 注释或 codegen 输入源并重新生成。

#### Scenario: Entity description 进入接口文档
- **WHEN** 生成的 Entity schema `description` 会进入 OpenAPI 文档
- **THEN** 对应 SQL 表/字段注释或生成源 MUST 提供英文源文本
- **AND** apidoc service MUST NOT 维护中文到英文的临时转换表

#### Scenario: 重新生成 DAO 工件
- **WHEN** SQL 注释或生成源被修改以治理 schema 文案
- **THEN** 开发者 MUST 运行项目约定的 DAO 生成流程
- **AND** 生成结果 MUST 保持可重复生成，不依赖手工修改

### Requirement: 防回归扫描必须覆盖高风险位置

项目 MUST 提供后端运行时中文硬编码扫描门禁，覆盖调用端错误、用户可见字段、导出表头、状态 label、插件诊断和结构化错误 fallback 等高风险位置。

#### Scenario: 新增中文 gerror 被阻断
- **WHEN** 开发者在手写非测试 Go 文件中新增中文 `gerror.New*`、`gerror.Wrap*`、`errors.New` 或 `fmt.Errorf`
- **THEN** 扫描工具 MUST 标记该新增项为违规
- **AND** 违规项 MUST 改为 `bizerr`、英文开发者诊断或带原因的 allowlist

#### Scenario: allowlist 必须说明边界
- **WHEN** 某个中文字符串被允许保留
- **THEN** allowlist MUST 记录文件、分类、保留原因和适用范围
- **AND** 用户可见错误和用户交付物文案 MUST NOT 仅靠 allowlist 豁免
