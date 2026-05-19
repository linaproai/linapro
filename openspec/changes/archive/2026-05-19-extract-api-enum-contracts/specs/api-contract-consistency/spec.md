## MODIFIED Requirements

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
