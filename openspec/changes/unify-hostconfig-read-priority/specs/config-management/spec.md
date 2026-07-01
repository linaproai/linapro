## ADDED Requirements

### Requirement: HostConfig 原始读取必须使用统一来源优先级

系统 SHALL 对非 root 配置键使用统一的宿主配置读取优先级。读取顺序 MUST 为当前上下文可见的`sys_config`有效快照、GoFrame 当前静态配置源中的`config.yaml`值、系统已有默认值、`nil`。通用读取流程、运行时快照解析和受保护配置校验调度 MUST NOT 通过具体配置键常量、`IsManagedSysConfigKey()`或等价白名单分支决定来源顺序或处理策略；具体 key 的默认值、校验器和解析器 MUST 归属到配置元数据或等价 owner 中。空 key 和`.`不属于普通配置键，MUST 继续按宿主配置组件语义返回完整静态配置快照。

#### Scenario: sys_config 优先于静态配置和默认值

- **WHEN** 当前上下文可见的`sys_config`中存在 key 为`custom.feature.limit`的记录
- **AND** 静态`config.yaml`和系统默认值元数据也存在同名 key
- **THEN** 宿主配置服务通过`GetRaw(ctx, "custom.feature.limit")`返回`sys_config`中的有效值
- **AND** 不读取静态配置或系统默认值覆盖该值

#### Scenario: 静态配置优先于系统默认值

- **WHEN** 当前上下文可见的`sys_config`中不存在`workspace.basePath`
- **AND** 静态`config.yaml`中存在`workspace.basePath`
- **AND** 系统默认值元数据也存在`workspace.basePath`
- **THEN** 宿主配置服务通过`GetRaw(ctx, "workspace.basePath")`返回静态配置值
- **AND** 不提前返回系统默认值

#### Scenario: 缺少动态和静态配置时返回系统默认值

- **WHEN** 当前上下文可见的`sys_config`中不存在`sys.jwt.expire`
- **AND** 静态`config.yaml`中不存在`sys.jwt.expire`
- **AND** 系统默认值元数据存在`sys.jwt.expire`
- **THEN** 宿主配置服务通过`GetRaw(ctx, "sys.jwt.expire")`返回系统默认值
- **AND** 该读取不依赖`GetRaw()`中的具体 key 分支

#### Scenario: 所有来源缺失时返回 nil

- **WHEN** 当前上下文可见的`sys_config`中不存在`custom.missing.key`
- **AND** 静态`config.yaml`中不存在`custom.missing.key`
- **AND** 系统默认值元数据中不存在`custom.missing.key`
- **THEN** 宿主配置服务通过`GetRaw(ctx, "custom.missing.key")`返回`nil`
- **AND** 不因 key 未登记到白名单而返回权限错误

#### Scenario: root 配置快照保持静态配置语义

- **WHEN** 调用方通过`GetRaw(ctx, "")`或`GetRaw(ctx, ".")`读取宿主配置
- **THEN** 系统按宿主配置组件语义返回完整静态配置快照
- **AND** 该读取不要求逐个 key 进入`sys_config`快照或默认值元数据

### Requirement: 系统默认值必须由通用元数据提供

系统 SHALL 将宿主已有硬编码默认值维护为可按 key 查询的通用默认值元数据或等价 resolver。`HostConfig`通用读取流程 MUST 只调用默认值查询入口，不得在读取流程中直接判断具体配置键。新增宿主默认值时，系统 MUST 更新默认值元数据和测试，而不是在`GetRaw()`中增加新的 key 分支。

#### Scenario: 新增默认值不修改通用读取分支

- **WHEN** 主框架为新的宿主配置 key 增加系统默认值
- **THEN** 开发者在默认值元数据或等价 resolver 中登记该 key 和默认值
- **AND** 不在`GetRaw()`读取流程中增加该 key 的专用判断

#### Scenario: 专用 getter 保留类型校验

- **WHEN** 专用 getter 读取具有系统默认值的配置键
- **AND** `sys_config`和静态`config.yaml`都没有提供该值
- **THEN** 专用 getter 使用系统默认值作为输入
- **AND** 继续执行该 getter 已有的类型解析、归一化和业务校验

#### Scenario: sys_config freshness 错误不被 fallback 掩盖

- **WHEN** 宿主读取非 root 配置键时无法确认`sys_config`快照 freshness
- **THEN** 系统返回可见错误
- **AND** 不继续回退到静态配置或系统默认值来掩盖运行时配置一致性故障
