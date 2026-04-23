## ADDED Requirements

### Requirement: 配置管理组件单元测试覆盖率门槛
系统 SHALL 为 `apps/lina-core/internal/service/config` 配置管理组件维护可重复执行的单元测试，并以包级覆盖率验证作为该组件交付前的质量门槛。

#### Scenario: 包级覆盖率达到交付要求
- **WHEN** 维护者在 `apps/lina-core` 目录执行 `go test ./internal/service/config -cover`
- **THEN** 命令执行成功
- **AND** 输出的包级 statements 覆盖率不低于 `80%`

### Requirement: 配置管理关键分支具备自动化回归保护
系统 SHALL 为配置管理组件中的关键辅助逻辑补充自动化单元测试，至少覆盖默认值/回退、缓存或快照复用、以及非法输入或异常传播等高风险分支。

#### Scenario: 插件与公共前端配置辅助逻辑被修改
- **WHEN** 变更涉及插件动态存储路径、公共前端受保护配置键判断或统一校验入口
- **THEN** 单元测试覆盖正常读取路径
- **AND** 覆盖默认值或兼容回退路径
- **AND** 覆盖非法输入或空值防御路径

#### Scenario: 运行时参数缓存与修订同步逻辑被修改
- **WHEN** 变更涉及运行时参数 snapshot 缓存、revision 控制器或共享 KV 同步逻辑
- **THEN** 单元测试覆盖缓存命中或本地复用路径
- **AND** 覆盖 revision 变化后的重建路径
- **AND** 覆盖共享 KV 读取失败、无效缓存值或等价异常场景的错误传播与防御行为
