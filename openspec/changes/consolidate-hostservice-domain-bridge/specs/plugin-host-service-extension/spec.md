## ADDED Requirements

### Requirement: WASM host service dispatch 必须由显式 registry 驱动

系统 SHALL 将动态插件 WASM host service dispatch 收敛为显式注册的 registry 驱动结构。`wasm_host_service.go`入口 MUST 只负责 envelope 解码、调用上下文构造、授权校验、registry lookup 和统一错误响应；`internal/service/plugin/internal/wasm/hostservicedispatch`MUST 拥有 registry、handler context、注册校验和通用响应辅助。具体 service/method 处理逻辑 MAY 继续保留在`wasm`父包作为显式注册适配层，避免为了迁移目录扩大`hostCallContext`、运行时快照和插件执行状态的公开面；若后续领域 handler 迁移到子包，MUST 先抽取窄上下文契约并保持 DI 来源清晰。registry 注册 MUST 使用显式装配函数，不得使用`init()`隐式注册。

#### Scenario: 已注册 method 正常分发

- **WHEN** 动态插件调用一个已在 registry 注册且已授权的 service/method
- **THEN** `wasm_host_service.go`通过 registry lookup 定位 handler
- **AND** handler 或父包适配层接收统一 host call context、resource identifier、method 和 payload
- **AND** handler 返回统一 host call response envelope

#### Scenario: 未知 service 或 method 被拒绝

- **WHEN** 动态插件调用未在 registry 注册的 service/method
- **THEN** 宿主返回结构化“不支持”或“未找到”错误
- **AND** 宿主不得进入任何实际领域能力、数据访问、缓存、网络或外部资源调用

#### Scenario: 入口文件不维护 service 级 switch

- **WHEN** 静态检索`internal/service/plugin/internal/wasm/wasm_host_service.go`
- **THEN** 不得存在按 host service family 分发到`dispatch<X>HostService`的 service 级大 switch
- **AND** 新增领域 host service 不需要修改该入口文件的分发分支

#### Scenario: 注册方式保持显式依赖注入

- **WHEN** 宿主启动并构造 WASM host service registry
- **THEN** 注册入口显式接收 handler 需要的共享运行期依赖
- **AND** handler 不得通过`init()`、包级默认实例或调用关键服务`New()`自行获得依赖
- **AND** 缺失依赖必须在构造或注册阶段返回错误

### Requirement: 领域 dispatch handler 必须保持宿主治理边界

系统 SHALL 要求每个 host service dispatch handler 在 registry 驱动结构下继续保持既有授权、数据权限、租户边界、缓存一致性、审计和错误 envelope 语义。普通领域 handler 只负责 transport DTO 与`capability/<x>cap`领域契约之间的转换，不得直接依赖宿主 DAO、DO、Entity、私有缓存快照或未发布内部 service 实现。

#### Scenario: 数据访问能力通过等价数据权限边界

- **WHEN** 动态插件通过 host service handler 读取列表、详情、批量信息、候选项或执行写操作
- **THEN** handler 必须保持与宿主 API 等价的数据权限、租户边界和目标可见性校验
- **AND** 不得因为 registry 重构绕过授权快照、数据范围过滤或目标记录可见性检查

#### Scenario: 缓存敏感能力复用共享实例

- **WHEN** handler 访问 cache、session、权限快照、插件状态、运行时配置或其他缓存敏感能力
- **THEN** handler 必须复用启动期注入的共享服务实例或共享后端
- **AND** 不得在插件调用路径中创建仅当前节点可见的默认实例
- **AND** registry 重构不得改变缓存权威源、失效触发点、跨实例同步或可接受陈旧窗口

#### Scenario: 普通领域 handler 不暴露宿主内部模型

- **WHEN** 普通领域 handler 调用`usercap`、`dictcap`、`filecap`、`sessioncap`或其他`capability/<x>cap`契约
- **THEN** handler 只传递插件可见 DTO、投影或值对象
- **AND** 不得把`*gdb.Model`、DAO、DO、Entity、HTTP request 或宿主私有 service 实例暴露给 guest client

#### Scenario: 错误响应保持结构化

- **WHEN** handler 因授权、参数、数据权限、租户边界、缓存后端或领域服务失败返回错误
- **THEN** 宿主必须返回现有 host service 错误 envelope 或等价结构化错误
- **AND** 不得把裸 Go error 文本、敏感请求体、完整响应体或密钥写入插件可见响应
