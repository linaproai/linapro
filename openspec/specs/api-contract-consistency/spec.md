# api-contract-consistency Specification

## Purpose
统一后端 API 的 REST 语义、路径参数绑定方式与文档标签约束，确保接口合同在全仓库保持一致。
## Requirements
### Requirement: REST 语义与路径风格统一
后端 API SHALL 在路径、HTTP 方法和资源语义上保持一致，读操作、写操作和删除操作分别使用仓库约定的 REST 方式表达。

#### Scenario: 定义读取接口
- **WHEN** 定义列表查询、详情查询、选项查询、树查询或导出接口
- **THEN** 接口使用 `GET`
- **AND** 路径采用资源化命名而不是动作化命名

#### Scenario: 定义写操作接口
- **WHEN** 定义创建、更新、状态切换、导入或删除接口
- **THEN** 接口分别使用 `POST`、`PUT`、`DELETE`
- **AND** 同类资源的路径风格在全仓库保持一致

### Requirement: 路径参数绑定统一
后端 API DTO SHALL 统一路径参数声明和绑定方式，避免同一接口同时混用多种参数风格。

#### Scenario: 声明路径参数
- **WHEN** 接口路径包含资源标识或子资源标识
- **THEN** `g.Meta` 中使用 `{param}` 形式声明路径参数
- **AND** 输入 DTO 字段统一使用 `json:"param"` 声明参数名，不在同一仓库混用 `p` 与 `json` 标签
- **AND** 输出 DTO 继续使用 `json` 标签定义响应字段

### Requirement: API 文档标签完整
后端 API 输入输出结构 SHALL 为所有可见字段补齐清晰的 `dc` 和 `eg` 标签，确保自动生成的 OpenAPI 文档可直接理解和调试。

#### Scenario: 定义输入输出字段
- **WHEN** 在 API DTO 中新增或整改输入输出字段
- **THEN** 每个字段包含 `dc` 标签说明业务含义
- **AND** 每个字段包含可直接用于调试的 `eg` 示例值

#### Scenario: 定义枚举或可选字段
- **WHEN** 字段表示状态、类型、开关或可选筛选条件
- **THEN** `dc` 中明确列出取值含义或默认行为
- **AND** 文档不要求调用方依赖阅读实现代码推断参数语义
