# 接口契约一致性规范

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

#### Scenario: 复用跨模块公共枚举契约
- **WHEN** 后端 API DTO 字段使用跨多个 API 模块共享的稳定枚举或值类型
- **THEN** 该 DTO 使用 `apps/lina-core/pkg` 中按契约边界划分的公共组件或已有稳定公共组件
- **AND** 公共组件 SHALL 以小型领域契约命名，禁止新增大一统 `pkg/enums`
- **AND** 领域私有枚举 SHALL 保留在所属 API 或领域包内，不因字面值重复而强行合并
- **AND** 类型抽象不得改变既有 JSON 字段名、字段取值、默认值或 API 文档示例语义

### Requirement: API 响应不得直接暴露数据库实体

后端 API 响应 DTO SHALL 使用独立响应结构定义调用端可见字段，不得直接嵌入或返回 DAO 生成的数据库实体类型。响应边界 SHALL 通过显式字段映射只暴露当前接口需要的字段，避免数据库字段新增、字段重命名或内部治理字段变化被动改变外部 API 合同。

#### Scenario: 定义读取响应 DTO
- **WHEN** API 定义列表、详情、选项、树、文件信息或当前用户资料等读取响应
- **THEN** 响应类型 SHALL 使用 API 包内的独立 DTO
- **AND** 响应类型 SHALL 不嵌入 `entity.*`
- **AND** 响应列表 SHALL 不直接使用 `[]*entity.*`

#### Scenario: 映射实体到响应 DTO
- **WHEN** 控制器或响应组装逻辑从服务层实体、投影或领域对象生成 API 响应
- **THEN** 响应组装 SHALL 逐字段映射允许暴露的字段
- **AND** 不得直接把数据库实体指针赋值到响应字段

#### Scenario: 隐藏内部和敏感字段
- **WHEN** 数据库实体包含密码、凭据、软删除时间、存储路径、哈希、内部租户治理字段、缓存版本或其他内部实现字段
- **THEN** API 响应 DTO SHALL 默认不暴露这些字段
- **AND** 只有明确属于当前接口合同且经过权限、数据权限和业务边界评估的字段才可加入响应 DTO

### Requirement: Response instant fields use Unix timestamps in milliseconds
Public HTTP JSON response DTO fields that represent exact instants SHALL return Unix timestamps in milliseconds. The JSON type MUST be numeric, and the Go DTO type MUST be `int64` or `*int64`. Public response DTOs MUST NOT directly expose `time.Time`, `*time.Time`, `gtime.Time`, `*gtime.Time`, or formatted strings for instant fields. This rule only constrains the API response DTO boundary and does not constrain internal `dao`, `entity`, `do`, or `service` models.

#### Scenario: Defining an instant response field
- **WHEN** a developer adds or modifies `createdAt`, `updatedAt`, `deletedAt`, `loginAt`, `startedAt`, `endedAt`, `expiredAt`, or another public HTTP JSON response DTO field that represents an exact instant
- **THEN** the field type is `int64` or `*int64`
- **AND** the field value represents a Unix timestamp in milliseconds
- **AND** the field `dc` tag or API documentation includes `Unix timestamp in milliseconds`

#### Scenario: Defining a calendar-date response field
- **WHEN** a developer defines `birthday`, `businessDate`, `periodDate`, or another public HTTP JSON response DTO field that represents only a calendar date and not an exact instant
- **THEN** the field may use a `YYYY-MM-DD` string
- **AND** the field `dc` tag or API documentation clearly states `date-only` semantics
- **AND** callers do not need to infer an exact instant in the server time zone from the field

#### Scenario: Migrating an existing time-field contract
- **WHEN** a developer plans to migrate an existing public API instant field from a string or time-object response to a Unix timestamp in milliseconds
- **THEN** the migration is evaluated through the active OpenSpec change that owns the API contract change
- **AND** frontend parsing, page formatting, `apidoc i18n` resources, and related test impact are evaluated together

