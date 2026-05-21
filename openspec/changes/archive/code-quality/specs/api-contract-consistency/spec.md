## MODIFIED Requirements

### Requirement: API 文档标签完整
后端 API 输入输出结构 SHALL 为所有可见字段补齐清晰的 `dc` 和 `eg` 标签，确保自动生成的 OpenAPI 文档可直接理解和调试。

#### Scenario: 复用跨模块公共枚举契约
- **WHEN** 后端 API DTO 字段使用跨多个 API 模块共享的稳定枚举或值类型
- **THEN** 该 DTO 使用 `apps/lina-core/pkg` 中按契约边界划分的公共组件
- **AND** 公共组件 SHALL 以小型领域契约命名，禁止新增大一统 `pkg/enums`
- **AND** 类型抽象不得改变既有 JSON 字段名、字段取值、默认值或 API 文档示例语义

## ADDED Requirements

### Requirement: API 响应不得直接暴露数据库实体
后端 API 响应 DTO SHALL 使用独立响应结构定义调用端可见字段，不得直接嵌入或返回 DAO 生成的数据库实体类型。响应边界 SHALL 通过显式字段映射只暴露当前接口需要的字段。

#### Scenario: 定义读取响应 DTO
- **WHEN** API 定义列表、详情、选项、树、文件信息或当前用户资料等读取响应
- **THEN** 响应类型 SHALL 使用 API 包内的独立 DTO
- **AND** 响应类型 SHALL 不嵌入 `entity.*`

#### Scenario: 隐藏内部和敏感字段
- **WHEN** 数据库实体包含密码、凭据、软删除时间、存储路径、哈希或内部实现字段
- **THEN** API 响应 DTO SHALL 默认不暴露这些字段
