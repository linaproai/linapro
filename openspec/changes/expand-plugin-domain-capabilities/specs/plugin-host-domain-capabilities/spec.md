## ADDED Requirements

### Requirement: 插件领域能力扩展必须先冻结阶段矩阵

系统 SHALL 在新增或动态发布插件领域能力方法前冻结方法发布、错误语义、规模上限和授权资源四类矩阵。未完成冻结的方法 MUST 只保留为路线图，不得进入当批普通能力实现或动态`host service`实现。

#### Scenario: 方法缺少发布矩阵

- **WHEN** 变更计划新增插件可消费领域方法
- **THEN** OpenSpec 设计必须声明该方法是源码插件专属、同步动态发布、延后发布还是不发布
- **AND** 未声明发布决策的方法不得修改`capability.Services`或动态`hostServices`目录

#### Scenario: 动态方法缺少授权资源矩阵

- **WHEN** 变更计划把普通领域方法发布为动态`host service`
- **THEN** OpenSpec 设计必须声明对应`service`、`method`、`resource`和`plugin.yaml hostServices`授权形态
- **AND** 未声明授权资源的方法不得进入动态协议 catalog、guest client 或 dispatcher

#### Scenario: 高频方法缺少规模上限

- **WHEN** 变更计划新增批量、搜索、候选、树形、聚合或资源枚举类领域方法
- **THEN** OpenSpec 设计必须声明输入数量、分页、key/path 长度或总字节数上限
- **AND** 超限行为必须映射为结构化能力错误

### Requirement: 阶段一必须只发布冻结的高频只读领域能力

系统 SHALL 将`expand-plugin-domain-capabilities`第一批实现范围限定为当前用户投影、用户批量解析、批量权限判断、字典值可见性校验和当前在线会话投影。候选搜索、组织/租户投影、插件治理搜索、私有资源批量、通知类型化和`AI`方法状态 MUST 保留为后续阶段，除非另行更新 OpenSpec 并完成阶段矩阵冻结。

#### Scenario: 阶段一实现当前用户投影

- **WHEN** 源码插件或已授权动态插件调用用户当前投影方法
- **THEN** 系统返回当前 actor 可见的用户最小投影
- **AND** 缺少用户 actor 或系统上下文调用必须 fail-closed

#### Scenario: 阶段一实现用户批量解析

- **WHEN** 插件按用户 ID、用户名、手机号或邮箱批量解析用户
- **THEN** 系统在数据库查询阶段应用租户和数据权限过滤
- **AND** 不存在、不可见、租户外或未授权目标统一进入`MissingIDs`
- **AND** 实现不得对每个解析键执行一次用户详情查询

#### Scenario: 阶段一实现批量权限判断

- **WHEN** 插件一次判断多个权限 key
- **THEN** 系统返回每个权限 key 的布尔结果
- **AND** 实现必须复用权限快照、集合化权限服务或等价批量路径
- **AND** 不得循环调用单权限判断作为常规实现

#### Scenario: 阶段一发布字典值可见性校验

- **WHEN** 插件在写入或执行动作前校验一组字典值
- **THEN** 系统按字典类型和值集合执行可见性校验
- **AND** 任一值不存在或不可见时整体拒绝
- **AND** 错误不得区分不存在和不可见

#### Scenario: 阶段一实现当前在线会话投影

- **WHEN** 插件请求当前在线会话投影
- **THEN** 系统只返回当前 token 对应的会话投影
- **AND** 缺少请求型 token/session 上下文时必须 fail-closed
- **AND** 实现不得扫描全部在线会话来推断当前用户最新会话

### Requirement: 阶段一领域能力必须记录影响和验证证据

系统 SHALL 在任务记录和审查结论中记录阶段一领域能力对`i18n`、缓存一致性、数据权限、数据库、开发工具、测试和 E2E 的影响判断，并提供匹配验证证据。

#### Scenario: 无 HTTP API 或 UI 变化

- **WHEN** 阶段一只修改 Go 领域契约、动态 host service 协议和 README
- **THEN** 任务记录必须说明无静态 HTTP API、前端 UI、插件清单、语言包和 E2E 影响
- **AND** 仍必须运行 OpenSpec strict 校验、相关 Go 包测试和 README 静态检查

#### Scenario: 缓存和数据权限影响

- **WHEN** 阶段一方法复用权限、字典、用户或 session owner 数据
- **THEN** 任务记录必须说明权威数据源、共享实例或快照复用方式
- **AND** 数据权限读取或校验路径必须通过测试、静态检索或审查证据证明未退化为内存过滤或`N+1`
